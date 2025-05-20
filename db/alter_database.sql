-- Migration script to add missing columns to existing tables
-- Run this script to update your existing database structure

-- Add missing fields to suppliers table
ALTER TABLE suppliers 
ADD COLUMN IF NOT EXISTS total_purchases INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_spent DECIMAL(10, 2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_order_at TIMESTAMP WITH TIME ZONE;

-- Add missing fields to categories table
ALTER TABLE categories 
ADD COLUMN IF NOT EXISTS status INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS sort_order INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS image_url TEXT;

-- Add missing fields to customers table
ALTER TABLE customers
ADD COLUMN IF NOT EXISTS total_orders INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_spent DECIMAL(10, 2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS last_order_at TIMESTAMP WITH TIME ZONE;

-- Add missing fields to sales table
ALTER TABLE sales
ADD COLUMN IF NOT EXISTS paid DECIMAL(10, 2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS balance DECIMAL(10, 2) GENERATED ALWAYS AS (grand_total - paid) STORED;

-- Add missing fields to sale_items table
ALTER TABLE sale_items
ADD COLUMN IF NOT EXISTS tax DECIMAL(10, 2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS discount DECIMAL(10, 2) DEFAULT 0;

-- Create indexes for new fields
CREATE INDEX IF NOT EXISTS idx_categories_status ON categories(status);
CREATE INDEX IF NOT EXISTS idx_categories_sort_order ON categories(sort_order);
