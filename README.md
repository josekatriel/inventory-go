# Inventory Management System

A comprehensive inventory management API built with Go and Supabase PostgreSQL, designed to handle product management, stock control, sales tracking, more.

## ✨ Features

- **Product Management**

  - Product variants and attributes
  - Categories and subcategories
  - Barcode/SKU management
  - Stock level tracking

- **Inventory Control**

  - Stock-in management
  - Stock rejections/write-offs
  - Low stock alerts
  - Inventory valuation

- **Sales & Customers**

  - Sales order processing
  - Customer management
  - Sales reporting
  - Customer purchase history

- **Supplier Management**

  - Supplier directory
  - Purchase order tracking
  - Supplier performance metrics

- **Reporting**
  - Sales reports
  - Inventory valuation
  - Stock movement history
  - Daily transaction summaries

## 🚀 Tech Stack

- **Backend**: Go 1.18+
- **Router**: gorilla/mux
- **Database**: PostgreSQL (Supabase)
- **ORM**: pgx/v5
- **API**: RESTful JSON API

## 📁 Project Structure

```
inventory-go/
├── db/                # Database setup and migrations
├── documentations/    # System documentation
│   ├── API_DOCUMENTATION.md
│   ├── database_triggers.md
│   └── system_capabilities.md
├── handlers/          # HTTP request handlers
├── models/            # Data models and business logic
├── repositories/      # Database operations
├── routes/            # API route definitions
├── .env               # Environment variables (gitignored)
├── env.example        # Example environment variables
├── go.mod             # Go module definition
├── go.sum             # Dependency checksums
├── main.go            # Application entry point
└── README.md          # This file
```

## 🛠️ Setup & Installation

### Prerequisites

- Go 1.18 or newer
- Supabase account with a project
- Git (for version control)

### Database Setup

1. Create a new Supabase project at [Supabase](https://supabase.com/)
2. Run the SQL migration script from `db/schema.sql` in the Supabase SQL editor
3. Set up any required triggers and functions as described in `documentations/database_triggers.md`

### Environment Configuration

1. Copy the example environment file:

   ```bash
   cp env.example .env
   ```

2. Update `.env` with your Supabase credentials:

   ```env
   # Database Configuration
   DB_HOST=aws-0-us-east-2.pooler.supabase.com
   DB_PORT=6543
   DB_USER=postgres.YOUR_PROJECT_REF
   DB_PASSWORD=YOUR_PASSWORD
   DB_NAME=postgres

   # Server Configuration
   PORT=8080

   # Complete connection string (alternative to individual settings above)
   DATABASE_URL=postgresql://postgres.YOUR_PROJECT_REF:YOUR_PASSWORD@aws-0-us-east-2.pooler.supabase.com:6543/postgres
   ```

### Running the Application

1. Install dependencies:

   ```bash
   go mod download
   ```

2. Start the server:

   ```bash
   go run main.go
   ```

   The API will be available at `http://localhost:8080/api`

3. (Optional) Build for production:
   ```bash
   go build -o inventory-api
   ./inventory-api
   ```

## 📚 API Documentation

For complete API documentation including request/response examples, please see:

- [API Documentation](./documentations/API_DOCUMENTATION.md)
- [Database Triggers](./documentations/database_triggers.md)
- [System Capabilities](./documentations/system_capabilities.md)

## 🔍 API Overview

### Core Resources

- **Products**: `GET|POST /api/products`
- **Categories**: `GET|POST /api/categories`
- **Customers**: `GET|POST /api/customers`
- **Suppliers**: `GET|POST /api/suppliers`
- **Sales**: `GET|POST /api/sales`
- **Stock In**: `GET|POST /api/stockins`
- **Rejects**: `GET|POST /api/rejects`

### Example Request

```bash
# Create a new product
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Wireless Mouse",
    "sku": "WM-001",
    "description": "Ergonomic wireless mouse",
    "price": 29.99,
    "quantity": 100
  }'
```

## 📊 Features in Development

- [ ] User Authentication & Authorization
- [ ] Barcode/QR Code Support
- [ ] Multi-warehouse Support
- [ ] Advanced Reporting
- [ ] Mobile App

## 🤝 Contributing

Contributions are welcome! Please read our [contributing guidelines](CONTRIBUTING.md) to get started.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

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

### Suppliers

- `GET /api/suppliers` - Get all suppliers
- `GET /api/suppliers/{id}` - Get a supplier by ID
- `GET /api/suppliers/search` - Search suppliers
- `GET /api/suppliers/top` - Get top suppliers
- `POST /api/suppliers` - Create a new supplier
- `PUT /api/suppliers/{id}` - Update a supplier
- `DELETE /api/suppliers/{id}` - Delete a supplier

### Stock-In

- `GET /api/stockins` - Get all stock-ins
- `GET /api/stockins/{id}` - Get a stock-in by ID
- `GET /api/stockins/reference/{reference}` - Get a stock-in by reference number
- `GET /api/stockins/summary` - Get stock-in summary
- `GET /api/stockins/daily` - Get daily stock-in data
- `GET /api/suppliers/{id}/stockins` - Get stock-ins by supplier
- `POST /api/stockins` - Create a new stock-in
- `PUT /api/stockins/{id}` - Update a stock-in
- `DELETE /api/stockins/{id}` - Delete a stock-in
- `POST /api/stockins/{id}/items` - Add a stock-in item
- `PUT /api/stockins/{stockInId}/items/{itemId}` - Update a stock-in item
- `DELETE /api/stockins/{stockInId}/items/{itemId}` - Delete a stock-in item

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
curl -X GET http://localhost:8080/api/categories
