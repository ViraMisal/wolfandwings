package model

import "time"

type Artist struct {
	ID         int64             `json:"id"`
	Slug       string            `json:"slug"`
	Name       string            `json:"name"`
	Bio        string            `json:"bio"`
	AvatarURL  string            `json:"avatar_url"`
	Socials    map[string]string `json:"socials"`
	WorksCount int               `json:"works_count"`
	Status     string            `json:"status"`
	Sort       int               `json:"-"`
	CreatedAt  time.Time         `json:"-"`
	UpdatedAt  time.Time         `json:"-"`
}

type Fandom struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Sort int    `json:"-"`
}

type Category struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Sort int    `json:"-"`
}

type Variant struct {
	ID         int64  `json:"id"`
	ProductID  int64  `json:"-"`
	Name       string `json:"name"`
	SKU        string `json:"sku"`
	PriceDelta int    `json:"price_delta"`
	StockQty   int    `json:"stock_qty"`
	Sort       int    `json:"-"`
}

type Image struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"-"`
	URL       string `json:"url"`
	Alt       string `json:"alt"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	IsPrimary bool   `json:"is_primary"`
	Sort      int    `json:"-"`
}

type Product struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Artist      Artist    `json:"artist"`
	Fandom      Fandom    `json:"fandom"`
	Category    Category  `json:"category"`
	BasePrice   int       `json:"base_price"`
	OldPrice    int       `json:"old_price"`
	Status      string    `json:"status"`
	ETA         string    `json:"eta"`
	Tint        string    `json:"tint"`
	Featured    bool      `json:"featured"`
	Sort        int       `json:"-"`
	Variants    []Variant `json:"variants"`
	Images      []Image   `json:"images"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

type Page struct {
	ID              int64  `json:"id"`
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	Body            string `json:"body"`
	MetaDescription string `json:"meta_description"`
}

// PreorderRequest — заявка-лид (без оплаты)
type PreorderRequest struct {
	ID           int64  `json:"id"`
	ProductID    int64  `json:"product_id"`
	ProductSlug  string `json:"product_slug"`
	ProductTitle string `json:"product_title"`
	VariantID    int64  `json:"variant_id"`
	VariantName  string `json:"variant_name"`
	ContactName  string `json:"contact_name"`
	Contact      string `json:"contact"`
	ContactKind  string `json:"contact_kind"`
	Status       string `json:"status"`
	Note         string `json:"note"`
	CreatedAt    string `json:"created_at"`
}

// Order — заказ
type Order struct {
	ID             int64        `json:"id"`
	Number         string       `json:"number"`
	Status         string       `json:"status"`
	PaymentStatus  string       `json:"payment_status"`
	IsPreorder     bool         `json:"is_preorder"`
	CustomerName   string       `json:"customer_name"`
	CustomerEmail  string       `json:"customer_email"`
	CustomerPhone  string       `json:"customer_phone"`
	Total          int          `json:"total"`
	ETA            string       `json:"eta"`
	DeliveryMethod string       `json:"delivery_method"`
	TrackingNumber string       `json:"tracking_number"`
	Items          []OrderItem  `json:"items"`
	Events         []OrderEvent `json:"events"`
	CreatedAt      string       `json:"created_at"`
}

type OrderItem struct {
	ID          int64  `json:"id"`
	ProductSlug string `json:"product_slug"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Image       string `json:"image"`
	VariantName string `json:"variant_name"`
	Price       int    `json:"price"`
	Qty         int    `json:"qty"`
	LineTotal   int    `json:"line_total"`
}

type OrderEvent struct {
	Status    string `json:"status"`
	Note      string `json:"note"`
	CreatedAt string `json:"created_at"`
}

// User — аккаунт (админ / покупатель).
type User struct {
	ID           int64     `json:"-"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"-"`
}
