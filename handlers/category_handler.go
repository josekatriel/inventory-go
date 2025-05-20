// category_handler.go
package handlers

import (
	"encoding/json"
	"inventory-go/models"
	"inventory-go/repositories"

	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CategoryHandler handles category-related operations
type CategoryHandler struct {
	*BaseHandler
	repo repositories.CategoryRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(db *pgxpool.Pool) *CategoryHandler {
	return &CategoryHandler{
		BaseHandler: &BaseHandler{DB: db},
		repo:        repositories.NewCategoryRepository(db),
	}
}

// GetAll returns all categories
func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.repo.GetAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch categories: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, categories)
}

// GetByIDOrSlug returns a category by ID or slug
func (h *CategoryHandler) GetCategoryByIDOrSlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idOrSlug := vars["idOrSlug"]
	var category *models.Category
	var err error

	// Check if the parameter is a UUID
	if len(idOrSlug) == 36 && strings.Contains(idOrSlug, "-") {
		category, err = h.repo.GetByID(idOrSlug)
	} else {
		category, err = h.repo.GetBySlug(idOrSlug)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch category: "+err.Error())
		return
	}

	if category == nil {
		respondWithError(w, http.StatusNotFound, "Category not found")
		return
	}

	respondWithJSON(w, http.StatusOK, category)
}

// GetByParentID returns all categories with specific parent ID
func (h *CategoryHandler) GetCategoriesByParentID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	parentID := vars["parentID"]
	var ptrParentID *string

	if parentID != "root" {
		ptrParentID = &parentID
	}

	categories, err := h.repo.GetByParentID(ptrParentID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch categories: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, categories)
}

// Create creates a new category
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&category); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	if err := h.repo.Create(&category); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create category: "+err.Error())
		return
	}

	// If this is a child category, generate breadcrumbs
	if category.ParentID != nil {
		breadcrumbs, err := h.repo.GetBreadcrumbs(category.ID)
		if err == nil {
			category.Breadcrumbs = breadcrumbs
		}
	}

	respondWithJSON(w, http.StatusCreated, category)
}

// Update updates an existing category
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if category exists
	existingCategory, err := h.repo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch category: "+err.Error())
		return
	}

	if existingCategory == nil {
		respondWithError(w, http.StatusNotFound, "Category not found")
		return
	}

	// Bind updated data
	var updatedCategory models.Category
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&updatedCategory); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Ensure ID is maintained
	updatedCategory.ID = id

	if err := h.repo.Update(&updatedCategory); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update category: "+err.Error())
		return
	}

	// If this is a child category, generate breadcrumbs
	if updatedCategory.ParentID != nil {
		breadcrumbs, err := h.repo.GetBreadcrumbs(updatedCategory.ID)
		if err == nil {
			updatedCategory.Breadcrumbs = breadcrumbs
		}
	}

	respondWithJSON(w, http.StatusOK, updatedCategory)
}

// Delete deletes a category
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if category exists
	existingCategory, err := h.repo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch category: "+err.Error())
		return
	}

	if existingCategory == nil {
		respondWithError(w, http.StatusNotFound, "Category not found")
		return
	}

	if err := h.repo.Delete(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete category: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// GetWithChildren returns a category along with its children
func (h *CategoryHandler) GetWithChildren(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	category, err := h.repo.GetWithChildren(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch category: "+err.Error())
		return
	}

	if category == nil {
		respondWithError(w, http.StatusNotFound, "Category not found")
		return
	}

	respondWithJSON(w, http.StatusOK, category)
}

// GetBreadcrumbs returns the breadcrumbs for a category
func (h *CategoryHandler) GetBreadcrumbs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	breadcrumbs, err := h.repo.GetBreadcrumbs(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch breadcrumbs: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, breadcrumbs)
}
