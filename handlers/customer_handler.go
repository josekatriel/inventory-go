package handlers

import (
	"encoding/json"
	"inventory-go/models"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *CustomerHandler) GetAllCustomers(w http.ResponseWriter, r *http.Request) {
	customers, err := h.repo.GetAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, customers)
}

func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	customer, err := h.repo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if customer == nil {
		respondWithError(w, http.StatusNotFound, "Customer not found")
		return
	}

	respondWithJSON(w, http.StatusOK, customer)
}

func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if email already exists
	if customer.Email != "" {
		existing, _ := h.repo.GetByEmail(customer.Email)
		if existing != nil {
			respondWithError(w, http.StatusBadRequest, "Email already in use")
			return
		}
	}

	if err := h.repo.Create(&customer); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, customer)
}

func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get existing customer
	existing, err := h.repo.GetByID(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if existing == nil {
		respondWithError(w, http.StatusNotFound, "Customer not found")
		return
	}

	var customer models.Customer
	if err := json.NewDecoder(r.Body).Decode(&customer); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Check if email is being updated and if it's already in use
	if customer.Email != "" && customer.Email != existing.Email {
		existingWithEmail, _ := h.repo.GetByEmail(customer.Email)
		if existingWithEmail != nil && existingWithEmail.ID != id {
			respondWithError(w, http.StatusBadRequest, "Email already in use")
			return
		}
	}

	// Update fields
	existing.Name = customer.Name
	existing.Email = customer.Email
	existing.Phone = customer.Phone
	existing.Address = customer.Address
	existing.Notes = customer.Notes

	if err := h.repo.Update(existing); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, existing)
}

func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := h.repo.Delete(id); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}

func (h *CustomerHandler) SearchCustomers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	customers, total, err := h.repo.Search(query, offset, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"data": customers,
		"pagination": map[string]interface{}{
			"total":  total,
			"page":   page,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *CustomerHandler) GetTopCustomers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 10
	}

	customers, err := h.repo.GetTopCustomers(limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, customers)
}
