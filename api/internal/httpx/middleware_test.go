package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func ok(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestSecurityHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	SecurityHeaders(http.HandlerFunc(ok)).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	h := rec.Header()
	for _, k := range []string{"X-Content-Type-Options", "X-Frame-Options", "Referrer-Policy", "Permissions-Policy"} {
		if h.Get(k) == "" {
			t.Errorf("missing header %s", k)
		}
	}
}

func TestHostGuard(t *testing.T) {
	h := HostGuard(map[string]bool{"panel.example": true})(http.HandlerFunc(ok))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "http://panel.example/admin", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("allowed host blocked: %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "http://evil.example/admin", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("foreign host passed: %d", rec.Code)
	}
}

func TestRateLimit(t *testing.T) {
	h := RateLimit(1, 2)(http.HandlerFunc(ok))
	for i := 0; i < 2; i++ {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d throttled too early: %d", i, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("want 429, got %d", rec.Code)
	}
}

func TestReadJSON(t *testing.T) {
	type in struct {
		Name string `json:"name"`
	}
	var v in
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"x"}`))
	if err := ReadJSON(req, &v); err != nil || v.Name != "x" {
		t.Fatalf("read json: %v %q", err, v.Name)
	}

	// неизвестные поля отклоняются
	req = httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"x","evil":1}`))
	if err := ReadJSON(req, &v); err == nil {
		t.Fatal("unknown field accepted")
	}

	// тело >1МБ отклоняется
	big := `{"name":"` + strings.Repeat("a", 2<<20) + `"}`
	req = httptest.NewRequest("POST", "/", strings.NewReader(big))
	if err := ReadJSON(req, &v); err == nil {
		t.Fatal("oversized body accepted")
	}
}

func TestCORS(t *testing.T) {
	h := CORS([]string{"https://site.ru"})(http.HandlerFunc(ok))

	req := httptest.NewRequest("OPTIONS", "/", nil)
	req.Header.Set("Origin", "https://site.ru")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://site.ru" {
		t.Fatalf("allowed origin missing: %v", rec.Header())
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight: %d", rec.Code)
	}

	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://evil.ru")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("foreign origin allowed")
	}
}
