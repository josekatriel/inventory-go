# Inventory Go

A RESTful inventory management API built with Go, using Supabase PostgreSQL for data storage.

## Overview

This project provides a comprehensive inventory management system with the following features:
- Product management with variants and categories
- Customer management
- Sales tracking
- RESTful API endpoints

## Tech Stack

- **Backend**: Go with gorilla/mux for routing
- **Database**: PostgreSQL (Supabase)
- **Database Driver**: pgx/v5

## Project Structure

```
inventory-go/
├── db/              # Database connection setup
├── handlers/        # HTTP request handlers
├── models/          # Data models
├── repositories/    # Database operations
├── routes/          # API route definitions
├── .env             # Environment variables (not committed to version control)
├── env.example      # Example environment variables template
├── main.go          # Application entry point
└── README.md        # This file
```

## Setup Instructions

### Prerequisites

- Go 1.18 or newer
- Supabase account with a project

### Database Setup

1. Create a new Supabase project
2. Run the SQL migration scripts to create the necessary tables:

```sql
-- Create extension for UUID support
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create categories table
CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36) REFERENCES categories(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    slug VARCHAR(255) UNIQUE,
    status INTEGER DEFAULT 1,
    sort_order INTEGER DEFAULT 0,
    image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(36) PRIMARY KEY,
    parent_id VARCHAR(36) REFERENCES products(id),
    stock INTEGER NOT NULL DEFAULT 0,
    child_category_id VARCHAR(36) REFERENCES categories(id),
    basic JSONB NOT NULL,
    price JSONB NOT NULL,
    weight JSONB NOT NULL,
    inventory_activity JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create images table
CREATE TABLE IF NOT EXISTS images (
    id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) REFERENCES products(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT false,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create customers table
CREATE TABLE IF NOT EXISTS customers (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(50),
    address TEXT,
    total_orders INTEGER DEFAULT 0,
    total_spent DECIMAL(15,2) DEFAULT 0,
    last_order_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create sales table
CREATE TABLE IF NOT EXISTS sales (
    id VARCHAR(36) PRIMARY KEY,
    reference_no VARCHAR(50) UNIQUE NOT NULL,
    status INTEGER DEFAULT 1,
    sale_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    note TEXT,
    total DECIMAL(15,2) NOT NULL DEFAULT 0,
    paid DECIMAL(15,2) NOT NULL DEFAULT 0,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    customer_id VARCHAR(36) REFERENCES customers(id),
    platform VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create sale_items table
CREATE TABLE IF NOT EXISTS sale_items (
    id VARCHAR(36) PRIMARY KEY,
    sale_id VARCHAR(36) REFERENCES sales(id) ON DELETE CASCADE,
    product_id VARCHAR(36) REFERENCES products(id),
    product_name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    unit_price DECIMAL(15,2) NOT NULL,
    tax DECIMAL(15,2) DEFAULT 0,
    discount DECIMAL(15,2) DEFAULT 0,
    subtotal DECIMAL(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

### Environment Setup

1. Copy `env.example` to `.env`
2. Configure your Supabase connection details in `.env`:

```
# Database Configuration
DB_HOST=aws-0-us-east-2.pooler.supabase.com
DB_PORT=6543
DB_USER=postgres.YOUR_PROJECT_REF
DB_PASSWORD=YOUR_PASSWORD
DB_NAME=postgres

# Server Configuration
PORT=8080

DATABASE_URL=postgresql://postgres.YOUR_PROJECT_REF:YOUR_PASSWORD@aws-0-us-east-2.pooler.supabase.com:6543/postgres
```

### Running the Application

```bash
# Install dependencies
go mod download

# Run the application
go run main.go
```

By default, the server will run on port 8080 (or the port specified in your `.env` file).

## API Endpoints

### Products

- `GET /api/products` - Get all products
- `GET /api/products/{id}` - Get a product by ID
- `POST /api/products` - Create a new product
- `PUT /api/products/{id}` - Update a product
- `DELETE /api/products/{id}` - Delete a product

### Categories

- `GET /api/categories` - Get all categories
- `GET /api/categories/{idOrSlug}` - Get a category by ID or slug
- `GET /api/categories/{id}/children` - Get a category with its children
- `GET /api/categories/parent/{parentID}` - Get categories by parent ID
- `GET /api/categories/{id}/breadcrumbs` - Get breadcrumbs for a category
- `POST /api/categories` - Create a new category
- `PUT /api/categories/{id}` - Update a category
- `DELETE /api/categories/{id}` - Delete a category

### Customers

- `GET /api/customers` - Get all customers
- `GET /api/customers/search` - Search customers
- `GET /api/customers/top` - Get top customers
- `GET /api/customers/{id}` - Get a customer by ID
- `POST /api/customers` - Create a new customer
- `PUT /api/customers/{id}` - Update a customer
- `DELETE /api/customers/{id}` - Delete a customer

### Sales

- `GET /api/sales` - Get all sales
- `GET /api/sales/summary` - Get sales summary
- `GET /api/sales/daily` - Get daily sales
- `GET /api/sales/{id}` - Get a sale by ID
- `GET /api/sales/reference/{reference}` - Get a sale by reference number
- `GET /api/customers/{id}/sales` - Get sales by customer
- `POST /api/sales` - Create a new sale
- `PUT /api/sales/{id}` - Update a sale
- `DELETE /api/sales/{id}` - Delete a sale

## Example API Requests

### Create a Category

```bash
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Electronics",
    "description": "Electronic devices and gadgets",
    "slug": "electronics",
    "status": 1,
    "sort_order": 10,
    "image_url": "https://example.com/images/electronics.jpg"
  }'
```

### Create a Product

```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "stock": 100,
    "basic": {
      "name": "Test Product",
      "description": "This is a test product description",
      "status": 1,
      "condition": 1,
      "sku": "TP-001",
      "is_variant": false
    },
    "price": {
      "price": 150000,
      "currency": "IDR",
      "last_update_unix": 1747048000
    },
    "weight": {
      "weight": 500,
      "unit": 1
    },
    "child_category_id": "your-category-id",
    "pictures": []
  }'
```

## Troubleshooting

If you encounter connection issues with Supabase:

1. Verify your DATABASE_URL format: `postgresql://postgres.YOUR_PROJECT_REF:YOUR_PASSWORD@aws-0-us-east-2.pooler.supabase.com:6543/postgres`
2. Make sure all tables are created in the database
3. Check the logs for detailed error messages

## License

MIT
