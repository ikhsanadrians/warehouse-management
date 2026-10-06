package repository

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"warehouse-management/internal/model"
	"warehouse-management/internal/service"
)

func TestItemRepositoryCreateDuplicateSKUMapsConflict(t *testing.T) {
	db := newIntegrationPostgres(t)
	repo := NewItemRepository(db)

	input := model.ItemInput{
		SKU:      "SKU-DUP-001",
		Name:     "Safety Gloves",
		Category: "Safety",
		Unit:     "pcs",
	}
	if _, err := repo.Create(context.Background(), input); err != nil {
		t.Fatalf("create first item: %v", err)
	}

	_, err := repo.Create(context.Background(), input)

	if !errors.Is(err, service.ErrDuplicateSKU) {
		t.Fatalf("expected ErrDuplicateSKU, got %v", err)
	}
}

func TestItemRepositoryUpdateMissingRowMapsNotFound(t *testing.T) {
	db := newIntegrationPostgres(t)
	repo := NewItemRepository(db)

	_, err := repo.Update(context.Background(), 999, model.ItemInput{
		SKU:      "SKU-MISSING-001",
		Name:     "Missing",
		Category: "Safety",
		Unit:     "pcs",
	})

	if !errors.Is(err, service.ErrItemNotFound) {
		t.Fatalf("expected ErrItemNotFound, got %v", err)
	}
}

func TestItemRepositorySoftDeleteHidesItemFromNormalQueries(t *testing.T) {
	db := newIntegrationPostgres(t)
	repo := NewItemRepository(db)

	item, err := repo.Create(context.Background(), model.ItemInput{
		SKU:      "SKU-DELETE-001",
		Name:     "Safety Gloves",
		Category: "Safety",
		Unit:     "pcs",
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if err := repo.SoftDelete(context.Background(), item.ID); err != nil {
		t.Fatalf("soft delete item: %v", err)
	}

	deleted, err := repo.GetByID(context.Background(), item.ID)
	if err != nil {
		t.Fatalf("get deleted item: %v", err)
	}
	if deleted != nil {
		t.Fatalf("expected deleted item to be hidden, got %+v", deleted)
	}

	items, total, err := repo.List(context.Background(), model.ItemListFilter{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("expected deleted item hidden from list, total=%d len=%d", total, len(items))
	}
}

func newIntegrationPostgres(t *testing.T) *Postgres {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(pool.Close)

	resetIntegrationSchema(t, pool)
	return &Postgres{pool: pool}
}

func resetIntegrationSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS stock, locations, items CASCADE"); err != nil {
		t.Fatalf("drop tables: %v", err)
	}

	migrationPath := filepath.Join("..", "..", "migrations", "000001_init.sql")
	sqlBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}
