package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID          string    `gorm:"type:char(36);primaryKey" json:"id"`
	ParentID    *string   `gorm:"type:char(36);index" json:"parent_id,omitempty"`
	Name        string    `json:"name" gorm:"not null;index"`
	Description string    `json:"description,omitempty"`
	Slug        string    `json:"slug" gorm:"uniqueIndex"`
	Status      int       `json:"status" gorm:"default:1"`
	SortOrder   int       `json:"sort_order" gorm:"default:0"`
	ImageURL    string    `json:"image_url,omitempty"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Parent   *Category  `json:"-" gorm:"foreignKey:ParentID"`
	Children []Category `json:"children,omitempty" gorm:"foreignKey:ParentID"`

	// For API responses
	Breadcrumbs []Breadcrumb `gorm:"-" json:"breadcrumbs,omitempty"`
}

type Breadcrumb struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Level int    `json:"level"`
}

// BeforeCreate will set a UUID
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	return nil
}

// AfterFind hook to populate breadcrumbs
func (c *Category) AfterFind(tx *gorm.DB) (err error) {
	if c.ParentID != nil {
		breadcrumbs, err := BuildCategoryBreadcrumbs(tx, c)
		if err != nil {
			return err
		}
		c.Breadcrumbs = breadcrumbs
	}
	return nil
}

func BuildCategoryBreadcrumbs(db *gorm.DB, category *Category) ([]Breadcrumb, error) {
	var breadcrumbs []Breadcrumb
	current := category

	// First, count the total depth
	depth := 0
	temp := current
	for temp != nil {
		depth++
		if temp.Parent != nil {
			temp = temp.Parent
		} else if temp.ParentID != nil {
			var parent Category
			if err := db.First(&parent, "id = ?", *temp.ParentID).Error; err != nil {
				return nil, err
			}
			temp = &parent
		} else {
			break
		}
	}

	// Now build breadcrumbs with correct levels
	level := 0
	for current != nil {
		breadcrumbs = append(breadcrumbs, Breadcrumb{
			ID:    current.ID,
			Name:  current.Name,
			Slug:  current.Slug,
			Level: depth - level - 1, // Calculate level from root
		})

		if current.Parent != nil {
			current = current.Parent
		} else if current.ParentID != nil {
			var parent Category
			if err := db.First(&parent, "id = ?", *current.ParentID).Error; err != nil {
				return breadcrumbs, err
			}
			current = &parent
		} else {
			current = nil
		}
		level++
	}

	return breadcrumbs, nil
}
