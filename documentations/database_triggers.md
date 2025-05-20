# Database Triggers and Functions

This document provides an overview of all triggers and functions in the inventory management system database.

## Timestamp Management

### Function: `update_timestamp()`
Updates the `updated_at` field to the current timestamp.

### Triggers:
- `trigger_update_categories_timestamp` on `categories`
- `trigger_update_products_timestamp` on `products`
- `trigger_update_customers_timestamp` on `customers`
- `trigger_update_suppliers_timestamp` on `suppliers`
- `trigger_update_stock_ins_timestamp` on `stock_ins`
- `trigger_update_stock_in_items_timestamp` on `stock_in_items`
- `trigger_update_sales_timestamp` on `sales`
- `trigger_update_sale_items_timestamp` on `sale_items`

## UUID Generation

### Function: `generate_uuid_for_id()`
Automatically generates a UUID for VARCHAR(36) ID columns if not provided.

### Triggers:
- `trigger_categories_generate_uuid` on `categories`
- `trigger_products_generate_uuid` on `products`
- `trigger_images_generate_uuid` on `images`
- `trigger_customers_generate_uuid` on `customers`
- `trigger_suppliers_generate_uuid` on `suppliers`
- `trigger_stock_ins_generate_uuid` on `stock_ins`
- `trigger_stock_in_items_generate_uuid` on `stock_in_items`
- `trigger_sales_generate_uuid` on `sales`
- `trigger_sale_items_generate_uuid` on `sale_items`

## Inventory Management

### Function: `update_product_quantity_after_stock_in()`
Increases product quantity when a stock-in is marked as completed.

Behavior:
- Only runs when stock-in status changes to 'completed'
- Increases the `stock` field in the products table
- Aggregates quantities from all items in the stock-in

### Trigger:
- `trigger_update_product_quantity_after_stock_in` AFTER UPDATE on `stock_ins`

### Function: `update_product_quantity_after_sale()`
Decreases product quantity when a sale is marked as completed.

Behavior:
- Only runs when sale status changes to 'completed'
- Decreases the `stock` field in the products table
- Aggregates quantities from all items in the sale

### Trigger:
- `trigger_update_product_quantity_after_sale` AFTER UPDATE on `sales`

## Potential Additions

The following are potential triggers/functions that could be added:

1. **Prevent Negative Stock**
   - Prevent selling more items than available in stock

2. **Stock Reservation**
   - Temporarily reserve stock during checkout process

3. **Stock-in Item Update Trigger**
   - Update product stock when a stock-in item is updated after completion

4. **Stock History Tracking**
   - Record all stock changes in a separate history table

5. **Low Stock Alerts**
   - Check and flag products that reach low stock levels

6. **Automatic Pricing Updates**
   - Update product prices based on latest stock-in costs

7. **Total Recalculation**
   - Automatically recalculate totals when items are added/updated/removed
