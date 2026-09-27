package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/wolfandwings/api/internal/model"
)

type ProductFilter struct {
	ArtistSlug   string
	FandomSlug   string
	CategorySlug string
	Availability string // all | in_stock | preorder
	Sort         string // pop | new | cheap | exp
	Limit        int
	Offset       int
}

func (s *Store) ListFandoms(ctx context.Context) ([]model.Fandom, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, slug, name, sort FROM fandoms ORDER BY sort, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Fandom
	for rows.Next() {
		var f model.Fandom
		if err := rows.Scan(&f.ID, &f.Slug, &f.Name, &f.Sort); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, nil
}

func (s *Store) ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, slug, name, sort FROM categories ORDER BY sort, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Sort); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *Store) ListArtists(ctx context.Context) ([]model.Artist, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, slug, name, bio, avatar_url, socials, works_count, status, sort
		FROM artists WHERE status='published' ORDER BY sort, name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArtists(rows)
}

func (s *Store) GetArtistBySlug(ctx context.Context, slug string) (*model.Artist, error) {
	// как и товары: неопубликованный автор снаружи не существует
	row := s.DB.QueryRowContext(ctx, `SELECT id, slug, name, bio, avatar_url, socials, works_count, status, sort
		FROM artists WHERE slug=`+s.ph(1)+` AND status='published'`, slug)
	a, err := scanArtistRow(row.Scan)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return a, nil
}

func scanArtists(rows *sql.Rows) ([]model.Artist, error) {
	var out []model.Artist
	for rows.Next() {
		a, err := scanArtistRow(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, *a)
	}
	return out, rows.Err()
}

func scanArtistRow(scan func(...any) error) (*model.Artist, error) {
	var a model.Artist
	var socials string
	if err := scan(&a.ID, &a.Slug, &a.Name, &a.Bio, &a.AvatarURL, &socials, &a.WorksCount, &a.Status, &a.Sort); err != nil {
		return nil, err
	}
	a.Socials = parseSocials(socials)
	return &a, nil
}

func parseSocials(s string) map[string]string {
	m := map[string]string{}
	if s == "" {
		return m
	}
	_ = json.Unmarshal([]byte(s), &m)
	return m
}

func (s *Store) ListProducts(ctx context.Context, f ProductFilter) ([]model.Product, error) {
	var where []string
	var args []any
	n := 1
	if f.ArtistSlug != "" {
		where = append(where, "a.slug="+s.ph(n))
		n++
		args = append(args, f.ArtistSlug)
	}
	if f.FandomSlug != "" {
		where = append(where, "f.slug="+s.ph(n))
		n++
		args = append(args, f.FandomSlug)
	}
	if f.CategorySlug != "" {
		where = append(where, "c.slug="+s.ph(n))
		n++
		args = append(args, f.CategorySlug)
	}
	switch f.Availability {
	case "in_stock":
		where = append(where, "p.status='in_stock'")
	case "preorder":
		where = append(where, "p.status='preorder'")
	}
	where = append(where, "p.status!='draft'")

	order := "p.featured DESC, p.sort ASC, p.id DESC"
	switch f.Sort {
	case "cheap":
		order = "p.base_price ASC"
	case "exp":
		order = "p.base_price DESC"
	case "new":
		order = "CASE WHEN p.status='preorder' THEN 0 ELSE 1 END, p.id DESC"
	}

	q := `SELECT p.id,p.slug,p.title,p.description,p.base_price,p.old_price,p.status,p.eta,p.tint,p.featured,p.sort,
		a.id,a.slug,a.name,a.bio,a.avatar_url,a.socials,a.works_count,a.status,a.sort,
		f.id,f.slug,f.name,f.sort,
		c.id,c.slug,c.name,c.sort
		FROM products p
		JOIN artists a ON a.id=p.artist_id
		JOIN fandoms f ON f.id=p.fandom_id
		JOIN categories c ON c.id=p.category_id
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY ` + order

	if f.Limit > 0 {
		q += " LIMIT " + s.ph(n)
		n++
		args = append(args, f.Limit)
		if f.Offset > 0 {
			q += " OFFSET " + s.ph(n)
			n++
			args = append(args, f.Offset)
		}
	}

	rows, err := s.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Product
	for rows.Next() {
		p, err := s.scanProduct(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// подгружаем варианты и картинки пачкой
	if len(out) > 0 {
		ids := make([]any, len(out))
		for i, p := range out {
			ids[i] = p.ID
		}
		variants, _ := s.loadVariants(ctx, ids)
		images, _ := s.loadImages(ctx, ids)
		for i := range out {
			out[i].Variants = variants[out[i].ID]
			out[i].Images = images[out[i].ID]
		}
	}
	return out, nil
}

func (s *Store) GetProductBySlug(ctx context.Context, slug string) (*model.Product, error) {
	// публичный lookup: драфт не виден нигде, включая прямую ссылку (админка ходит по id)
	q := `SELECT p.id,p.slug,p.title,p.description,p.base_price,p.old_price,p.status,p.eta,p.tint,p.featured,p.sort,
		a.id,a.slug,a.name,a.bio,a.avatar_url,a.socials,a.works_count,a.status,a.sort,
		f.id,f.slug,f.name,f.sort,
		c.id,c.slug,c.name,c.sort
		FROM products p
		JOIN artists a ON a.id=p.artist_id
		JOIN fandoms f ON f.id=p.fandom_id
		JOIN categories c ON c.id=p.category_id
		WHERE p.slug=` + s.ph(1) + ` AND p.status!='draft'`
	row := s.DB.QueryRowContext(ctx, q, slug)
	p, err := s.scanProduct(row.Scan)
	if err != nil {
		return nil, err
	}
	vs, _ := s.loadVariants(ctx, []any{p.ID})
	ps, _ := s.loadImages(ctx, []any{p.ID})
	p.Variants = vs[p.ID]
	p.Images = ps[p.ID]
	return p, nil
}

func (s *Store) scanProduct(scan func(...any) error) (*model.Product, error) {
	var p model.Product
	var a model.Artist
	var f model.Fandom
	var c model.Category
	var socials string
	var featured int
	if err := scan(
		&p.ID, &p.Slug, &p.Title, &p.Description, &p.BasePrice, &p.OldPrice, &p.Status, &p.ETA, &p.Tint, &featured, &p.Sort,
		&a.ID, &a.Slug, &a.Name, &a.Bio, &a.AvatarURL, &socials, &a.WorksCount, &a.Status, &a.Sort,
		&f.ID, &f.Slug, &f.Name, &f.Sort,
		&c.ID, &c.Slug, &c.Name, &c.Sort,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	a.Socials = parseSocials(socials)
	p.Featured = featured == 1
	p.Artist = a
	p.Fandom = f
	p.Category = c
	return &p, nil
}

func (s *Store) loadVariants(ctx context.Context, ids []any) (map[int64][]model.Variant, error) {
	out := map[int64][]model.Variant{}
	if len(ids) == 0 {
		return out, nil
	}
	q := `SELECT id, product_id, name, sku, price_delta, stock_qty, sort FROM product_variants WHERE product_id IN (` + s.phList(1, len(ids)) + ") ORDER BY sort, id"
	rows, err := s.DB.QueryContext(ctx, q, ids...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var v model.Variant
		if err := rows.Scan(&v.ID, &v.ProductID, &v.Name, &v.SKU, &v.PriceDelta, &v.StockQty, &v.Sort); err != nil {
			return out, err
		}
		out[v.ProductID] = append(out[v.ProductID], v)
	}
	return out, rows.Err()
}

func (s *Store) loadImages(ctx context.Context, ids []any) (map[int64][]model.Image, error) {
	out := map[int64][]model.Image{}
	if len(ids) == 0 {
		return out, nil
	}
	q := `SELECT id, product_id, url, alt, width, height, is_primary, sort FROM product_images WHERE product_id IN (` + s.phList(1, len(ids)) + ") ORDER BY is_primary DESC, sort, id"
	rows, err := s.DB.QueryContext(ctx, q, ids...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var im model.Image
		var prim int
		if err := rows.Scan(&im.ID, &im.ProductID, &im.URL, &im.Alt, &im.Width, &im.Height, &prim, &im.Sort); err != nil {
			return out, err
		}
		im.IsPrimary = prim == 1
		out[im.ProductID] = append(out[im.ProductID], im)
	}
	return out, rows.Err()
}
