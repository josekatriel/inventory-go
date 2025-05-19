package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Customer model
type Customer struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Email       string    `gorm:"uniqueIndex" json:"email,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	Address     string    `gorm:"type:text" json:"address,omitempty"`
	TotalOrders int       `gorm:"default:0" json:"total_orders"`
	TotalSpent  float64   `gorm:"default:0" json:"total_spent"`
	LastOrderAt time.Time `json:"last_order_at,omitempty"`
	Notes       string    `gorm:"type:text" json:"notes,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate hook for Customer
func (c *Customer) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}
