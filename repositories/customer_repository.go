package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"inventory-go/models"
	"time"

	"github.com/jackc/pgx/v5"
)

type CustomerRepository interface {
	// Basic CRUD
	GetAll() ([]models.Customer, error)
	GetByID(id string) (*models.Customer, error)
	GetByEmail(email string) (*models.Customer, error)
	Create(customer *models.Customer) error
	Update(customer *models.Customer) error
	Delete(id string) error

	// Additional queries
	Search(query string, offset, limit int) ([]models.Customer, int64, error)
	GetTopCustomers(limit int) ([]models.Customer, error)
}

type CustomerRepositoryImpl struct {
	db *pgx.Conn
}

func NewCustomerRepository(db *pgx.Conn) CustomerRepository {
	return &CustomerRepositoryImpl{db: db}
}

func (r *CustomerRepositoryImpl) GetAll() ([]models.Customer, error) {
	query := `SELECT id, name, email, phone, address, total_orders, total_spent, 
	          last_order_at, notes, created_at, updated_at 
	          FROM customers 
	          WHERE deleted_at IS NULL
	          ORDER BY name ASC`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		var lastOrderAt sql.NullTime
		
		err := rows.Scan(
			&customer.ID, &customer.Name, &customer.Email, &customer.Phone,
			&customer.Address, &customer.TotalOrders, &customer.TotalSpent,
			&lastOrderAt, &customer.Notes, &customer.CreatedAt, &customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		
		if lastOrderAt.Valid {
			customer.LastOrderAt = lastOrderAt.Time
		}
		
		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}

func (r *CustomerRepositoryImpl) GetByID(id string) (*models.Customer, error) {
	query := `SELECT id, name, email, phone, address, total_orders, total_spent, 
	          last_order_at, notes, created_at, updated_at 
	          FROM customers 
	          WHERE id = $1 AND deleted_at IS NULL`

	var customer models.Customer
	var lastOrderAt sql.NullTime

	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&customer.ID, &customer.Name, &customer.Email, &customer.Phone,
		&customer.Address, &customer.TotalOrders, &customer.TotalSpent,
		&lastOrderAt, &customer.Notes, &customer.CreatedAt, &customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if lastOrderAt.Valid {
		customer.LastOrderAt = lastOrderAt.Time
	}

	return &customer, nil
}

func (r *CustomerRepositoryImpl) GetByEmail(email string) (*models.Customer, error) {
	query := `SELECT id, name, email, phone, address, total_orders, total_spent, 
	          last_order_at, notes, created_at, updated_at 
	          FROM customers 
	          WHERE email = $1 AND deleted_at IS NULL`

	var customer models.Customer
	var lastOrderAt sql.NullTime

	err := r.db.QueryRow(context.Background(), query, email).Scan(
		&customer.ID, &customer.Name, &customer.Email, &customer.Phone,
		&customer.Address, &customer.TotalOrders, &customer.TotalSpent,
		&lastOrderAt, &customer.Notes, &customer.CreatedAt, &customer.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	if lastOrderAt.Valid {
		customer.LastOrderAt = lastOrderAt.Time
	}

	return &customer, nil
}

func (r *CustomerRepositoryImpl) Create(customer *models.Customer) error {
	query := `INSERT INTO customers (id, name, email, phone, address, total_orders, total_spent, 
	                   last_order_at, notes, created_at, updated_at) 
	                   VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(context.Background(), query,
		customer.ID, customer.Name, customer.Email, customer.Phone,
		customer.Address, customer.TotalOrders, customer.TotalSpent,
		customer.LastOrderAt, customer.Notes, customer.CreatedAt, customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *CustomerRepositoryImpl) Update(customer *models.Customer) error {
	customer.UpdatedAt = time.Now().UTC()

	query := `UPDATE customers SET
		name = $1, email = $2, phone = $3, address = $4, 
		total_orders = $5, total_spent = $6, last_order_at = $7, 
		notes = $8, updated_at = $9
		WHERE id = $10`

	_, err := r.db.Exec(context.Background(), query,
		customer.Name, customer.Email, customer.Phone, customer.Address,
		customer.TotalOrders, customer.TotalSpent, customer.LastOrderAt, 
		customer.Notes, customer.UpdatedAt, customer.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

func (r *CustomerRepositoryImpl) Delete(id string) error {
	// Soft delete
	query := `UPDATE customers SET deleted_at = $1 WHERE id = $2`
	result, err := r.db.Exec(context.Background(), query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}
	
	if result.RowsAffected() == 0 {
		return fmt.Errorf("customer not found")
	}
	
	return nil
}

func (r *CustomerRepositoryImpl) Search(query string, offset, limit int) ([]models.Customer, int64, error) {
	searchTerm := "%" + query + "%"
	
	// Count total matching records
	var count int64
	countQuery := `SELECT COUNT(*) FROM customers 
		WHERE deleted_at IS NULL AND 
		(name ILIKE $1 OR email ILIKE $2 OR phone ILIKE $3)`
	
	err := r.db.QueryRow(context.Background(), countQuery, 
		searchTerm, searchTerm, searchTerm).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Get paginated results
	searchQuery := `SELECT id, name, email, phone, address, total_orders, total_spent, 
		last_order_at, notes, created_at, updated_at 
		FROM customers 
		WHERE deleted_at IS NULL AND 
		(name ILIKE $1 OR email ILIKE $2 OR phone ILIKE $3)
		ORDER BY name ASC
		LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(context.Background(), searchQuery, 
		searchTerm, searchTerm, searchTerm, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		var lastOrderAt sql.NullTime
		
		err := rows.Scan(
			&customer.ID, &customer.Name, &customer.Email, &customer.Phone,
			&customer.Address, &customer.TotalOrders, &customer.TotalSpent,
			&lastOrderAt, &customer.Notes, &customer.CreatedAt, &customer.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		
		if lastOrderAt.Valid {
			customer.LastOrderAt = lastOrderAt.Time
		}
		
		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, count, nil
}

func (r *CustomerRepositoryImpl) GetTopCustomers(limit int) ([]models.Customer, error) {
	query := `SELECT id, name, email, phone, address, total_orders, total_spent, 
	          last_order_at, notes, created_at, updated_at 
	          FROM customers 
	          WHERE deleted_at IS NULL
	          ORDER BY total_spent DESC
	          LIMIT $1`

	rows, err := r.db.Query(context.Background(), query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query top customers: %w", err)
	}
	defer rows.Close()

	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		var lastOrderAt sql.NullTime
		
		err := rows.Scan(
			&customer.ID, &customer.Name, &customer.Email, &customer.Phone,
			&customer.Address, &customer.TotalOrders, &customer.TotalSpent,
			&lastOrderAt, &customer.Notes, &customer.CreatedAt, &customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		
		if lastOrderAt.Valid {
			customer.LastOrderAt = lastOrderAt.Time
		}
		
		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}
