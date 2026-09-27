package admin

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/wolfandwings/api/internal/config"
	"github.com/wolfandwings/api/internal/httpx"
	"github.com/wolfandwings/api/internal/platform"
	"github.com/wolfandwings/api/internal/store"
)

func newTestServer(t *testing.T) (*httptest.Server, *Handlers) {
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
	h := NewHandlers(st, cfg)
	r := chi.NewRouter()
	r.Mount("/admin", h.Routes())
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv, h
}

func login(t *testing.T, client *http.Client, url, email, pass string) *http.Response {
	t.Helper()
	// флоу с CSRF: GET формы кладёт double-submit cookie, POST обязан вернуть токен
	get, err := client.Get(url + "/admin/login")
	if err != nil {
		t.Fatalf("login form: %v", err)
	}
	token := csrfFrom(t, readAll(t, get))
	form := "email=" + email + "&password=" + pass + "&csrf=" + token
	resp, err := client.Post(url+"/admin/login", "application/x-www-form-urlencoded",
		strings.NewReader(form))
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	resp.Body.Close()
	return resp
}

// loginNoCSRF — POST логина без токена: должен быть отброшен до проверки пароля.
func loginNoCSRF(t *testing.T, client *http.Client, url, email, pass string) *http.Response {
	t.Helper()
	resp, err := client.Post(url+"/admin/login", "application/x-www-form-urlencoded",
		strings.NewReader("email="+email+"&password="+pass))
	if err != nil {
		t.Fatalf("login no csrf: %v", err)
	}
	resp.Body.Close()
	return resp
}

func csrfFrom(t *testing.T, body string) string {
	t.Helper()
	i := strings.Index(body, `name="csrf" value="`)
	if i < 0 {
		t.Fatalf("no csrf token in page: %.200s", body)
	}
	rest := body[i+len(`name="csrf" value="`):]
	return rest[:strings.Index(rest, `"`)]
}

func TestLoginFlow(t *testing.T) {
	srv, _ := newTestServer(t)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:           jar,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	// POST без csrf-токена — не пускаем даже с верным паролем
	if resp := loginNoCSRF(t, client, srv.URL, "admin@test.ru", "secret123"); resp.StatusCode != http.StatusOK {
		t.Fatalf("login without csrf: want 200 (form again), got %d", resp.StatusCode)
	}

	// верный пароль + токен — редирект в админку
	if resp := login(t, client, srv.URL, "admin@test.ru", "secret123"); resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("good login: %d", resp.StatusCode)
	}

	// без сессии — на логин
	fresh := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	resp, err := fresh.Get(srv.URL + "/admin/products")
	if err != nil {
		t.Fatalf("get products: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("unauth access: want 303, got %d", resp.StatusCode)
	}

	// с сессией — 200, CSRF в форме
	resp, err = client.Get(srv.URL + "/admin/products/new")
	if err != nil {
		t.Fatalf("get form: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("auth access: %d", resp.StatusCode)
	}
	csrf := csrfFrom(t, readAll(t, resp))

	// POST без CSRF — 403
	form := strings.NewReader("csrf=wrong&slug=x&title=X&artist_id=1&base_price=100&status=draft&tint=luna")
	resp, err = client.Post(srv.URL+"/admin/products", "application/x-www-form-urlencoded", form)
	if err != nil {
		t.Fatalf("post no csrf: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("post without csrf: want 403, got %d", resp.StatusCode)
	}

	// POST с CSRF — 303
	form = strings.NewReader("csrf=" + csrf + "&slug=test-prod&title=Тест&artist_id=1&fandom_id=1&category_id=1&base_price=1990.50&status=draft&tint=luna&variants=XL | 0")
	resp, err = client.Post(srv.URL+"/admin/products", "application/x-www-form-urlencoded", form)
	if err != nil {
		t.Fatalf("post with csrf: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("post with csrf: want 303, got %d", resp.StatusCode)
	}
}

func TestHostGuardOnAdmin(t *testing.T) {
	_, h := newTestServer(t)
	guarded := httptest.NewServer(httpx.HostGuard(map[string]bool{"panel.test": true})(h.Routes()))
	defer guarded.Close()
	resp, err := http.Get(guarded.URL + "/admin/login") // host 127.0.0.1:port ≠ panel.test
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("admin reachable via foreign host: %d", resp.StatusCode)
	}
}

func readAll(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}
