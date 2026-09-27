package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/wolfandwings/api/internal/model"
)

// ListAllProducts — все товары (вкл. draft), для админки.
func (s *Store) ListAllProducts(ctx context.Context) ([]model.Product, error) {
	q := `SELECT p.id,p.slug,p.title,p.description,p.base_price,p.old_price,p.status,p.eta,p.tint,p.featured,p.sort,
		a.id,a.slug,a.name,'','','{}',0,a.status,a.sort,
		f.id,f.slug,f.name,f.sort,
		c.id,c.slug,c.name,c.sort
		FROM products p
		JOIN artists a ON a.id=p.artist_id
		JOIN fandoms f ON f.id=p.fandom_id
		JOIN categories c ON c.id=p.category_id
		ORDER BY p.id DESC`
	rows, err := s.DB.QueryContext(ctx, q)
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
	if len(out) > 0 {
		ids := make([]any, len(out))
		for i, p := range out {
			ids[i] = p.ID
		}
		images, _ := s.loadImages(ctx, ids)
		for i := range out {
			out[i].Images = images[out[i].ID]
		}
	}
	return out, nil
}

// GetProductByID — для редактирования (с вариантами).
func (s *Store) GetProductByID(ctx context.Context, id int64) (*model.Product, error) {
	q := `SELECT p.id,p.slug,p.title,p.description,p.base_price,p.old_price,p.status,p.eta,p.tint,p.featured,p.sort,
		a.id,a.slug,a.name,'','','{}',0,a.status,a.sort,
		f.id,f.slug,f.name,f.sort,
		c.id,c.slug,c.name,c.sort
		FROM products p
		JOIN artists a ON a.id=p.artist_id
		JOIN fandoms f ON f.id=p.fandom_id
		JOIN categories c ON c.id=p.category_id
		WHERE p.id=` + s.ph(1)
	row := s.DB.QueryRowContext(ctx, q, id)
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

type ProductInput struct {
	Slug, Title, Description       string
	ArtistID, FandomID, CategoryID int64
	BasePrice, OldPrice            int
	Status, ETA, Tint              string
	Featured                       bool
	Variants                       []VariantInput
	Images                         []ImageInput
}
type VariantInput struct {
	Name, SKU            string
	PriceDelta, StockQty int
}
type ImageInput struct {
	URL, Alt  string
	IsPrimary bool
}

func (s *Store) CreateProduct(ctx context.Context, in ProductInput) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	feat := 0
	if in.Featured {
		feat = 1
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck
	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO products
		(slug,title,description,artist_id,fandom_id,category_id,base_price,old_price,status,eta,tint,featured,sort,created_at,updated_at)
		VALUES (`+s.phList(1, 15)+`) RETURNING id`,
		in.Slug, in.Title, in.Description, in.ArtistID, in.FandomID, in.CategoryID, in.BasePrice, in.OldPrice, in.Status, in.ETA, in.Tint, feat, 0, now, now).Scan(&id)
	if err != nil {
		return 0, err
	}
	if err := s.replaceVariants(ctx, tx, id, in.Variants); err != nil {
		return 0, err
	}
	if err := s.replaceImages(ctx, tx, id, in.Images); err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

func (s *Store) UpdateProduct(ctx context.Context, id int64, in ProductInput) error {
	now := time.Now().UTC().Format(time.RFC3339)
	feat := 0
	if in.Featured {
		feat = 1
	}
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	_, err = tx.ExecContext(ctx, `UPDATE products SET
		slug=`+s.ph(1)+`,title=`+s.ph(2)+`,description=`+s.ph(3)+`,artist_id=`+s.ph(4)+`,fandom_id=`+s.ph(5)+`,category_id=`+s.ph(6)+`,
		base_price=`+s.ph(7)+`,old_price=`+s.ph(8)+`,status=`+s.ph(9)+`,eta=`+s.ph(10)+`,tint=`+s.ph(11)+`,featured=`+s.ph(12)+`,updated_at=`+s.ph(13)+`
		WHERE id=`+s.ph(14),
		in.Slug, in.Title, in.Description, in.ArtistID, in.FandomID, in.CategoryID, in.BasePrice, in.OldPrice, in.Status, in.ETA, in.Tint, feat, now, id)
	if err != nil {
		return err
	}
	if err := s.replaceVariants(ctx, tx, id, in.Variants); err != nil {
		return err
	}
	if err := s.replaceImages(ctx, tx, id, in.Images); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) replaceVariants(ctx context.Context, tx *sql.Tx, productID int64, vs []VariantInput) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM product_variants WHERE product_id=`+s.ph(1), productID); err != nil {
		return err
	}
	for i, v := range vs {
		if v.Name == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO product_variants (product_id,name,sku,price_delta,stock_qty,sort) VALUES (`+s.phList(1, 6)+`)`,
			productID, v.Name, v.SKU, v.PriceDelta, v.StockQty, i); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) replaceImages(ctx context.Context, tx *sql.Tx, productID int64, ims []ImageInput) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM product_images WHERE product_id=`+s.ph(1), productID); err != nil {
		return err
	}
	for i, im := range ims {
		if im.URL == "" {
			continue
		}
		prim := 0
		if im.IsPrimary {
			prim = 1
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO product_images (product_id,url,alt,is_primary,sort) VALUES (`+s.phList(1, 5)+`)`,
			productID, im.URL, im.Alt, prim, i); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListArtistsAll(ctx context.Context) ([]model.Artist, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, slug, name, bio, avatar_url, socials, works_count, status, sort FROM artists ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanArtists(rows)
}

func (s *Store) UpdatePreorderStatus(ctx context.Context, id int64, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.ExecContext(ctx, `UPDATE preorder_requests SET status=`+s.ph(1)+`, updated_at=`+s.ph(2)+` WHERE id=`+s.ph(3), status, now, id)
	return err
}

func (s *Store) DeleteProduct(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM products WHERE id=`+s.ph(1), id)
	return err
}

func (s *Store) CreateArtist(ctx context.Context, name, slug, bio, avatar, socials string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := s.DB.QueryRowContext(ctx, `INSERT INTO artists (slug,name,bio,avatar_url,socials,works_count,status,sort,created_at,updated_at)
		VALUES (`+s.phList(1, 4)+", "+s.jsonParam(5)+", "+s.phList(6, 5)+`) RETURNING id`, slug, name, bio, avatar, socials, 0, "published", 0, now, now).Scan(&id)
	return id, err
}

func (s *Store) UpdateArtist(ctx context.Context, id int64, name, slug, bio, avatar, socials string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.ExecContext(ctx, `UPDATE artists SET slug=`+s.ph(1)+`,name=`+s.ph(2)+`,bio=`+s.ph(3)+
		`,avatar_url=`+s.ph(4)+`,socials=`+s.jsonParam(5)+`,updated_at=`+s.ph(6)+` WHERE id=`+s.ph(7),
		slug, name, bio, avatar, socials, now, id)
	return err
}

func (s *Store) DeleteArtist(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM artists WHERE id=`+s.ph(1), id)
	return err
}

// GetArtistByID — для формы редактирования.
func (s *Store) GetArtistByID(ctx context.Context, id int64) (*model.Artist, error) {
	var a model.Artist
	var socials string
	err := s.DB.QueryRowContext(ctx, `SELECT id, slug, name, bio, avatar_url, socials, works_count, status, sort
		FROM artists WHERE id=`+s.ph(1), id).Scan(&a.ID, &a.Slug, &a.Name, &a.Bio, &a.AvatarURL, &socials, &a.WorksCount, &a.Status, &a.Sort)
	if err != nil {
		return nil, ErrNotFound
	}
	a.Socials = parseSocials(socials)
	return &a, nil
}

// ListPagesAll — все страницы для админки.
func (s *Store) ListPagesAll(ctx context.Context) ([]model.Page, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT id, slug, title, body FROM pages ORDER BY slug`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.Page
	for rows.Next() {
		var p model.Page
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Body); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// GetPageByID — для формы редактирования.
func (s *Store) GetPageByID(ctx context.Context, id int64) (*model.Page, error) {
	var p model.Page
	err := s.DB.QueryRowContext(ctx, `SELECT id, slug, title, body, meta_description FROM pages WHERE id=`+s.ph(1), id).Scan(&p.ID, &p.Slug, &p.Title, &p.Body, &p.MetaDescription)
	if err != nil {
		return nil, ErrNotFound
	}
	return &p, nil
}

func (s *Store) CreatePage(ctx context.Context, slug, title, body, meta string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := s.DB.QueryRowContext(ctx, `INSERT INTO pages (slug,title,body,meta_description,updated_at)
		VALUES (`+s.phList(1, 5)+`) RETURNING id`, slug, title, body, meta, now).Scan(&id)
	return id, err
}

func (s *Store) UpdatePage(ctx context.Context, id int64, slug, title, body, meta string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.DB.ExecContext(ctx, `UPDATE pages SET slug=`+s.ph(1)+`,title=`+s.ph(2)+`,body=`+s.ph(3)+
		`,meta_description=`+s.ph(4)+`,updated_at=`+s.ph(5)+` WHERE id=`+s.ph(6),
		slug, title, body, meta, now, id)
	return err
}
