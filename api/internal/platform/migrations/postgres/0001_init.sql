-- 0001_init.sql (postgres)
-- Цены в копейках (INTEGER). Timestamps — TIMESTAMPTZ. JSON — JSONB.

CREATE TABLE artists (
  id          BIGSERIAL PRIMARY KEY,
  slug        TEXT NOT NULL UNIQUE,
  name        TEXT NOT NULL,
  bio         TEXT NOT NULL DEFAULT '',
  avatar_url  TEXT NOT NULL DEFAULT '',
  socials     JSONB NOT NULL DEFAULT '{}',
  works_count INTEGER NOT NULL DEFAULT 0,
  status      TEXT NOT NULL DEFAULT 'published',
  sort        INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE fandoms (
  id    BIGSERIAL PRIMARY KEY,
  slug  TEXT NOT NULL UNIQUE,
  name  TEXT NOT NULL,
  sort  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE categories (
  id    BIGSERIAL PRIMARY KEY,
  slug  TEXT NOT NULL UNIQUE,
  name  TEXT NOT NULL,
  sort  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE products (
  id          BIGSERIAL PRIMARY KEY,
  slug        TEXT NOT NULL UNIQUE,
  title       TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  artist_id   BIGINT NOT NULL REFERENCES artists(id) ON DELETE RESTRICT,
  fandom_id   BIGINT NOT NULL REFERENCES fandoms(id) ON DELETE RESTRICT,
  category_id BIGINT NOT NULL REFERENCES categories(id) ON DELETE RESTRICT,
  base_price  INTEGER NOT NULL,
  old_price   INTEGER NOT NULL DEFAULT 0,
  status      TEXT NOT NULL DEFAULT 'draft',
  eta         TEXT NOT NULL DEFAULT '',
  tint        TEXT NOT NULL DEFAULT 'luna',
  featured    INTEGER NOT NULL DEFAULT 0,
  sort        INTEGER NOT NULL DEFAULT 0,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_artist   ON products(artist_id);
CREATE INDEX idx_products_fandom   ON products(fandom_id);
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_status   ON products(status);

CREATE TABLE product_variants (
  id          BIGSERIAL PRIMARY KEY,
  product_id  BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  name        TEXT NOT NULL,
  sku         TEXT NOT NULL DEFAULT '',
  price_delta INTEGER NOT NULL DEFAULT 0,
  stock_qty   INTEGER NOT NULL DEFAULT 0,
  sort        INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_variants_product ON product_variants(product_id);

CREATE TABLE product_images (
  id         BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  url        TEXT NOT NULL,
  alt        TEXT NOT NULL DEFAULT '',
  width      INTEGER NOT NULL DEFAULT 0,
  height     INTEGER NOT NULL DEFAULT 0,
  is_primary INTEGER NOT NULL DEFAULT 0,
  sort       INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_images_product ON product_images(product_id);

CREATE TABLE pages (
  id              BIGSERIAL PRIMARY KEY,
  slug            TEXT NOT NULL UNIQUE,
  title           TEXT NOT NULL,
  body            TEXT NOT NULL DEFAULT '',
  meta_description TEXT NOT NULL DEFAULT '',
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE users (
  id            BIGSERIAL PRIMARY KEY,
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL DEFAULT '',
  name          TEXT NOT NULL DEFAULT '',
  role          TEXT NOT NULL DEFAULT 'customer',
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE preorder_requests (
  id                 BIGSERIAL PRIMARY KEY,
  product_id         BIGINT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  variant_id         BIGINT REFERENCES product_variants(id) ON DELETE SET NULL,
  contact_name       TEXT NOT NULL,
  contact            TEXT NOT NULL,
  contact_kind       TEXT NOT NULL DEFAULT 'email',
  status             TEXT NOT NULL DEFAULT 'new',
  note               TEXT NOT NULL DEFAULT '',
  converted_order_id BIGINT,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_preorders_status ON preorder_requests(status);

CREATE TABLE orders (
  id              BIGSERIAL PRIMARY KEY,
  number          TEXT NOT NULL UNIQUE,
  status          TEXT NOT NULL DEFAULT 'created',
  payment_status  TEXT NOT NULL DEFAULT 'pending',
  is_preorder     INTEGER NOT NULL DEFAULT 0,
  customer_name   TEXT NOT NULL DEFAULT '',
  customer_email  TEXT NOT NULL DEFAULT '',
  customer_phone  TEXT NOT NULL DEFAULT '',
  total           INTEGER NOT NULL DEFAULT 0,
  eta             TEXT NOT NULL DEFAULT '',
  delivery_method TEXT NOT NULL DEFAULT '',
  delivery_data   JSONB NOT NULL DEFAULT '{}',
  tracking_number TEXT NOT NULL DEFAULT '',
  user_id         BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_email  ON orders(customer_email);
CREATE INDEX idx_orders_status ON orders(status);

CREATE TABLE order_items (
  id           BIGSERIAL PRIMARY KEY,
  order_id     BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  product_id   BIGINT REFERENCES products(id) ON DELETE SET NULL,
  snapshot     JSONB NOT NULL,
  variant_name TEXT NOT NULL DEFAULT '',
  price        INTEGER NOT NULL,
  qty          INTEGER NOT NULL DEFAULT 1,
  line_total   INTEGER NOT NULL
);
CREATE INDEX idx_items_order ON order_items(order_id);

CREATE TABLE order_status_events (
  id         BIGSERIAL PRIMARY KEY,
  order_id   BIGINT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
  status     TEXT NOT NULL,
  note       TEXT NOT NULL DEFAULT '',
  created_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_events_order ON order_status_events(order_id);
