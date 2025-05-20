package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"inventory-go/models"
	"inventory-go/repositories"
	"inventory-go/utils"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xuri/excelize/v2"
)

// ExportImportHandler handles export and import operations for the application
type ExportImportHandler struct {
	db           *pgx.Conn
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
	customerRepo repositories.CustomerRepository
	supplierRepo repositories.SupplierRepository
}

// NewExportImportHandler creates a new export/import handler
func NewExportImportHandler(db *pgx.Conn) *ExportImportHandler {
	return &ExportImportHandler{
		db:           db,
		productRepo:  repositories.NewProductRepository(db),
		categoryRepo: repositories.NewCategoryRepository(db),
		customerRepo: repositories.NewCustomerRepository(db),
		supplierRepo: repositories.NewSupplierRepository(db),
	}
}

// ExportProducts exports products to CSV or Excel format
func (h *ExportImportHandler) ExportProducts(w http.ResponseWriter, r *http.Request) {
	// Get format from query parameters (default to CSV)
	format := r.URL.Query().Get("format")
	if format != "excel" && format != "csv" {
		format = "csv" // Default to CSV
	}

	// Get all products
	products, err := h.productRepo.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve products: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Define the fields to export
	headers := []string{"ID", "Name", "Description", "SKU", "Stock", "Reorder Level", "Price", "Currency"}

	// Convert products to a flat structure for CSV/Excel
	rows := make([][]string, len(products))
	for i, product := range products {
		// Extract values for each field
		row := make([]string, len(headers))

		// ID
		row[0] = product.ID

		// Basic fields
		row[1] = product.Basic.Name
		row[2] = product.Basic.Description
		row[3] = product.Basic.SKU

		// Stock
		row[4] = strconv.Itoa(product.Stock)

		// Reorder Level
		row[5] = strconv.Itoa(product.ReorderLevel)

		// Price and Currency
		row[6] = fmt.Sprintf("%.2f", product.Price.Price)
		row[7] = product.Price.Currency

		rows[i] = row
	}

	// Set export options
	exportFormat := utils.CSV
	if format == "excel" {
		exportFormat = utils.Excel
	}

	// Set filename
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("products-export-%s", timestamp)

	// Export data
	options := utils.ExportOptions{
		Format:    exportFormat,
		Headers:   headers,
		Data:      rows,
		FileName:  fileName,
		SheetName: "Products",
	}

	if err := utils.ExportToResponse(w, options); err != nil {
		http.Error(w, "Failed to export data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ExportCategories exports categories to CSV or Excel format
func (h *ExportImportHandler) ExportCategories(w http.ResponseWriter, r *http.Request) {
	// Get format from query parameters (default to CSV)
	format := r.URL.Query().Get("format")
	if format != "excel" && format != "csv" {
		format = "csv" // Default to CSV
	}

	// Get all categories
	categories, err := h.categoryRepo.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve categories: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Define headers for the export
	headers := []string{"ID", "Name", "Description", "Parent ID", "Slug", "Status", "Sort Order", "Image URL"}

	// Convert categories to a flat structure for CSV/Excel
	rows := make([][]string, len(categories))
	for i, category := range categories {
		// Convert parent ID to string, handling nil case
		parentID := ""
		if category.ParentID != nil {
			parentID = *category.ParentID
		}

		rows[i] = []string{
			category.ID,                      // ID
			category.Name,                    // Name
			category.Description,             // Description
			parentID,                         // Parent ID
			category.Slug,                    // Slug
			strconv.Itoa(category.Status),    // Status
			strconv.Itoa(category.SortOrder), // Sort Order
			category.ImageURL,                // Image URL
		}
	}

	// Set export options
	exportFormat := utils.CSV
	if format == "excel" {
		exportFormat = utils.Excel
	}

	// Set filename
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("categories-export-%s", timestamp)

	// Export data
	options := utils.ExportOptions{
		Format:    exportFormat,
		Headers:   headers,
		Data:      rows,
		FileName:  fileName,
		SheetName: "Categories",
	}

	if err := utils.ExportToResponse(w, options); err != nil {
		http.Error(w, "Failed to export data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ExportCustomers exports customers to CSV or Excel format
func (h *ExportImportHandler) ExportCustomers(w http.ResponseWriter, r *http.Request) {
	// Get format from query parameters (default to CSV)
	format := r.URL.Query().Get("format")
	if format != "excel" && format != "csv" {
		format = "csv" // Default to CSV
	}

	// Get all customers
	customers, err := h.customerRepo.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve customers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Define headers for the export
	headers := []string{"ID", "Name", "Email", "Phone", "Address", "Total Orders", "Total Spent", "Notes"}

	// Convert customers to a flat structure for CSV/Excel
	rows := make([][]string, len(customers))
	for i, customer := range customers {
		rows[i] = []string{
			customer.ID,                              // ID
			customer.Name,                            // Name
			customer.Email,                           // Email
			customer.Phone,                           // Phone
			customer.Address,                         // Address
			strconv.Itoa(customer.TotalOrders),       // Total Orders
			fmt.Sprintf("%.2f", customer.TotalSpent), // Total Spent
			customer.Notes,                           // Notes
		}
	}

	// Set export options
	exportFormat := utils.CSV
	if format == "excel" {
		exportFormat = utils.Excel
	}

	// Set filename
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("customers-export-%s", timestamp)

	// Export data
	options := utils.ExportOptions{
		Format:    exportFormat,
		Headers:   headers,
		Data:      rows,
		FileName:  fileName,
		SheetName: "Customers",
	}

	if err := utils.ExportToResponse(w, options); err != nil {
		http.Error(w, "Failed to export data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ExportSuppliers exports suppliers to CSV or Excel format
func (h *ExportImportHandler) ExportSuppliers(w http.ResponseWriter, r *http.Request) {
	// Get format from query parameters (default to CSV)
	format := r.URL.Query().Get("format")
	if format != "excel" && format != "csv" {
		format = "csv" // Default to CSV
	}

	// Get all suppliers
	suppliers, err := h.supplierRepo.GetAll()
	if err != nil {
		http.Error(w, "Failed to retrieve suppliers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Define headers for the export
	headers := []string{"ID", "Name", "Email", "Phone", "Address", "Contact Person", "Notes"}

	// Convert suppliers to a flat structure for CSV/Excel
	rows := make([][]string, len(suppliers))
	for i, supplier := range suppliers {
		rows[i] = []string{
			supplier.ID,            // ID
			supplier.Name,          // Name
			supplier.Email,         // Email
			supplier.Phone,         // Phone
			supplier.Address,       // Address
			supplier.ContactPerson, // Contact Person
			supplier.Notes,         // Notes
		}
	}

	// Set export options
	exportFormat := utils.CSV
	if format == "excel" {
		exportFormat = utils.Excel
	}

	// Set filename
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("suppliers-export-%s", timestamp)

	// Export data
	options := utils.ExportOptions{
		Format:    exportFormat,
		Headers:   headers,
		Data:      rows,
		FileName:  fileName,
		SheetName: "Suppliers",
	}

	if err := utils.ExportToResponse(w, options); err != nil {
		http.Error(w, "Failed to export data: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// ImportProducts imports products from CSV or Excel file
func (h *ExportImportHandler) ImportProducts(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with 32MB limit
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file extension to determine format
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".csv" && ext != ".xlsx" {
		http.Error(w, "Unsupported file format. Please upload a CSV or Excel file.", http.StatusBadRequest)
		return
	}

	// Import data from the file
	var records [][]string
	if ext == ".csv" {
		records, err = importFromCSV(file)
	} else {
		records, err = importFromExcel(file)
	}

	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate the imported data
	if len(records) < 2 { // At least header and one data row
		http.Error(w, "File contains insufficient data. Please check the file format.", http.StatusBadRequest)
		return
	}

	// Extract headers
	headers := records[0]
	expectedHeaders := []string{"ID", "Name", "Description", "SKU", "Stock", "Reorder Level", "Price", "Currency"}

	// Simplified header check (should be more robust in production)
	if len(headers) < len(expectedHeaders) {
		http.Error(w, "File has insufficient columns. Required columns: "+strings.Join(expectedHeaders, ", "), http.StatusBadRequest)
		return
	}

	// Process each row (skip the header row)
	successCount := 0
	errors := []string{}

	for i, row := range records[1:] {
		if len(row) < len(headers) {
			errors = append(errors, fmt.Sprintf("Row %d has insufficient columns", i+2))
			continue
		}

		// Extract fields (using index-based approach for simplicity)
		rowMap := make(map[string]string)
		for j, header := range headers {
			if j < len(row) {
				rowMap[header] = row[j]
			}
		}

		// Check if it's an update (ID present) or create (no ID)
		var product *models.Product
		isUpdate := rowMap["ID"] != ""

		if isUpdate {
			// Get existing product
			product, err = h.productRepo.GetByID(rowMap["ID"])
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: Product with ID %s not found", i+2, rowMap["ID"]))
				continue
			}
		} else {
			// Create new product with initialized struct
			product = models.NewProduct()
		}

		// Update basic fields
		product.Basic.Name = rowMap["Name"]
		product.Basic.Description = rowMap["Description"]
		product.Basic.SKU = rowMap["SKU"]

		// Update numeric fields
		stock, err := strconv.Atoi(rowMap["Stock"])
		if err == nil { // Only update if valid number
			product.Stock = stock
		}

		reorderLevel, err := strconv.Atoi(rowMap["Reorder Level"])
		if err == nil { // Only update if valid number
			product.ReorderLevel = reorderLevel
		}

		// Update price fields
		price, err := strconv.ParseFloat(rowMap["Price"], 64)
		if err == nil { // Only update if valid number
			product.Price.Price = price
		}
		product.Price.Currency = rowMap["Currency"]
		product.Price.LastUpdateUnix = time.Now().Unix()

		// Save the product
		if isUpdate {
			err = h.productRepo.Update(product)
		} else {
			err = h.productRepo.Create(product)
		}

		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Failed to save product: %s", i+2, err.Error()))
		} else {
			successCount++
		}
	}

	// Return response
	result := map[string]any{
		"success":  true,
		"total":    len(records) - 1, // Excluding header
		"imported": successCount,
		"errors":   errors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// ImportCategories imports categories from CSV or Excel file
func (h *ExportImportHandler) ImportCategories(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form with 32MB limit
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Failed to parse form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file from request: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Check file extension to determine format
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".csv" && ext != ".xlsx" {
		http.Error(w, "Unsupported file format. Please upload a CSV or Excel file.", http.StatusBadRequest)
		return
	}

	// Import data from the file
	var records [][]string
	if ext == ".csv" {
		records, err = importFromCSV(file)
	} else {
		records, err = importFromExcel(file)
	}

	if err != nil {
		http.Error(w, "Failed to read file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Validate the imported data
	if len(records) < 2 { // At least header and one data row
		http.Error(w, "File contains insufficient data. Please check the file format.", http.StatusBadRequest)
		return
	}

	// Extract headers
	headers := records[0]
	expectedHeaders := []string{"ID", "Name", "Description", "Parent ID", "Slug", "Status", "Sort Order", "Image URL"}

	// Simplified header check (should be more robust in production)
	if len(headers) < len(expectedHeaders) {
		http.Error(w, "File has insufficient columns. Required columns: "+strings.Join(expectedHeaders, ", "), http.StatusBadRequest)
		return
	}

	// Process each row (skip the header row)
	successCount := 0
	errors := []string{}

	for i, row := range records[1:] {
		if len(row) < len(headers) {
			errors = append(errors, fmt.Sprintf("Row %d has insufficient columns", i+2))
			continue
		}

		// Extract fields (using index-based approach for simplicity)
		rowMap := make(map[string]string)
		for j, header := range headers {
			if j < len(row) {
				rowMap[header] = row[j]
			}
		}

		// Check if it's an update (ID present) or create (no ID)
		var category *models.Category
		isUpdate := rowMap["ID"] != ""

		if isUpdate {
			// Get existing category
			category, err = h.categoryRepo.GetByID(rowMap["ID"])
			if err != nil {
				errors = append(errors, fmt.Sprintf("Row %d: Category with ID %s not found", i+2, rowMap["ID"]))
				continue
			}
		} else {
			// Create new category
			category = &models.Category{
				ID:        uuid.NewString(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
		}

		// Update category fields
		category.Name = rowMap["Name"]
		category.Description = rowMap["Description"]

		// Handle parent ID (optional)
		if rowMap["Parent ID"] != "" {
			parentID := rowMap["Parent ID"]
			category.ParentID = &parentID
		} else {
			category.ParentID = nil
		}

		// Handle slug (will be auto-generated if empty)
		category.Slug = rowMap["Slug"]

		// Update numeric fields
		status, err := strconv.Atoi(rowMap["Status"])
		if err == nil { // Only update if valid number
			category.Status = status
		}

		sortOrder, err := strconv.Atoi(rowMap["Sort Order"])
		if err == nil { // Only update if valid number
			category.SortOrder = sortOrder
		}

		// Update image URL
		category.ImageURL = rowMap["Image URL"]

		// Update timestamps
		category.UpdatedAt = time.Now()

		// Save the category
		if isUpdate {
			err = h.categoryRepo.Update(category)
		} else {
			err = h.categoryRepo.Create(category)
		}

		if err != nil {
			errors = append(errors, fmt.Sprintf("Row %d: Failed to save category: %s", i+2, err.Error()))
		} else {
			successCount++
		}
	}

	// Return response
	result := map[string]interface{}{
		"success":  true,
		"total":    len(records) - 1, // Excluding header
		"imported": successCount,
		"errors":   errors,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Helper functions for import/export

// importFromCSV imports data from a CSV file
func importFromCSV(file multipart.File) ([][]string, error) {
	// Create CSV reader
	reader := csv.NewReader(file)

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}

	return records, nil
}

// importFromExcel imports data from an Excel file
func importFromExcel(file multipart.File) ([][]string, error) {
	// Create a temporary file to store the uploaded Excel file
	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Open the Excel file
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	// Get all rows from the first sheet
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from Excel: %w", err)
	}

	return rows, nil
}
