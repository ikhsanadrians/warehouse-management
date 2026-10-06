package repository

import (
	"context"
	"testing"

	"warehouse-management/internal/model"
)

func TestStockRepositoryReceiveIncrementsExistingStock(t *testing.T) {
	db := newIntegrationPostgres(t)
	seedItemAndLocation(t, db)
	repo := NewStockRepository(db)

	first, err := repo.Receive(context.Background(), model.ReceiveStockInput{ItemID: 1, LocationID: 1, Qty: 10})
	if err != nil {
		t.Fatalf("receive first stock: %v", err)
	}
	if first.Qty != 10 {
		t.Fatalf("expected first qty 10, got %d", first.Qty)
	}

	second, err := repo.Receive(context.Background(), model.ReceiveStockInput{ItemID: 1, LocationID: 1, Qty: 5})
	if err != nil {
		t.Fatalf("receive second stock: %v", err)
	}
	if second.Qty != 15 {
		t.Fatalf("expected incremented qty 15, got %d", second.Qty)
	}
}

func seedItemAndLocation(t *testing.T, db *Postgres) {
	t.Helper()

	ctx := context.Background()
	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO items (sku, name, category, unit)
		VALUES ('SKU-STOCK-001', 'Safety Gloves', 'Safety', 'pcs')`); err != nil {
		t.Fatalf("seed item: %v", err)
	}

	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO locations (code, zone, type)
		VALUES ('A-01', 'A', 'RACK')`); err != nil {
		t.Fatalf("seed location: %v", err)
	}
}
