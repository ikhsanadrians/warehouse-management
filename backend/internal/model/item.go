package model

import "time"

type Item struct {
	ID        int64      `json:"id"`
	SKU       string     `json:"sku"`
	Name      string     `json:"name"`
	Category  string     `json:"category"`
	Unit      string     `json:"unit"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type ItemInput struct {
	SKU      string
	Name     string
	Category string
	Unit     string
}

type ItemListFilter struct {
	Category string
	Page     int
	Limit    int
}
