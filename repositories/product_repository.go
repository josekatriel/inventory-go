package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"inventory-go/models"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// Define Attribute type for product attributes
type Attribute struct {
	Key   string
	Value string
}

type ProductRepository interface {
	// Basic CRUD operations
	GetAll() ([]*models.Product, error)
	GetByID(id string) (*models.Product, error)
	Create(product *models.Product) error
	Update(product *models.Product) error
	Delete(id string) error

	// Variant operations
	GetVariants(parentID string) ([]*models.Product, error)
	AddVariant(parentID string, variant *models.Product) error

	// Image operations
	AddImage(productID string, image *models.Images) error
	RemoveImage(productID, imageID string) error
	SetPrimaryImage(productID, imageID string) error

	// Stock operations
	UpdateStock(id string, quantity int) error
	GetStockHistory(id string) ([]models.InventoryActivity, error)

	// Attribute operations
	AddAttribute(productID string, attribute Attribute) error

	// Pagination, search and filter
	List(offset, limit int, status string) ([]*models.Product, int64, error)
	Search(query string, categoryID *string) ([]*models.Product, error)
	GetBySKU(sku string) (*models.Product, error)
}

type ProductRepositoryImpl struct {
	*BaseRepository
}

func NewProductRepository(db *pgx.Conn) ProductRepository {
	return &ProductRepositoryImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// GetAll implements ProductRepository.
func (r *ProductRepositoryImpl) GetAll() ([]*models.Product, error) {
	query := `
		SELECT 
			p.id, p.parent_id, p.stock, p.child_category_id, 
			p.created_at, p.updated_at, p.deleted_at,
			p.basic->>'name' as name,
			p.basic->>'description' as description,
			(p.basic->>'status')::int as status,
			(p.basic->>'condition')::int as condition,
			p.basic->>'sku' as sku,
			(p.basic->>'is_variant')::boolean as is_variant,
			p.price->>'price' as price,
			p.price->>'currency' as currency
		FROM products p 
		WHERE p.deleted_at IS NULL`
	
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var priceStr string
		var deletedAt pgtype.Timestamp
		
		err := rows.Scan(
			&product.ID, &product.ParentID, &product.Stock, &product.ChildCategoryID,
			&product.CreatedAt, &product.UpdatedAt, &deletedAt,
			&product.Basic.Name, &product.Basic.Description, &product.Basic.Status,
			&product.Basic.Condition, &product.Basic.SKU, &product.Basic.IsVariant,
			&priceStr, &product.Price.Currency,
		)
		
		if err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}
		
		// Convert string values to proper types
		if priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err == nil {
				product.Price.Price = price
			}
		}
		
		if deletedAt.Valid {
			product.DeletedAt = &deletedAt.Time
		}
		
		products = append(products, product)
	}
	
	return products, nil
}

// Create implements ProductRepository.
func (r *ProductRepositoryImpl) Create(product *models.Product) error {
	if product.ID == "" {
		product.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Marshal structs to JSONB
	basicJSON, err := json.Marshal(product.Basic)
	if err != nil {
		return fmt.Errorf("failed to marshal basic info: %w", err)
	}
	
	priceJSON, err := json.Marshal(product.Price)
	if err != nil {
		return fmt.Errorf("failed to marshal price info: %w", err)
	}
	
	weightJSON, err := json.Marshal(product.Weight)
	if err != nil {
		return fmt.Errorf("failed to marshal weight info: %w", err)
	}
	
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}
	
	inventoryJSON, err := json.Marshal(product.InventoryActivity)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory activity: %w", err)
	}
	
	query := `
		INSERT INTO products (
			id, parent_id, stock, child_category_id, created_at, updated_at,
			basic, price, weight, images, inventory_activity
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = r.db.Exec(context.Background(), query,
		product.ID, product.ParentID, product.Stock, product.ChildCategoryID,
		product.CreatedAt, product.UpdatedAt,
		basicJSON, priceJSON, weightJSON, imagesJSON, inventoryJSON)
	
	return err
}

// Update implements ProductRepository.
func (r *ProductRepositoryImpl) Update(product *models.Product) error {
	product.UpdatedAt = time.Now()

	// Marshal structs to JSONB
	basicJSON, err := json.Marshal(product.Basic)
	if err != nil {
		return fmt.Errorf("failed to marshal basic info: %w", err)
	}
	
	priceJSON, err := json.Marshal(product.Price)
	if err != nil {
		return fmt.Errorf("failed to marshal price info: %w", err)
	}
	
	weightJSON, err := json.Marshal(product.Weight)
	if err != nil {
		return fmt.Errorf("failed to marshal weight info: %w", err)
	}
	
	imagesJSON, err := json.Marshal(product.Images)
	if err != nil {
		return fmt.Errorf("failed to marshal images: %w", err)
	}
	
	inventoryJSON, err := json.Marshal(product.InventoryActivity)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory activity: %w", err)
	}

	query := `
		UPDATE products SET 
			updated_at = $1, 
			parent_id = $2, 
			stock = $3, 
			child_category_id = $4, 
			basic = $5, 
			price = $6, 
			weight = $7, 
			images = $8, 
			inventory_activity = $9
		WHERE id = $10`

	result, err := r.db.Exec(context.Background(), query,
		product.UpdatedAt, 
		product.ParentID, 
		product.Stock, 
		product.ChildCategoryID,
		basicJSON, 
		priceJSON, 
		weightJSON, 
		imagesJSON, 
		inventoryJSON,
		product.ID)

	if err != nil {
		return fmt.Errorf("error updating product: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("product not found with id: %s", product.ID)
	}

	return nil
}

// Helper function to check if product exists
func (r *ProductRepositoryImpl) productExists(id string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM products WHERE id = $1 AND deleted_at IS NULL)"
	err := r.db.QueryRow(context.Background(), query, id).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking if product exists: %w", err)
	}

	return exists, nil
}

// Delete implements ProductRepository.
func (r *ProductRepositoryImpl) Delete(id string) error {
	// Check if product exists
	exists, err := r.productExists(id)
	if err != nil {
		return err
	}
	if !exists {
		return nil // Already deleted or doesn't exist
	}

	// Perform soft delete
	query := `UPDATE products SET deleted_at = $1 WHERE id = $2`
	_, err = r.db.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("error deleting product: %w", err)
	}

	return nil
}

// GetByID implements ProductRepository.
func (r *ProductRepositoryImpl) GetByID(id string) (*models.Product, error) {
	var product models.Product
	var parentID *string
	var attributesJSON []byte

	query := `
		SELECT 
		    p.id, p.parent_id, p.stock, p.child_category_id, 
		    p.created_at, p.updated_at, p.deleted_at,
		    p.basic->>'name', p.basic->>'description', 
		    (p.basic->>'status')::int, (p.basic->>'condition')::int,
		    p.basic->>'sku', (p.basic->>'is_variant')::boolean,
		    p.price->>'price', p.price->>'currency',
		    (p.price->>'last_update_unix')::bigint as last_update_unix,
		    p.weight->>'weight' as weight,
		    (p.weight->>'unit')::int as unit,
		    p.inventory_activity->>'sales_count' as sales_count,
		    p.inventory_activity->>'stock_in' as stock_in,
		    p.inventory_activity->>'reject' as reject,
		    p.attributes
		FROM products p 
		WHERE p.id = $1 AND p.deleted_at IS NULL`

	err := r.QueryRow(context.Background(), query, id).Scan(
		&product.ID, &parentID, &product.Stock, &product.ChildCategoryID,
		&product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		&product.Basic.Name, &product.Basic.Description, &product.Basic.Status,
		&product.Basic.Condition, &product.Basic.SKU, &product.Basic.IsVariant,
		&product.Price.Price, &product.Price.Currency, &product.Price.LastUpdateUnix,
		&product.Weight.Weight, &product.Weight.Unit,
		&product.InventoryActivity.SalesCount, &product.InventoryActivity.StockIn,
		&product.InventoryActivity.Reject, &attributesJSON,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("error getting product: %w", err)
	}

	// Skip attributes unmarshaling as it's not part of the Product struct
	// We would handle them separately if needed through the AddAttribute method
	_ = attributesJSON

	// Load variants if this is a parent product
	if !product.Basic.IsVariant {
		tx, err := r.db.Begin(context.Background())
		if err != nil {
			return nil, fmt.Errorf("error beginning transaction: %w", err)
		}
		defer tx.Rollback(context.Background())

		if err := product.LoadVariants(tx); err != nil {
			return nil, fmt.Errorf("error loading variants: %w", err)
		}

		if err := tx.Commit(context.Background()); err != nil {
			return nil, fmt.Errorf("error committing transaction: %w", err)
		}
	}

	return &product, nil
}

// List implements ProductRepository.
func (r *ProductRepositoryImpl) List(offset int, limit int, status string) ([]*models.Product, int64, error) {
	var products []*models.Product
	var total int64

	// Build query
	query := `
		SELECT 
			p.id, p.parent_id, p.stock, p.child_category_id,
			p.created_at, p.updated_at, p.deleted_at, 
			p.basic->>'name', p.basic->>'description', 
			(p.basic->>'status')::int, (p.basic->>'condition')::int,
			p.basic->>'sku', (p.basic->>'is_variant')::boolean,
			p.price->>'price', p.price->>'currency',
			(SELECT COUNT(*) FROM products WHERE deleted_at IS NULL) as total
		FROM products p 
		WHERE p.deleted_at IS NULL`
	
	// Add status filter if provided
	args := []interface{}{}
	if status != "" {
		query += " AND p.basic->>'status' = $1"
		args = append(args, status)
	}
	
	// Add pagination
	query += " ORDER BY p.created_at DESC LIMIT $" + strconv.Itoa(len(args)+1) + 
		" OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	for rows.Next() {
		product := &models.Product{}
		var priceStr string
		var deletedAt pgtype.Timestamp
		
		err := rows.Scan(
			&product.ID, &product.ParentID, &product.Stock, &product.ChildCategoryID,
			&product.CreatedAt, &product.UpdatedAt, &deletedAt,
			&product.Basic.Name, &product.Basic.Description, &product.Basic.Status,
			&product.Basic.Condition, &product.Basic.SKU, &product.Basic.IsVariant,
			&priceStr, &product.Price.Currency,
		)
		if err != nil {
			return nil, 0, err
		}
		if deletedAt.Valid {
			product.DeletedAt = &deletedAt.Time
		}
		
		// Convert string values to proper types
		if priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err == nil {
				product.Price.Price = price
			}
		}
		
		products = append(products, product)
	}
	
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	err = r.db.QueryRow(context.Background(), "SELECT COUNT(*) FROM products WHERE deleted_at IS NULL").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// Search implements ProductRepository.
func (r *ProductRepositoryImpl) Search(searchQuery string, categoryID *string) ([]*models.Product, error) {
	query := `
		SELECT 
			p.id, p.parent_id, p.stock, p.child_category_id,
			p.created_at, p.updated_at, p.deleted_at, 
			p.basic->>'name', p.basic->>'description', 
			(p.basic->>'status')::int, (p.basic->>'condition')::int,
			p.basic->>'sku', (p.basic->>'is_variant')::boolean,
			p.price->>'price', p.price->>'currency'
		FROM products p 
		WHERE p.deleted_at IS NULL`

	// Add search conditions
	args := []interface{}{}
	if searchQuery != "" {
		query += " AND (p.basic->>'name' ILIKE $1 OR p.basic->>'description' ILIKE $1 OR p.basic->>'sku' ILIKE $1)"
		args = append(args, "%"+searchQuery+"%")
	}

	if categoryID != nil && *categoryID != "" {
		query += fmt.Sprintf(" AND p.child_category_id = $%d", len(args)+1)
		args = append(args, *categoryID)
	}

	query += " ORDER BY p.created_at DESC LIMIT 100"

	rows, err := r.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing search query: %w", err)
	}
	defer rows.Close()
	
	var products []*models.Product
	for rows.Next() {
		product := &models.Product{}
		var priceStr string
		var deletedAt pgtype.Timestamp
		
		err := rows.Scan(
			&product.ID, &product.ParentID, &product.Stock, &product.ChildCategoryID,
			&product.CreatedAt, &product.UpdatedAt, &deletedAt,
			&product.Basic.Name, &product.Basic.Description, &product.Basic.Status,
			&product.Basic.Condition, &product.Basic.SKU, &product.Basic.IsVariant,
			&priceStr, &product.Price.Currency,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning product: %w", err)
		}
		if deletedAt.Valid {
			product.DeletedAt = &deletedAt.Time
		}
		
		// Convert string values to proper types
		if priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err == nil {
				product.Price.Price = price
			}
		}
		
		products = append(products, product)
	}
	
	return products, nil
}

// RemoveImage implements ProductRepository interface
func (r *ProductRepositoryImpl) RemoveImage(productID string, imageID string) error {
	// First, check if this is the primary image
	var image models.Images
	query := `
		SELECT id, product_id, is_primary
		FROM images 
		WHERE id = $1`

	err := r.db.QueryRow(context.Background(), query, imageID).Scan(&image.ID, &image.ProductID, &image.IsPrimary)
	if err != nil {
		return err
	}

	// If it's the primary image, we need to set another image as primary
	if image.IsPrimary {
		// Find another image for the same product to set as primary
		var otherImage models.Images
		query = `
			SELECT id, product_id, is_primary
			FROM images 
			WHERE product_id = $1 AND id != $2
			ORDER BY sort_order ASC
			LIMIT 1`

		err := r.db.QueryRow(context.Background(), query, image.ProductID, imageID).Scan(&otherImage.ID, &otherImage.ProductID, &otherImage.IsPrimary)
		if err == nil {
			// Found another image, set it as primary
			query = `
				UPDATE images 
				SET is_primary = TRUE 
				WHERE id = $1`

			_, err = r.db.Exec(context.Background(), query, otherImage.ID)
			if err != nil {
				return err
			}
		}
	}

	// Delete the image
	query = `
		DELETE FROM images 
		WHERE id = $1`

	_, err = r.db.Exec(context.Background(), query, imageID)
	if err != nil {
		return err
	}

	return nil
}

// SetPrimaryImage implements ProductRepository.
func (r *ProductRepositoryImpl) SetPrimaryImage(productID string, imageID string) error {
	// First, unset any existing primary image for this product
	query := `UPDATE images SET is_primary = false WHERE product_id = $1`
	_, err := r.db.Exec(context.Background(), query, productID)
	if err != nil {
		return err
	}

	// Then set this image as primary
	query = `UPDATE images SET is_primary = true WHERE id = $1 AND product_id = $2`
	_, err = r.db.Exec(context.Background(), query, imageID, productID)
	if err != nil {
		return fmt.Errorf("failed to update image: %w", err)
	}

	return nil
}



// AddAttribute implements ProductRepository.
func (r *ProductRepositoryImpl) AddAttribute(productID string, attribute Attribute) error {
	// Check if product exists
	exists, err := r.productExists(productID)
	if err != nil {
		return fmt.Errorf("failed to check product: %w", err)
	}
	if !exists {
		return fmt.Errorf("product not found: %s", productID)
	}
	
	// Create new attribute object
	attributeMap := map[string]string{
		attribute.Key: attribute.Value,
	}

	attrsJSON, err := json.Marshal(attributeMap)
	if err != nil {
		return fmt.Errorf("failed to marshal attribute: %w", err)
	}

	// Update product with new attribute
	query := `UPDATE products SET attributes = attributes || $1 WHERE id = $2`
	_, err = r.db.Exec(context.Background(), query, attrsJSON, productID)
	if err != nil {
		return fmt.Errorf("failed to update attributes: %w", err)
	}

	return nil
}

// UpdateStock implements ProductRepository.
func (r *ProductRepositoryImpl) UpdateStock(id string, quantity int) error {
	// Get the existing product
	product, err := r.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get product: %w", err)
	}

	// Calculate new stock
	newStock := product.Stock + quantity
	
	// Ensure stock doesn't go below 0
	if newStock < 0 {
		return fmt.Errorf("insufficient stock - current: %d, requested: %d", product.Stock, -quantity)
	}

	// Update the stock in the database
	query := "UPDATE products SET stock = $1 WHERE id = $2"
	_, err = r.db.Exec(context.Background(), query, newStock, id)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	return nil
}

// GetVariants implements ProductRepository.
func (r *ProductRepositoryImpl) GetVariants(parentID string) ([]*models.Product, error) {
	var variants []*models.Product

	query := `SELECT 
		id, parent_id, stock, child_category_id,
		basic->>'name' as name, 
		basic->>'description' as description, 
		(basic->>'status')::int as status, 
		(basic->>'condition')::int as condition,
		basic->>'sku' as sku,
		(basic->>'is_variant')::boolean as is_variant,
		price->>'price' as price
	  FROM products 
	  WHERE parent_id = $1 AND deleted_at IS NULL`
	
	rows, err := r.db.Query(context.Background(), query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query variants: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		product := &models.Product{}
		var priceStr string
		
		err := rows.Scan(
			&product.ID, &product.ParentID, &product.Stock, &product.ChildCategoryID,
			&product.Basic.Name, &product.Basic.Description, &product.Basic.Status,
			&product.Basic.Condition, &product.Basic.SKU, &product.Basic.IsVariant,
			&priceStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan variant: %w", err)
		}
		
		if priceStr != "" {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err == nil {
				product.Price.Price = price
			}
		}
		
		variants = append(variants, product)
	}

	return variants, nil
}

func (r *ProductRepositoryImpl) GetStockHistory(id string) ([]models.InventoryActivity, error) {
	// Query to get the inventory activity for a product
	query := `
		SELECT 
			p.inventory_activity->>'sales_count' as sales_count,
			p.inventory_activity->>'stock_in' as stock_in,
			p.inventory_activity->>'reject' as reject
		FROM products p 
		WHERE p.id = $1 AND p.deleted_at IS NULL`

	var activityStr struct {
		salesCount string
		stockIn    string
		reject     string
	}

	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&activityStr.salesCount, &activityStr.stockIn, &activityStr.reject,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("error getting stock history: %w", err)
	}
	
	// Convert string values to integer
	salesCount, _ := strconv.Atoi(activityStr.salesCount)
	stockIn, _ := strconv.Atoi(activityStr.stockIn)
	reject, _ := strconv.Atoi(activityStr.reject)
	
	// Create an inventory activity entry
	activity := models.InventoryActivity{
		SalesCount: salesCount,
		StockIn:    stockIn,
		Reject:     reject,
	}

	// Return as a slice since the interface expects a slice
	return []models.InventoryActivity{activity}, nil
}

func (r *ProductRepositoryImpl) GetBySKU(sku string) (*models.Product, error) {
	query := `
		SELECT 
			p.id, p.parent_id, p.stock, p.child_category_id, 
			p.created_at, p.updated_at, p.deleted_at,
			p.basic->>'name' as name,
			p.basic->>'description' as description,
			(p.basic->>'status')::int as status,
			(p.basic->>'condition')::int as condition,
			p.basic->>'sku' as sku,
			(p.basic->>'is_variant')::boolean as is_variant,
			p.price->>'price' as price,
			p.price->>'currency' as currency,
			(p.price->>'last_update_unix')::bigint as last_update_unix,
			p.weight->>'weight' as weight,
			(p.weight->>'unit')::int as unit
		FROM products p 
		WHERE p.basic->>'sku' = $1 AND p.deleted_at IS NULL`

	var product models.Product
	var priceStr, weightStr string
	var lastUpdateUnix int64
	
	err := r.db.QueryRow(context.Background(), query, sku).Scan(
		&product.ID, &product.ParentID, &product.Stock, &product.ChildCategoryID,
		&product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		&product.Basic.Name, &product.Basic.Description, &product.Basic.Status,
		&product.Basic.Condition, &product.Basic.SKU, &product.Basic.IsVariant,
		&priceStr, &product.Price.Currency, &lastUpdateUnix,
		&weightStr, &product.Weight.Unit,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("error getting product by SKU: %w", err)
	}

	// Convert string values to their proper types
	if priceStr != "" {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err == nil {
			product.Price.Price = price
		}
	}

	if weightStr != "" {
		weight, err := strconv.ParseFloat(weightStr, 64)
		if err == nil {
			product.Weight.Weight = weight
		}
	}

	product.Price.LastUpdateUnix = lastUpdateUnix

	return &product, nil
}

// AddVariant adds a product variant to a parent product
func (r *ProductRepositoryImpl) AddVariant(parentID string, variant *models.Product) error {
	// Generate ID if not set
	if variant.ID == "" {
		variant.ID = uuid.NewString()
	}

	// Set variant fields
	variant.ParentID = &parentID
	variant.Basic.IsVariant = true
	
	// Create variant in database
	now := time.Now()
	variant.CreatedAt = now
	variant.UpdatedAt = now
	
	query := `INSERT INTO products (
		id, parent_id, stock, created_at, updated_at,
		basic, price, weight
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	basicJSON, err := json.Marshal(variant.Basic)
	if err != nil {
		return fmt.Errorf("failed to marshal basic info: %w", err)
	}
	
	priceJSON, err := json.Marshal(variant.Price)
	if err != nil {
		return fmt.Errorf("failed to marshal price: %w", err)
	}
	
	weightJSON, err := json.Marshal(variant.Weight)
	if err != nil {
		return fmt.Errorf("failed to marshal weight: %w", err)
	}
	
	_, err = r.db.Exec(context.Background(), query,
		variant.ID, parentID, variant.Stock, variant.CreatedAt, variant.UpdatedAt,
		basicJSON, priceJSON, weightJSON,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create product variant: %w", err)
	}
	
	return nil
}

func (r *ProductRepositoryImpl) AddImage(productID string, image *models.Images) error {
	image.ProductID = productID
	if image.ID == "" {
		image.ID = uuid.NewString()
	}
	
	query := `INSERT INTO images (
		id, product_id, url, is_primary, sort_order, created_at
	) VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := r.db.Exec(context.Background(), query,
		image.ID, image.ProductID, image.URL, image.IsPrimary, image.SortOrder, time.Now(),
	)
	
	if err != nil {
		return fmt.Errorf("failed to add image: %w", err)
	}
	
	return nil
}

// ... implement other methods following the same pattern
