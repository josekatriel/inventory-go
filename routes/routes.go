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
	saleHandler := handlers.NewSaleHandler(db) // Update these when repositories are ready

	// Product routes
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

	// Customer routes
	r.HandleFunc("/api/customers", customerHandler.GetAllCustomers).Methods("GET")
	r.HandleFunc("/api/customers/search", customerHandler.SearchCustomers).Methods("GET")
	r.HandleFunc("/api/customers/top", customerHandler.GetTopCustomers).Methods("GET")
	r.HandleFunc("/api/customers/{id}", customerHandler.GetCustomer).Methods("GET")
	r.HandleFunc("/api/customers", customerHandler.CreateCustomer).Methods("POST")
	r.HandleFunc("/api/customers/{id}", customerHandler.UpdateCustomer).Methods("PUT")
	r.HandleFunc("/api/customers/{id}", customerHandler.DeleteCustomer).Methods("DELETE")

	// Add sale routes
	r.HandleFunc("/api/sales", saleHandler.GetSales).Methods("GET")
	r.HandleFunc("/api/sales/summary", saleHandler.GetSalesSummary).Methods("GET")
	r.HandleFunc("/api/sales/daily", saleHandler.GetDailySales).Methods("GET")
	r.HandleFunc("/api/sales/{id}", saleHandler.GetSale).Methods("GET")
	r.HandleFunc("/api/sales/reference/{reference}", saleHandler.GetSaleByReference).Methods("GET")
	r.HandleFunc("/api/sales", saleHandler.CreateSale).Methods("POST")
	r.HandleFunc("/api/sales/{id}", saleHandler.UpdateSale).Methods("PUT")
	r.HandleFunc("/api/sales/{id}", saleHandler.DeleteSale).Methods("DELETE")
	r.HandleFunc("/api/customers/{id}/sales", saleHandler.GetCustomerSales).Methods("GET")
}
