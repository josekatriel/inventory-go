package handlers

import (
	"encoding/json"
	"inventory-go/repositories"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// ---- utils ----
// respondWithError writes an error response as JSON
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondWithJSON writes the response as JSON
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// ---- handlers ----
// BaseHandler contains common handler dependencies
type BaseHandler struct {
	DB *pgx.Conn
}

// CategoryHandler handles category-related operations
type CategoryHandler struct {
	*BaseHandler
	repo repositories.CategoryRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(db *pgx.Conn) *CategoryHandler {
	return &CategoryHandler{
		BaseHandler: &BaseHandler{DB: db},
		repo:        repositories.NewCategoryRepository(db),
	}
}

// ProductHandler handles product-related operations
type ProductHandler struct {
	*BaseHandler
	prodRepo repositories.ProductRepository
}

// NewProductHandler creates a new ProductHandler
func NewProductHandler(db *pgx.Conn) *ProductHandler {
	return &ProductHandler{
		BaseHandler: &BaseHandler{DB: db},
		prodRepo:    repositories.NewProductRepository(db),
	}
}

// CustomerHandler handles customer-related operations
type CustomerHandler struct {
	*BaseHandler
	repo repositories.CustomerRepository
}

// NewCustomerHandler creates a new CustomerHandler
func NewCustomerHandler(db *pgx.Conn) *CustomerHandler {
	return &CustomerHandler{
		BaseHandler: &BaseHandler{DB: db},
		repo:        repositories.NewCustomerRepository(db),
	}
}

// SaleHandler handles sale-related operations
type SaleHandler struct {
	*BaseHandler
	saleRepo     repositories.SaleRepository
	customerRepo repositories.CustomerRepository
	prodRepo     repositories.ProductRepository
}

// NewSaleHandler creates a new SaleHandler
func NewSaleHandler(db *pgx.Conn) *SaleHandler {
	return &SaleHandler{
		BaseHandler:  &BaseHandler{DB: db},
		saleRepo:     repositories.NewSaleRepository(db),
		customerRepo: repositories.NewCustomerRepository(db),
		prodRepo:     repositories.NewProductRepository(db),
	}
}
