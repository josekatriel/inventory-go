package routes

import (
	"inventory-go/handlers"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

func SetupRoutes(r *mux.Router, db *pgx.Conn) {
	// Initialize handlers with database connection
	productHandler := handlers.NewProductHandler(db)
	categoryHandler := handlers.NewCategoryHandler(db)
	customerHandler := handlers.NewCustomerHandler(db)
	saleHandler := handlers.NewSaleHandler(db)
	supplierHandler := handlers.NewSupplierHandler(db)
	stockInHandler := handlers.NewStockInHandler(db)
	rejectHandler := handlers.NewRejectHandler(db)
	exportImportHandler := handlers.NewExportImportHandler(db)

	// Product routes
	r.HandleFunc("/api/products/low-stock", productHandler.GetLowStockProducts).Methods("GET") // Specific route first
	r.HandleFunc("/api/products", productHandler.CreateProduct).Methods("POST")
	r.HandleFunc("/api/products", productHandler.GetAllProducts).Methods("GET")
	r.HandleFunc("/api/products/{id}", productHandler.GetProduct).Methods("GET")
	r.HandleFunc("/api/products/{id}", productHandler.UpdateProduct).Methods("PUT")
	r.HandleFunc("/api/products/{id}", productHandler.DeleteProduct).Methods("DELETE")

	// Category routes
	r.HandleFunc("/api/categories", categoryHandler.CreateCategory).Methods("POST")
	r.HandleFunc("/api/categories", categoryHandler.GetAllCategories).Methods("GET")
	r.HandleFunc("/api/categories/{idOrSlug}", categoryHandler.GetCategoryByIDOrSlug).Methods("GET")
	r.HandleFunc("/api/categories/{id}/children", categoryHandler.GetWithChildren).Methods("GET")
	r.HandleFunc("/api/categories/parent/{parentID}", categoryHandler.GetCategoriesByParentID).Methods("GET")
	r.HandleFunc("/api/categories/{id}/breadcrumbs", categoryHandler.GetBreadcrumbs).Methods("GET")
	r.HandleFunc("/api/categories/{id}", categoryHandler.UpdateCategory).Methods("PUT")
	r.HandleFunc("/api/categories/{id}", categoryHandler.DeleteCategory).Methods("DELETE")

	// Export/Import routes
	r.HandleFunc("/api/export/products", exportImportHandler.ExportProducts).Methods("GET")
	r.HandleFunc("/api/export/categories", exportImportHandler.ExportCategories).Methods("GET")
	r.HandleFunc("/api/export/customers", exportImportHandler.ExportCustomers).Methods("GET")
	r.HandleFunc("/api/export/suppliers", exportImportHandler.ExportSuppliers).Methods("GET")
	r.HandleFunc("/api/import/products", exportImportHandler.ImportProducts).Methods("POST")
	r.HandleFunc("/api/import/categories", exportImportHandler.ImportCategories).Methods("POST")

	// Customer routes
	r.HandleFunc("/api/customers", customerHandler.GetAllCustomers).Methods("GET")
	r.HandleFunc("/api/customers/{id}", customerHandler.GetCustomer).Methods("GET")
	r.HandleFunc("/api/customers", customerHandler.CreateCustomer).Methods("POST")
	r.HandleFunc("/api/customers/{id}", customerHandler.UpdateCustomer).Methods("PUT")
	r.HandleFunc("/api/customers/{id}", customerHandler.DeleteCustomer).Methods("DELETE")
	r.HandleFunc("/api/customers/search", customerHandler.SearchCustomers).Methods("GET")
	r.HandleFunc("/api/customers/top", customerHandler.GetTopCustomers).Methods("GET")

	// Sale routes
	r.HandleFunc("/api/sales", saleHandler.GetSales).Methods("GET")
	r.HandleFunc("/api/sales/{id}", saleHandler.GetSale).Methods("GET")
	r.HandleFunc("/api/sales", saleHandler.CreateSale).Methods("POST")
	r.HandleFunc("/api/sales/{id}", saleHandler.UpdateSale).Methods("PUT")
	r.HandleFunc("/api/sales/{id}", saleHandler.DeleteSale).Methods("DELETE")
	r.HandleFunc("/api/sales/reference/{reference}", saleHandler.GetSaleByReference).Methods("GET")
	r.HandleFunc("/api/customers/{id}/sales", saleHandler.GetCustomerSales).Methods("GET")
	r.HandleFunc("/api/sales/summary", saleHandler.GetSalesSummary).Methods("GET")
	r.HandleFunc("/api/sales/daily", saleHandler.GetDailySales).Methods("GET")

	// Supplier routes
	r.HandleFunc("/api/suppliers", supplierHandler.GetAllSuppliers).Methods("GET")
	r.HandleFunc("/api/suppliers/{id}", supplierHandler.GetSupplier).Methods("GET")
	r.HandleFunc("/api/suppliers", supplierHandler.CreateSupplier).Methods("POST")
	r.HandleFunc("/api/suppliers/{id}", supplierHandler.UpdateSupplier).Methods("PUT")
	r.HandleFunc("/api/suppliers/{id}", supplierHandler.DeleteSupplier).Methods("DELETE")
	r.HandleFunc("/api/suppliers/search", supplierHandler.SearchSuppliers).Methods("GET")
	r.HandleFunc("/api/suppliers/top", supplierHandler.GetTopSuppliers).Methods("GET")

	// Stock-in routes
	r.HandleFunc("/api/stockins", stockInHandler.GetStockIns).Methods("GET")
	r.HandleFunc("/api/stockins/{id}", stockInHandler.GetStockIn).Methods("GET")
	r.HandleFunc("/api/stockins", stockInHandler.CreateStockIn).Methods("POST")
	r.HandleFunc("/api/stockins/{id}", stockInHandler.UpdateStockIn).Methods("PUT")
	r.HandleFunc("/api/stockins/{id}", stockInHandler.DeleteStockIn).Methods("DELETE")
	r.HandleFunc("/api/stockins/reference/{reference}", stockInHandler.GetStockInByReference).Methods("GET")
	r.HandleFunc("/api/stockins/{id}/items", stockInHandler.AddStockInItem).Methods("POST")
	r.HandleFunc("/api/stockins/{stockInId}/items/{itemId}", stockInHandler.UpdateStockInItem).Methods("PUT")
	r.HandleFunc("/api/stockins/{stockInId}/items/{itemId}", stockInHandler.DeleteStockInItem).Methods("DELETE")
	r.HandleFunc("/api/suppliers/{id}/stockins", stockInHandler.GetStockInsBySupplier).Methods("GET")
	r.HandleFunc("/api/stockins/summary", stockInHandler.GetStockInSummary).Methods("GET")
	r.HandleFunc("/api/stockins/daily", stockInHandler.GetDailyStockIn).Methods("GET")

	// Reject routes (for inventory decreases/write-offs)
	r.HandleFunc("/api/rejects", rejectHandler.GetRejects).Methods("GET")
	r.HandleFunc("/api/rejects/{id}", rejectHandler.GetReject).Methods("GET")
	r.HandleFunc("/api/rejects", rejectHandler.CreateReject).Methods("POST")
	r.HandleFunc("/api/rejects/{id}", rejectHandler.UpdateReject).Methods("PUT")
	r.HandleFunc("/api/rejects/{id}", rejectHandler.DeleteReject).Methods("DELETE")
	r.HandleFunc("/api/rejects/reference/{reference}", rejectHandler.GetRejectByReference).Methods("GET")
	r.HandleFunc("/api/rejects/{id}/items", rejectHandler.AddRejectItem).Methods("POST")
	r.HandleFunc("/api/rejects/{rejectId}/items/{itemId}", rejectHandler.UpdateRejectItem).Methods("PUT")
	r.HandleFunc("/api/rejects/{rejectId}/items/{itemId}", rejectHandler.DeleteRejectItem).Methods("DELETE")
	r.HandleFunc("/api/rejects/summary", rejectHandler.GetRejectSummary).Methods("GET")
	r.HandleFunc("/api/rejects/daily", rejectHandler.GetDailyReject).Methods("GET")
}
