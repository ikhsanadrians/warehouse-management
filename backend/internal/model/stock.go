package model

import "time"

type Stock struct {
	ID         int64     `json:"id"`
	ItemID     int64     `json:"item_id"`
	LocationID int64     `json:"location_id"`
	Qty        int       `json:"qty"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ReceiveStockInput struct {
	ItemID     int64
	LocationID int64
	Qty        int
}

type StockLocation struct {
	LocationID int64  `json:"location_id"`
	Code       string `json:"code"`
	Zone       string `json:"zone"`
	Type       string `json:"type"`
	Qty        int    `json:"qty"`
}
