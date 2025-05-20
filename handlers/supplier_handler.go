package handlers

import (
	"encoding/json"
	"inventory-go/models"
	"inventory-go/repositories"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

// SupplierHandler handles supplier-related operations
type SupplierHandler struct {
	*BaseHandler
	repo repositories.SupplierRepository
}

// NewSupplierHandler creates a new SupplierHandler
func NewSupplierHandler(db *pgx.Conn) *SupplierHandler {
	return &SupplierHandler{
		BaseHandler: &BaseHandler{DB: db},
		repo:        repositories.NewSupplierRepository(db),
	}
}

// GetAllSuppliers handles GET /suppliers
func (h *SupplierHandler) GetAllSuppliers(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.repo.GetAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get suppliers: "+err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, suppliers)
}

// GetSupplier handles GET /suppliers/{id}
func (h *SupplierHandler) GetSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	supplier, err := h.repo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get supplier: "+err.Error())
		return
	}

	if supplier == nil {
		respondWithError(w, http.StatusNotFound, "Supplier not found")
		return
	}

	respondWithJSON(w, http.StatusOK, supplier)
}

// CreateSupplier handles POST /suppliers
func (h *SupplierHandler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	var supplier models.Supplier
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&supplier); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	if err := h.repo.Create(&supplier); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create supplier: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, supplier)
}

// UpdateSupplier handles PUT /suppliers/{id}
func (h *SupplierHandler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var supplier models.Supplier
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&supplier); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Ensure ID in path matches body
	supplier.ID = id

	if err := h.repo.Update(&supplier); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update supplier: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, supplier)
}

// DeleteSupplier handles DELETE /suppliers/{id}
func (h *SupplierHandler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.repo.Delete(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete supplier: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Supplier deleted successfully"})
}

// SearchSuppliers handles GET /suppliers/search
func (h *SupplierHandler) SearchSuppliers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondWithError(w, http.StatusBadRequest, "Search query is required")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	suppliers, total, err := h.repo.Search(query, offset, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to search suppliers: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"suppliers": suppliers,
		"total":     total,
		"page":      page,
		"limit":     limit,
		"pages":     (total + int64(limit) - 1) / int64(limit),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetTopSuppliers handles GET /suppliers/top
func (h *SupplierHandler) GetTopSuppliers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 5
	}

	suppliers, err := h.repo.GetTopSuppliers(limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get top suppliers: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, suppliers)
}
