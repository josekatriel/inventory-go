# Inventory Management API Documentation

## Table of Contents

1. [Authentication](#authentication)
2. [Products](#products)
3. [Categories](#categories)
4. [Customers](#customers)
5. [Sales](#sales)
6. [Suppliers](#suppliers)
7. [Export/Import](#exportimport)

## Base URL

All API endpoints are prefixed with `/api`

## Authentication

> Note: Authentication will be implemented in a future update.

## Response Format

All API responses follow a standard format:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {}
}
```

## Error Responses

### 400 Bad Request

```json
{
  "success": false,
  "message": "Validation error",
  "errors": ["field_name: error description"]
}
```

### 404 Not Found

```json
{
  "success": false,
  "message": "Resource not found"
}
```

### 500 Internal Server Error

```json
{
  "success": false,
  "message": "Internal server error"
}
```

## API Endpoints

### Products

#### Get All Products

```
GET /products
```

**Query Parameters:**

- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 10)
- `category` (optional): Filter by category ID
- `search` (optional): Search term for product name or SKU

#### Get Low Stock Products

```
GET /products/low-stock
```

Returns products where current stock is at or below the reorder level. Useful for identifying items that need to be restocked.

**Response:**

```json
{
  "data": [
    {
      "id": "uuid-here",
      "stock": 5,
      "reorder_level": 10,
      "basic": {
        "name": "Product Name",
        "sku": "SKU001"
      },
      "price": {
        "price": 29.99
      }
    }
  ],
  "total": 1,
  "message": "Products below their reorder threshold. These items need to be restocked."
}
```

#### Get Product by ID

```
GET /products/{id}
```

#### Create Product

```
POST /products
```

**Request Body:**

```json
{
  "basic": {
    "name": "Product Name",
    "sku": "PROD-001",
    "description": "Product description"
  },
  "price": {
    "price": 99.99,
    "currency": "IDR"
  },
  "reorder_level": 5,
  "category_id": "uuid-here"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": {
    "id": "generated-uuid",
    "basic": {
      "name": "Product Name",
      "sku": "PROD-001",
      "description": "Product description"
    },
    "price": {
      "price": 99.99,
      "currency": "IDR",
      "last_update_unix": 1621500000
    },
    "stock": 0,
    "reorder_level": 5,
    "child_category_id": "uuid-here",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Categories

#### Get All Categories

```
GET /categories
```

#### Get Category by ID or Slug

```
GET /categories/{idOrSlug}
```

#### Create Category

```
POST /categories
```

**Request Body:**

```json
{
  "name": "Electronics",
  "description": "Electronic devices",
  "parent_id": "uuid-or-null",
  "sort_order": 1,
  "status": 1,
  "image_url": "https://example.com/image.jpg"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Category created successfully",
  "data": {
    "id": "generated-uuid",
    "name": "Electronics",
    "description": "Electronic devices",
    "slug": "electronics",
    "status": 1,
    "sort_order": 1,
    "parent_id": "uuid-or-null",
    "image_url": "https://example.com/image.jpg",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Customers

#### Get All Customers

```
GET /customers
```

#### Create Customer

```
POST /customers
```

**Request Body:**

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "address": "123 Main St, City, Country",
  "notes": "VIP customer"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Customer created successfully",
  "data": {
    "id": "generated-uuid",
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+1234567890",
    "address": "123 Main St, City, Country",
    "total_orders": 0,
    "total_spent": 0,
    "notes": "VIP customer",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Sales

#### Get All Sales

```
GET /sales
```

#### Get Sale by ID

```
GET /sales/{id}
```

#### Create Sale

```
POST /sales
```

**Request Body:**

```json
{
  "status": "draft",
  "sale_date": "2023-01-01T10:00:00Z",
  "note": "Customer order",
  "customer_id": "customer-uuid",
  "platform": "Website",
  "items": [
    {
      "product_id": "product-uuid",
      "product_name": "Product Name",
      "quantity": 2,
      "unit_price": 49.99,
      "tax": 5.0,
      "discount": 0,
      "subtotal": 104.98
    }
  ],
  "payments": [
    {
      "amount": 104.98,
      "payment_method": "credit_card",
      "reference": "TRANS123456",
      "note": "Full payment",
      "payment_date": "2023-01-01T10:05:00Z"
    }
  ]
}
```

**Response:**

```json
{
  "success": true,
  "message": "Sale created successfully",
  "data": {
    "id": "generated-uuid",
    "reference_no": "SALE-20230101-abc123",
    "status": "draft",
    "sale_date": "2023-01-01T10:00:00Z",
    "note": "Customer order",
    "total": 104.98,
    "paid": 104.98,
    "balance": 0,
    "customer_id": "customer-uuid",
    "customer": {
      "id": "customer-uuid",
      "name": "John Doe"
    },
    "platform": "Website",
    "items": [
      {
        "id": "item-uuid",
        "sale_id": "generated-uuid",
        "product_id": "product-uuid",
        "product_name": "Product Name",
        "quantity": 2,
        "unit_price": 49.99,
        "tax": 5.0,
        "discount": 0,
        "subtotal": 104.98,
        "created_at": "2023-01-01T10:00:00Z",
        "updated_at": "2023-01-01T10:00:00Z"
      }
    ],
    "payments": [
      {
        "id": "payment-uuid",
        "sale_id": "generated-uuid",
        "amount": 104.98,
        "payment_method": "credit_card",
        "reference": "TRANS123456",
        "note": "Full payment",
        "payment_date": "2023-01-01T10:05:00Z",
        "created_at": "2023-01-01T10:00:00Z",
        "updated_at": "2023-01-01T10:00:00Z"
      }
    ],
    "created_at": "2023-01-01T10:00:00Z",
    "updated_at": "2023-01-01T10:00:00Z"
  }
}
```

#### Update Sale Status

```
PATCH /sales/{id}/status
```

**Request Body:**

```json
{
  "status": "completed"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Sale status updated successfully",
  "data": {
    "id": "sale-uuid",
    "status": "completed"
  }
}
```

#### Add Payment to Sale

```
POST /sales/{id}/payments
```

**Request Body:**

```json
{
  "amount": 50.0,
  "payment_method": "cash",
  "reference": "CASH12345",
  "note": "Partial payment",
  "payment_date": "2023-01-01T12:00:00Z"
}
```

### Suppliers

#### Get All Suppliers

```
GET /suppliers
```

#### Get Supplier by ID

```
GET /suppliers/{id}
```

#### Create Supplier

```
POST /suppliers
```

**Request Body:**

```json
{
  "name": "ABC Supplies Inc.",
  "email": "contact@abcsupplies.com",
  "phone": "+1234567890",
  "address": "456 Business Ave, Industry Zone",
  "contact_person": "Jane Smith",
  "notes": "Preferred supplier for electronics"
}
```

**Response:**

```json
{
  "success": true,
  "message": "Supplier created successfully",
  "data": {
    "id": "generated-uuid",
    "name": "ABC Supplies Inc.",
    "email": "contact@abcsupplies.com",
    "phone": "+1234567890",
    "address": "456 Business Ave, Industry Zone",
    "contact_person": "Jane Smith",
    "total_purchases": 0,
    "total_spent": 0,
    "notes": "Preferred supplier for electronics",
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Stock In

#### Create Stock In

```
POST /stockins
```

**Request Body:**

```json
{
  "supplier_id": "uuid-here",
  "reference_no": "STOCKIN-001",
  "status": "received",
  "total": 500.0,
  "items": [
    {
      "product_id": "uuid-here",
      "quantity": 10,
      "unit_cost": 50.0,
      "total": 500.0
    }
  ]
}
```

### Rejects (Stock Decrease)

#### Create Reject

```
POST /rejects
```

**Request Body:**

```json
{
  "reference_no": "REJ-001",
  "status": "completed",
  "reason": "Damaged during handling",
  "total": 100.0,
  "items": [
    {
      "product_id": "uuid-here",
      "quantity": 2,
      "unit_cost": 50.0,
      "total": 100.0,
      "reason": "Damaged packaging"
    }
  ]
}
```

## Rate Limiting

> Note: Rate limiting will be implemented in a future update.

## Versioning

API versioning will be implemented in a future update.

## Export/Import

### Export Products

```
GET /export/products
```

Exports product data in CSV or Excel format.

**Query Parameters**:

- `format` (optional): Export format, either `csv` (default) or `excel`

**Response**:
Returns a downloadable file in the requested format containing product data.

### Export Categories

```
GET /export/categories
```

Exports category data in CSV or Excel format.

**Query Parameters**:

- `format` (optional): Export format, either `csv` (default) or `excel`

**Response**:
Returns a downloadable file in the requested format containing category data.

### Export Customers

```
GET /export/customers
```

Exports customer data in CSV or Excel format.

**Query Parameters**:

- `format` (optional): Export format, either `csv` (default) or `excel`

**Response**:
Returns a downloadable file in the requested format containing customer data.

### Export Suppliers

```
GET /export/suppliers
```

Exports supplier data in CSV or Excel format.

**Query Parameters**:

- `format` (optional): Export format, either `csv` (default) or `excel`

**Response**:
Returns a downloadable file in the requested format containing supplier data.

### Import Products

```
POST /import/products
```

Imports products from a CSV or Excel file.

**Request**:
Multipart form data with a file field named `file` containing the CSV or Excel file.

**Expected File Format**:
The file should contain the following columns:

- ID (optional, if provided will update existing products)
- Name (required)
- Description
- SKU
- Stock
- Reorder Level
- Price
- Currency

**Response**:

```json
{
  "success": true,
  "total": 10,
  "imported": 8,
  "errors": [
    "Row 3: Product with ID abc123 not found",
    "Row 5: Failed to save product: invalid price format"
  ]
}
```

### Import Categories

```
POST /import/categories
```

Imports categories from a CSV or Excel file.

**Request**:
Multipart form data with a file field named `file` containing the CSV or Excel file.

**Expected File Format**:
The file should contain the following columns:

- ID (optional, if provided will update existing categories)
- Name (required)
- Description
- Parent ID
- Slug (optional, will be auto-generated if not provided)
- Status
- Sort Order
- Image URL

**Response**:

```json
{
  "success": true,
  "total": 5,
  "imported": 5,
  "errors": []
}
```
