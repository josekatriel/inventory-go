# Inventory Management API Documentation

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
  "name": "Product Name",
  "sku": "PROD-001",
  "description": "Product description",
  "category_id": "uuid-here",
  "price": 99.99,
  "cost_price": 49.99,
  "quantity": 100,
  "reorder_level": 10,
  "attributes": {
    "color": "red",
    "size": "M"
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
  "image_url": "https://example.com/image.jpg"
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
  "address": "123 Main St",
  "city": "New York",
  "country": "USA",
  "postal_code": "10001"
}
```

### Sales

#### Create Sale
```
POST /sales
```

**Request Body:**
```json
{
  "customer_id": "uuid-here",
  "reference_no": "SALE-001",
  "status": "completed",
  "discount": 0,
  "tax": 10.5,
  "shipping": 5.0,
  "total": 115.5,
  "items": [
    {
      "product_id": "uuid-here",
      "quantity": 2,
      "unit_price": 50.0,
      "total": 100.0
    }
  ]
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
