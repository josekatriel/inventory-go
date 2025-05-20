-- Database schema for inventory management system

-- Create extension for UUID support (if not already created)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    parent_id VARCHAR(36) REFERENCES categories(id),
    breadcrumbs JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36) REFERENCES products(id),
    stock INTEGER NOT NULL DEFAULT 0,
    reorder_level INTEGER DEFAULT 0,
    child_category_id VARCHAR(36) REFERENCES categories(id),
    basic JSONB NOT NULL,
    price JSONB NOT NULL,
    weight JSONB NOT NULL,
    inventory_activity JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Product images table
CREATE TABLE IF NOT EXISTS images (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Customers table
CREATE TABLE IF NOT EXISTS customers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Suppliers table
CREATE TABLE IF NOT EXISTS suppliers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(20),
    address TEXT,
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    postal_code VARCHAR(20),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- StockIn table
CREATE TABLE IF NOT EXISTS stock_ins (
    id VARCHAR(36) PRIMARY KEY,
    reference_no VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    order_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    supplier_id VARCHAR(36) REFERENCES suppliers(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- StockIn items table
CREATE TABLE IF NOT EXISTS stock_in_items (
    id VARCHAR(36) PRIMARY KEY,
    stock_in_id VARCHAR(36) NOT NULL REFERENCES stock_ins(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Sales table
CREATE TABLE IF NOT EXISTS sales (
    id VARCHAR(36) PRIMARY KEY,
    reference_no VARCHAR(100) NOT NULL UNIQUE,
    customer_id VARCHAR(36) REFERENCES customers(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    sale_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    discount DECIMAL(10, 2) DEFAULT 0,
    tax DECIMAL(10, 2) DEFAULT 0,
    shipping_fee DECIMAL(10, 2) DEFAULT 0,
    grand_total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    payment_status VARCHAR(20) DEFAULT 'unpaid',
    payment_method VARCHAR(50),
    platform VARCHAR(50) DEFAULT 'pos',
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Sale items table
CREATE TABLE IF NOT EXISTS sale_items (
    id VARCHAR(36) PRIMARY KEY,
    sale_id VARCHAR(36) NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Rejects table (for stock decreases/inventory write-offs)
CREATE TABLE IF NOT EXISTS rejects (
    id VARCHAR(36) PRIMARY KEY,
    reference_no VARCHAR(100) NOT NULL UNIQUE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    reject_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reason TEXT,
    total DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Reject items table
CREATE TABLE IF NOT EXISTS reject_items (
    id VARCHAR(36) PRIMARY KEY,
    reject_id VARCHAR(36) NOT NULL REFERENCES rejects(id) ON DELETE CASCADE,
    product_id VARCHAR(36) NOT NULL REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    subtotal DECIMAL(10, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(child_category_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(((basic->>'status')::integer));
CREATE INDEX IF NOT EXISTS idx_images_product_id ON images(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_ins_supplier_id ON stock_ins(supplier_id);
CREATE INDEX IF NOT EXISTS idx_stock_ins_status ON stock_ins(status);
CREATE INDEX IF NOT EXISTS idx_stock_in_items_stock_in_id ON stock_in_items(stock_in_id);
CREATE INDEX IF NOT EXISTS idx_stock_in_items_product_id ON stock_in_items(product_id);
CREATE INDEX IF NOT EXISTS idx_sales_customer_id ON sales(customer_id);
CREATE INDEX IF NOT EXISTS idx_sales_status ON sales(status);
CREATE INDEX IF NOT EXISTS idx_sale_items_sale_id ON sale_items(sale_id);
CREATE INDEX IF NOT EXISTS idx_sale_items_product_id ON sale_items(product_id);
CREATE INDEX IF NOT EXISTS idx_rejects_status ON rejects(status);
CREATE INDEX IF NOT EXISTS idx_reject_items_reject_id ON reject_items(reject_id);
CREATE INDEX IF NOT EXISTS idx_reject_items_product_id ON reject_items(product_id);

-- Create triggers for updating the timestamps
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for all tables
CREATE TRIGGER trigger_update_categories_timestamp
BEFORE UPDATE ON categories
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_products_timestamp
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Images table doesn't need an update trigger as it doesn't have updated_at

CREATE TRIGGER trigger_update_customers_timestamp
BEFORE UPDATE ON customers
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_suppliers_timestamp
BEFORE UPDATE ON suppliers
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_stock_ins_timestamp
BEFORE UPDATE ON stock_ins
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_stock_in_items_timestamp
BEFORE UPDATE ON stock_in_items
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_sales_timestamp
BEFORE UPDATE ON sales
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_sale_items_timestamp
BEFORE UPDATE ON sale_items
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_rejects_timestamp
BEFORE UPDATE ON rejects
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trigger_update_reject_items_timestamp
BEFORE UPDATE ON reject_items
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- Create a function to generate UUIDs for VARCHAR(36) ID columns
CREATE OR REPLACE FUNCTION generate_uuid_for_id()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.id IS NULL OR NEW.id = '' THEN
        NEW.id := gen_random_uuid()::text;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add UUID generation triggers for all tables
CREATE TRIGGER trigger_categories_generate_uuid
BEFORE INSERT ON categories
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_products_generate_uuid
BEFORE INSERT ON products
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_images_generate_uuid
BEFORE INSERT ON images
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_customers_generate_uuid
BEFORE INSERT ON customers
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_suppliers_generate_uuid
BEFORE INSERT ON suppliers
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_stock_ins_generate_uuid
BEFORE INSERT ON stock_ins
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_stock_in_items_generate_uuid
BEFORE INSERT ON stock_in_items
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_sales_generate_uuid
BEFORE INSERT ON sales
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_sale_items_generate_uuid
BEFORE INSERT ON sale_items
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_rejects_generate_uuid
BEFORE INSERT ON rejects
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

CREATE TRIGGER trigger_reject_items_generate_uuid
BEFORE INSERT ON reject_items
FOR EACH ROW EXECUTE FUNCTION generate_uuid_for_id();

-- Add trigger for updating product quantity after stock-in
CREATE OR REPLACE FUNCTION update_product_quantity_after_stock_in()
RETURNS TRIGGER AS $$
BEGIN
    -- Only increase quantity for completed stock-ins
    IF NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed') THEN
        UPDATE products
        SET quantity = quantity + (
            SELECT COALESCE(SUM(quantity), 0)
            FROM stock_in_items
            WHERE stock_in_id = NEW.id AND deleted_at IS NULL
        )
        WHERE id IN (
            SELECT product_id
            FROM stock_in_items
            WHERE stock_in_id = NEW.id AND deleted_at IS NULL
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_product_quantity_after_stock_in
AFTER UPDATE ON stock_ins
FOR EACH ROW EXECUTE FUNCTION update_product_quantity_after_stock_in();

-- Add trigger for updating product quantity after sale
CREATE OR REPLACE FUNCTION update_product_quantity_after_sale()
RETURNS TRIGGER AS $$
BEGIN
    -- Only decrease quantity for completed sales
    IF NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed') THEN
        UPDATE products
        SET quantity = quantity - (
            SELECT COALESCE(SUM(quantity), 0)
            FROM sale_items
            WHERE sale_id = NEW.id AND deleted_at IS NULL
        )
        WHERE id IN (
            SELECT product_id
            FROM sale_items
            WHERE sale_id = NEW.id AND deleted_at IS NULL
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_product_quantity_after_sale
AFTER UPDATE ON sales
FOR EACH ROW EXECUTE FUNCTION update_product_quantity_after_sale();

-- Add trigger for updating product quantity after reject
CREATE OR REPLACE FUNCTION update_product_quantity_after_reject()
RETURNS TRIGGER AS $$
BEGIN
    -- Only decrease quantity for completed rejects
    IF NEW.status = 'completed' AND (OLD.status IS NULL OR OLD.status != 'completed') THEN
        UPDATE products
        SET stock = stock - (
            SELECT COALESCE(SUM(quantity), 0)
            FROM reject_items
            WHERE reject_id = NEW.id AND deleted_at IS NULL
        )
        WHERE id IN (
            SELECT product_id
            FROM reject_items
            WHERE reject_id = NEW.id AND deleted_at IS NULL
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_product_quantity_after_reject
AFTER UPDATE ON rejects
FOR EACH ROW EXECUTE FUNCTION update_product_quantity_after_reject();
