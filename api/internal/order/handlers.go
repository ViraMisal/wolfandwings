package order

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/wolfandwings/api/internal/httpx"
	"github.com/wolfandwings/api/internal/store"
)

type Handlers struct {
	Store *store.Store
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	number := strings.TrimSpace(chi.URLParam(r, "number"))
	email := strings.TrimSpace(r.URL.Query().Get("email"))
	if number == "" || email == "" {
		httpx.RespondError(w, http.StatusBadRequest, "укажите номер заказа и email")
		return
	}
	o, err := h.Store.GetOrderByNumber(r.Context(), number, email)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.RespondError(w, http.StatusNotFound, "заказ не найден")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "order error")
		return
	}
	httpx.RespondJSON(w, http.StatusOK, o)
}
