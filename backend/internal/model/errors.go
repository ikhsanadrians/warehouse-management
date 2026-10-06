package model

import "errors"

var (
	ErrDuplicateSKU     = errors.New("duplicate sku")
	ErrItemNotFound     = errors.New("item not found")
	ErrLocationNotFound = errors.New("location not found")
	ErrInvalidStockQty  = errors.New("stock quantity must be greater than zero")
)
