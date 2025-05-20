package models

import (
	"time"

	"github.com/google/uuid"
)

// Customer model
type Customer struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Email       string    `json:"email,omitempty" db:"email"`
	Phone       string    `json:"phone,omitempty" db:"phone"`
	Address     string    `json:"address,omitempty" db:"address"`
	TotalOrders int       `json:"total_orders" db:"total_orders"`
	TotalSpent  float64   `json:"total_spent" db:"total_spent"`
	LastOrderAt time.Time `json:"last_order_at,omitempty" db:"last_order_at"`
	Notes       string    `json:"notes,omitempty" db:"notes"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// GenerateID sets a UUID if ID is empty
func (c *Customer) GenerateID() {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
}
