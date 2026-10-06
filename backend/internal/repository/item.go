package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"warehouse-management/internal/model"
)

const activeSKUConstraintName = "idx_items_sku_active"

type ItemRepository struct {
	db *Postgres
}

func NewItemRepository(db *Postgres) *ItemRepository {
	return &ItemRepository{db: db}
}

func (r *ItemRepository) Create(ctx context.Context, input model.ItemInput) (model.Item, error) {
	const query = `
		INSERT INTO items (sku, name, category, unit)
		VALUES ($1, $2, $3, $4)
		RETURNING id, sku, name, category, unit, created_at, updated_at, deleted_at`

	var item model.Item
	err := r.db.Pool().QueryRow(ctx, query, input.SKU, input.Name, input.Category, input.Unit).
		Scan(&item.ID, &item.SKU, &item.Name, &item.Category, &item.Unit, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if isUniqueConstraint(err, activeSKUConstraintName) {
		return model.Item{}, model.ErrDuplicateSKU
	}
	return item, err
}

func (r *ItemRepository) List(ctx context.Context, filter model.ItemListFilter) ([]model.Item, int, error) {
	offset := (filter.Page - 1) * filter.Limit
	where := "WHERE deleted_at IS NULL"
	countArgs := []any{}
	listArgs := []any{filter.Limit, offset}
	if filter.Category != "" {
		where += " AND category = $3"
		countArgs = append(countArgs, filter.Category)
		listArgs = append(listArgs, filter.Category)
	}

	countWhere := "WHERE deleted_at IS NULL"
	if filter.Category != "" {
		countWhere += " AND category = $1"
	}

	countQuery := "SELECT COUNT(*) FROM items " + countWhere
	var total int
	if err := r.db.Pool().QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, sku, name, category, unit, created_at, updated_at, deleted_at
		FROM items ` + where + `
		ORDER BY id DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Pool().Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Item, 0)
	for rows.Next() {
		var item model.Item
		if err := rows.Scan(&item.ID, &item.SKU, &item.Name, &item.Category, &item.Unit, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

func (r *ItemRepository) GetByID(ctx context.Context, id int64) (*model.Item, error) {
	const query = `
		SELECT id, sku, name, category, unit, created_at, updated_at, deleted_at
		FROM items
		WHERE id = $1 AND deleted_at IS NULL`

	var item model.Item
	err := r.db.Pool().QueryRow(ctx, query, id).
		Scan(&item.ID, &item.SKU, &item.Name, &item.Category, &item.Unit, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func (r *ItemRepository) Update(ctx context.Context, id int64, input model.ItemInput) (model.Item, error) {
	const query = `
		UPDATE items
		SET sku = $2, name = $3, category = $4, unit = $5, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING id, sku, name, category, unit, created_at, updated_at, deleted_at`

	var item model.Item
	err := r.db.Pool().QueryRow(ctx, query, id, input.SKU, input.Name, input.Category, input.Unit).
		Scan(&item.ID, &item.SKU, &item.Name, &item.Category, &item.Unit, &item.CreatedAt, &item.UpdatedAt, &item.DeletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Item{}, model.ErrItemNotFound
	}
	if isUniqueConstraint(err, activeSKUConstraintName) {
		return model.Item{}, model.ErrDuplicateSKU
	}
	return item, err
}

func (r *ItemRepository) SoftDelete(ctx context.Context, id int64) error {
	const query = `
		UPDATE items
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`

	_, err := r.db.Pool().Exec(ctx, query, id)
	return err
}

func (r *ItemRepository) ActiveSKUExists(ctx context.Context, sku string, excludeID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM items
			WHERE sku = $1 AND deleted_at IS NULL AND ($2 = 0 OR id <> $2)
		)`

	var exists bool
	if err := r.db.Pool().QueryRow(ctx, query, sku, excludeID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func isUniqueConstraint(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == constraintName
}
