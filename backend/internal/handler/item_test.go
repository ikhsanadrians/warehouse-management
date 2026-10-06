package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"warehouse-management/internal/model"
	"warehouse-management/internal/service"
)

func TestItemHandlerListUsesCategoryFilterAndPaginationMeta(t *testing.T) {
	fakeService := &fakeItemService{
		listItems: []model.Item{{ID: 7, SKU: "SKU-007", Name: "Tablet", Category: "electronics", Unit: "pcs"}},
		listTotal: 42,
	}
	handler := NewItemHandler(fakeService)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?category=electronics&page=2&limit=10", nil)
	rec := httptest.NewRecorder()

	AppHandler(handler.List).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if fakeService.listFilter.Category != "electronics" {
		t.Fatalf("expected category electronics, got %q", fakeService.listFilter.Category)
	}
	if fakeService.listFilter.Page != 2 {
		t.Fatalf("expected page 2, got %d", fakeService.listFilter.Page)
	}
	if fakeService.listFilter.Limit != 10 {
		t.Fatalf("expected limit 10, got %d", fakeService.listFilter.Limit)
	}

	var body struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Data    []any    `json:"data"`
		Meta    listMeta `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success {
		t.Fatal("expected success true")
	}
	if body.Message != "Items retrieved successfully" {
		t.Fatalf("expected list message, got %q", body.Message)
	}
	if len(body.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(body.Data))
	}
	if body.Meta.Page != 2 || body.Meta.Limit != 10 || body.Meta.Total != 42 {
		t.Fatalf("unexpected meta: %+v", body.Meta)
	}
}

func TestItemHandlerListRejectsInvalidPagination(t *testing.T) {
	handler := NewItemHandler(&fakeItemService{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?page=0&limit=10", nil)
	rec := httptest.NewRecorder()

	AppHandler(handler.List).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestItemHandlerListRejectsLimitAboveMaximum(t *testing.T) {
	handler := NewItemHandler(&fakeItemService{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/items?page=1&limit=101", nil)
	rec := httptest.NewRecorder()

	AppHandler(handler.List).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestItemHandlerCreateValidatesRequiredFields(t *testing.T) {
	handler := NewItemHandler(&fakeItemService{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(`{"sku":"","name":"","category":"","unit":""}`))
	rec := httptest.NewRecorder()

	AppHandler(handler.Create).ServeHTTP(rec, req)

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
	if len(body.Errors) != 4 {
		t.Fatalf("expected 4 validation errors, got %d", len(body.Errors))
	}
}

func TestItemHandlerCreateMapsDuplicateSKUToConflict(t *testing.T) {
	handler := NewItemHandler(&fakeItemService{createErr: service.ErrDuplicateSKU})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/items", bytes.NewBufferString(`{"sku":"SKU-001","name":"Safety Gloves","category":"Safety","unit":"pcs"}`))
	rec := httptest.NewRecorder()

	AppHandler(handler.Create).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

type fakeItemService struct {
	createErr  error
	listFilter service.ItemListFilter
	listItems  []model.Item
	listTotal  int
}

func (f *fakeItemService) Create(ctx context.Context, input service.CreateItemInput) (model.Item, error) {
	if f.createErr != nil {
		return model.Item{}, f.createErr
	}
	return model.Item{ID: 1, SKU: input.SKU, Name: input.Name, Category: input.Category, Unit: input.Unit}, nil
}

func (f *fakeItemService) List(ctx context.Context, filter service.ItemListFilter) ([]model.Item, int, error) {
	f.listFilter = filter
	return f.listItems, f.listTotal, nil
}

func (f *fakeItemService) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeItemService) Update(ctx context.Context, id int64, input service.UpdateItemInput) (model.Item, error) {
	return model.Item{}, errors.New("not implemented")
}

func (f *fakeItemService) Delete(ctx context.Context, id int64) error {
	return errors.New("not implemented")
}
