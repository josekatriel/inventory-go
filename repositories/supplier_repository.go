package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"inventory-go/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SupplierRepository interface {
	// Basic CRUD
	GetAll() ([]models.Supplier, error)
	GetByID(id string) (*models.Supplier, error)
	GetByEmail(email string) (*models.Supplier, error)
	Create(supplier *models.Supplier) error
	Update(supplier *models.Supplier) error
	Delete(id string) error
	
	// Additional queries
	Search(query string, offset, limit int) ([]models.Supplier, int64, error)
	GetTopSuppliers(limit int) ([]models.Supplier, error)
}

type SupplierRepositoryImpl struct {
	db *pgx.Conn
}

func NewSupplierRepository(db *pgx.Conn) SupplierRepository {
	return &SupplierRepositoryImpl{db: db}
}

func (r *SupplierRepositoryImpl) GetAll() ([]models.Supplier, error) {
	query := `SELECT id, name, email, phone, address, contact_person, total_purchases, 
	          total_spent, last_order_at, notes, created_at, updated_at 
	          FROM suppliers 
	          WHERE deleted_at IS NULL
	          ORDER BY name ASC`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		var lastOrderAt sql.NullTime
		
		err := rows.Scan(
			&supplier.ID, &supplier.Name, &supplier.Email, &supplier.Phone,
			&supplier.Address, &supplier.ContactPerson, &supplier.TotalPurchases, &supplier.TotalSpent,
			&lastOrderAt, &supplier.Notes, &supplier.CreatedAt, &supplier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		
		if lastOrderAt.Valid {
			supplier.LastOrderAt = lastOrderAt.Time
		}
		
		suppliers = append(suppliers, supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating suppliers: %w", err)
	}

	return suppliers, nil
}

func (r *SupplierRepositoryImpl) GetByID(id string) (*models.Supplier, error) {
	query := `SELECT id, name, email, phone, address, contact_person, total_purchases, 
	          total_spent, last_order_at, notes, created_at, updated_at 
	          FROM suppliers 
	          WHERE id = $1 AND deleted_at IS NULL`

	var supplier models.Supplier
	var lastOrderAt sql.NullTime

	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&supplier.ID, &supplier.Name, &supplier.Email, &supplier.Phone,
		&supplier.Address, &supplier.ContactPerson, &supplier.TotalPurchases, &supplier.TotalSpent,
		&lastOrderAt, &supplier.Notes, &supplier.CreatedAt, &supplier.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	if lastOrderAt.Valid {
		supplier.LastOrderAt = lastOrderAt.Time
	}

	return &supplier, nil
}

func (r *SupplierRepositoryImpl) GetByEmail(email string) (*models.Supplier, error) {
	query := `SELECT id, name, email, phone, address, contact_person, total_purchases, 
	          total_spent, last_order_at, notes, created_at, updated_at 
	          FROM suppliers 
	          WHERE email = $1 AND deleted_at IS NULL`

	var supplier models.Supplier
	var lastOrderAt sql.NullTime

	err := r.db.QueryRow(context.Background(), query, email).Scan(
		&supplier.ID, &supplier.Name, &supplier.Email, &supplier.Phone,
		&supplier.Address, &supplier.ContactPerson, &supplier.TotalPurchases, &supplier.TotalSpent,
		&lastOrderAt, &supplier.Notes, &supplier.CreatedAt, &supplier.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get supplier by email: %w", err)
	}

	if lastOrderAt.Valid {
		supplier.LastOrderAt = lastOrderAt.Time
	}

	return &supplier, nil
}

func (r *SupplierRepositoryImpl) Create(supplier *models.Supplier) error {
	if supplier.ID == "" {
		supplier.ID = uuid.NewString()
	}

	now := time.Now().UTC()
	supplier.CreatedAt = now
	supplier.UpdatedAt = now

	query := `INSERT INTO suppliers (
		id, name, email, phone, address, contact_person, 
		total_purchases, total_spent, last_order_at, notes, 
		created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := r.db.Exec(context.Background(), query,
		supplier.ID, supplier.Name, supplier.Email, supplier.Phone,
		supplier.Address, supplier.ContactPerson, supplier.TotalPurchases,
		supplier.TotalSpent, supplier.LastOrderAt, supplier.Notes,
		supplier.CreatedAt, supplier.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	return nil
}

func (r *SupplierRepositoryImpl) Update(supplier *models.Supplier) error {
	supplier.UpdatedAt = time.Now().UTC()

	query := `UPDATE suppliers SET
		name = $1, email = $2, phone = $3, address = $4, 
		contact_person = $5, total_purchases = $6, total_spent = $7, 
		last_order_at = $8, notes = $9, updated_at = $10
		WHERE id = $11`

	_, err := r.db.Exec(context.Background(), query,
		supplier.Name, supplier.Email, supplier.Phone, supplier.Address,
		supplier.ContactPerson, supplier.TotalPurchases, supplier.TotalSpent, 
		supplier.LastOrderAt, supplier.Notes, supplier.UpdatedAt, supplier.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update supplier: %w", err)
	}

	return nil
}

func (r *SupplierRepositoryImpl) Delete(id string) error {
	// Soft delete
	query := `UPDATE suppliers SET deleted_at = $1 WHERE id = $2`
	result, err := r.db.Exec(context.Background(), query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("supplier not found")
	}

	return nil
}

func (r *SupplierRepositoryImpl) Search(query string, offset, limit int) ([]models.Supplier, int64, error) {
	searchTerm := "%" + query + "%"
	
	// Count total matching records
	var count int64
	countQuery := `SELECT COUNT(*) FROM suppliers 
		WHERE deleted_at IS NULL AND 
		(name ILIKE $1 OR email ILIKE $2 OR phone ILIKE $3 OR contact_person ILIKE $4)`
	
	err := r.db.QueryRow(context.Background(), countQuery, 
		searchTerm, searchTerm, searchTerm, searchTerm).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count suppliers: %w", err)
	}

	// Get paginated results
	searchQuery := `SELECT id, name, email, phone, address, contact_person, total_purchases, 
		total_spent, last_order_at, notes, created_at, updated_at 
		FROM suppliers 
		WHERE deleted_at IS NULL AND 
		(name ILIKE $1 OR email ILIKE $2 OR phone ILIKE $3 OR contact_person ILIKE $4)
		ORDER BY name ASC
		LIMIT $5 OFFSET $6`

	rows, err := r.db.Query(context.Background(), searchQuery, 
		searchTerm, searchTerm, searchTerm, searchTerm, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		var lastOrderAt sql.NullTime
		
		err := rows.Scan(
			&supplier.ID, &supplier.Name, &supplier.Email, &supplier.Phone,
			&supplier.Address, &supplier.ContactPerson, &supplier.TotalPurchases, &supplier.TotalSpent,
			&lastOrderAt, &supplier.Notes, &supplier.CreatedAt, &supplier.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan supplier: %w", err)
		}
		
		if lastOrderAt.Valid {
			supplier.LastOrderAt = lastOrderAt.Time
		}
		
		suppliers = append(suppliers, supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating suppliers: %w", err)
	}

	return suppliers, count, nil
}

func (r *SupplierRepositoryImpl) GetTopSuppliers(limit int) ([]models.Supplier, error) {
	query := `SELECT id, name, email, phone, address, contact_person, total_purchases, 
	          total_spent, last_order_at, notes, created_at, updated_at 
	          FROM suppliers 
	          WHERE deleted_at IS NULL
	          ORDER BY total_spent DESC
	          LIMIT $1`

	rows, err := r.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []models.Supplier
	for rows.Next() {
		var supplier models.Supplier
		var lastOrderAt sql.NullTime
		
		err := rows.Scan(
			&supplier.ID, &supplier.Name, &supplier.Email, &supplier.Phone,
			&supplier.Address, &supplier.ContactPerson, &supplier.TotalPurchases, &supplier.TotalSpent,
			&lastOrderAt, &supplier.Notes, &supplier.CreatedAt, &supplier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		
		if lastOrderAt.Valid {
			supplier.LastOrderAt = lastOrderAt.Time
		}
		
		suppliers = append(suppliers, supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating suppliers: %w", err)
	}

	return suppliers, nil
}
