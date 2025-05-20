package handlers

import (
	"encoding/json"
	"inventory-go/models"
	"inventory-go/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

// StockInHandler handles stock-in related operations
type StockInHandler struct {
	*BaseHandler
	stockInRepo  repositories.StockInRepository
	productRepo  repositories.ProductRepository
	supplierRepo repositories.SupplierRepository
}

// NewStockInHandler creates a new StockInHandler
func NewStockInHandler(db *pgxpool.Pool) *StockInHandler {
	return &StockInHandler{
		BaseHandler:  &BaseHandler{DB: db},
		stockInRepo:  repositories.NewStockInRepository(db),
		productRepo:  repositories.NewProductRepository(db),
		supplierRepo: repositories.NewSupplierRepository(db),
	}
}

// GetStockIns handles GET /stockins
func (h *StockInHandler) GetStockIns(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit
	status := r.URL.Query().Get("status")
	
	var supplierID *string
	if supID := r.URL.Query().Get("supplier_id"); supID != "" {
		supplierID = &supID
	}
	
	var startDate, endDate *time.Time
	if sd := r.URL.Query().Get("start_date"); sd != "" {
		t, err := time.Parse("2006-01-02", sd)
		if err == nil {
			startDate = &t
		}
	}
	
	if ed := r.URL.Query().Get("end_date"); ed != "" {
		t, err := time.Parse("2006-01-02", ed)
		if err == nil {
			// Set to end of day
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endDate = &t
		}
	}

	stockIns, total, err := h.stockInRepo.List(offset, limit, status, supplierID, startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stock-ins: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"stock_ins": stockIns,
		"total":     total,
		"page":      page,
		"limit":     limit,
		"pages":     (total + int64(limit) - 1) / int64(limit),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetStockIn handles GET /stockins/{id}
func (h *StockInHandler) GetStockIn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	stockIn, err := h.stockInRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stock-in: "+err.Error())
		return
	}

	if stockIn == nil {
		respondWithError(w, http.StatusNotFound, "Stock-in not found")
		return
	}

	respondWithJSON(w, http.StatusOK, stockIn)
}

// GetStockInByReference handles GET /stockins/reference/{reference}
func (h *StockInHandler) GetStockInByReference(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reference := vars["reference"]

	stockIn, err := h.stockInRepo.GetByReference(reference)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stock-in: "+err.Error())
		return
	}

	if stockIn == nil {
		respondWithError(w, http.StatusNotFound, "Stock-in not found")
		return
	}

	respondWithJSON(w, http.StatusOK, stockIn)
}

// CreateStockIn handles POST /stockins
func (h *StockInHandler) CreateStockIn(w http.ResponseWriter, r *http.Request) {
	var stockIn models.StockIn
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&stockIn); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Validate the stock-in
	if stockIn.ReferenceNo == "" {
		respondWithError(w, http.StatusBadRequest, "Reference number is required")
		return
	}

	// Check if supplier exists if provided
	if stockIn.SupplierID != nil {
		supplier, err := h.supplierRepo.GetByID(*stockIn.SupplierID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to verify supplier: "+err.Error())
			return
		}
		if supplier == nil {
			respondWithError(w, http.StatusBadRequest, "Supplier not found")
			return
		}
	}

	// Validate and process items
	for i, item := range stockIn.Items {
		if item.ProductID == "" {
			respondWithError(w, http.StatusBadRequest, "Product ID is required for all items")
			return
		}

		if item.Quantity <= 0 {
			respondWithError(w, http.StatusBadRequest, "Quantity must be greater than zero")
			return
		}

		// Get product details
		product, err := h.productRepo.GetByID(item.ProductID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get product: "+err.Error())
			return
		}
		if product == nil {
			respondWithError(w, http.StatusBadRequest, "Product not found: "+item.ProductID)
			return
		}

		// Set product name if not provided
		if item.ProductName == "" {
			stockIn.Items[i].ProductName = product.Basic.Name
		}

		// Calculate subtotal if not provided
		if item.Subtotal == 0 {
			stockIn.Items[i].Subtotal = item.UnitCost * float64(item.Quantity)
		}
	}

	// Create the stock-in
	if err := h.stockInRepo.Create(&stockIn); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create stock-in: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, stockIn)
}

// UpdateStockIn handles PUT /stockins/{id}
func (h *StockInHandler) UpdateStockIn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get existing stock-in
	existingStockIn, err := h.stockInRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stock-in: "+err.Error())
		return
	}
	if existingStockIn == nil {
		respondWithError(w, http.StatusNotFound, "Stock-in not found")
		return
	}

	// Parse update data
	var stockIn models.StockIn
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&stockIn); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Update fields
	stockIn.ID = id
	stockIn.Items = existingStockIn.Items // Keep existing items

	// Update the stock-in
	if err := h.stockInRepo.Update(&stockIn); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update stock-in: "+err.Error())
		return
	}

	// Get updated stock-in
	updatedStockIn, err := h.stockInRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get updated stock-in: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, updatedStockIn)
}

// DeleteStockIn handles DELETE /stockins/{id}
func (h *StockInHandler) DeleteStockIn(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if stock-in exists
	stockIn, err := h.stockInRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stock-in: "+err.Error())
		return
	}
	if stockIn == nil {
		respondWithError(w, http.StatusNotFound, "Stock-in not found")
		return
	}

	// Delete the stock-in
	if err := h.stockInRepo.Delete(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete stock-in: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Stock-in deleted successfully"})
}

// AddStockInItem handles POST /stockins/{id}/items
func (h *StockInHandler) AddStockInItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stockInID := vars["id"]

	// Check if stock-in exists
	stockIn, err := h.stockInRepo.GetByID(stockInID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stock-in: "+err.Error())
		return
	}
	if stockIn == nil {
		respondWithError(w, http.StatusNotFound, "Stock-in not found")
		return
	}

	// Parse item data
	var item models.StockInItem
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Set stock-in ID
	item.StockInID = stockInID

	// Validate item
	if item.ProductID == "" {
		respondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}
	if item.Quantity <= 0 {
		respondWithError(w, http.StatusBadRequest, "Quantity must be greater than zero")
		return
	}

	// Get product details
	product, err := h.productRepo.GetByID(item.ProductID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get product: "+err.Error())
		return
	}
	if product == nil {
		respondWithError(w, http.StatusBadRequest, "Product not found")
		return
	}

	// Set product name if not provided
	if item.ProductName == "" {
		item.ProductName = product.Basic.Name
	}

	// Calculate subtotal if not provided
	if item.Subtotal == 0 {
		item.Subtotal = item.UnitCost * float64(item.Quantity)
	}

	// Add the item
	if err := h.stockInRepo.AddStockInItem(&item); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add stock-in item: "+err.Error())
		return
	}

	// Get updated stock-in
	updatedStockIn, err := h.stockInRepo.GetByID(stockInID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get updated stock-in: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, updatedStockIn)
}

// UpdateStockInItem handles PUT /stockins/{stockInId}/items/{itemId}
func (h *StockInHandler) UpdateStockInItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stockInID := vars["stockInId"]
	itemID := vars["itemId"]

	// Parse item data
	var item models.StockInItem
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Set IDs
	item.ID = itemID
	item.StockInID = stockInID

	// Validate item
	if item.ProductID == "" {
		respondWithError(w, http.StatusBadRequest, "Product ID is required")
		return
	}
	if item.Quantity <= 0 {
		respondWithError(w, http.StatusBadRequest, "Quantity must be greater than zero")
		return
	}

	// Calculate subtotal if not provided
	if item.Subtotal == 0 {
		item.Subtotal = item.UnitCost * float64(item.Quantity)
	}

	// Update the item
	if err := h.stockInRepo.UpdateStockInItem(&item); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update stock-in item: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, item)
}

// DeleteStockInItem handles DELETE /stockins/{stockInId}/items/{itemId}
func (h *StockInHandler) DeleteStockInItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	itemID := vars["itemId"]

	// Delete the item
	if err := h.stockInRepo.DeleteStockInItem(itemID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete stock-in item: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Stock-in item deleted successfully"})
}

// GetStockInsBySupplier handles GET /suppliers/{id}/stockins
func (h *StockInHandler) GetStockInsBySupplier(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	supplierID := vars["id"]

	// Check if supplier exists
	supplier, err := h.supplierRepo.GetByID(supplierID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get supplier: "+err.Error())
		return
	}
	if supplier == nil {
		respondWithError(w, http.StatusNotFound, "Supplier not found")
		return
	}

	// Get stock-ins for the supplier
	stockIns, err := h.stockInRepo.GetStockInsBySupplier(supplierID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get supplier stock-ins: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, stockIns)
}

// GetStockInSummary handles GET /stockins/summary
func (h *StockInHandler) GetStockInSummary(w http.ResponseWriter, r *http.Request) {
	// Parse date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid start date format (use YYYY-MM-DD)")
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid end date format (use YYYY-MM-DD)")
			return
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		// Default to current time
		endDate = time.Now()
	}

	// Get summary
	summary, err := h.stockInRepo.GetStockInSummary(startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get stock-in summary: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, summary)
}

// GetDailyStockIn handles GET /stockins/daily
func (h *StockInHandler) GetDailyStockIn(w http.ResponseWriter, r *http.Request) {
	// Parse date range
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	var startDate, endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid start date format (use YYYY-MM-DD)")
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid end date format (use YYYY-MM-DD)")
			return
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	} else {
		// Default to current time
		endDate = time.Now()
	}

	// Get daily stock-in data
	dailyData, err := h.stockInRepo.GetDailyStockIn(startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get daily stock-in data: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, dailyData)
}
