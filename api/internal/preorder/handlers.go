package preorder

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/wolfandwings/api/internal/httpx"
	"github.com/wolfandwings/api/internal/notify"
	"github.com/wolfandwings/api/internal/store"
)

type Handlers struct {
	Store  *store.Store
	Notify *notify.Telegram
}

type createReq struct {
	ProductSlug string `json:"product_slug"`
	VariantID   int64  `json:"variant_id"`
	ContactName string `json:"contact_name"`
	Contact     string `json:"contact"`
	ContactKind string `json:"contact_kind"` // email | phone | tg
	Consent     bool   `json:"consent"`
}

var (
	emailRe = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	phoneRe = regexp.MustCompile(`^\+?\d[\d\s\-\(\)]{6,}$`)
	// белый список: только эти способы связи принимаем, остальное — 422
	contactKinds = []string{"email", "phone", "tg"}
)

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var in createReq
	if err := httpx.ReadJSON(r, &in); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "неверный формат запроса")
		return
	}
	in.ContactName = strings.TrimSpace(in.ContactName)
	in.Contact = strings.TrimSpace(in.Contact)
	in.ContactKind = strings.TrimSpace(in.ContactKind)
	if in.ContactKind == "" {
		in.ContactKind = "email"
	}

	if in.ProductSlug == "" || in.ContactName == "" || in.Contact == "" {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "заполните все поля")
		return
	}
	if !in.Consent {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "отметьте галочку — она ничего не подтверждает, но нужна")
		return
	}
	if !slices.Contains(contactKinds, in.ContactKind) {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "недопустимый способ связи")
		return
	}
	if in.ContactKind == "email" && !emailRe.MatchString(in.Contact) {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "некорректный email")
		return
	}
	if in.ContactKind == "phone" && !phoneRe.MatchString(in.Contact) {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "некорректный телефон")
		return
	}

	p, err := h.Store.GetProductBySlug(r.Context(), in.ProductSlug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.RespondError(w, http.StatusNotFound, "товар не найден")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "catalog error")
		return
	}
	// вариант должен принадлежать этому товару
	if in.VariantID != 0 && !h.Store.VariantBelongsTo(r.Context(), in.VariantID, p.ID) {
		httpx.RespondError(w, http.StatusUnprocessableEntity, "указан неверный размер")
		return
	}

	req, err := h.Store.CreatePreorder(r.Context(), store.PreorderInput{
		ProductID:   p.ID,
		ProductSlug: p.Slug,
		VariantID:   in.VariantID,
		ContactName: in.ContactName,
		Contact:     in.Contact,
		ContactKind: in.ContactKind,
	})
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "не удалось сохранить заявку")
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, map[string]any{
		"id":           req.ID,
		"status":       req.Status,
		"product_slug": req.ProductSlug,
	})

	// уведомление в Telegram — асинхронно, не блокирует ответ
	h.Notify.NotifyAsync(fmt.Sprintf("🆕 Новая заявка #%d\nТовар: %s\nКонтакт: %s (%s)",
		req.ID, p.Title, in.Contact, in.ContactKind))
}
