package migrations

import (
	"os"
	"strings"
	"testing"
)

func TestInitialMigrationDefinesRequiredWarehouseSchema(t *testing.T) {
	sqlBytes, err := os.ReadFile("000001_init.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}

	sql := normalizeSQL(string(sqlBytes))
	requiredSnippets := []string{
		"create table if not exists items",
		"id bigserial primary key",
		"sku varchar(64) not null",
		"name varchar(255) not null",
		"category varchar(100) not null",
		"unit varchar(50) not null",
		"created_at timestamptz not null default now()",
		"updated_at timestamptz not null default now()",
		"deleted_at timestamptz",
		"create unique index if not exists idx_items_sku_active on items (sku) where deleted_at is null",
		"create table if not exists locations",
		"code varchar(64) not null unique",
		"zone varchar(100) not null",
		"type varchar(50) not null",
		"create table if not exists stock",
		"item_id bigint not null references items(id)",
		"location_id bigint not null references locations(id)",
		"qty integer not null default 0",
		"constraint chk_stock_qty_non_negative check (qty >= 0)",
		"constraint uq_stock_item_location unique (item_id, location_id)",
		"create index if not exists idx_stock_item_id on stock (item_id)",
		"create index if not exists idx_stock_location_id on stock (location_id)",
	}

	for _, snippet := range requiredSnippets {
		if !strings.Contains(sql, snippet) {
			t.Fatalf("expected migration to contain %q\nsql:\n%s", snippet, string(sqlBytes))
		}
	}
}

func normalizeSQL(sql string) string {
	fields := strings.Fields(strings.ToLower(sql))
	return strings.Join(fields, " ")
}
