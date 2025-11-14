-- +goose Up
-- Optimal-style migration: Cross-database compatible products

-- Create products table
CREATE TABLE products (
    id INTEGER PRIMARY KEY,
    organization_id INTEGER NOT NULL,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    category TEXT,
    brand TEXT,
    unit_of_measure TEXT NOT NULL DEFAULT 'each',
    weight REAL,
    dimensions TEXT,
    barcode TEXT,
    tax_rate REAL DEFAULT 0.0,
    is_active INTEGER NOT NULL DEFAULT 1,
    is_trackable INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE
);

-- Create indexes for products table
CREATE UNIQUE INDEX idx_products_org_sku ON products(organization_id, sku);
CREATE INDEX idx_products_organization_id ON products(organization_id);
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category ON products(category);
CREATE INDEX idx_products_brand ON products(brand);
CREATE INDEX idx_products_barcode ON products(barcode);
CREATE INDEX idx_products_is_active ON products(is_active);
CREATE INDEX idx_products_is_trackable ON products(is_trackable);
CREATE INDEX idx_products_created_at ON products(created_at);
CREATE INDEX idx_products_org_category_active ON products(organization_id, category, is_active);

-- Create product_variants table
CREATE TABLE product_variants (
    id INTEGER PRIMARY KEY,
    product_id INTEGER NOT NULL,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    attributes TEXT, -- JSON as TEXT for cross-compatibility
    weight REAL,
    dimensions TEXT,
    barcode TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Create indexes for product_variants table
CREATE UNIQUE INDEX idx_product_variants_sku ON product_variants(sku);
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_name ON product_variants(name);
CREATE INDEX idx_product_variants_barcode ON product_variants(barcode);
CREATE INDEX idx_product_variants_is_active ON product_variants(is_active);
CREATE INDEX idx_product_variants_created_at ON product_variants(created_at);

-- Create product_prices table
CREATE TABLE product_prices (
    id INTEGER PRIMARY KEY,
    product_id INTEGER,
    product_variant_id INTEGER,
    price_type TEXT NOT NULL CHECK (price_type IN ('base', 'sale', 'wholesale', 'retail', 'cost')),
    currency TEXT NOT NULL DEFAULT 'USD',
    amount REAL NOT NULL,
    min_quantity INTEGER DEFAULT 1,
    max_quantity INTEGER,
    valid_from TEXT,
    valid_until TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE,
    FOREIGN KEY (product_variant_id) REFERENCES product_variants(id) ON DELETE CASCADE,
    -- Ensure either product_id or product_variant_id is set, but not both
    CHECK ((product_id IS NOT NULL AND product_variant_id IS NULL) OR 
           (product_id IS NULL AND product_variant_id IS NOT NULL))
);

-- Create indexes for product_prices table
CREATE INDEX idx_product_prices_product_id ON product_prices(product_id);
CREATE INDEX idx_product_prices_product_variant_id ON product_prices(product_variant_id);
CREATE INDEX idx_product_prices_price_type ON product_prices(price_type);
CREATE INDEX idx_product_prices_currency ON product_prices(currency);
CREATE INDEX idx_product_prices_is_active ON product_prices(is_active);
CREATE INDEX idx_product_prices_valid_from ON product_prices(valid_from);
CREATE INDEX idx_product_prices_valid_until ON product_prices(valid_until);
CREATE INDEX idx_product_prices_created_at ON product_prices(created_at);
CREATE INDEX idx_product_prices_product_type_active ON product_prices(product_id, price_type, is_active);
CREATE INDEX idx_product_prices_variant_type_active ON product_prices(product_variant_id, price_type, is_active);
CREATE INDEX idx_product_prices_validity ON product_prices(valid_from, valid_until, is_active);

-- +goose Down
-- Optimal-style rollback: Clean and simple

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS product_prices;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;