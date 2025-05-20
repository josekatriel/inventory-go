package models

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Product struct {
	ID           string  `json:"id" db:"id"`
	ParentID     *string `json:"parent_id,omitempty" db:"parent_id"`
	Stock        int     `json:"stock" db:"stock"`
	ReorderLevel int     `json:"reorder_level" db:"reorder_level"`

	// Embedded fields
	Basic             BasicInfo         `json:"basic"`
	Price             Price             `json:"price"`
	Weight            Weight            `json:"weight"`
	InventoryActivity InventoryActivity `json:"inventory_activity,omitempty"`

	// Relations
	Images     []*Images  `json:"pictures,omitempty" db:"-"`
	Parent     *Product   `json:"-" db:"-"`
	Variants   []*Product `json:"variants,omitempty" db:"-"`
	CategoryID *string    `json:"category_id,omitempty" db:"category_id"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

/*
Product Status Codes Reference:
-2  = Banned (Product is prohibited/banned from sale)
-1  = Pending (Product awaiting approval)
0   = Deleted (Product has been removed/soft-deleted)
1   = Active (Product is available for sale)
2   = Featured (Product is marked as "Best"/featured item)
3   = WarehouseInactive (Product exists but not currently sellable)
*/

type BasicInfo struct {
	Name        string `json:"name" db:"name"`
	Description string `json:"description" db:"description"`
	Status      int    `json:"status" db:"status"`
	Condition   int    `json:"condition" db:"condition"` // 1: New, 2: Used
	SKU         string `json:"sku" db:"sku"`
	IsVariant   bool   `json:"is_variant" db:"is_variant"`
}

type Price struct {
	Price          float64 `json:"price" db:"price"`
	Currency       string  `json:"currency" db:"currency"`
	LastUpdateUnix int64   `json:"last_update_unix" db:"last_update_unix"`
}

type Weight struct {
	Weight float64 `json:"weight" db:"weight"`
	Unit   int     `json:"unit" db:"unit"` // 1 = gram, 2 = kilogram
}

type InventoryActivity struct {
	SalesCount int `json:"sales_count" db:"sales_count"`
	StockIn    int `json:"stock_in" db:"stock_in"`
	Reject     int `json:"reject" db:"reject"`
}

type Images struct {
	ID        string    `json:"id" db:"id"`
	ProductID string    `json:"product_id" db:"product_id"`
	URL       string    `json:"url" db:"url"`
	IsPrimary bool      `json:"is_primary" db:"is_primary"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// NewProduct creates a new product with a generated UUID if not provided
func NewProduct() *Product {
	now := time.Now()
	return &Product{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
		Basic: BasicInfo{
			Status: 1, // Default status: Active
		},
		Price: Price{
			Currency: "IDR",
		},
		Weight: Weight{
			Unit: 1, // Default to grams
		},
	}
}

// NewImage creates a new image with a generated UUID if not provided
func NewImage() *Images {
	return &Images{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
	}
}

// Scan implements the sql.Scanner interface for Product
func (p *Product) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return nil
	}
}

// Value implements the driver.Valuer interface for Product
func (p Product) Value() (driver.Value, error) {
	return json.Marshal(p)
}

// Scan implements the sql.Scanner interface for Images
func (i *Images) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, i)
	case string:
		return json.Unmarshal([]byte(v), i)
	default:
		return nil
	}
}

// Value implements the driver.Valuer interface for Images
func (i Images) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// LoadVariants loads variants for the product
func (p *Product) LoadVariants(db pgx.Tx) error {
	if p.ID == "" {
		return nil
	}

	// Explicitly select only the columns that exist in the database
	query := `
		SELECT 
			id, parent_id, stock, reorder_level, category_id, 
			created_at, updated_at, deleted_at,
			basic->>'name', basic->>'description', 
			(basic->>'status')::int, (basic->>'condition')::int,
			basic->>'sku', (basic->>'is_variant')::boolean,
			price->>'price', price->>'currency'
		FROM products 
		WHERE parent_id = $1 AND deleted_at IS NULL`

	rows, err := db.Query(context.Background(), query, p.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var variants []*Product
	for rows.Next() {
		var variant Product
		// Scan only into the fields that match the columns we selected
		err := rows.Scan(
			&variant.ID, &variant.ParentID, &variant.Stock, &variant.ReorderLevel, &variant.CategoryID,
			&variant.CreatedAt, &variant.UpdatedAt, &variant.DeletedAt,
			&variant.Basic.Name, &variant.Basic.Description, &variant.Basic.Status,
			&variant.Basic.Condition, &variant.Basic.SKU, &variant.Basic.IsVariant,
			&variant.Price.Price, &variant.Price.Currency,
		)
		if err != nil {
			return err
		}
		variants = append(variants, &variant)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	p.Variants = variants
	return nil
}

// BeforeDelete handles cleanup before deleting a product
func (p *Product) BeforeDelete(db pgx.Tx) error {
	// Delete all variants
	_, err := db.Exec(context.Background(),
		"DELETE FROM products WHERE parent_id = $1", p.ID)
	if err != nil {
		return fmt.Errorf("error deleting variants: %w", err)
	}

	// Delete all images
	_, err = db.Exec(context.Background(),
		"DELETE FROM images WHERE product_id = $1", p.ID)
	if err != nil {
		return fmt.Errorf("error deleting images: %w", err)
	}

	return nil
}
