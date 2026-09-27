package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

var (
	startTime = time.Now()
	requests  atomic.Int64
	err5xx    atomic.Int64
)

// Inc учитывает запрос; код >= 500 считается ошибкой.
func Inc(status int) {
	requests.Add(1)
	if status >= 500 {
		err5xx.Add(1)
	}
}

// Handler отдаёт /metrics в текстовом формате Prometheus — без внешних зависимостей.
// Эндпоинт живёт только на 127.0.0.1:8080 (API_ADDR) и наружу Caddy не проксируется.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ms runtime.MemStats
		runtime.ReadMemStats(&ms)
		var b strings.Builder
		fmt.Fprintf(&b, "# HELP wolf_api_uptime_seconds Seconds since start.\n")
		fmt.Fprintf(&b, "# TYPE wolf_api_uptime_seconds gauge\n")
		fmt.Fprintf(&b, "wolf_api_uptime_seconds %.0f\n", time.Since(startTime).Seconds())
		fmt.Fprintf(&b, "# HELP wolf_api_requests_total Total HTTP requests.\n")
		fmt.Fprintf(&b, "# TYPE wolf_api_requests_total counter\n")
		fmt.Fprintf(&b, "wolf_api_requests_total %d\n", requests.Load())
		fmt.Fprintf(&b, "# HELP wolf_api_errors_total HTTP 5xx responses.\n")
		fmt.Fprintf(&b, "# TYPE wolf_api_errors_total counter\n")
		fmt.Fprintf(&b, "wolf_api_errors_total %d\n", err5xx.Load())
		fmt.Fprintf(&b, "# HELP wolf_api_goroutines Current goroutines.\n")
		fmt.Fprintf(&b, "# TYPE wolf_api_goroutines gauge\n")
		fmt.Fprintf(&b, "wolf_api_goroutines %d\n", runtime.NumGoroutine())
		fmt.Fprintf(&b, "# HELP wolf_api_mem_alloc_bytes Heap allocation in bytes.\n")
		fmt.Fprintf(&b, "# TYPE wolf_api_mem_alloc_bytes gauge\n")
		fmt.Fprintf(&b, "wolf_api_mem_alloc_bytes %d\n", ms.HeapAlloc)
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(b.String()))
	}
}
