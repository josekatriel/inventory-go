-- Add the missing reorder_level column to products table
ALTER TABLE products
ADD COLUMN IF NOT EXISTS reorder_level INTEGER DEFAULT 0;

-- Create an index for better performance on low stock queries
CREATE INDEX IF NOT EXISTS idx_products_reorder_level ON products(reorder_level);
