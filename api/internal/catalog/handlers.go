package catalog

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/wolfandwings/api/internal/httpx"
	"github.com/wolfandwings/api/internal/model"
	"github.com/wolfandwings/api/internal/store"
)

type Handlers struct {
	Store *store.Store
}

func Routes(h *Handlers) chi.Router {
	r := chi.NewRouter()
	r.Get("/products", h.ListProducts)
	r.Get("/products/{slug}", h.GetProduct)
	r.Get("/artists", h.ListArtists)
	r.Get("/artists/{slug}", h.GetArtist)
	r.Get("/fandoms", h.ListFandoms)
	r.Get("/categories", h.ListCategories)
	r.Get("/pages/{slug}", h.GetPage)
	return r
}

func (h *Handlers) ListProducts(w http.ResponseWriter, r *http.Request) {
	f := store.ProductFilter{
		ArtistSlug:   httpx.QueryStr(r, "artist"),
		FandomSlug:   httpx.QueryStr(r, "fandom"),
		CategorySlug: httpx.QueryStr(r, "category"),
		Availability: httpx.QueryStr(r, "availability"),
		Sort:         httpx.QueryStr(r, "sort"),
		Limit:        httpx.QueryInt(r, "limit", 60),
		Offset:       httpx.QueryInt(r, "offset", 0),
	}
	products, err := h.Store.ListProducts(r.Context(), f)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "catalog error")
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"items":   toCardDTOs(products),
		"count":   len(products),
		"filters": filterFacets(f),
	})
}

func (h *Handlers) GetProduct(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.Store.GetProductBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.RespondError(w, http.StatusNotFound, "product not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "catalog error")
		return
	}
	httpx.RespondJSON(w, http.StatusOK, toDetailDTO(p))
}

func (h *Handlers) ListArtists(w http.ResponseWriter, r *http.Request) {
	artists, err := h.Store.ListArtists(r.Context())
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "artists error")
		return
	}
	out := make([]map[string]any, 0, len(artists))
	for _, a := range artists {
		out = append(out, map[string]any{
			"slug": a.Slug, "name": a.Name, "bio": a.Bio,
			"avatar_url": a.AvatarURL, "socials": a.Socials,
		})
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handlers) GetArtist(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	a, err := h.Store.GetArtistBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			httpx.RespondError(w, http.StatusNotFound, "artist not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "artist error")
		return
	}
	works, _ := h.Store.ListProducts(r.Context(), store.ProductFilter{ArtistSlug: slug, Limit: 100})
	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"slug":        a.Slug,
		"name":        a.Name,
		"bio":         a.Bio,
		"avatar_url":  a.AvatarURL,
		"socials":     a.Socials,
		"works_count": len(works),
		"products":    toCardDTOs(works),
	})
}

func (h *Handlers) ListFandoms(w http.ResponseWriter, r *http.Request) {
	fs, err := h.Store.ListFandoms(r.Context())
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "fandoms error")
		return
	}
	out := make([]map[string]any, 0, len(fs))
	for _, f := range fs {
		out = append(out, map[string]any{"slug": f.Slug, "name": f.Name})
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handlers) ListCategories(w http.ResponseWriter, r *http.Request) {
	cs, err := h.Store.ListCategories(r.Context())
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "categories error")
		return
	}
	out := make([]map[string]any, 0, len(cs))
	for _, c := range cs {
		out = append(out, map[string]any{"slug": c.Slug, "name": c.Name})
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handlers) GetPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.Store.GetPageBySlug(r.Context(), slug)
	if err != nil {
		httpx.RespondError(w, http.StatusNotFound, "page not found")
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"slug": p.Slug, "title": p.Title, "body": p.Body, "meta_description": p.MetaDescription,
	})
}

// ---- DTO ----

func toCardDTOs(ps []model.Product) []map[string]any {
	out := make([]map[string]any, 0, len(ps))
	for _, p := range ps {
		out = append(out, cardDTO(p))
	}
	return out
}

func cardDTO(p model.Product) map[string]any {
	m := map[string]any{
		"slug":        p.Slug,
		"title":       p.Title,
		"description": p.Description,
		"base_price":  p.BasePrice,
		"currency":    "RUB",
		"status":      p.Status,
		"tint":        p.Tint,
		"artist":      map[string]any{"slug": p.Artist.Slug, "name": p.Artist.Name},
		"fandom":      map[string]any{"slug": p.Fandom.Slug, "name": p.Fandom.Name},
		"category":    map[string]any{"slug": p.Category.Slug, "name": p.Category.Name},
		"images":      imageDTOs(p.Images),
		"variants":    variantDTOs(p.Variants),
	}
	if p.OldPrice > 0 {
		m["old_price"] = p.OldPrice
	}
	if p.ETA != "" {
		m["eta"] = p.ETA
	}
	return m
}

func toDetailDTO(p *model.Product) map[string]any {
	m := cardDTO(*p)
	m["featured"] = p.Featured
	return m
}

func variantDTOs(vs []model.Variant) []map[string]any {
	out := make([]map[string]any, 0, len(vs))
	for _, v := range vs {
		out = append(out, map[string]any{
			"id": v.ID, "name": v.Name, "sku": v.SKU,
			"price_delta": v.PriceDelta, "stock_qty": v.StockQty,
		})
	}
	return out
}

func imageDTOs(ims []model.Image) []map[string]any {
	out := make([]map[string]any, 0, len(ims))
	for _, im := range ims {
		out = append(out, map[string]any{
			"url": im.URL, "alt": im.Alt, "width": im.Width, "height": im.Height,
			"is_primary": im.IsPrimary,
		})
	}
	return out
}

func filterFacets(f store.ProductFilter) map[string]any {
	return map[string]any{
		"artist": f.ArtistSlug, "fandom": f.FandomSlug, "category": f.CategorySlug,
		"availability": f.Availability, "sort": f.Sort,
	}
}
