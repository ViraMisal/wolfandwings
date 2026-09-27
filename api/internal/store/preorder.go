package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/wolfandwings/api/internal/model"
)

type PreorderInput struct {
	ProductID   int64
	ProductSlug string
	VariantID   int64
	ContactName string
	Contact     string
	ContactKind string
}

// VariantBelongsTo — вариант относится к этому товару (заявка не должна
// ссылаться на чужой variant_id).
func (s *Store) VariantBelongsTo(ctx context.Context, variantID, productID int64) bool {
	var n int
	if err := s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM product_variants
		WHERE id=`+s.ph(1)+` AND product_id=`+s.ph(2), variantID, productID).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

func (s *Store) CreatePreorder(ctx context.Context, in PreorderInput) (*model.PreorderRequest, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	var id int64
	err := s.DB.QueryRowContext(ctx, `INSERT INTO preorder_requests
		(product_id, variant_id, contact_name, contact, contact_kind, status, created_at, updated_at)
		VALUES (`+s.ph(1)+","+s.ph(2)+","+s.ph(3)+","+s.ph(4)+","+s.ph(5)+",'new',"+s.ph(6)+","+s.ph(7)+`)
		RETURNING id`,
		in.ProductID, nullInt64(in.VariantID), in.ContactName, in.Contact, in.ContactKind, now, now).Scan(&id)
	if err != nil {
		return nil, err
	}
	return &model.PreorderRequest{
		ID:          id,
		ProductID:   in.ProductID,
		ProductSlug: in.ProductSlug,
		ContactName: in.ContactName,
		Contact:     in.Contact,
		ContactKind: in.ContactKind,
		Status:      "new",
		CreatedAt:   now,
	}, nil
}

func (s *Store) ListPreorders(ctx context.Context) ([]model.PreorderRequest, error) {
	q := `SELECT pr.id, pr.product_id, p.slug, p.title, pr.variant_id, COALESCE(pv.name,''),
		pr.contact_name, pr.contact, pr.contact_kind, pr.status, pr.created_at
		FROM preorder_requests pr
		JOIN products p ON p.id=pr.product_id
		LEFT JOIN product_variants pv ON pv.id=pr.variant_id
		ORDER BY pr.created_at DESC`
	rows, err := s.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []model.PreorderRequest
	for rows.Next() {
		var r model.PreorderRequest
		var vid sql.NullInt64
		if err := rows.Scan(&r.ID, &r.ProductID, &r.ProductSlug, &r.ProductTitle, &vid, &r.VariantName,
			&r.ContactName, &r.Contact, &r.ContactKind, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		r.VariantID = vid.Int64
		out = append(out, r)
	}
	return out, rows.Err()
}

func nullInt64(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}
