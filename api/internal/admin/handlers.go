package admin

import (
	"encoding/json"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/wolfandwings/api/internal/config"
	"github.com/wolfandwings/api/internal/httpx"
	"github.com/wolfandwings/api/internal/model"
	"github.com/wolfandwings/api/internal/store"
)

type Handlers struct {
	Store *store.Store
	Auth  *Auth
	Tpl   *Templates
}

func NewHandlers(s *store.Store, cfg config.Config) *Handlers {
	return &Handlers{Store: s, Auth: NewAuth(s), Tpl: parseTemplates()}
}

var productStatuses = []string{"draft", "in_stock", "preorder", "low", "archived"}
var tints = []string{"luna", "violet", "amber", "teal", "rose"}
var preorderStatuses = []string{"new", "contacted", "confirmed", "converted", "cancelled"}

// Routes — монтируется под /admin с HostGuard(panel.*) в app.
func (h *Handlers) Routes() http.Handler {
	r := chi.NewRouter()
	r.Get("/login", h.loginGet)
	r.With(httpx.RateLimit(1, 5)).Post("/login", h.loginPost)
	r.Get("/logout", h.logout)

	r.Group(func(r chi.Router) {
		r.Use(h.requireAuth)
		r.Use(h.csrfCheck)
		r.Get("/", h.dashboard)
		r.Get("/products", h.productsList)
		r.Get("/products/new", h.productForm)
		r.Post("/products", h.productSave)
		r.Get("/products/{id}/edit", h.productForm)
		r.Post("/products/{id}", h.productSave)
		r.Post("/products/{id}/delete", h.productDelete)
		r.Get("/preorders", h.preordersList)
		r.Post("/preorders/{id}/status", h.preorderStatus)
		r.Get("/artists", h.artistsList)
		r.Get("/artists/new", h.artistForm)
		r.Post("/artists", h.artistSave)
		r.Get("/artists/{id}/edit", h.artistForm)
		r.Post("/artists/{id}", h.artistSave)
		r.Post("/artists/{id}/delete", h.artistDelete)
		r.Get("/pages", h.pagesList)
		r.Get("/pages/new", h.pageForm)
		r.Post("/pages", h.pageSave)
		r.Get("/pages/{id}/edit", h.pageForm)
		r.Post("/pages/{id}", h.pageSave)
	})
	return r
}

// ---- middleware ----

func (h *Handlers) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, u, err := h.Auth.sessionFromReq(r)
		if err != nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		ctx := withUser(r.Context(), u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handlers) csrfCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			s, _, _ := h.Auth.sessionFromReq(r)
			if s == nil || r.FormValue("csrf") != s.csrf {
				http.Error(w, "неверный CSRF", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// ---- auth ----

func (h *Handlers) loginGet(w http.ResponseWriter, r *http.Request) {
	csrf := h.Auth.loginCSRF(w)
	h.Tpl.WritePage(w, "login", map[string]any{"Title": "Вход", "CSRF": csrf})
}

func (h *Handlers) loginPost(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if !h.Auth.checkLoginCSRF(r) {
		h.Tpl.WritePage(w, "login", map[string]any{"Title": "Вход", "Error": "Форма устарела — обновите страницу", "Email": email})
		return
	}

	u, err := h.Store.GetUserByEmail(r.Context(), email)
	if err != nil {
		// считаем bcrypt и для несуществующего email — время ответа одинаковое
		_ = bcrypt.CompareHashAndPassword(dummyBcryptHash, []byte(password))
		h.Tpl.WritePage(w, "login", map[string]any{"Title": "Вход", "Error": "Неверный email или пароль", "Email": email})
		return
	}
	if !h.Auth.VerifyPassword(u, password) {
		h.Tpl.WritePage(w, "login", map[string]any{"Title": "Вход", "Error": "Неверный email или пароль", "Email": email})
		return
	}
	_, tok := h.Auth.newSession(u.ID)
	h.Auth.setCookie(w, tok)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	h.Auth.logout(w, r)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// ---- dashboard ----

func (h *Handlers) dashboard(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	products, _ := h.Store.ListAllProducts(r.Context())
	preorders, _ := h.Store.ListPreorders(r.Context())
	newP := 0
	for _, p := range preorders {
		if p.Status == "new" {
			newP++
		}
	}
	h.Tpl.WritePage(w, "dashboard", map[string]any{
		"Title":         "Дашборд",
		"User":          u,
		"ProductCount":  len(products),
		"PreorderCount": len(preorders),
		"NewPreorders":  newP,
	})
}

// ---- products ----

func (h *Handlers) productsList(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	products, _ := h.Store.ListAllProducts(r.Context())
	h.Tpl.WritePage(w, "products", map[string]any{"Title": "Товары", "User": u, "CSRF": s.csrf, "Products": products})
}

func (h *Handlers) productForm(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	idStr := chi.URLParam(r, "id")
	isNew := idStr == ""
	var p *model.Product
	if !isNew {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		p, _ = h.Store.GetProductByID(r.Context(), id)
	}
	if p == nil {
		p = &model.Product{Tint: "luna", Status: "draft", Variants: nil}
	}
	artists, _ := h.Store.ListArtistsAll(r.Context())
	fandoms, _ := h.Store.ListFandoms(r.Context())
	categories, _ := h.Store.ListCategories(r.Context())
	h.Tpl.WritePage(w, "product_form", map[string]any{
		"Title": ternary(isNew, "Новый товар", "Редактирование"),
		"User":  u, "CSRF": s.csrf, "IsNew": isNew, "Product": p,
		"Artists": artists, "Fandoms": fandoms, "Categories": categories,
		"Tints": tints, "Statuses": productStatuses,
	})
}

func (h *Handlers) productSave(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	r.ParseForm()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	isNew := idStr == ""

	artistID, _ := strconv.ParseInt(r.FormValue("artist_id"), 10, 64)
	fandomID, _ := strconv.ParseInt(r.FormValue("fandom_id"), 10, 64)
	categoryID, _ := strconv.ParseInt(r.FormValue("category_id"), 10, 64)
	basePrice := parseRubles(r.FormValue("base_price"))
	oldPrice := parseRubles(r.FormValue("old_price"))

	in := store.ProductInput{
		Slug:        strings.TrimSpace(r.FormValue("slug")),
		Title:       strings.TrimSpace(r.FormValue("title")),
		Description: r.FormValue("description"),
		ArtistID:    artistID, FandomID: fandomID, CategoryID: categoryID,
		BasePrice: basePrice, OldPrice: oldPrice,
		Status: r.FormValue("status"), ETA: strings.TrimSpace(r.FormValue("eta")),
		Tint:     r.FormValue("tint"),
		Featured: r.FormValue("featured") == "1",
		Variants: parseVariants(r.FormValue("variants")),
		Images:   parseImages(r.FormValue("images")),
	}
	if in.Slug == "" || in.Title == "" || artistID == 0 {
		artists, _ := h.Store.ListArtistsAll(r.Context())
		fandoms, _ := h.Store.ListFandoms(r.Context())
		categories, _ := h.Store.ListCategories(r.Context())
		h.Tpl.WritePage(w, "product_form", map[string]any{
			"Title": ternary(isNew, "Новый товар", "Редактирование"), "User": u, "CSRF": s.csrf,
			"IsNew": isNew, "Product": &model.Product{Slug: in.Slug, Title: in.Title, Description: in.Description,
				BasePrice: in.BasePrice, OldPrice: in.OldPrice, Status: in.Status, ETA: in.ETA, Tint: in.Tint, Featured: in.Featured},
			"Artists": artists, "Fandoms": fandoms, "Categories": categories, "Tints": tints, "Statuses": productStatuses,
			"Error": "Заполните slug, название и художника",
		})
		return
	}
	if !slices.Contains(productStatuses, in.Status) || !slices.Contains(tints, in.Tint) {
		http.Error(w, "недопустимый статус или цвет", http.StatusBadRequest)
		return
	}
	var err error
	if isNew {
		_, err = h.Store.CreateProduct(r.Context(), in)
	} else {
		err = h.Store.UpdateProduct(r.Context(), id, in)
	}
	if err != nil {
		slog.Error("product save", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

// ---- preorders ----

func (h *Handlers) preordersList(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	preorders, _ := h.Store.ListPreorders(r.Context())
	h.Tpl.WritePage(w, "preorders", map[string]any{
		"Title": "Заявки", "User": u, "CSRF": s.csrf,
		"Preorders": preorders, "PreorderStatuses": preorderStatuses,
	})
}

func (h *Handlers) preorderStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	status := r.FormValue("status")
	if id > 0 && slices.Contains(preorderStatuses, status) {
		if err := h.Store.UpdatePreorderStatus(r.Context(), id, status); err != nil {
			slog.Error("preorder status", "err", err)
		}
	}
	http.Redirect(w, r, "/admin/preorders", http.StatusSeeOther)
}

// ---- helpers ----

func parseRubles(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f < 0 || f > 10_000_000 { // не цена: отрицательная или за пределами разумного
		return 0
	}
	return int(math.Round(f * 100.0))
}

func parseImages(text string) []store.ImageInput {
	var out []store.ImageInput
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		im := store.ImageInput{}
		if len(parts) >= 1 {
			im.URL = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			im.Alt = strings.TrimSpace(parts[1])
		}
		if len(parts) >= 3 {
			im.IsPrimary = strings.TrimSpace(parts[2]) == "1"
		}
		out = append(out, im)
	}
	return out
}

func (h *Handlers) productDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	_ = h.Store.DeleteProduct(r.Context(), id)
	http.Redirect(w, r, "/admin/products", http.StatusSeeOther)
}

// ---- artists ----

func (h *Handlers) artistsList(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	artists, _ := h.Store.ListArtistsAll(r.Context())
	h.Tpl.WritePage(w, "artists", map[string]any{"Title": "Художники", "User": u, "CSRF": s.csrf, "Artists": artists})
}

func (h *Handlers) artistForm(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	idStr := chi.URLParam(r, "id")
	isNew := idStr == ""
	var a *model.Artist
	if !isNew {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		a, _ = h.Store.GetArtistByID(r.Context(), id)
	}
	if a == nil {
		a = &model.Artist{Status: "published", Socials: map[string]string{}}
	}
	h.Tpl.WritePage(w, "artist_form", map[string]any{
		"Title": ternary(isNew, "Новый художник", "Редактирование"), "User": u, "CSRF": s.csrf, "IsNew": isNew, "Artist": a,
	})
}

func (h *Handlers) artistSave(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	r.ParseForm()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	isNew := idStr == ""

	name := strings.TrimSpace(r.FormValue("name"))
	slug := strings.TrimSpace(r.FormValue("slug"))
	bio := r.FormValue("bio")
	avatar := r.FormValue("avatar_url")
	socials := mustJSON(map[string]string{
		"vk": strings.TrimSpace(r.FormValue("vk")),
		"tg": strings.TrimSpace(r.FormValue("tg")),
	})

	if name == "" || slug == "" {
		a := &model.Artist{Name: name, Slug: slug, Bio: bio, AvatarURL: avatar, Socials: map[string]string{}}
		h.Tpl.WritePage(w, "artist_form", map[string]any{"Title": "Новый художник", "User": u, "CSRF": s.csrf, "IsNew": isNew, "Artist": a, "Error": "Имя и slug обязательны"})
		return
	}
	var err error
	if isNew {
		_, err = h.Store.CreateArtist(r.Context(), name, slug, bio, avatar, socials)
	} else {
		err = h.Store.UpdateArtist(r.Context(), id, name, slug, bio, avatar, socials)
	}
	if err != nil {
		slog.Error("artist save", "err", err)
		a := &model.Artist{Name: name, Slug: slug, Bio: bio, AvatarURL: avatar, Socials: map[string]string{}}
		h.Tpl.WritePage(w, "artist_form", map[string]any{"Title": ternary(isNew, "Новый художник", "Редактирование"), "User": u, "CSRF": s.csrf, "IsNew": isNew, "Artist": a, "Error": "Не удалось сохранить (ошибка БД)"})
		return
	}
	http.Redirect(w, r, "/admin/artists", http.StatusSeeOther)
}

func (h *Handlers) artistDelete(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	_ = h.Store.DeleteArtist(r.Context(), id)
	http.Redirect(w, r, "/admin/artists", http.StatusSeeOther)
}

// ---- pages ----

func (h *Handlers) pagesList(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	pages, _ := h.Store.ListPagesAll(r.Context())
	h.Tpl.WritePage(w, "pages", map[string]any{"Title": "Страницы", "User": u, "CSRF": s.csrf, "Pages": pages})
}

func (h *Handlers) pageForm(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	idStr := chi.URLParam(r, "id")
	isNew := idStr == ""
	var p *model.Page
	if !isNew {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		p, _ = h.Store.GetPageByID(r.Context(), id)
	}
	if p == nil {
		p = &model.Page{}
	}
	h.Tpl.WritePage(w, "page_form", map[string]any{
		"Title": ternary(isNew, "Новая страница", "Редактирование"), "User": u, "CSRF": s.csrf, "IsNew": isNew, "Page": p,
	})
}

func (h *Handlers) pageSave(w http.ResponseWriter, r *http.Request) {
	u := userFromCtx(r.Context())
	s, _, _ := h.Auth.sessionFromReq(r)
	r.ParseForm()
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	isNew := idStr == ""

	slug := strings.TrimSpace(r.FormValue("slug"))
	title := strings.TrimSpace(r.FormValue("title"))
	body := r.FormValue("body")
	meta := r.FormValue("meta")

	if slug == "" || title == "" {
		p := &model.Page{Slug: slug, Title: title, Body: body, MetaDescription: meta}
		h.Tpl.WritePage(w, "page_form", map[string]any{"Title": "Новая страница", "User": u, "CSRF": s.csrf, "IsNew": isNew, "Page": p, "Error": "Заголовок и slug обязательны"})
		return
	}
	var err error
	if isNew {
		_, err = h.Store.CreatePage(r.Context(), slug, title, body, meta)
	} else {
		err = h.Store.UpdatePage(r.Context(), id, slug, title, body, meta)
	}
	if err != nil {
		slog.Error("page save", "err", err)
		p := &model.Page{Slug: slug, Title: title, Body: body, MetaDescription: meta}
		h.Tpl.WritePage(w, "page_form", map[string]any{"Title": ternary(isNew, "Новая страница", "Редактирование"), "User": u, "CSRF": s.csrf, "IsNew": isNew, "Page": p, "Error": "Не удалось сохранить (ошибка БД)"})
		return
	}
	http.Redirect(w, r, "/admin/pages", http.StatusSeeOther)
}

func parseVariants(text string) []store.VariantInput {
	var out []store.VariantInput
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name := line
		delta := 0
		if i := strings.LastIndex(line, "|"); i >= 0 {
			name = strings.TrimSpace(line[:i])
			delta, _ = strconv.Atoi(strings.TrimSpace(line[i+1:]))
		}
		out = append(out, store.VariantInput{Name: name, PriceDelta: delta})
	}
	return out
}

func ternary(b bool, a, c string) string {
	if b {
		return a
	}
	return c
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
