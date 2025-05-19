package repositories

import (
	"context"
	"errors"
	"fmt"
	"inventory-go/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SaleRepository interface {
	// Basic CRUD
	GetByID(id string) (*models.Sale, error)
	GetByReference(referenceNo string) (*models.Sale, error)
	Create(sale *models.Sale) error
	Update(sale *models.Sale) error
	Delete(id string) error

	// Queries
	List(offset, limit int, status string, customerID *string, startDate, endDate *time.Time) ([]models.Sale, int64, error)
	GetSalesByCustomer(customerID string) ([]models.Sale, error)
	GetSalesSummary(startDate, endDate time.Time) (*models.SalesSummary, error)
	GetDailySales(startDate, endDate time.Time) ([]models.DailySales, error)
}

type SaleRepositoryImpl struct {
	db *pgx.Conn
}

func NewSaleRepository(db *pgx.Conn) SaleRepository {
	return &SaleRepositoryImpl{db: db}
}

func (r *SaleRepositoryImpl) GetByID(id string) (*models.Sale, error) {
	ctx := context.Background()
	var sale models.Sale

	// Get sale details
	saleQuery := `SELECT id, reference_no, status, sale_date, note, total, paid, balance, 
		customer_id, platform, created_at, updated_at 
		FROM sales WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, saleQuery, id).Scan(
		&sale.ID, &sale.ReferenceNo, &sale.Status, &sale.SaleDate, &sale.Note,
		&sale.Total, &sale.Paid, &sale.Balance, &sale.CustomerID, &sale.Platform,
		&sale.CreatedAt, &sale.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sale: %w", err)
	}

	// Get sale items
	itemsQuery := `SELECT id, product_id, product_name, quantity, unit_price, tax, discount, subtotal
		FROM sale_items WHERE sale_id = $1 AND deleted_at IS NULL`
		
	rows, err := r.db.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get sale items: %w", err)
	}
	defer rows.Close()

	var items []models.SaleItem
	for rows.Next() {
		var item models.SaleItem
		item.SaleID = id
		
		err := rows.Scan(
			&item.ID, &item.ProductID, &item.ProductName, &item.Quantity,
			&item.UnitPrice, &item.Tax, &item.Discount, &item.Subtotal,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sale item: %w", err)
		}
		items = append(items, item)
	}
	sale.Items = items

	return &sale, nil
}

func (r *SaleRepositoryImpl) GetByReference(referenceNo string) (*models.Sale, error) {
	// First get the ID of the sale by reference
	var id string
	query := `SELECT id FROM sales WHERE reference_no = $1 AND deleted_at IS NULL`
	
	err := r.db.QueryRow(context.Background(), query, referenceNo).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get sale by reference: %w", err)
	}
	
	// Then use GetByID to get the full sale
	return r.GetByID(id)
}

func (r *SaleRepositoryImpl) Create(sale *models.Sale) error {
	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Generate ID if not set
	if sale.ID == "" {
		sale.ID = uuid.NewString()
	}
	
	// Insert sale
	saleQuery := `INSERT INTO sales (
		id, reference_no, status, sale_date, note, total, paid, balance,
		customer_id, platform, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	
	_, err = tx.Exec(ctx, saleQuery,
		sale.ID, sale.ReferenceNo, sale.Status, sale.SaleDate, sale.Note,
		sale.Total, sale.Paid, sale.Balance, sale.CustomerID, sale.Platform,
		time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert sale: %w", err)
	}

	// Update customer stats if customer exists
	if sale.CustomerID != nil {
		updateCustomerQuery := `UPDATE customers SET 
			total_orders = total_orders + 1,
			total_spent = total_spent + $1,
			last_order_at = $2,
			updated_at = $3
			WHERE id = $4`
		
		_, err = tx.Exec(ctx, updateCustomerQuery,
			sale.Total, time.Now(), time.Now(), *sale.CustomerID,
		)
		if err != nil {
			return fmt.Errorf("failed to update customer stats: %w", err)
		}
	}
	
	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

func (r *SaleRepositoryImpl) Update(sale *models.Sale) error {
	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	sale.UpdatedAt = time.Now()
	
	// Update the basic sale information
	updateQuery := `UPDATE sales SET 
		reference_no = $1, status = $2, sale_date = $3, note = $4, 
		total = $5, paid = $6, balance = $7, customer_id = $8, 
		platform = $9, updated_at = $10
		WHERE id = $11`
	
	_, err = tx.Exec(ctx, updateQuery,
		sale.ReferenceNo, sale.Status, sale.SaleDate, sale.Note,
		sale.Total, sale.Paid, sale.Balance, sale.CustomerID,
		sale.Platform, sale.UpdatedAt, sale.ID,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update sale: %w", err)
	}
	
	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

func (r *SaleRepositoryImpl) Delete(id string) error {
	// Soft delete the sale
	query := `UPDATE sales SET deleted_at = $1 WHERE id = $2`
	_, err := r.db.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete sale: %w", err)
	}
	return nil
}

func (r *SaleRepositoryImpl) List(offset, limit int, status string, customerID *string, startDate, endDate *time.Time) ([]models.Sale, int64, error) {
	var sales []models.Sale
	var total int64 = 0
	// Basic implementation to allow compilation
	return sales, total, nil
}

func (r *SaleRepositoryImpl) GetSalesByCustomer(customerID string) ([]models.Sale, error) {
	var sales []models.Sale
	// Basic implementation to allow compilation
	return sales, nil
}

func (r *SaleRepositoryImpl) GetSalesSummary(startDate, endDate time.Time) (*models.SalesSummary, error) {
	var summary models.SalesSummary
	// Basic implementation to allow compilation
	return &summary, nil
}

func (r *SaleRepositoryImpl) GetDailySales(startDate, endDate time.Time) ([]models.DailySales, error) {
	var results []models.DailySales
	// Basic implementation to allow compilation
	return results, nil
}
