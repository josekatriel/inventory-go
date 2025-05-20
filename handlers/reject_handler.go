package handlers

import (
	"encoding/json"
	"inventory-go/models"
	"inventory-go/repositories"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

// RejectHandler handles reject-related operations
type RejectHandler struct {
	*BaseHandler
	rejectRepo  repositories.RejectRepository
	productRepo repositories.ProductRepository
}

// NewRejectHandler creates a new RejectHandler
func NewRejectHandler(db *pgx.Conn) *RejectHandler {
	return &RejectHandler{
		BaseHandler: &BaseHandler{DB: db},
		rejectRepo:  repositories.NewRejectRepository(db),
		productRepo: repositories.NewProductRepository(db),
	}
}

// GetRejects handles GET /rejects
func (h *RejectHandler) GetRejects(w http.ResponseWriter, r *http.Request) {
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

	rejects, total, err := h.rejectRepo.List(offset, limit, status, startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get rejects: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"rejects": rejects,
		"total":   total,
		"page":    page,
		"limit":   limit,
		"pages":   (total + int64(limit) - 1) / int64(limit),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// GetReject handles GET /rejects/{id}
func (h *RejectHandler) GetReject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	reject, err := h.rejectRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject: "+err.Error())
		return
	}

	if reject == nil {
		respondWithError(w, http.StatusNotFound, "Reject not found")
		return
	}

	respondWithJSON(w, http.StatusOK, reject)
}

// GetRejectByReference handles GET /rejects/reference/{reference}
func (h *RejectHandler) GetRejectByReference(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reference := vars["reference"]

	reject, err := h.rejectRepo.GetByReference(reference)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject: "+err.Error())
		return
	}

	if reject == nil {
		respondWithError(w, http.StatusNotFound, "Reject not found")
		return
	}

	respondWithJSON(w, http.StatusOK, reject)
}

// CreateReject handles POST /rejects
func (h *RejectHandler) CreateReject(w http.ResponseWriter, r *http.Request) {
	var reject models.Reject
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reject); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Validate the reject
	if reject.ReferenceNo == "" {
		respondWithError(w, http.StatusBadRequest, "Reference number is required")
		return
	}

	// Validate and process items
	for i, item := range reject.Items {
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
			reject.Items[i].ProductName = product.Basic.Name
		}

		// Calculate subtotal if not provided
		if item.Subtotal == 0 {
			reject.Items[i].Subtotal = item.UnitCost * float64(item.Quantity)
		}

		// Check stock availability if status is completed
		if reject.Status == models.RejectStatusCompleted {
			if product.Stock < item.Quantity {
				respondWithError(w, http.StatusBadRequest, "Insufficient stock for product: "+product.Basic.Name)
				return
			}
		}
	}

	// Create the reject
	if err := h.rejectRepo.Create(&reject); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create reject: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, reject)
}

// UpdateReject handles PUT /rejects/{id}
func (h *RejectHandler) UpdateReject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get existing reject
	existingReject, err := h.rejectRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject: "+err.Error())
		return
	}
	if existingReject == nil {
		respondWithError(w, http.StatusNotFound, "Reject not found")
		return
	}

	// Parse update data
	var reject models.Reject
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reject); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Check for status transition to completed
	statusTransitionToCompleted := existingReject.Status != models.RejectStatusCompleted && 
		reject.Status == models.RejectStatusCompleted

	// Check stock availability if transitioning to completed
	if statusTransitionToCompleted {
		for _, item := range existingReject.Items {
			product, err := h.productRepo.GetByID(item.ProductID)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to get product: "+err.Error())
				return
			}
			if product == nil {
				respondWithError(w, http.StatusBadRequest, "Product not found: "+item.ProductID)
				return
			}

			if product.Stock < item.Quantity {
				respondWithError(w, http.StatusBadRequest, "Insufficient stock for product: "+product.Basic.Name)
				return
			}
		}
	}

	// Update fields
	reject.ID = id
	reject.Items = existingReject.Items // Keep existing items

	// Update the reject
	if err := h.rejectRepo.Update(&reject); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update reject: "+err.Error())
		return
	}

	// Get updated reject
	updatedReject, err := h.rejectRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get updated reject: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, updatedReject)
}

// DeleteReject handles DELETE /rejects/{id}
func (h *RejectHandler) DeleteReject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Check if reject exists
	reject, err := h.rejectRepo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject: "+err.Error())
		return
	}
	if reject == nil {
		respondWithError(w, http.StatusNotFound, "Reject not found")
		return
	}

	// Don't allow deletion of completed rejects
	if reject.Status == models.RejectStatusCompleted {
		respondWithError(w, http.StatusBadRequest, "Cannot delete a completed reject")
		return
	}

	// Delete the reject
	if err := h.rejectRepo.Delete(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete reject: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Reject deleted successfully"})
}

// AddRejectItem handles POST /rejects/{id}/items
func (h *RejectHandler) AddRejectItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rejectID := vars["id"]

	// Check if reject exists
	reject, err := h.rejectRepo.GetByID(rejectID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject: "+err.Error())
		return
	}
	if reject == nil {
		respondWithError(w, http.StatusNotFound, "Reject not found")
		return
	}

	// Don't allow modification of completed or cancelled rejects
	if reject.Status == models.RejectStatusCompleted || reject.Status == models.RejectStatusCancelled {
		respondWithError(w, http.StatusBadRequest, "Cannot modify a completed or cancelled reject")
		return
	}

	// Parse item data
	var item models.RejectItem
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Set reject ID
	item.RejectID = rejectID

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
	if err := h.rejectRepo.AddRejectItem(&item); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to add reject item: "+err.Error())
		return
	}

	// Get updated reject
	updatedReject, err := h.rejectRepo.GetByID(rejectID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get updated reject: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, updatedReject)
}

// UpdateRejectItem handles PUT /rejects/{rejectId}/items/{itemId}
func (h *RejectHandler) UpdateRejectItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rejectID := vars["rejectId"]
	itemID := vars["itemId"]

	// Check if reject exists
	reject, err := h.rejectRepo.GetByID(rejectID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject: "+err.Error())
		return
	}
	if reject == nil {
		respondWithError(w, http.StatusNotFound, "Reject not found")
		return
	}

	// Don't allow modification of completed or cancelled rejects
	if reject.Status == models.RejectStatusCompleted || reject.Status == models.RejectStatusCancelled {
		respondWithError(w, http.StatusBadRequest, "Cannot modify a completed or cancelled reject")
		return
	}

	// Parse item data
	var item models.RejectItem
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&item); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
		return
	}
	defer r.Body.Close()

	// Set IDs
	item.ID = itemID
	item.RejectID = rejectID

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
	if err := h.rejectRepo.UpdateRejectItem(&item); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update reject item: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, item)
}

// DeleteRejectItem handles DELETE /rejects/{rejectId}/items/{itemId}
func (h *RejectHandler) DeleteRejectItem(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rejectID := vars["rejectId"]
	itemID := vars["itemId"]

	// Check if reject exists
	reject, err := h.rejectRepo.GetByID(rejectID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject: "+err.Error())
		return
	}
	if reject == nil {
		respondWithError(w, http.StatusNotFound, "Reject not found")
		return
	}

	// Don't allow modification of completed or cancelled rejects
	if reject.Status == models.RejectStatusCompleted || reject.Status == models.RejectStatusCancelled {
		respondWithError(w, http.StatusBadRequest, "Cannot modify a completed or cancelled reject")
		return
	}

	// Delete the item
	if err := h.rejectRepo.DeleteRejectItem(itemID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete reject item: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{"message": "Reject item deleted successfully"})
}

// GetRejectSummary handles GET /rejects/summary
func (h *RejectHandler) GetRejectSummary(w http.ResponseWriter, r *http.Request) {
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
	summary, err := h.rejectRepo.GetRejectSummary(startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get reject summary: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, summary)
}

// GetDailyReject handles GET /rejects/daily
func (h *RejectHandler) GetDailyReject(w http.ResponseWriter, r *http.Request) {
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

	// Get daily reject data
	dailyData, err := h.rejectRepo.GetDailyReject(startDate, endDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get daily reject data: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, dailyData)
}
