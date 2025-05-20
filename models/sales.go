// models/sale.go
package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
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
	ID          string     `json:"id" db:"id"`
	ReferenceNo string     `json:"reference_no" db:"reference_no"`
	Status      SaleStatus `json:"status" db:"status"`
	SaleDate    time.Time  `json:"sale_date" db:"sale_date"`
	Note        string     `json:"note,omitempty" db:"note"`
	Total       float64    `json:"total" db:"total"`
	Paid        float64    `json:"paid" db:"paid"`
	Balance     float64    `json:"balance" db:"balance"`

	// Relations
	CustomerID *string       `json:"customer_id,omitempty" db:"customer_id"`
	Customer   *Customer     `json:"customer,omitempty" db:"-"`
	Items      []SaleItem    `json:"items" db:"-"`
	Payments   []SalePayment `json:"payments,omitempty" db:"-"`
	Platform   PlatformType  `json:"platform" db:"platform"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

type SaleItem struct {
	ID          string  `json:"id" db:"id"`
	SaleID      string  `json:"sale_id" db:"sale_id"`
	ProductID   string  `json:"product_id" db:"product_id"`
	ProductName string  `json:"product_name" db:"product_name"`
	Quantity    int     `json:"quantity" db:"quantity"`
	UnitPrice   float64 `json:"unit_price" db:"unit_price"`
	Tax         float64 `json:"tax" db:"tax"`
	Discount    float64 `json:"discount" db:"discount"`
	Subtotal    float64 `json:"subtotal" db:"subtotal"`

	// Relations
	Product *Product `json:"product,omitempty" db:"-"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

type SalePayment struct {
	ID            string    `json:"id" db:"id"`
	SaleID        string    `json:"sale_id" db:"sale_id"`
	Amount        float64   `json:"amount" db:"amount"`
	PaymentMethod string    `json:"payment_method" db:"payment_method"` // e.g., "cash", "credit_card", "bank_transfer"
	Reference     string    `json:"reference,omitempty" db:"reference"`
	Note          string    `json:"note,omitempty" db:"note"`
	PaymentDate   time.Time `json:"payment_date" db:"payment_date"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"-" db:"deleted_at"`
}

// GenerateID sets a UUID if ID is empty and ensures reference number exists
func (s *Sale) GenerateID() {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.ReferenceNo == "" {
		s.ReferenceNo = "SALE-" + time.Now().Format("20060102") + "-" + s.ID[:6]
	}
	if s.SaleDate.IsZero() {
		s.SaleDate = time.Now()
	}
}

// PrepareSave sets the SaleID for all items and payments and calculates totals
func (s *Sale) PrepareSave() {
	// Set SaleID for all items
	for i := range s.Items {
		s.Items[i].SaleID = s.ID
	}

	// Set SaleID for all payments
	for i := range s.Payments {
		s.Payments[i].SaleID = s.ID
	}

	s.CalculateTotals()
}

// GenerateID sets a UUID if ID is empty
func (i *SaleItem) GenerateID() {
	if i.ID == "" {
		i.ID = uuid.NewString()
	}
	i.CalculateSubtotal()
}

// GenerateID sets a UUID if ID is empty
func (s *SalePayment) GenerateID() {
	if s.ID == "" {
		s.ID = uuid.NewString()
	}
	if s.PaymentDate.IsZero() {
		s.PaymentDate = time.Now()
	}
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
