package metrics

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	Inc(200)
	Inc(500)

	rec := httptest.NewRecorder()
	Handler()(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Fatalf("content-type: %s", ct)
	}

	vals := map[string]int{}
	for _, line := range strings.Split(rec.Body.String(), "\n") {
		if strings.HasPrefix(line, "wolf_") {
			parts := strings.Fields(line)
			if len(parts) == 2 {
				v, _ := strconv.Atoi(parts[1])
				vals[parts[0]] = v
			}
		}
	}
	if vals["wolf_api_requests_total"] < 2 {
		t.Fatalf("requests_total not counted: %v", vals)
	}
	if vals["wolf_api_errors_total"] < 1 {
		t.Fatalf("errors_total not counted: %v", vals)
	}
	for _, key := range []string{"wolf_api_uptime_seconds", "wolf_api_goroutines", "wolf_api_mem_alloc_bytes"} {
		if _, ok := vals[key]; !ok {
			t.Fatalf("missing metric %s in:\n%s", key, rec.Body.String())
		}
	}
}
