package handlers

import (
	"encoding/json"
	"errors"
	"inventory-go/models"
	"inventory-go/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

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

func (h *SaleHandler) GetSales(w http.ResponseWriter, r *http.Request) {
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

	status := r.URL.Query().Get("status")
	customerID := r.URL.Query().Get("customer_id")

	// Parse date range
	var startDate, endDate *time.Time
	if start := r.URL.Query().Get("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = &t
		}
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &t
		}
	}

	sales, total, err := h.saleRepo.List(offset, limit, status, &customerID, startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": sales,
		"pagination": map[string]interface{}{
			"total":  total,
			"page":   page,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *SaleHandler) GetSale(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sale, err := h.saleRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if sale == nil {
		respondWithError(w, http.StatusNotFound, "Sale not found")
		return
	}

	respondWithJSON(w, http.StatusOK, sale)
}

func (h *SaleHandler) GetSaleByReference(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reference := vars["reference"]

	sale, err := h.saleRepo.GetByReference(reference)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if sale == nil {
		respondWithError(w, http.StatusNotFound, "Sale not found")
		return
	}

	respondWithJSON(w, http.StatusOK, sale)
}

func (h *SaleHandler) CreateSale(w http.ResponseWriter, r *http.Request) {
	var sale models.Sale
	if err := json.NewDecoder(r.Body).Decode(&sale); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate the sale
	if err := sale.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if customer exists if customer_id is provided
	if sale.CustomerID != nil && *sale.CustomerID != "" {
		customer, err := h.customerRepo.GetByID(*sale.CustomerID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if customer == nil {
			respondWithError(w, http.StatusBadRequest, "Customer not found")
			return
		}
	}

	// Calculate totals if not provided
	if sale.Total == 0 {
		sale.CalculateTotals()
	}

	if err := h.saleRepo.Create(&sale); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload to get relationships
	createdSale, err := h.saleRepo.GetByID(sale.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving created sale")
		return
	}

	respondWithJSON(w, http.StatusCreated, createdSale)
}

func (h *SaleHandler) UpdateSale(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get existing sale
	existing, err := h.saleRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Sale not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var sale models.Sale
	if err := json.NewDecoder(r.Body).Decode(&sale); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate the sale
	if err := sale.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Check if customer exists if customer_id is being updated
	if sale.CustomerID != nil && *sale.CustomerID != "" && (existing.CustomerID == nil || *sale.CustomerID != *existing.CustomerID) {
		customer, err := h.customerRepo.GetByID(*sale.CustomerID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if customer == nil {
			respondWithError(w, http.StatusBadRequest, "Customer not found")
			return
		}
	}

	// Update fields
	existing.ReferenceNo = sale.ReferenceNo
	existing.Status = sale.Status
	existing.SaleDate = sale.SaleDate
	existing.Note = sale.Note
	existing.CustomerID = sale.CustomerID
	existing.Items = sale.Items
	existing.Payments = sale.Payments
	existing.Platform = sale.Platform

	// Recalculate totals
	existing.CalculateTotals()

	if err := h.saleRepo.Update(existing); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Reload to get relationships
	updatedSale, err := h.saleRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving updated sale")
		return
	}

	respondWithJSON(w, http.StatusOK, updatedSale)
}

func (h *SaleHandler) DeleteSale(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if sale exists
	if _, err := h.saleRepo.GetByID(id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Sale not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.saleRepo.Delete(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *SaleHandler) GetCustomerSales(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID := vars["id"]

	// Check if customer exists
	customer, err := h.customerRepo.GetByID(customerID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if customer == nil {
		respondWithError(w, http.StatusNotFound, "Customer not found")
		return
	}

	sales, err := h.saleRepo.GetSalesByCustomer(customerID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, sales)
}

func (h *SaleHandler) GetSalesSummary(w http.ResponseWriter, r *http.Request) {
	// Default to last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	// Parse date range from query params if provided
	if start := r.URL.Query().Get("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = t
		}
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = t
		}
	}

	summary, err := h.saleRepo.GetSalesSummary(startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, summary)
}

func (h *SaleHandler) GetDailySales(w http.ResponseWriter, r *http.Request) {
	// Default to last 30 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, -1, 0)

	// Parse date range from query params if provided
	if start := r.URL.Query().Get("start_date"); start != "" {
		if t, err := time.Parse("2006-01-02", start); err == nil {
			startDate = t
		}
	}
	if end := r.URL.Query().Get("end_date"); end != "" {
		if t, err := time.Parse("2006-01-02", end); err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = t
		}
	}

	dailySales, err := h.saleRepo.GetDailySales(startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, dailySales)
}
