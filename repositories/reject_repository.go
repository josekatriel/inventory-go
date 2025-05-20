package repositories

import (
	"context"
	"errors"
	"fmt"
	"inventory-go/models"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RejectRepository defines methods for reject operations
type RejectRepository interface {
	GetByID(id string) (*models.Reject, error)
	GetByReference(reference string) (*models.Reject, error)
	Create(reject *models.Reject) error
	Update(reject *models.Reject) error
	Delete(id string) error
	List(offset, limit int, status string, startDate, endDate *time.Time) ([]models.Reject, int64, error)
	AddRejectItem(item *models.RejectItem) error
	UpdateRejectItem(item *models.RejectItem) error
	DeleteRejectItem(id string) error
	GetRejectSummary(startDate, endDate time.Time) (*models.RejectSummary, error)
	GetDailyReject(startDate, endDate time.Time) ([]models.DailyReject, error)
}

// RejectRepositoryImpl implements the RejectRepository interface
type RejectRepositoryImpl struct {
	db *pgxpool.Pool
}

// NewRejectRepository creates a new RejectRepository
func NewRejectRepository(db *pgxpool.Pool) RejectRepository {
	return &RejectRepositoryImpl{db: db}
}

// GetByID retrieves a reject by its ID
func (r *RejectRepositoryImpl) GetByID(id string) (*models.Reject, error) {
	// First get the reject header
	query := `
		SELECT id, reference_no, status, reject_date, reason, total, created_at, updated_at 
		FROM rejects
		WHERE id = $1 AND deleted_at IS NULL
	`
	var reject models.Reject
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&reject.ID, &reject.ReferenceNo, &reject.Status, &reject.RejectDate,
		&reject.Reason, &reject.Total, &reject.CreatedAt, &reject.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("reject not found")
		}
		return nil, fmt.Errorf("error getting reject: %w", err)
	}

	// Get the reject items
	itemsQuery := `
		SELECT id, reject_id, product_id, product_name, quantity, unit_cost, subtotal, created_at, updated_at
		FROM reject_items
		WHERE reject_id = $1 AND deleted_at IS NULL
	`
	rows, err := r.db.Query(context.Background(), itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("error getting reject items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.RejectItem
		err := rows.Scan(
			&item.ID, &item.RejectID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.UnitCost, &item.Subtotal, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reject.Items = append(reject.Items, item)
	}

	return &reject, nil
}

// GetByReference retrieves a reject by its reference number
func (r *RejectRepositoryImpl) GetByReference(reference string) (*models.Reject, error) {
	query := `
		SELECT id
		FROM rejects
		WHERE reference_no = $1 AND deleted_at IS NULL
	`
	var id string
	err := r.db.QueryRow(context.Background(), query, reference).Scan(&id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return r.GetByID(id)
}

// Create creates a new reject
func (r *RejectRepositoryImpl) Create(reject *models.Reject) error {
	// Start a transaction
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Generate ID if not provided
	if reject.ID == "" {
		reject.ID = uuid.NewString()
	}

	// Insert the reject
	query := `
		INSERT INTO rejects (id, reference_no, status, reject_date, reason, total, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	var id string
	err = tx.QueryRow(context.Background(), query,
		reject.ID, reject.ReferenceNo, reject.Status, reject.RejectDate,
		reject.Reason, reject.Total, time.Now(), time.Now(),
	).Scan(&id)
	if err != nil {
		return err
	}

	// Insert the reject items
	for i := range reject.Items {
		item := &reject.Items[i]
		item.RejectID = id

		if item.ID == "" {
			item.ID = uuid.NewString()
		}

		itemQuery := `
			INSERT INTO reject_items (id, reject_id, product_id, product_name, quantity, unit_cost, subtotal, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err = tx.Exec(context.Background(), itemQuery,
			item.ID, item.RejectID, item.ProductID, item.ProductName,
			item.Quantity, item.UnitCost, item.Subtotal, time.Now(), time.Now(),
		)
		if err != nil {
			return err
		}
	}

	// Update the total if not provided
	if reject.Total == 0 && len(reject.Items) > 0 {
		var total float64
		for _, item := range reject.Items {
			total += item.Subtotal
		}

		updateQuery := `UPDATE rejects SET total = $1 WHERE id = $2`
		_, err = tx.Exec(context.Background(), updateQuery, total, id)
		if err != nil {
			return err
		}
		reject.Total = total
	}

	// Commit the transaction
	return tx.Commit(context.Background())
}

// Update updates an existing reject
func (r *RejectRepositoryImpl) Update(reject *models.Reject) error {
	query := `
		UPDATE rejects
		SET reference_no = $1, status = $2, reject_date = $3, reason = $4, total = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`
	_, err := r.db.Exec(context.Background(), query,
		reject.ReferenceNo, reject.Status, reject.RejectDate,
		reject.Reason, reject.Total, time.Now(), reject.ID,
	)
	return err
}

// Delete soft-deletes a reject
func (r *RejectRepositoryImpl) Delete(id string) error {
	// Start a transaction
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Delete the reject items
	itemsQuery := `UPDATE reject_items SET deleted_at = $1 WHERE reject_id = $2`
	_, err = tx.Exec(context.Background(), itemsQuery, time.Now(), id)
	if err != nil {
		return err
	}

	// Delete the reject
	query := `UPDATE rejects SET deleted_at = $1 WHERE id = $2`
	_, err = tx.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit(context.Background())
}

// List retrieves a list of rejects with pagination and filtering
func (r *RejectRepositoryImpl) List(offset, limit int, status string, startDate, endDate *time.Time) ([]models.Reject, int64, error) {
	// Build the query conditions
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if startDate != nil {
		conditions = append(conditions, fmt.Sprintf("reject_date >= $%d", argIndex))
		args = append(args, startDate)
		argIndex++
	}

	if endDate != nil {
		conditions = append(conditions, fmt.Sprintf("reject_date <= $%d", argIndex))
		args = append(args, endDate)
		argIndex++
	}

	// Count the total number of records
	whereClause := "WHERE " + strings.Join(conditions, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM rejects %s", whereClause)
	var total int64
	err := r.db.QueryRow(context.Background(), countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get the paginated results
	query := fmt.Sprintf(`
		SELECT id, reference_no, status, reject_date, reason, total, created_at, updated_at 
		FROM rejects 
		%s
		ORDER BY reject_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(context.Background(), query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	rejects := []models.Reject{}
	for rows.Next() {
		var reject models.Reject
		err := rows.Scan(
			&reject.ID, &reject.ReferenceNo, &reject.Status, &reject.RejectDate,
			&reject.Reason, &reject.Total, &reject.CreatedAt, &reject.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rejects = append(rejects, reject)
	}

	return rejects, total, nil
}

// AddRejectItem adds a new item to a reject
func (r *RejectRepositoryImpl) AddRejectItem(item *models.RejectItem) error {
	// Generate ID if not provided
	if item.ID == "" {
		item.ID = uuid.NewString()
	}

	// Start a transaction
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Insert the item
	query := `
		INSERT INTO reject_items (id, reject_id, product_id, product_name, quantity, unit_cost, subtotal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err = tx.Exec(context.Background(), query,
		item.ID, item.RejectID, item.ProductID, item.ProductName,
		item.Quantity, item.UnitCost, item.Subtotal, time.Now(), time.Now(),
	)
	if err != nil {
		return err
	}

	// Update the reject total
	updateQuery := `
		UPDATE rejects 
		SET total = (SELECT COALESCE(SUM(subtotal), 0) FROM reject_items WHERE reject_id = $1 AND deleted_at IS NULL),
		    updated_at = $2
		WHERE id = $1
	`
	_, err = tx.Exec(context.Background(), updateQuery, item.RejectID, time.Now())
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit(context.Background())
}

// UpdateRejectItem updates an existing reject item
func (r *RejectRepositoryImpl) UpdateRejectItem(item *models.RejectItem) error {
	// Start a transaction
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Update the item
	query := `
		UPDATE reject_items
		SET product_id = $1, product_name = $2, quantity = $3, unit_cost = $4, subtotal = $5, updated_at = $6
		WHERE id = $7 AND deleted_at IS NULL
	`
	_, err = tx.Exec(context.Background(), query,
		item.ProductID, item.ProductName, item.Quantity,
		item.UnitCost, item.Subtotal, time.Now(), item.ID,
	)
	if err != nil {
		return err
	}

	// Update the reject total
	updateQuery := `
		UPDATE rejects 
		SET total = (SELECT COALESCE(SUM(subtotal), 0) FROM reject_items WHERE reject_id = $1 AND deleted_at IS NULL),
		    updated_at = $2
		WHERE id = $1
	`
	_, err = tx.Exec(context.Background(), updateQuery, item.RejectID, time.Now())
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit(context.Background())
}

// DeleteRejectItem soft-deletes a reject item
func (r *RejectRepositoryImpl) DeleteRejectItem(id string) error {
	// Start a transaction
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}
	defer tx.Rollback(context.Background())

	// Get the reject ID for the item
	var rejectID string
	idQuery := `SELECT reject_id FROM reject_items WHERE id = $1 AND deleted_at IS NULL`
	err = tx.QueryRow(context.Background(), idQuery, id).Scan(&rejectID)
	if err != nil {
		return err
	}

	// Delete the item
	query := `UPDATE reject_items SET deleted_at = $1 WHERE id = $2`
	_, err = tx.Exec(context.Background(), query, time.Now(), id)
	if err != nil {
		return err
	}

	// Update the reject total
	updateQuery := `
		UPDATE rejects 
		SET total = (SELECT COALESCE(SUM(subtotal), 0) FROM reject_items WHERE reject_id = $1 AND deleted_at IS NULL),
		    updated_at = $2
		WHERE id = $1
	`
	_, err = tx.Exec(context.Background(), updateQuery, rejectID, time.Now())
	if err != nil {
		return err
	}

	// Commit the transaction
	return tx.Commit(context.Background())
}

// GetRejectSummary retrieves summary statistics for rejects within a date range
func (r *RejectRepositoryImpl) GetRejectSummary(startDate, endDate time.Time) (*models.RejectSummary, error) {
	query := `
		SELECT 
			COUNT(*) as total_rejects,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as total_completed_rejects,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as total_pending_rejects,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as total_cancelled_rejects,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN total ELSE 0 END), 0) as total_value,
			(SELECT COUNT(*) FROM reject_items ri JOIN rejects r ON ri.reject_id = r.id 
				WHERE r.reject_date BETWEEN $1 AND $2 AND r.status = 'completed' AND ri.deleted_at IS NULL) as total_rejected_items
		FROM rejects
		WHERE reject_date BETWEEN $1 AND $2 AND deleted_at IS NULL
	`

	var summary models.RejectSummary
	err := r.db.QueryRow(context.Background(), query, startDate, endDate).Scan(
		&summary.TotalRejects,
		&summary.TotalCompletedRejects,
		&summary.TotalPendingRejects,
		&summary.TotalCancelledRejects,
		&summary.TotalValue,
		&summary.TotalRejectedItems,
	)
	if err != nil {
		return nil, err
	}

	summary.PeriodStart = startDate.Format("2006-01-02")
	summary.PeriodEnd = endDate.Format("2006-01-02")

	return &summary, nil
}

// GetDailyReject retrieves daily reject data for a date range
func (r *RejectRepositoryImpl) GetDailyReject(startDate, endDate time.Time) ([]models.DailyReject, error) {
	query := `
		SELECT 
			TO_CHAR(reject_date, 'YYYY-MM-DD') as date,
			COUNT(*) as total_rejects,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN total ELSE 0 END), 0) as total_value,
			(SELECT COUNT(*) FROM reject_items ri JOIN rejects r ON ri.reject_id = r.id 
				WHERE DATE(r.reject_date) = DATE(ro.reject_date) AND r.status = 'completed' AND ri.deleted_at IS NULL) as item_count
		FROM rejects ro
		WHERE reject_date BETWEEN $1 AND $2 AND deleted_at IS NULL
		GROUP BY DATE(reject_date), TO_CHAR(reject_date, 'YYYY-MM-DD')
		ORDER BY DATE(reject_date)
	`

	rows, err := r.db.Query(context.Background(), query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.DailyReject
	for rows.Next() {
		var daily models.DailyReject
		err := rows.Scan(
			&daily.Date,
			&daily.TotalRejects,
			&daily.TotalValue,
			&daily.ItemCount,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, daily)
	}

	return results, nil
}
