package handler

import "github.com/go-chi/chi/v5"

func NewRouter(itemService ItemService, stockService StockService) *chi.Mux {
	router := chi.NewRouter()
	health := NewHealthHandler()

	router.Get("/health", health.Check)
	if itemService != nil {
		item := NewItemHandler(itemService)
		router.Get("/api/v1/items", AppHandler(item.List).ServeHTTP)
		router.Post("/api/v1/items", AppHandler(item.Create).ServeHTTP)
		router.Get("/api/v1/items/{id}", AppHandler(item.Detail).ServeHTTP)
		router.Put("/api/v1/items/{id}", AppHandler(item.Update).ServeHTTP)
		router.Delete("/api/v1/items/{id}", AppHandler(item.Delete).ServeHTTP)
	}
	if stockService != nil {
		stock := NewStockHandler(stockService)
		router.Post("/api/v1/stock/receive", AppHandler(stock.Receive).ServeHTTP)
		router.Get("/api/v1/stock/{item_id}", AppHandler(stock.Detail).ServeHTTP)
	}

	return router
}
