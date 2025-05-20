// category_repository.go
package repositories

import (
	"context"
	"fmt"
	"inventory-go/models"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CategoryRepository interface {
	GetAll() ([]models.Category, error)
	GetByID(id string) (*models.Category, error)
	GetBySlug(slug string) (*models.Category, error)
	GetByParentID(parentID *string) ([]models.Category, error)
	Create(category *models.Category) error
	Update(category *models.Category) error
	Delete(id string) error
	GetWithChildren(id string) (*models.Category, error)
	GetBreadcrumbs(id string) ([]models.Breadcrumb, error)
}

type CategoryRepositoryImpl struct {
	*BaseRepository
}

func NewCategoryRepository(db *pgx.Conn) CategoryRepository {
	return &CategoryRepositoryImpl{
		BaseRepository: NewBaseRepository(db),
	}
}

// GetAll retrieves all categories
func (r *CategoryRepositoryImpl) GetAll() ([]models.Category, error) {
	var categories []models.Category
	query := `SELECT id, name, slug, description, parent_id, image_url, status, sort_order, 
	          created_at, updated_at 
	          FROM categories 
	          WHERE deleted_at IS NULL
	          ORDER BY sort_order ASC, name ASC`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Slug, &category.Description,
			&category.ParentID, &category.ImageURL, &category.Status, &category.SortOrder,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// GetByID retrieves a category by its ID
func (r *CategoryRepositoryImpl) GetByID(id string) (*models.Category, error) {
	var category models.Category
	var deletedAt *time.Time

	query := `SELECT id, name, slug, description, parent_id, image_url, status, sort_order, 
	          created_at, updated_at, deleted_at 
	          FROM categories 
	          WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.ParentID, &category.ImageURL, &category.Status, &category.SortOrder,
		&category.CreatedAt, &category.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

// GetBySlug retrieves a category by its slug
func (r *CategoryRepositoryImpl) GetBySlug(slug string) (*models.Category, error) {
	var category models.Category
	var deletedAt *time.Time

	query := `SELECT id, name, slug, description, parent_id, image_url, status, sort_order, 
	          created_at, updated_at, deleted_at 
	          FROM categories 
	          WHERE slug = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(context.Background(), query, slug).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.ParentID, &category.ImageURL, &category.Status, &category.SortOrder,
		&category.CreatedAt, &category.UpdatedAt, &deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &category, nil
}

// GetByParentID retrieves categories by their parent ID
func (r *CategoryRepositoryImpl) GetByParentID(parentID *string) ([]models.Category, error) {
	var categories []models.Category
	var rows pgx.Rows
	var err error

	if parentID == nil {
		query := `SELECT id, name, slug, description, parent_id, image_url, status, sort_order, 
		          created_at, updated_at 
		          FROM categories 
		          WHERE parent_id IS NULL AND deleted_at IS NULL`
		rows, err = r.db.Query(context.Background(), query)
	} else {
		query := `SELECT id, name, slug, description, parent_id, image_url, status, sort_order, 
		          created_at, updated_at 
		          FROM categories 
		          WHERE parent_id = $1 AND deleted_at IS NULL`
		rows, err = r.db.Query(context.Background(), query, *parentID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Slug, &category.Description,
			&category.ParentID, &category.ImageURL, &category.Status, &category.SortOrder,
			&category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// Create creates a new category
func (r *CategoryRepositoryImpl) Create(category *models.Category) error {
	// Generate UUID if not provided
	if category.ID == "" {
		category.ID = uuid.New().String()
	}

	// Generate slug from name if not provided
	category.GenerateSlug()

	// Check if slug already exists
	existing, err := r.GetBySlug(category.Slug)
	if err != nil {
		return fmt.Errorf("error checking for duplicate slug: %w", err)
	}
	if existing != nil {
		// Append a unique identifier to make the slug unique
		category.Slug = fmt.Sprintf("%s-%s", category.Slug, category.ID[:8])
	}

	now := time.Now().UTC()
	category.CreatedAt = now
	category.UpdatedAt = now

	query := `INSERT INTO categories (id, name, slug, description, parent_id, image_url, status, sort_order, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	          RETURNING id, created_at, updated_at`

	return r.db.QueryRow(context.Background(), query,
		category.ID, category.Name, category.Slug, category.Description,
		category.ParentID, category.ImageURL, category.Status, category.SortOrder,
		category.CreatedAt, category.UpdatedAt,
	).Scan(&category.ID, &category.CreatedAt, &category.UpdatedAt)
}

// Update updates an existing category
func (r *CategoryRepositoryImpl) Update(category *models.Category) error {
	category.UpdatedAt = time.Now().UTC()
	
	// Get existing category to check if name changed
	existing, err := r.GetByID(category.ID)
	if err != nil {
		return fmt.Errorf("error retrieving existing category: %w", err)
	}
	if existing == nil {
		return fmt.Errorf("category not found")
	}
	
	// Only regenerate slug if name changed or slug is empty
	if existing.Name != category.Name || category.Slug == "" {
		category.GenerateSlug()
		
		// Check if new slug already exists (excluding this category)
		duplicate, err := r.GetBySlug(category.Slug)
		if err != nil {
			return fmt.Errorf("error checking for duplicate slug: %w", err)
		}
		if duplicate != nil && duplicate.ID != category.ID {
			// Append a unique identifier to make the slug unique
			category.Slug = fmt.Sprintf("%s-%s", category.Slug, category.ID[:8])
		}
	}

	query := `UPDATE categories 
	          SET name = $1, slug = $2, description = $3, parent_id = $4, 
	              image_url = $5, status = $6, sort_order = $7, updated_at = $8
	          WHERE id = $9
	          RETURNING updated_at`

	return r.db.QueryRow(context.Background(), query,
		category.Name, category.Slug, category.Description,
		category.ParentID, category.ImageURL, category.Status, category.SortOrder,
		category.UpdatedAt, category.ID,
	).Scan(&category.UpdatedAt)
}

// Delete deletes a category by its ID
func (r *CategoryRepositoryImpl) Delete(id string) error {
	query := `UPDATE categories SET deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(context.Background(), query, id)
	return err
}

// GetWithChildren retrieves a category by its ID and its children
func (r *CategoryRepositoryImpl) GetWithChildren(id string) (*models.Category, error) {
	// Get the parent category
	var category models.Category
	query := `SELECT id, name, slug, description, parent_id, image_url, status, sort_order, 
	          created_at, updated_at 
	          FROM categories 
	          WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&category.ID, &category.Name, &category.Slug, &category.Description,
		&category.ParentID, &category.ImageURL, &category.Status, &category.SortOrder,
		&category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get children
	children, err := r.GetByParentID(&category.ID)
	if err != nil {
		return nil, err
	}

	// Append all children at once using the spread operator
	category.Children = append(category.Children, children...)

	return &category, nil
}

// GetBreadcrumbs retrieves the breadcrumbs for a category by its ID
func (r *CategoryRepositoryImpl) GetBreadcrumbs(id string) ([]models.Breadcrumb, error) {
	var breadcrumbs []models.Breadcrumb
	currentID := id
	level := 0

	for currentID != "" {
		var category models.Category
		query := `SELECT id, name, slug, parent_id FROM categories WHERE id = $1 AND deleted_at IS NULL`

		err := r.db.QueryRow(context.Background(), query, currentID).Scan(
			&category.ID, &category.Name, &category.Slug, &category.ParentID,
		)

		if err != nil {
			if err == pgx.ErrNoRows {
				break
			}
			return nil, fmt.Errorf("error getting category %s: %w", currentID, err)
		}

		breadcrumbs = append([]models.Breadcrumb{{
			ID:    category.ID,
			Name:  category.Name,
			Slug:  category.Slug,
			Level: level,
		}}, breadcrumbs...)

		if category.ParentID == nil {
			break
		}
		currentID = *category.ParentID
		level++
	}

	// Reverse the breadcrumbs to have them from root to current
	for i, j := 0, len(breadcrumbs)-1; i < j; i, j = i+1, j-1 {
		breadcrumbs[i], breadcrumbs[j] = breadcrumbs[j], breadcrumbs[i]
	}

	// Update levels after reversal
	for i := range breadcrumbs {
		breadcrumbs[i].Level = i
	}

	return breadcrumbs, nil
}
