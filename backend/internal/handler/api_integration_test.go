package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"warehouse-management/internal/repository"
	"warehouse-management/internal/service"
)

func TestAPIIntegrationBusinessRules(t *testing.T) {
	db := newHandlerIntegrationPostgres(t)
	seedHandlerIntegrationLocation(t, db)

	itemService := service.NewItemService(repository.NewItemRepository(db))
	stockService := service.NewStockService(repository.NewStockRepository(db))
	router := NewRouter(itemService, stockService)

	createResp := performJSON(router, http.MethodPost, "/api/v1/items", `{"sku":"SKU-API-001","name":"Safety Gloves","category":"Safety","unit":"pcs"}`)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d body=%s", http.StatusCreated, createResp.Code, createResp.Body.String())
	}

	duplicateResp := performJSON(router, http.MethodPost, "/api/v1/items", `{"sku":"SKU-API-001","name":"Other Gloves","category":"Safety","unit":"pcs"}`)
	if duplicateResp.Code != http.StatusConflict {
		t.Fatalf("expected duplicate status %d, got %d body=%s", http.StatusConflict, duplicateResp.Code, duplicateResp.Body.String())
	}

	receiveResp := performJSON(router, http.MethodPost, "/api/v1/stock/receive", `{"item_id":1,"location_id":1,"qty":10}`)
	if receiveResp.Code != http.StatusOK {
		t.Fatalf("expected receive status %d, got %d body=%s", http.StatusOK, receiveResp.Code, receiveResp.Body.String())
	}
	var receiveBody struct {
		Data struct {
			Qty int `json:"qty"`
		} `json:"data"`
	}
	if err := json.NewDecoder(receiveResp.Body).Decode(&receiveBody); err != nil {
		t.Fatalf("decode receive response: %v", err)
	}
	if receiveBody.Data.Qty != 10 {
		t.Fatalf("expected received qty 10, got %d", receiveBody.Data.Qty)
	}

	invalidQtyResp := performJSON(router, http.MethodPost, "/api/v1/stock/receive", `{"item_id":1,"location_id":1,"qty":0}`)
	if invalidQtyResp.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid qty status %d, got %d body=%s", http.StatusBadRequest, invalidQtyResp.Code, invalidQtyResp.Body.String())
	}

	missingItemResp := performJSON(router, http.MethodPost, "/api/v1/stock/receive", `{"item_id":999,"location_id":1,"qty":1}`)
	if missingItemResp.Code != http.StatusNotFound {
		t.Fatalf("expected missing item status %d, got %d body=%s", http.StatusNotFound, missingItemResp.Code, missingItemResp.Body.String())
	}

	deleteResp := performJSON(router, http.MethodDelete, "/api/v1/items/1", "")
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("expected delete status %d, got %d body=%s", http.StatusOK, deleteResp.Code, deleteResp.Body.String())
	}

	listResp := performJSON(router, http.MethodGet, "/api/v1/items", "")
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected list status %d, got %d body=%s", http.StatusOK, listResp.Code, listResp.Body.String())
	}
	var listBody struct {
		Data []any `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&listBody); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listBody.Data) != 0 || listBody.Meta.Total != 0 {
		t.Fatalf("expected deleted item hidden, len=%d total=%d", len(listBody.Data), listBody.Meta.Total)
	}

	detailResp := performJSON(router, http.MethodGet, "/api/v1/items/1", "")
	if detailResp.Code != http.StatusNotFound {
		t.Fatalf("expected deleted detail status %d, got %d body=%s", http.StatusNotFound, detailResp.Code, detailResp.Body.String())
	}
}

func newHandlerIntegrationPostgres(t *testing.T) *repository.Postgres {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	resetHandlerIntegrationSchema(t, pool)
	pool.Close()

	db, err := repository.NewPostgres(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("connect repository postgres: %v", err)
	}
	t.Cleanup(db.Close)

	return db
}

func resetHandlerIntegrationSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(context.Background(), "DROP TABLE IF EXISTS stock, locations, items CASCADE"); err != nil {
		t.Fatalf("drop tables: %v", err)
	}

	sqlBytes, err := os.ReadFile("../../migrations/000001_init.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := pool.Exec(context.Background(), string(sqlBytes)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

func seedHandlerIntegrationLocation(t *testing.T, db *repository.Postgres) {
	t.Helper()

	if _, err := db.Pool().Exec(context.Background(), `
		INSERT INTO locations (code, zone, type)
		VALUES ('A-01', 'A', 'RACK')`); err != nil {
		t.Fatalf("seed location: %v", err)
	}
}

func performJSON(handler http.Handler, method string, path string, body string) *httptest.ResponseRecorder {
	var requestBody *bytes.Buffer
	if body == "" {
		requestBody = bytes.NewBuffer(nil)
	} else {
		requestBody = bytes.NewBufferString(body)
	}

	req := httptest.NewRequest(method, path, requestBody)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
