package service

import (
	"context"
	"errors"
	"testing"

	"warehouse-management/internal/model"
)

func TestStockServiceReceiveRejectsMissingItem(t *testing.T) {
	repo := &fakeStockRepository{locationExists: true}
	service := NewStockService(repo)

	_, err := service.Receive(context.Background(), ReceiveStockInput{ItemID: 1, LocationID: 2, Qty: 10})

	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
	if repo.received {
		t.Fatal("expected missing item to stop before stock update")
	}
}

func TestStockServiceReceiveRejectsMissingLocation(t *testing.T) {
	repo := &fakeStockRepository{itemExists: true}
	service := NewStockService(repo)

	_, err := service.Receive(context.Background(), ReceiveStockInput{ItemID: 1, LocationID: 2, Qty: 10})

	if !errors.Is(err, ErrLocationNotFound) {
		t.Fatalf("expected ErrLocationNotFound, got %v", err)
	}
	if repo.received {
		t.Fatal("expected missing location to stop before stock update")
	}
}

func TestStockServiceReceiveRejectsNonPositiveQty(t *testing.T) {
	repo := &fakeStockRepository{itemExists: true, locationExists: true}
	service := NewStockService(repo)

	_, err := service.Receive(context.Background(), ReceiveStockInput{ItemID: 1, LocationID: 2, Qty: 0})

	if !errors.Is(err, ErrInvalidStockQty) {
		t.Fatalf("expected ErrInvalidStockQty, got %v", err)
	}
	if repo.received {
		t.Fatal("expected invalid qty to stop before stock update")
	}
}

func TestStockServiceReceiveIncrementsStock(t *testing.T) {
	repo := &fakeStockRepository{
		itemExists:     true,
		locationExists: true,
		stock:          model.Stock{ID: 3, ItemID: 1, LocationID: 2, Qty: 15},
	}
	service := NewStockService(repo)

	stock, err := service.Receive(context.Background(), ReceiveStockInput{ItemID: 1, LocationID: 2, Qty: 10})
	if err != nil {
		t.Fatalf("receive stock: %v", err)
	}

	if !repo.received {
		t.Fatal("expected repository receive to be called")
	}
	if stock.Qty != 15 {
		t.Fatalf("expected qty 15, got %d", stock.Qty)
	}
}

func TestStockServiceGetByItemIDRejectsMissingItem(t *testing.T) {
	repo := &fakeStockRepository{}
	service := NewStockService(repo)

	_, err := service.GetByItemID(context.Background(), 1)

	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
	if repo.queried {
		t.Fatal("expected missing item to stop before stock query")
	}
}

func TestStockServiceGetByItemIDReturnsLocationStock(t *testing.T) {
	repo := &fakeStockRepository{
		itemExists: true,
		stockByLocation: []model.StockLocation{
			{LocationID: 2, Code: "A-01", Zone: "A", Type: "RACK", Qty: 20},
		},
	}
	service := NewStockService(repo)

	stock, err := service.GetByItemID(context.Background(), 1)
	if err != nil {
		t.Fatalf("get stock: %v", err)
	}

	if !repo.queried {
		t.Fatal("expected repository stock query to be called")
	}
	if len(stock) != 1 {
		t.Fatalf("expected 1 stock row, got %d", len(stock))
	}
	if stock[0].Code != "A-01" || stock[0].Qty != 20 {
		t.Fatalf("unexpected stock row: %+v", stock[0])
	}
}

type fakeStockRepository struct {
	itemExists      bool
	locationExists  bool
	received        bool
	queried         bool
	stock           model.Stock
	stockByLocation []model.StockLocation
}

func (f *fakeStockRepository) ActiveItemExists(ctx context.Context, id int64) (bool, error) {
	return f.itemExists, nil
}

func (f *fakeStockRepository) LocationExists(ctx context.Context, id int64) (bool, error) {
	return f.locationExists, nil
}

func (f *fakeStockRepository) Receive(ctx context.Context, input ReceiveStockInput) (model.Stock, error) {
	f.received = true
	return f.stock, nil
}

func (f *fakeStockRepository) GetByItemID(ctx context.Context, itemID int64) ([]model.StockLocation, error) {
	f.queried = true
	return f.stockByLocation, nil
}
