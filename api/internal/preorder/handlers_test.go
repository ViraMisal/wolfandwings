package preorder

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wolfandwings/api/internal/config"
	"github.com/wolfandwings/api/internal/platform"
	"github.com/wolfandwings/api/internal/store"
)

func newTestHandlers(t *testing.T) *Handlers {
	t.Helper()
	dsn := "sqlite://" + filepath.Join(t.TempDir(), "test.db")
	cfg := config.Config{Dialect: "sqlite", DatabaseURL: dsn}
	db, err := platform.Open(context.Background(), cfg, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	st := store.New(db, "sqlite")
	if err := st.SeedIfEmpty(context.Background(), "admin@test.ru", "secret123", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return &Handlers{Store: st}
}

func post(t *testing.T, h *Handlers, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/preorders", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Create(rec, req)
	return rec
}

func TestCreateValidation(t *testing.T) {
	h := newTestHandlers(t)

	cases := []struct {
		name string
		body string
		code int
	}{
		{"нет согласия", `{"product_slug":"lappland-lapplandia","contact_name":"A","contact":"a@b.ru","consent":false}`, http.StatusUnprocessableEntity},
		{"нет контакта", `{"product_slug":"lappland-lapplandia","contact_name":"","contact":"","consent":true}`, http.StatusUnprocessableEntity},
		{"плохой email", `{"product_slug":"lappland-lapplandia","contact_name":"A","contact":"not-an-email","contact_kind":"email","consent":true}`, http.StatusUnprocessableEntity},
		{"плохой телефон", `{"product_slug":"lappland-lapplandia","contact_name":"A","contact":"abc","contact_kind":"phone","consent":true}`, http.StatusUnprocessableEntity},
		{"нет товара", `{"product_slug":"no-such","contact_name":"A","contact":"a@b.ru","consent":true}`, http.StatusNotFound},
		{"битый json", `{not json`, http.StatusBadRequest},
	}
	for _, c := range cases {
		if rec := post(t, h, c.body); rec.Code != c.code {
			t.Errorf("%s: want %d, got %d (%s)", c.name, c.code, rec.Code, rec.Body.String())
		}
	}
}

func TestCreateHappyPath(t *testing.T) {
	h := newTestHandlers(t)
	rec := post(t, h, `{"product_slug":"lappland-lapplandia","contact_name":"Тест","contact":"a@b.ru","contact_kind":"email","consent":true}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("want 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var out struct {
		ID     int64  `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if out.Status != "new" || out.ID == 0 {
		t.Fatalf("unexpected response: %+v", out)
	}
	list, err := h.Store.ListPreorders(context.Background())
	if err != nil || len(list) != 1 {
		t.Fatalf("preorder not persisted: %v %d", err, len(list))
	}
}
