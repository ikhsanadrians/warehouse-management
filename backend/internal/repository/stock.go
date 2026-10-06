package repository

import (
	"context"

	"warehouse-management/internal/model"
)

type StockRepository struct {
	db *Postgres
}

func NewStockRepository(db *Postgres) *StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) ActiveItemExists(ctx context.Context, id int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM items
			WHERE id = $1 AND deleted_at IS NULL
		)`

	var exists bool
	if err := r.db.Pool().QueryRow(ctx, query, id).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *StockRepository) LocationExists(ctx context.Context, id int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM locations
			WHERE id = $1
		)`

	var exists bool
	if err := r.db.Pool().QueryRow(ctx, query, id).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *StockRepository) Receive(ctx context.Context, input model.ReceiveStockInput) (model.Stock, error) {
	const query = `
		INSERT INTO stock (item_id, location_id, qty)
		VALUES ($1, $2, $3)
		ON CONFLICT (item_id, location_id)
		DO UPDATE SET qty = stock.qty + EXCLUDED.qty, updated_at = NOW()
		RETURNING id, item_id, location_id, qty, updated_at`

	var stock model.Stock
	err := r.db.Pool().QueryRow(ctx, query, input.ItemID, input.LocationID, input.Qty).
		Scan(&stock.ID, &stock.ItemID, &stock.LocationID, &stock.Qty, &stock.UpdatedAt)
	return stock, err
}

func (r *StockRepository) GetByItemID(ctx context.Context, itemID int64) ([]model.StockLocation, error) {
	const query = `
		SELECT l.id, l.code, l.zone, l.type, s.qty
		FROM stock s
		INNER JOIN locations l ON l.id = s.location_id
		WHERE s.item_id = $1
		ORDER BY l.code ASC`

	rows, err := r.db.Pool().Query(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stockRows := make([]model.StockLocation, 0)
	for rows.Next() {
		var stock model.StockLocation
		if err := rows.Scan(&stock.LocationID, &stock.Code, &stock.Zone, &stock.Type, &stock.Qty); err != nil {
			return nil, err
		}
		stockRows = append(stockRows, stock)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stockRows, nil
}
