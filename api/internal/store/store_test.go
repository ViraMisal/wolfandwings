package store

import (
	"context"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/wolfandwings/api/internal/config"
	"github.com/wolfandwings/api/internal/platform"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := "sqlite://" + filepath.Join(t.TempDir(), "test.db")
	cfg := config.Config{Dialect: "sqlite", DatabaseURL: dsn}
	db, err := platform.Open(context.Background(), cfg, slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return New(db, "sqlite")
}

func TestSeedIfEmpty(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)

	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "secret123", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("seed: %v", err)
	}
	products, err := s.ListAllProducts(ctx)
	if err != nil {
		t.Fatalf("list products: %v", err)
	}
	if len(products) != 12 {
		t.Fatalf("want 12 products, got %d", len(products))
	}
	// драфты не видны публично
	pub, err := s.ListProducts(ctx, ProductFilter{})
	if err != nil {
		t.Fatalf("list public products: %v", err)
	}
	for _, p := range pub {
		if p.Status == "draft" {
			t.Fatalf("draft leaked to public catalog: %s", p.Slug)
		}
	}

	// повторный сид не должен дублировать контент
	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "secret123", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("reseed: %v", err)
	}
	again, _ := s.ListAllProducts(ctx)
	if len(again) != 12 {
		t.Fatalf("reseed duplicated products: %d", len(again))
	}

	// админ создан
	if _, err := s.GetUserByEmail(ctx, "admin@test.ru"); err != nil {
		t.Fatalf("admin user missing: %v", err)
	}
}

func TestListProductsFilters(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "secret123", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	byFandom, err := s.ListProducts(ctx, ProductFilter{FandomSlug: "arknights"})
	if err != nil {
		t.Fatalf("filter fandom: %v", err)
	}
	for _, p := range byFandom {
		if p.Fandom.Slug != "arknights" {
			t.Fatalf("wrong fandom in filter result: %s", p.Fandom.Slug)
		}
	}
	if len(byFandom) == 0 {
		t.Fatal("fandom filter returned nothing")
	}

	pre, err := s.ListProducts(ctx, ProductFilter{Availability: "preorder"})
	if err != nil {
		t.Fatalf("filter availability: %v", err)
	}
	for _, p := range pre {
		if p.Status != "preorder" {
			t.Fatalf("availability filter leaked %s (%s)", p.Slug, p.Status)
		}
	}

	if _, err := s.GetProductBySlug(ctx, "lappland-lapplandia"); err != nil {
		t.Fatalf("get by slug: %v", err)
	}
	if _, err := s.GetProductBySlug(ctx, "no-such-slug"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestCreatePreorderAndList(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "secret123", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	p, _ := s.GetProductBySlug(ctx, "ellen-posle-smeny")
	req, err := s.CreatePreorder(ctx, PreorderInput{
		ProductID: p.ID, ProductSlug: p.Slug, VariantID: 0,
		ContactName: "Тест", Contact: "a@b.ru", ContactKind: "email",
	})
	if err != nil {
		t.Fatalf("create preorder: %v", err)
	}
	if req.Status != "new" {
		t.Fatalf("want status new, got %s", req.Status)
	}

	list, err := s.ListPreorders(ctx)
	if err != nil {
		t.Fatalf("list preorders: %v", err)
	}
	if len(list) != 1 || list[0].Contact != "a@b.ru" {
		t.Fatalf("unexpected preorders: %+v", list)
	}

	if err := s.UpdatePreorderStatus(ctx, req.ID, "contacted"); err != nil {
		t.Fatalf("update status: %v", err)
	}
	list, _ = s.ListPreorders(ctx)
	if list[0].Status != "contacted" {
		t.Fatalf("status not updated: %s", list[0].Status)
	}
}

func TestGetOrderByNumber(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "secret123", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("seed: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	snap, _ := json.Marshal(map[string]string{
		"title": "Лаппланд в Лаппландии", "artist": "не.сая", "image": "", "product_slug": "lappland-lapplandia",
	})

	var orderID int64
	if err := s.DB.QueryRowContext(ctx, `INSERT INTO orders
		(number,status,payment_status,is_preorder,customer_email,total,created_at,updated_at)
		VALUES ('WW-2026-7777','paid','paid',0,'buyer@test.ru',249000,?,?) RETURNING id`, now, now).Scan(&orderID); err != nil {
		t.Fatalf("insert order: %v", err)
	}
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO order_items (order_id,snapshot,variant_name,price,qty,line_total)
		VALUES (?,?,'XL · 90×40',249000,1,249000)`, orderID, string(snap)); err != nil {
		t.Fatalf("insert item: %v", err)
	}
	if _, err := s.DB.ExecContext(ctx, `INSERT INTO order_status_events (order_id,status,note,created_at)
		VALUES (?,'shipped','СДЭК',?)`, orderID, now); err != nil {
		t.Fatalf("insert event: %v", err)
	}

	// верная почта — заказ найден, снапшот распакован
	o, err := s.GetOrderByNumber(ctx, "WW-2026-7777", "Buyer@test.ru")
	if err != nil {
		t.Fatalf("get order: %v", err)
	}
	if len(o.Items) != 1 || o.Items[0].Title != "Лаппланд в Лаппландии" {
		t.Fatalf("items not restored: %+v", o.Items)
	}
	if len(o.Events) != 1 || o.Events[0].Status != "shipped" {
		t.Fatalf("events not restored: %+v", o.Events)
	}

	// чужая почта — IDOR-защита
	if _, err := s.GetOrderByNumber(ctx, "WW-2026-7777", "other@test.ru"); err != ErrNotFound {
		t.Fatalf("want ErrNotFound for foreign email, got %v", err)
	}

	// демо-заказ из сида тоже находится по паре номер + email
	if _, err := s.GetOrderByNumber(ctx, "WW-2026-0001", "buyer@example.com"); err != nil {
		t.Fatalf("seeded demo order missing: %v", err)
	}
}

func TestDraftNotPublic(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "secret123", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// draft-товар не существует для публичного lookup по slug
	_, err := s.DB.ExecContext(ctx, `INSERT INTO products
		(slug,title,description,artist_id,fandom_id,category_id,base_price,old_price,status,eta,tint,featured,sort,created_at,updated_at)
		SELECT 'secret-draft','Секретный','',artist_id,fandom_id,category_id,100,0,'draft','','luna',0,99,created_at,updated_at
		FROM products LIMIT 1`)
	if err != nil {
		t.Fatalf("insert draft: %v", err)
	}
	if _, err := s.GetProductBySlug(ctx, "secret-draft"); err != ErrNotFound {
		t.Fatalf("draft leaked via direct slug lookup: %v", err)
	}
	// обычный товар при этом находится
	if _, err := s.GetProductBySlug(ctx, "lappland-lapplandia"); err != nil {
		t.Fatalf("public product missing: %v", err)
	}

	// неопубликованный художник тоже скрыт
	_, err = s.DB.ExecContext(ctx, `INSERT INTO artists (slug,name,bio,socials,works_count,status,sort,created_at,updated_at)
		VALUES ('ghost','призрак','','{}',0,'draft',99,'2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`)
	if err != nil {
		t.Fatalf("insert draft artist: %v", err)
	}
	if _, err := s.GetArtistBySlug(ctx, "ghost"); err != ErrNotFound {
		t.Fatalf("draft artist leaked: %v", err)
	}
}

func TestAdminPasswordRotation(t *testing.T) {
	ctx := context.Background()
	s := openTestStore(t)
	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "first-secret", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("seed: %v", err)
	}
	// повторный сид с новым паролем из env — ротация существующего админа
	if err := s.SeedIfEmpty(ctx, "admin@test.ru", "second-secret", slog.New(slog.DiscardHandler)); err != nil {
		t.Fatalf("reseed: %v", err)
	}
	u, err := s.GetUserByEmail(ctx, "admin@test.ru")
	if err != nil {
		t.Fatalf("get admin: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("second-secret")) != nil {
		t.Fatal("password was not rotated")
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("first-secret")) == nil {
		t.Fatal("old password still valid")
	}
}
