// models/sale.go
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SalesSummary represents a summary of sales data
// for a given time period
type SalesSummary struct {
	TotalOrders       int     `json:"total_orders"`
	TotalSales        float64 `json:"total_sales"`
	CompletedOrders   int     `json:"completed_orders"`
	CompletedSales    float64 `json:"completed_sales"`
	AverageOrderValue float64 `json:"average_order_value"`
}

// DailySales represents the sales data for a single day
type DailySales struct {
	Date       string  `json:"date"`
	OrderCount int     `json:"order_count"`
	TotalSales float64 `json:"total_sales"`
}

type SaleStatus string

const (
	SaleStatusDraft     SaleStatus = "draft"
	SaleStatusCompleted SaleStatus = "completed"
	SaleStatusCancelled SaleStatus = "cancelled"
)

type PlatformType string

const (
	PlatformOfflineStore PlatformType = "OfflineStore"
	PlatformWebsite      PlatformType = "Website"
	PlatformTokopedia    PlatformType = "Tokopedia"
	PlatformShopee       PlatformType = "Shopee"
	PlatformLazada       PlatformType = "Lazada"
	PlatformBlibli       PlatformType = "Blibli"
	PlatformBukalapak    PlatformType = "Bukalapak"
	PlatformOther        PlatformType = "Other"
)

type Sale struct {
	ID          string     `gorm:"type:char(36);primaryKey" json:"id"`
	ReferenceNo string     `gorm:"index" json:"reference_no"`
	Status      SaleStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	SaleDate    time.Time  `json:"sale_date"`
	Note        string     `gorm:"type:text" json:"note,omitempty"`
	Total       float64    `gorm:"not null" json:"total"`
	Paid        float64    `gorm:"default:0" json:"paid"`
	Balance     float64    `gorm:"default:0" json:"balance"`

	// Relations
	CustomerID *string       `gorm:"type:char(36);index" json:"customer_id,omitempty"`
	Customer   *Customer     `json:"customer,omitempty" gorm:"foreignKey:CustomerID"`
	Items      []SaleItem    `json:"items" gorm:"foreignKey:SaleID"`
	Payments   []SalePayment `json:"payments,omitempty" gorm:"foreignKey:SaleID"`
	Platform   PlatformType  `gorm:"type:varchar(50);not null;uniqueIndex" json:"platform"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type SaleItem struct {
	ID          string  `gorm:"type:char(36);primaryKey" json:"id"`
	SaleID      string  `gorm:"type:char(36);index" json:"sale_id"`
	ProductID   string  `gorm:"type:char(36);index" json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `gorm:"not null;check:quantity > 0"`
	UnitPrice   float64 `gorm:"not null;check:unit_price >= 0"`
	Tax         float64 `gorm:"default:0" json:"tax"`
	Discount    float64 `gorm:"default:0" json:"discount"`
	Subtotal    float64 `json:"subtotal"`

	// Relations
	Product *Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type SalePayment struct {
	ID            string    `gorm:"type:char(36);primaryKey" json:"id"`
	SaleID        string    `gorm:"type:char(36);index" json:"sale_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"` // e.g., "cash", "credit_card", "bank_transfer"
	Reference     string    `json:"reference,omitempty"`
	Note          string    `json:"note,omitempty"`
	PaymentDate   time.Time `json:"payment_date"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate hook for Sale
func (s *Sale) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.ReferenceNo == "" {
		s.ReferenceNo = "SALE-" + time.Now().Format("20060102") + "-" + s.ID[:6]
	}
	if s.SaleDate.IsZero() {
		s.SaleDate = time.Now()
	}

	// Set SaleID for all items
	for i := range s.Items {
		s.Items[i].SaleID = s.ID
	}

	// Set SaleID for all payments
	for i := range s.Payments {
		s.Payments[i].SaleID = s.ID
	}

	s.CalculateTotals()
	return nil
}

func (i *SaleItem) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = uuid.NewString()
	}
	i.CalculateSubtotal()
	return nil
}

func (s *SalePayment) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.PaymentDate.IsZero() {
		s.PaymentDate = time.Now()
	}
	return nil
}

// Calculate methods
func (s *Sale) CalculateTotals() {
	var subtotal float64

	// Calculate subtotal from items
	for _, item := range s.Items {
		item.CalculateSubtotal()
		subtotal += item.Subtotal
	}

	s.Total = subtotal
	s.Balance = s.Total - s.Paid
}

func (i *SaleItem) CalculateSubtotal() {
	// Calculate subtotal: (unit_price * quantity) + tax - discount
	itemTotal := (i.UnitPrice * float64(i.Quantity))
	i.Subtotal = itemTotal + i.Tax - i.Discount
}

// Helper methods
func (s *Sale) AddPayment(amount float64, method string, reference string, note string) {
	payment := SalePayment{
		Amount:        amount,
		PaymentMethod: method,
		Reference:     reference,
		Note:          note,
		PaymentDate:   time.Now(),
	}
	s.Payments = append(s.Payments, payment)
	s.Paid += amount
	s.Balance = s.Total - s.Paid
}

func (s *Sale) UpdateStatus(status SaleStatus) {
	s.Status = status
	if status == SaleStatusCompleted {
		s.Balance = 0 // Ensure balance is zero when completed
		s.Paid = s.Total
	}
}

func (s *Sale) Validate() error {
	if len(s.Items) == 0 {
		return errors.New("sale must have at least one item")
	}
	return nil
}
