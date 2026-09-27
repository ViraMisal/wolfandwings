package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/wolfandwings/api/internal/admin"
	"github.com/wolfandwings/api/internal/catalog"
	"github.com/wolfandwings/api/internal/config"
	"github.com/wolfandwings/api/internal/httpx"
	"github.com/wolfandwings/api/internal/metrics"
	"github.com/wolfandwings/api/internal/order"
	"github.com/wolfandwings/api/internal/preorder"
	"github.com/wolfandwings/api/internal/store"
)

// Deps — зависимости сервера.
type Deps struct {
	Cfg      config.Config
	Catalog  *catalog.Handlers
	Preorder *preorder.Handlers
	Order    *order.Handlers
	Admin    *admin.Handlers
	Store    *store.Store
}

// New собирает итоговый http.Handler с middleware и маршрутами.
// RealIP не подключён: реальный клиентский IP достаётся в httpx.clientIP
// из правого элемента X-Forwarded-For (его дописывает доверенный Caddy).
func New(d Deps, log *slog.Logger) http.Handler {
	r := chi.NewRouter()
	r.Use(httpx.Metrics)
	r.Use(httpx.Recover(log))
	r.Use(httpx.RequestID)
	r.Use(httpx.SecurityHeaders)
	r.Use(httpx.CORS([]string{d.Cfg.WebOrigin}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.RespondJSON(w, 200, map[string]string{"status": "ok"})
	})
	// метрики — только локально (API слушает 127.0.0.1), Prometheus может снимать через SSH-туннель
	r.Get("/metrics", metrics.Handler())

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/catalog", catalog.Routes(d.Catalog))
		r.With(httpx.RateLimit(5, 10)).Post("/preorders", d.Preorder.Create)
		r.With(httpx.RateLimit(10, 20)).Get("/orders/{number}", d.Order.Get)
	})

	// Админка наружу не проксируется вовсе: Caddy смотрит только витрину,
	// публичной DNS-записи у panel.* нет. HostGuard страхует сам API.
	r.With(httpx.HostGuard(map[string]bool{d.Cfg.AdminHost: true})).Mount("/admin", d.Admin.Routes())

	return r
}
