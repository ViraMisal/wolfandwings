package store

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/wolfandwings/api/internal/model"
)

// GetOrderByNumber — гостевой просмотр: номер + email (защита от IDOR).
func (s *Store) GetOrderByNumber(ctx context.Context, number, email string) (*model.Order, error) {
	var o model.Order
	var created string
	err := s.DB.QueryRowContext(ctx, `SELECT id, number, status, payment_status, is_preorder, customer_email,
		total, eta, delivery_method, tracking_number, CAST(created_at AS TEXT)
		FROM orders WHERE number=`+s.ph(1)+` AND LOWER(customer_email)=LOWER(`+s.ph(2)+`)`, number, email).
		Scan(&o.ID, &o.Number, &o.Status, &o.PaymentStatus, &o.IsPreorder, &o.CustomerEmail,
			&o.Total, &o.ETA, &o.DeliveryMethod, &o.TrackingNumber, &created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	o.CreatedAt = created

	items, err := s.orderItems(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Items = items

	events, err := s.orderEvents(ctx, o.ID)
	if err != nil {
		return nil, err
	}
	o.Events = events
	return &o, nil
}

func (s *Store) orderItems(ctx context.Context, orderID int64) ([]model.OrderItem, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT snapshot, variant_name, price, qty, line_total
		FROM order_items WHERE order_id=`+s.ph(1), orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// не nil: фронт делает items.length и не должен падать на заказе без позиций
	out := []model.OrderItem{}
	for rows.Next() {
		var snap string
		var it model.OrderItem
		if err := rows.Scan(&snap, &it.VariantName, &it.Price, &it.Qty, &it.LineTotal); err != nil {
			return nil, err
		}
		var sv struct {
			Title  string `json:"title"`
			Artist string `json:"artist"`
			Image  string `json:"image"`
			Slug   string `json:"product_slug"`
		}
		_ = json.Unmarshal([]byte(snap), &sv)
		it.Title = sv.Title
		it.Artist = sv.Artist
		it.Image = sv.Image
		it.ProductSlug = sv.Slug
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) orderEvents(ctx context.Context, orderID int64) ([]model.OrderEvent, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT status, note, CAST(created_at AS TEXT)
		FROM order_status_events WHERE order_id=`+s.ph(1)+` ORDER BY created_at`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []model.OrderEvent{}
	for rows.Next() {
		var e model.OrderEvent
		if err := rows.Scan(&e.Status, &e.Note, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
