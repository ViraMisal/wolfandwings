package store

import (
	"context"

	"github.com/wolfandwings/api/internal/model"
)

func (s *Store) GetPageBySlug(ctx context.Context, slug string) (*model.Page, error) {
	var p model.Page
	err := s.DB.QueryRowContext(ctx, `SELECT id, slug, title, body, meta_description FROM pages WHERE slug=`+s.ph(1), slug).
		Scan(&p.ID, &p.Slug, &p.Title, &p.Body, &p.MetaDescription)
	if err != nil {
		return nil, ErrNotFound
	}
	return &p, nil
}
