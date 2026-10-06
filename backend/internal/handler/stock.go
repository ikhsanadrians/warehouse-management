package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
	"warehouse-management/internal/model"
	"warehouse-management/internal/service"
)

type StockService interface {
	Receive(ctx context.Context, input service.ReceiveStockInput) (model.Stock, error)
	GetByItemID(ctx context.Context, itemID int64) ([]model.StockLocation, error)
}

type StockHandler struct {
	service StockService
}

type receiveStockRequest struct {
	ItemID     *int64 `json:"item_id"`
	LocationID *int64 `json:"location_id"`
	Qty        *int   `json:"qty"`
}

func NewStockHandler(service StockService) *StockHandler {
	return &StockHandler{service: service}
}

func (h *StockHandler) Receive(w http.ResponseWriter, r *http.Request) error {
	var request receiveStockRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return validationError("", "Invalid JSON body")
	}

	if details := validateReceiveStockRequest(request); len(details) > 0 {
		return AppError{Status: http.StatusBadRequest, Message: "Invalid request body", Errors: details}
	}

	stock, err := h.service.Receive(r.Context(), service.ReceiveStockInput{
		ItemID:     *request.ItemID,
		LocationID: *request.LocationID,
		Qty:        *request.Qty,
	})
	if err != nil {
		return mapStockError(err)
	}

	writeJSON(w, http.StatusOK, Envelope{
		Success: true,
		Message: "Stock received successfully",
		Data:    stock,
	})
	return nil
}

func (h *StockHandler) Detail(w http.ResponseWriter, r *http.Request) error {
	itemID, err := strconv.ParseInt(chi.URLParam(r, "item_id"), 10, 64)
	if err != nil || itemID <= 0 {
		return validationError("item_id", "Must be a positive integer")
	}

	stock, err := h.service.GetByItemID(r.Context(), itemID)
	if err != nil {
		return mapStockError(err)
	}

	writeJSON(w, http.StatusOK, Envelope{
		Success: true,
		Message: "Stock retrieved successfully",
		Data:    stock,
	})
	return nil
}

func validateReceiveStockRequest(request receiveStockRequest) []ErrorDetail {
	details := make([]ErrorDetail, 0)
	if request.ItemID == nil {
		details = append(details, ErrorDetail{Field: "item_id", Reason: "Required"})
	} else if *request.ItemID <= 0 {
		details = append(details, ErrorDetail{Field: "item_id", Reason: "Must be a positive integer"})
	}

	if request.LocationID == nil {
		details = append(details, ErrorDetail{Field: "location_id", Reason: "Required"})
	} else if *request.LocationID <= 0 {
		details = append(details, ErrorDetail{Field: "location_id", Reason: "Must be a positive integer"})
	}

	if request.Qty == nil {
		details = append(details, ErrorDetail{Field: "qty", Reason: "Required"})
	}

	return details
}

func mapStockError(err error) error {
	switch {
	case errors.Is(err, service.ErrItemNotFound):
		return AppError{Status: http.StatusNotFound, Message: "Item not found"}
	case errors.Is(err, service.ErrLocationNotFound):
		return AppError{Status: http.StatusNotFound, Message: "Location not found"}
	case errors.Is(err, service.ErrInvalidStockQty):
		return AppError{
			Status:  http.StatusBadRequest,
			Message: "Invalid stock quantity",
			Errors:  []ErrorDetail{{Field: "qty", Reason: "Must be greater than zero"}},
		}
	default:
		return err
	}
}
