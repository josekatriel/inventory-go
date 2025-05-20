package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type Category struct {
	ID          string    `json:"id" db:"id"`
	ParentID    *string   `json:"parent_id,omitempty" db:"parent_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description,omitempty" db:"description"`
	Slug        string    `json:"slug" db:"slug"`
	Status      int       `json:"status" db:"status"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	ImageURL    string    `json:"image_url,omitempty" db:"image_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"-" db:"deleted_at"`

	// Relations
	Parent   *Category  `json:"-" db:"-"`
	Children []Category `json:"children,omitempty" db:"-"`

	// For API responses
	Breadcrumbs []Breadcrumb `json:"breadcrumbs,omitempty" db:"-"`
}

type Breadcrumb struct {
	ID    string `json:"id" db:"id"`
	Name  string `json:"name" db:"name"`
	Slug  string `json:"slug" db:"slug"`
	Level int    `json:"level" db:"level"`
}

// GenerateID sets a UUID if ID is empty
func (c *Category) GenerateID() {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
}

// GenerateSlug creates a slug from the category name if one doesn't exist
func (c *Category) GenerateSlug() {
	if c.Slug == "" && c.Name != "" {
		c.Slug = slug.Make(c.Name)
	}
}

// PopulateBreadcrumbs generates breadcrumbs using the provided data
// This should be called by the repository after retrieving category data
func (c *Category) PopulateBreadcrumbs(breadcrumbs []Breadcrumb) {
	c.Breadcrumbs = breadcrumbs
}

// Note: BuildCategoryBreadcrumbs has been moved to the category repository
// This is a placeholder to avoid breaking interfaces
func BuildCategoryBreadcrumbs(categoryID string, categories map[string]Category) []Breadcrumb {
	var breadcrumbs []Breadcrumb
	var currentID = categoryID
	level := 0
	
	// Build the breadcrumbs by traversing up the category tree
	for {
		category, exists := categories[currentID]
		if !exists {
			break
		}

		// Add to breadcrumbs
		breadcrumbs = append([]Breadcrumb{{
			ID:    category.ID,
			Name:  category.Name,
			Slug:  category.Slug,
			Level: level,
		}}, breadcrumbs...)

		// Break if this is a root category
		if category.ParentID == nil {
			break
		}
		
		// Move to parent
		currentID = *category.ParentID
		level++
	}

	// Reverse breadcrumbs order to have root first
	for i, j := 0, len(breadcrumbs)-1; i < j; i, j = i+1, j-1 {
		breadcrumbs[i], breadcrumbs[j] = breadcrumbs[j], breadcrumbs[i]
	}

	// Update levels after reversal
	for i := range breadcrumbs {
		breadcrumbs[i].Level = i
	}

	return breadcrumbs
}
