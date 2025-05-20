package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// RejectStatus defines the status of a reject
type RejectStatus string

const (
	// RejectStatusPending means the reject is pending
	RejectStatusPending RejectStatus = "pending"
	// RejectStatusCompleted means the reject has been completed
	RejectStatusCompleted RejectStatus = "completed"
	// RejectStatusCancelled means the reject has been cancelled
	RejectStatusCancelled RejectStatus = "cancelled"
)

// Reject represents a stock rejection (removal from inventory)
type Reject struct {
	ID          string       `json:"id" db:"id"`
	ReferenceNo string       `json:"reference_no" db:"reference_no"`
	Status      RejectStatus `json:"status" db:"status"`
	RejectDate  time.Time    `json:"reject_date" db:"reject_date"`
	Reason      string       `json:"reason" db:"reason"`
	Total       float64      `json:"total" db:"total"`
	Items       []RejectItem `json:"items" db:"-"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time   `json:"-" db:"deleted_at"`
}

// RejectItem represents a line item in a stock rejection
type RejectItem struct {
	ID          string     `json:"id" db:"id"`
	RejectID    string     `json:"reject_id" db:"reject_id"`
	ProductID   string     `json:"product_id" db:"product_id"`
	ProductName string     `json:"product_name" db:"product_name"`
	Quantity    int        `json:"quantity" db:"quantity"`
	UnitCost    float64    `json:"unit_cost" db:"unit_cost"`
	Subtotal    float64    `json:"subtotal" db:"subtotal"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`
}

// RejectSummary represents summary statistics for stock rejections
type RejectSummary struct {
	TotalRejects          int     `json:"total_rejects"`
	TotalCompletedRejects int     `json:"total_completed_rejects"`
	TotalPendingRejects   int     `json:"total_pending_rejects"`
	TotalCancelledRejects int     `json:"total_cancelled_rejects"`
	TotalValue            float64 `json:"total_value"`
	TotalRejectedItems    int     `json:"total_rejected_items"`
	PeriodStart           string  `json:"period_start"`
	PeriodEnd             string  `json:"period_end"`
}

// DailyReject represents daily stock rejection data for reporting
type DailyReject struct {
	Date         string  `json:"date"`
	TotalRejects int     `json:"total_rejects"`
	TotalValue   float64 `json:"total_value"`
	ItemCount    int     `json:"item_count"`
}

// NewReject creates a new rejection entry with default values
func NewReject() *Reject {
	return &Reject{
		ID:         uuid.NewString(),
		Status:     RejectStatusPending,
		RejectDate: time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// NewRejectItem creates a new reject item with default values
func NewRejectItem() *RejectItem {
	return &RejectItem{
		ID:        uuid.NewString(),
		Quantity:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Scan implements the sql.Scanner interface for Reject
func (r *Reject) Scan(value any) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, r)
	case string:
		return json.Unmarshal([]byte(v), r)
	default:
		return nil
	}
}

// Value implements the driver.Valuer interface for Reject
func (r Reject) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements the sql.Scanner interface for RejectItem
func (ri *RejectItem) Scan(value any) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ri)
	case string:
		return json.Unmarshal([]byte(v), ri)
	default:
		return nil
	}
}

// Value implements the driver.Valuer interface for RejectItem
func (ri RejectItem) Value() (driver.Value, error) {
	return json.Marshal(ri)
}
