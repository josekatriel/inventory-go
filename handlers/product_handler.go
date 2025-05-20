// product_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"inventory-go/models"
	"inventory-go/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ProductHandler handles product-related operations
type ProductHandler struct {
	*BaseHandler
	prodRepo repositories.ProductRepository
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(db *pgxpool.Pool) *ProductHandler {
	return &ProductHandler{
		BaseHandler: &BaseHandler{DB: db},
		prodRepo:    repositories.NewProductRepository(db),
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Set default values
	product.Basic.Status = 1 // Active by default
	product.Price.Currency = "IDR"
	product.Price.LastUpdateUnix = time.Now().Unix()
	product.Weight.Unit = 1 // gram

	if err := h.prodRepo.Create(&product); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload to get relationships
	createdProduct, err := h.prodRepo.GetByID(product.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving created product")
		return
	}

	respondWithJSON(w, http.StatusCreated, createdProduct)
}

// GetProduct retrieves a single product by ID
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		respondWithError(w, http.StatusBadRequest, "Invalid product ID")
		return
	}

	// Get product by ID
	product, err := h.prodRepo.GetByID(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Product not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get existing product
	existing, err := h.prodRepo.GetByID(id)
	if err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Product not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var product models.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Preserve the ID and timestamps
	product.ID = existing.ID
	product.CreatedAt = existing.CreatedAt
	product.UpdatedAt = time.Now()

	if err := h.prodRepo.Update(&product); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get updated product
	updated, err := h.prodRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving updated product")
		return
	}

	respondWithJSON(w, http.StatusOK, updated)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if product exists
	if _, err := h.prodRepo.GetByID(id); err != nil {
		if err == pgx.ErrNoRows {
			respondWithError(w, http.StatusNotFound, "Product not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.prodRepo.Delete(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	var status *int
	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = &s
		}
	}

	var (
		productPtrs []*models.Product
		total       int64
		err         error
	)

	// Convert status from *int to string if not nil
	statusStr := ""
	if status != nil {
		statusStr = fmt.Sprintf("%d", *status)
	}

	// Check if category ID is provided
	if categoryID := r.URL.Query().Get("category_id"); categoryID != "" {
		productPtrs, total, err = h.prodRepo.List(offset, limit, statusStr)
	} else {
		productPtrs, total, err = h.prodRepo.List(offset, limit, statusStr)
	}

	// Convert []*models.Product to []models.Product for compatibility
	products := make([]models.Product, 0, len(productPtrs))
	for _, p := range productPtrs {
		if p != nil {
			products = append(products, *p)
		}
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]any{
		"data": products,
		"pagination": map[string]any{
			"total":  total,
			"page":   page,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// GetLowStockProducts retrieves products where stock is at or below the reorder level
func (h *ProductHandler) GetLowStockProducts(w http.ResponseWriter, r *http.Request) {
	// Get all products with low stock
	productPtrs, err := h.prodRepo.GetLowStockProducts()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Convert []*models.Product to []models.Product for consistency
	products := make([]models.Product, 0, len(productPtrs))
	for _, p := range productPtrs {
		if p != nil {
			products = append(products, *p)
		}
	}

	// Add some helpful metadata to the response
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data":    products,
		"total":   len(products),
		"message": "Products below their reorder threshold. These items need to be restocked.",
	})
}
