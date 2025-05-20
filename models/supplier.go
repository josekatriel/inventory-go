package models

import (
	"time"

	"github.com/google/uuid"
)

// Supplier represents a supplier/vendor that provides products
type Supplier struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Email          string    `json:"email,omitempty" db:"email"`
	Phone          string    `json:"phone,omitempty" db:"phone"`
	Address        string    `json:"address,omitempty" db:"address"`
	ContactPerson  string    `json:"contact_person,omitempty" db:"contact_person"`
	TotalPurchases int       `json:"total_purchases" db:"total_purchases"`
	TotalSpent     float64   `json:"total_spent" db:"total_spent"`
	LastOrderAt    time.Time `json:"last_order_at,omitempty" db:"last_order_at"`
	Notes          string    `json:"notes,omitempty" db:"notes"`
	
	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// NewSupplier creates a new supplier with a generated UUID
func NewSupplier() *Supplier {
	now := time.Now()
	return &Supplier{
		ID:        uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
	}
}
