package service

import (
	"context"
	"log"

	"warehouse-management/internal/model"
)

var (
	ErrLocationNotFound = model.ErrLocationNotFound
	ErrInvalidStockQty  = model.ErrInvalidStockQty
)

type StockRepository interface {
	ActiveItemExists(ctx context.Context, id int64) (bool, error)
	LocationExists(ctx context.Context, id int64) (bool, error)
	Receive(ctx context.Context, input ReceiveStockInput) (model.Stock, error)
	GetByItemID(ctx context.Context, itemID int64) ([]model.StockLocation, error)
}

type StockService struct {
	repo StockRepository
}

type ReceiveStockInput = model.ReceiveStockInput

func NewStockService(repo StockRepository) *StockService {
	return &StockService{repo: repo}
}

func (s *StockService) Receive(ctx context.Context, input ReceiveStockInput) (model.Stock, error) {
	if input.Qty <= 0 {
		return model.Stock{}, ErrInvalidStockQty
	}

	itemExists, err := s.repo.ActiveItemExists(ctx, input.ItemID)
	if err != nil {
		return model.Stock{}, err
	}
	if !itemExists {
		return model.Stock{}, ErrItemNotFound
	}

	locationExists, err := s.repo.LocationExists(ctx, input.LocationID)
	if err != nil {
		return model.Stock{}, err
	}
	if !locationExists {
		return model.Stock{}, ErrLocationNotFound
	}

	stock, err := s.repo.Receive(ctx, input)
	if err != nil {
		return model.Stock{}, err
	}

	log.Printf("stock received: item_id=%d location_id=%d qty=%d", input.ItemID, input.LocationID, input.Qty)
	return stock, nil
}

func (s *StockService) GetByItemID(ctx context.Context, itemID int64) ([]model.StockLocation, error) {
	itemExists, err := s.repo.ActiveItemExists(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if !itemExists {
		return nil, ErrItemNotFound
	}

	return s.repo.GetByItemID(ctx, itemID)
}
