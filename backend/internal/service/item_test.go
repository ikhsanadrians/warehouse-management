package service

import (
	"context"
	"errors"
	"testing"

	"warehouse-management/internal/model"
)

func TestItemServiceCreateRejectsDuplicateSKU(t *testing.T) {
	repo := &fakeItemRepository{skuExists: true}
	service := NewItemService(repo)

	_, err := service.Create(context.Background(), CreateItemInput{
		SKU:      "SKU-001",
		Name:     "Safety Gloves",
		Category: "Safety",
		Unit:     "pcs",
	})

	if !errors.Is(err, ErrDuplicateSKU) {
		t.Fatalf("expected ErrDuplicateSKU, got %v", err)
	}
	if repo.created {
		t.Fatal("expected duplicate SKU to stop before repository create")
	}
}

func TestItemServiceDeleteSoftDeletesExistingItem(t *testing.T) {
	repo := &fakeItemRepository{item: &model.Item{ID: 1, SKU: "SKU-001"}}
	service := NewItemService(repo)

	if err := service.Delete(context.Background(), 1); err != nil {
		t.Fatalf("delete item: %v", err)
	}

	if repo.deletedID != 1 {
		t.Fatalf("expected deleted id 1, got %d", repo.deletedID)
	}
}

type fakeItemRepository struct {
	skuExists bool
	created   bool
	deletedID int64
	item      *model.Item
}

func (f *fakeItemRepository) Create(ctx context.Context, input CreateItemInput) (model.Item, error) {
	f.created = true
	return model.Item{ID: 1, SKU: input.SKU, Name: input.Name, Category: input.Category, Unit: input.Unit}, nil
}

func (f *fakeItemRepository) List(ctx context.Context, filter ItemListFilter) ([]model.Item, int, error) {
	return nil, 0, nil
}

func (f *fakeItemRepository) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	return f.item, nil
}

func (f *fakeItemRepository) Update(ctx context.Context, id int64, input UpdateItemInput) (model.Item, error) {
	return model.Item{ID: id, SKU: input.SKU, Name: input.Name, Category: input.Category, Unit: input.Unit}, nil
}

func (f *fakeItemRepository) SoftDelete(ctx context.Context, id int64) error {
	f.deletedID = id
	return nil
}

func (f *fakeItemRepository) ActiveSKUExists(ctx context.Context, sku string, excludeID int64) (bool, error) {
	return f.skuExists, nil
}
