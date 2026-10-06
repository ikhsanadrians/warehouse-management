package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"warehouse-management/internal/model"
	"warehouse-management/internal/service"
)

func TestStockHandlerReceiveValidatesRequiredFields(t *testing.T) {
	handler := NewStockHandler(&fakeStockService{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/receive", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()

	AppHandler(handler.Receive).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var body Envelope
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Success {
		t.Fatal("expected success false")
	}
	if len(body.Errors) != 3 {
		t.Fatalf("expected 3 validation errors, got %d", len(body.Errors))
	}
}

func TestStockHandlerReceiveMapsMissingItemToNotFound(t *testing.T) {
	handler := NewStockHandler(&fakeStockService{receiveErr: service.ErrItemNotFound})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/receive", bytes.NewBufferString(`{"item_id":1,"location_id":2,"qty":10}`))
	rec := httptest.NewRecorder()

	AppHandler(handler.Receive).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestStockHandlerReceiveReturnsStockEnvelope(t *testing.T) {
	handler := NewStockHandler(&fakeStockService{
		stock: model.Stock{ID: 1, ItemID: 1, LocationID: 2, Qty: 10},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/stock/receive", bytes.NewBufferString(`{"item_id":1,"location_id":2,"qty":10}`))
	rec := httptest.NewRecorder()

	AppHandler(handler.Receive).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestStockHandlerDetailReturnsLocationStockEnvelope(t *testing.T) {
	handler := NewStockHandler(&fakeStockService{
		stockByLocation: []model.StockLocation{
			{LocationID: 1, Code: "A-01", Zone: "A", Type: "RACK", Qty: 20},
		},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stock/1", nil)
	req = withURLParam(req, "item_id", "1")
	rec := httptest.NewRecorder()

	AppHandler(handler.Detail).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body struct {
		Success bool                  `json:"success"`
		Message string                `json:"message"`
		Data    []model.StockLocation `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success {
		t.Fatal("expected success true")
	}
	if body.Message != "Stock retrieved successfully" {
		t.Fatalf("expected stock message, got %q", body.Message)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 stock row, got %d", len(body.Data))
	}
	if body.Data[0].Code != "A-01" || body.Data[0].Qty != 20 {
		t.Fatalf("unexpected stock row: %+v", body.Data[0])
	}
}

func TestStockHandlerDetailMapsMissingItemToNotFound(t *testing.T) {
	handler := NewStockHandler(&fakeStockService{detailErr: service.ErrItemNotFound})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stock/1", nil)
	req = withURLParam(req, "item_id", "1")
	rec := httptest.NewRecorder()

	AppHandler(handler.Detail).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

type fakeStockService struct {
	receiveErr      error
	detailErr       error
	stock           model.Stock
	stockByLocation []model.StockLocation
}

func (f *fakeStockService) Receive(ctx context.Context, input service.ReceiveStockInput) (model.Stock, error) {
	if f.receiveErr != nil {
		return model.Stock{}, f.receiveErr
	}
	return f.stock, nil
}

func (f *fakeStockService) GetByItemID(ctx context.Context, itemID int64) ([]model.StockLocation, error) {
	if f.detailErr != nil {
		return nil, f.detailErr
	}
	return f.stockByLocation, nil
}

func withURLParam(req *http.Request, key string, value string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeContext))
}
