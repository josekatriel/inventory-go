package repositories

import (
	"context"
	"errors"
	"fmt"
	"inventory-go/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StockInRepository interface {
	// Basic CRUD
	GetByID(id string) (*models.StockIn, error)
	GetByReference(referenceNo string) (*models.StockIn, error)
	Create(stockIn *models.StockIn) error
	Update(stockIn *models.StockIn) error
	Delete(id string) error

	// Items
	AddStockInItem(item *models.StockInItem) error
	UpdateStockInItem(item *models.StockInItem) error
	DeleteStockInItem(id string) error
	GetStockInItems(stockInID string) ([]models.StockInItem, error)

	// Queries
	List(offset, limit int, status string, supplierID *string, startDate, endDate *time.Time) ([]models.StockIn, int64, error)
	GetStockInsBySupplier(supplierID string) ([]models.StockIn, error)
	GetStockInSummary(startDate, endDate time.Time) (*models.StockInSummary, error)
	GetDailyStockIn(startDate, endDate time.Time) ([]models.DailyStockIn, error)
}

type StockInRepositoryImpl struct {
	db *pgxpool.Pool
}

func NewStockInRepository(db *pgxpool.Pool) StockInRepository {
	return &StockInRepositoryImpl{db: db}
}

func (r *StockInRepositoryImpl) GetByID(id string) (*models.StockIn, error) {
	ctx := context.Background()
	var stockIn models.StockIn

	// Get stockIn details
	stockInQuery := `SELECT id, reference_no, status, order_date, note, total, paid, balance, 
		supplier_id, created_at, updated_at 
		FROM stock_ins WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, stockInQuery, id).Scan(
		&stockIn.ID, &stockIn.ReferenceNo, &stockIn.Status, &stockIn.OrderDate, &stockIn.Note,
		&stockIn.Total, &stockIn.Paid, &stockIn.Balance, &stockIn.SupplierID,
		&stockIn.CreatedAt, &stockIn.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get stock-in: %w", err)
	}

	// Get stock-in items
	items, err := r.GetStockInItems(id)
	if err != nil {
		return nil, err
	}
	stockIn.Items = items

	// Get supplier if exists
	if stockIn.SupplierID != nil {
		supplierRepo := NewSupplierRepository(r.db)
		supplier, err := supplierRepo.GetByID(*stockIn.SupplierID)
		if err != nil {
			return nil, fmt.Errorf("failed to get supplier: %w", err)
		}
		stockIn.Supplier = supplier
	}

	return &stockIn, nil
}

func (r *StockInRepositoryImpl) GetByReference(referenceNo string) (*models.StockIn, error) {
	// First get the ID of the stockIn by reference
	var id string
	query := `SELECT id FROM stock_ins WHERE reference_no = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(context.Background(), query, referenceNo).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get stock-in by reference: %w", err)
	}

	// Then use GetByID to get the full stockIn
	return r.GetByID(id)
}

func (r *StockInRepositoryImpl) Create(stockIn *models.StockIn) error {
	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Generate ID if not set
	if stockIn.ID == "" {
		stockIn.ID = uuid.NewString()
	}

	// Insert stockIn
	stockInQuery := `INSERT INTO stock_ins (
		id, reference_no, status, order_date, note, total, paid, balance,
		supplier_id, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(ctx, stockInQuery,
		stockIn.ID, stockIn.ReferenceNo, stockIn.Status, stockIn.OrderDate, stockIn.Note,
		stockIn.Total, stockIn.Paid, stockIn.Balance, stockIn.SupplierID,
		time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock-in: %w", err)
	}

	// Insert stock-in items
	for i := range stockIn.Items {
		item := &stockIn.Items[i]
		item.StockInID = stockIn.ID

		if item.ID == "" {
			item.ID = uuid.NewString()
		}

		itemQuery := `INSERT INTO stock_in_items (
			id, stock_in_id, product_id, product_name, quantity, 
			unit_cost, tax, discount, subtotal, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

		_, err = tx.Exec(ctx, itemQuery,
			item.ID, item.StockInID, item.ProductID, item.ProductName, item.Quantity,
			item.UnitCost, item.Tax, item.Discount, item.Subtotal, time.Now(), time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert stock-in item: %w", err)
		}

		// Update product stock
		_, err = tx.Exec(ctx, `UPDATE products SET stock = stock + $1 WHERE id = $2`,
			item.Quantity, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	// Update supplier stats if supplier exists
	if stockIn.SupplierID != nil {
		updateSupplierQuery := `UPDATE suppliers SET 
			total_purchases = total_purchases + 1,
			total_spent = total_spent + $1,
			last_order_at = $2,
			updated_at = $3
			WHERE id = $4`

		_, err = tx.Exec(ctx, updateSupplierQuery,
			stockIn.Total, time.Now(), time.Now(), *stockIn.SupplierID,
		)
		if err != nil {
			return fmt.Errorf("failed to update supplier stats: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *StockInRepositoryImpl) Update(stockIn *models.StockIn) error {
	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	stockIn.UpdatedAt = time.Now()

	// Update the basic stockIn information
	updateQuery := `UPDATE stock_ins SET 
		reference_no = $1, status = $2, order_date = $3, note = $4, 
		total = $5, paid = $6, balance = $7, supplier_id = $8, 
		updated_at = $9
		WHERE id = $10`

	_, err = tx.Exec(ctx, updateQuery,
		stockIn.ReferenceNo, stockIn.Status, stockIn.OrderDate, stockIn.Note,
		stockIn.Total, stockIn.Paid, stockIn.Balance, stockIn.SupplierID,
		stockIn.UpdatedAt, stockIn.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update stock-in: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *StockInRepositoryImpl) Delete(id string) error {
	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get all items to revert stock
	rows, err := tx.Query(ctx,
		`SELECT product_id, quantity FROM stock_in_items WHERE stock_in_id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to get stock-in items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var productID string
		var quantity int
		if err := rows.Scan(&productID, &quantity); err != nil {
			return fmt.Errorf("failed to scan item: %w", err)
		}

		// Revert stock change
		_, err = tx.Exec(ctx, `UPDATE products SET stock = stock - $1 WHERE id = $2`,
			quantity, productID)
		if err != nil {
			return fmt.Errorf("failed to revert product stock: %w", err)
		}
	}

	// Soft delete the items
	_, err = tx.Exec(ctx,
		`UPDATE stock_in_items SET deleted_at = $1 WHERE stock_in_id = $2`,
		time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete stock-in items: %w", err)
	}

	// Soft delete the stock-in
	_, err = tx.Exec(ctx,
		`UPDATE stock_ins SET deleted_at = $1 WHERE id = $2`,
		time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete stock-in: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *StockInRepositoryImpl) AddStockInItem(item *models.StockInItem) error {
	if item.ID == "" {
		item.ID = uuid.NewString()
	}

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now

	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert the item
	query := `INSERT INTO stock_in_items (
		id, stock_in_id, product_id, product_name, quantity, 
		unit_cost, tax, discount, subtotal, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = tx.Exec(ctx, query,
		item.ID, item.StockInID, item.ProductID, item.ProductName, item.Quantity,
		item.UnitCost, item.Tax, item.Discount, item.Subtotal, item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock-in item: %w", err)
	}

	// Update product stock
	_, err = tx.Exec(ctx, `UPDATE products SET stock = stock + $1 WHERE id = $2`,
		item.Quantity, item.ProductID)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}

	// Update stock-in total
	_, err = tx.Exec(ctx, `
		UPDATE stock_ins SET 
		total = (SELECT SUM(subtotal) FROM stock_in_items WHERE stock_in_id = $1 AND deleted_at IS NULL),
		updated_at = $2
		WHERE id = $1`,
		item.StockInID, now,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock-in total: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *StockInRepositoryImpl) UpdateStockInItem(item *models.StockInItem) error {
	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get original quantity
	var originalQuantity int
	err = tx.QueryRow(ctx,
		`SELECT quantity FROM stock_in_items WHERE id = $1`, item.ID).Scan(&originalQuantity)
	if err != nil {
		return fmt.Errorf("failed to get original quantity: %w", err)
	}

	// Calculate quantity difference
	quantityDiff := item.Quantity - originalQuantity

	// Update the item
	item.UpdatedAt = time.Now()
	query := `UPDATE stock_in_items SET 
		product_id = $1, product_name = $2, quantity = $3, 
		unit_cost = $4, tax = $5, discount = $6, subtotal = $7, updated_at = $8
		WHERE id = $9`

	_, err = tx.Exec(ctx, query,
		item.ProductID, item.ProductName, item.Quantity,
		item.UnitCost, item.Tax, item.Discount, item.Subtotal, item.UpdatedAt, item.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock-in item: %w", err)
	}

	// Update product stock
	if quantityDiff != 0 {
		_, err = tx.Exec(ctx, `UPDATE products SET stock = stock + $1 WHERE id = $2`,
			quantityDiff, item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	// Update stock-in total
	_, err = tx.Exec(ctx, `
		UPDATE stock_ins SET 
		total = (SELECT SUM(subtotal) FROM stock_in_items WHERE stock_in_id = $1 AND deleted_at IS NULL),
		updated_at = $2
		WHERE id = $1`,
		item.StockInID, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock-in total: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *StockInRepositoryImpl) DeleteStockInItem(id string) error {
	// Begin transaction
	ctx := context.Background()
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get item details before deleting
	var item models.StockInItem
	var stockInID string
	err = tx.QueryRow(ctx, `
		SELECT id, stock_in_id, product_id, quantity FROM stock_in_items WHERE id = $1
	`, id).Scan(&item.ID, &stockInID, &item.ProductID, &item.Quantity)
	if err != nil {
		return fmt.Errorf("failed to get item details: %w", err)
	}

	// Soft delete the item
	now := time.Now()
	_, err = tx.Exec(ctx, `UPDATE stock_in_items SET deleted_at = $1 WHERE id = $2`,
		now, id)
	if err != nil {
		return fmt.Errorf("failed to delete stock-in item: %w", err)
	}

	// Update product stock to revert quantity
	_, err = tx.Exec(ctx, `UPDATE products SET stock = stock - $1 WHERE id = $2`,
		item.Quantity, item.ProductID)
	if err != nil {
		return fmt.Errorf("failed to revert product stock: %w", err)
	}

	// Update stock-in total
	_, err = tx.Exec(ctx, `
		UPDATE stock_ins SET 
		total = (SELECT COALESCE(SUM(subtotal), 0) FROM stock_in_items WHERE stock_in_id = $1 AND deleted_at IS NULL),
		updated_at = $2
		WHERE id = $1`,
		stockInID, now,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock-in total: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *StockInRepositoryImpl) GetStockInItems(stockInID string) ([]models.StockInItem, error) {
	query := `SELECT id, stock_in_id, product_id, product_name, quantity, 
		unit_cost, tax, discount, subtotal, created_at, updated_at
		FROM stock_in_items 
		WHERE stock_in_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.Query(context.Background(), query, stockInID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock-in items: %w", err)
	}
	defer rows.Close()

	var items []models.StockInItem
	for rows.Next() {
		var item models.StockInItem

		err := rows.Scan(
			&item.ID, &item.StockInID, &item.ProductID, &item.ProductName, &item.Quantity,
			&item.UnitCost, &item.Tax, &item.Discount, &item.Subtotal,
			&item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock-in item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock-in items: %w", err)
	}

	return items, nil
}

func (r *StockInRepositoryImpl) List(offset, limit int, status string, supplierID *string, startDate, endDate *time.Time) ([]models.StockIn, int64, error) {
	ctx := context.Background()

	// Build query conditions
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if supplierID != nil {
		conditions = append(conditions, fmt.Sprintf("supplier_id = $%d", argIndex))
		args = append(args, supplierID)
		argIndex++
	}

	if startDate != nil {
		conditions = append(conditions, fmt.Sprintf("order_date >= $%d", argIndex))
		args = append(args, startDate)
		argIndex++
	}

	if endDate != nil {
		conditions = append(conditions, fmt.Sprintf("order_date <= $%d", argIndex))
		args = append(args, endDate)
		argIndex++
	}

	// Join conditions
	whereClause := "WHERE " + conditions[0]
	for i := 1; i < len(conditions); i++ {
		whereClause += " AND " + conditions[i]
	}

	// Count total matching records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM stock_ins %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count stock-ins: %w", err)
	}

	// Get paginated results
	query := fmt.Sprintf(`
		SELECT id, reference_no, status, order_date, note, total, paid, balance, 
		supplier_id, created_at, updated_at 
		FROM stock_ins 
		%s
		ORDER BY order_date DESC
		LIMIT $%d OFFSET $%d`,
		whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query stock-ins: %w", err)
	}
	defer rows.Close()

	var stockIns []models.StockIn
	for rows.Next() {
		var stockIn models.StockIn

		err := rows.Scan(
			&stockIn.ID, &stockIn.ReferenceNo, &stockIn.Status, &stockIn.OrderDate, &stockIn.Note,
			&stockIn.Total, &stockIn.Paid, &stockIn.Balance, &stockIn.SupplierID,
			&stockIn.CreatedAt, &stockIn.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan stock-in: %w", err)
		}

		stockIns = append(stockIns, stockIn)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating stock-ins: %w", err)
	}

	return stockIns, total, nil
}

func (r *StockInRepositoryImpl) GetStockInsBySupplier(supplierID string) ([]models.StockIn, error) {
	query := `
		SELECT id, reference_no, status, order_date, note, total, paid, balance, 
		supplier_id, created_at, updated_at 
		FROM stock_ins 
		WHERE supplier_id = $1 AND deleted_at IS NULL
		ORDER BY order_date DESC`

	rows, err := r.db.Query(context.Background(), query, supplierID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock-ins by supplier: %w", err)
	}
	defer rows.Close()

	var stockIns []models.StockIn
	for rows.Next() {
		var stockIn models.StockIn

		err := rows.Scan(
			&stockIn.ID, &stockIn.ReferenceNo, &stockIn.Status, &stockIn.OrderDate, &stockIn.Note,
			&stockIn.Total, &stockIn.Paid, &stockIn.Balance, &stockIn.SupplierID,
			&stockIn.CreatedAt, &stockIn.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stock-in: %w", err)
		}

		stockIns = append(stockIns, stockIn)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock-ins: %w", err)
	}

	return stockIns, nil
}

func (r *StockInRepositoryImpl) GetStockInSummary(startDate, endDate time.Time) (*models.StockInSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_orders,
			COALESCE(SUM(total), 0) as total_cost,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_orders,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN total ELSE 0 END), 0) as completed_cost
		FROM stock_ins
		WHERE order_date BETWEEN $1 AND $2
		AND deleted_at IS NULL`

	var summary models.StockInSummary
	err := r.db.QueryRow(context.Background(), query, startDate, endDate).Scan(
		&summary.TotalOrders, &summary.TotalCost,
		&summary.CompletedOrders, &summary.CompletedCost,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stock-in summary: %w", err)
	}

	// Calculate average order value
	if summary.TotalOrders > 0 {
		summary.AverageOrderValue = summary.TotalCost / float64(summary.TotalOrders)
	}

	return &summary, nil
}

func (r *StockInRepositoryImpl) GetDailyStockIn(startDate, endDate time.Time) ([]models.DailyStockIn, error) {
	query := `
		SELECT 
			TO_CHAR(order_date, 'YYYY-MM-DD') as date,
			COUNT(*) as order_count,
			COALESCE(SUM(total), 0) as total_cost
		FROM stock_ins
		WHERE order_date BETWEEN $1 AND $2
		AND deleted_at IS NULL
		GROUP BY TO_CHAR(order_date, 'YYYY-MM-DD')
		ORDER BY date`

	rows, err := r.db.Query(context.Background(), query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily stock-in data: %w", err)
	}
	defer rows.Close()

	var results []models.DailyStockIn
	for rows.Next() {
		var day models.DailyStockIn

		err := rows.Scan(&day.Date, &day.OrderCount, &day.TotalCost)
		if err != nil {
			return nil, fmt.Errorf("failed to scan daily stock-in data: %w", err)
		}

		results = append(results, day)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating daily stock-in data: %w", err)
	}

	return results, nil
}
