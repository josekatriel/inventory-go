package models

import (
	"time"

	"github.com/google/uuid"
)

// StockInStatus represents the status of a stock-in transaction
type StockInStatus string

const (
	StockInStatusDraft     StockInStatus = "draft"
	StockInStatusCompleted StockInStatus = "completed"
	StockInStatusCancelled StockInStatus = "cancelled"
)

// StockInSummary represents a summary of stock-in data
// for a given time period
type StockInSummary struct {
	TotalOrders       int     `json:"total_orders"`
	TotalCost         float64 `json:"total_cost"`
	CompletedOrders   int     `json:"completed_orders"`
	CompletedCost     float64 `json:"completed_cost"`
	AverageOrderValue float64 `json:"average_order_value"`
}

// DailyStockIn represents the stock-in data for a single day
type DailyStockIn struct {
	Date       string  `json:"date"`
	OrderCount int     `json:"order_count"`
	TotalCost  float64 `json:"total_cost"`
}

// StockIn represents a stock-in transaction (purchase order)
type StockIn struct {
	ID          string       `json:"id" db:"id"`
	ReferenceNo string       `json:"reference_no" db:"reference_no"`
	Status      StockInStatus `json:"status" db:"status"`
	OrderDate   time.Time    `json:"order_date" db:"order_date"`
	Note        string       `json:"note,omitempty" db:"note"`
	Total       float64      `json:"total" db:"total"`
	Paid        float64      `json:"paid" db:"paid"`
	Balance     float64      `json:"balance" db:"balance"`

	// Relations
	SupplierID *string       `json:"supplier_id,omitempty" db:"supplier_id"`
	Supplier   *Supplier     `json:"supplier,omitempty" db:"-"`
	Items      []StockInItem `json:"items" db:"-"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// StockInItem represents an item in a stock-in transaction
type StockInItem struct {
	ID          string  `json:"id" db:"id"`
	StockInID   string  `json:"stock_in_id" db:"stock_in_id"`
	ProductID   string  `json:"product_id" db:"product_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	Quantity    int     `json:"quantity" db:"quantity"`
	UnitCost    float64 `json:"unit_cost" db:"unit_cost"`
	Tax         float64 `json:"tax,omitempty" db:"tax"`
	Discount    float64 `json:"discount,omitempty" db:"discount"`
	Subtotal    float64 `json:"subtotal" db:"subtotal"`

	// Relations
	Product *Product `json:"product,omitempty" db:"-"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// NewStockIn creates a new stock-in with a generated UUID
func NewStockIn() *StockIn {
	now := time.Now()
	return &StockIn{
		ID:        uuid.NewString(),
		Status:    StockInStatusDraft,
		OrderDate: now,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewStockInItem creates a new stock-in item with a generated UUID
func NewStockInItem() *StockInItem {
	now := time.Now()
	return &StockInItem{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}
