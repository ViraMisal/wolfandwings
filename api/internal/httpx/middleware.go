package httpx

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/wolfandwings/api/internal/metrics"
)

type ctxKey string

const CtxRequestID ctxKey = "reqid"

// Recover ловит панику, логирует и отдаёт 500.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic", "err", rec, "stack", string(debug.Stack()))
					RespondError(w, http.StatusInternalServerError, "internal error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID добавляет X-Request-ID; клиентский принимается только вразумительной длины.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), CtxRequestID, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SecurityHeaders — базовые заголовки API. CSP ставит Next (nonce на каждый
// запрос), HSTS — Caddy; здесь только то, что должно быть и на прямом доступе.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}

// CORS — минималистичный, origin из allowed.
func CORS(allowed []string) func(http.Handler) http.Handler {
	allow := map[string]bool{}
	for _, a := range allowed {
		allow[strings.TrimRight(a, "/")] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allow[strings.TrimRight(origin, "/")] {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Vary", "Origin")
				h.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				h.Set("Access-Control-Allow-Headers", "Content-Type")
				h.Set("Access-Control-Max-Age", "600")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit — токен-бакет на IP, отдельный лимит для «тяжёлых» маршрутов.
// Бакеты чистятся поштучно по давности использования, а не все разом.
type bucket struct {
	lim  *rate.Limiter
	seen time.Time
}

func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	var mu sync.Mutex
	buckets := map[string]*bucket{}
	lastClean := time.Now()

	get := func(ip string) *rate.Limiter {
		mu.Lock()
		defer mu.Unlock()
		if time.Since(lastClean) > 5*time.Minute {
			for k, b := range buckets {
				if time.Since(b.seen) > 10*time.Minute {
					delete(buckets, k)
				}
			}
			lastClean = time.Now()
		}
		b, ok := buckets[ip]
		if !ok {
			b = &bucket{lim: rate.NewLimiter(rate.Limit(rps), burst)}
			buckets[ip] = b
		}
		b.seen = time.Now()
		return b.lim
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			if !get(ip).Allow() {
				RespondError(w, http.StatusTooManyRequests, "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	// правый элемент XFF дописывает наш доверенный прокси (Caddy), левый — можно подделать
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		parts := strings.Split(h, ",")
		ip := strings.TrimSpace(parts[len(parts)-1])
		if ip != "" {
			return ip
		}
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return host
}

// Metrics считает запросы и 5xx для /metrics.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		metrics.Inc(rec.status)
	})
}

// statusRecorder запоминает статус ответа.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// HostGuard пускает только на разрешённых хостах (для админки на panel.*).
func HostGuard(allowed map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Host
			if i := strings.IndexByte(h, ':'); i > 0 {
				h = h[:i]
			}
			if !allowed[h] {
				RespondError(w, http.StatusNotFound, "not found")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
