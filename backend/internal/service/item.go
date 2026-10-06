package service

import (
	"context"

	"warehouse-management/internal/model"
)

var (
	ErrDuplicateSKU = model.ErrDuplicateSKU
	ErrItemNotFound = model.ErrItemNotFound
)

type ItemRepository interface {
	Create(ctx context.Context, input CreateItemInput) (model.Item, error)
	List(ctx context.Context, filter ItemListFilter) ([]model.Item, int, error)
	GetByID(ctx context.Context, id int64) (*model.Item, error)
	Update(ctx context.Context, id int64, input UpdateItemInput) (model.Item, error)
	SoftDelete(ctx context.Context, id int64) error
	ActiveSKUExists(ctx context.Context, sku string, excludeID int64) (bool, error)
}

type ItemService struct {
	repo ItemRepository
}

type CreateItemInput = model.ItemInput
type UpdateItemInput = model.ItemInput
type ItemListFilter = model.ItemListFilter

func NewItemService(repo ItemRepository) *ItemService {
	return &ItemService{repo: repo}
}

func (s *ItemService) Create(ctx context.Context, input CreateItemInput) (model.Item, error) {
	exists, err := s.repo.ActiveSKUExists(ctx, input.SKU, 0)
	if err != nil {
		return model.Item{}, err
	}
	if exists {
		return model.Item{}, ErrDuplicateSKU
	}

	return s.repo.Create(ctx, input)
}

func (s *ItemService) List(ctx context.Context, filter ItemListFilter) ([]model.Item, int, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	return s.repo.List(ctx, filter)
}

func (s *ItemService) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrItemNotFound
	}

	return item, nil
}

func (s *ItemService) Update(ctx context.Context, id int64, input UpdateItemInput) (model.Item, error) {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return model.Item{}, err
	}
	if item == nil {
		return model.Item{}, ErrItemNotFound
	}

	exists, err := s.repo.ActiveSKUExists(ctx, input.SKU, id)
	if err != nil {
		return model.Item{}, err
	}
	if exists {
		return model.Item{}, ErrDuplicateSKU
	}

	return s.repo.Update(ctx, id, input)
}

func (s *ItemService) Delete(ctx context.Context, id int64) error {
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if item == nil {
		return ErrItemNotFound
	}

	return s.repo.SoftDelete(ctx, id)
}
