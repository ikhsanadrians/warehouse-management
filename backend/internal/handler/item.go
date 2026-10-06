package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"warehouse-management/internal/model"
	"warehouse-management/internal/service"
)

const maxListLimit = 100

type ItemService interface {
	Create(ctx context.Context, input service.CreateItemInput) (model.Item, error)
	List(ctx context.Context, filter service.ItemListFilter) ([]model.Item, int, error)
	GetByID(ctx context.Context, id int64) (*model.Item, error)
	Update(ctx context.Context, id int64, input service.UpdateItemInput) (model.Item, error)
	Delete(ctx context.Context, id int64) error
}

type ItemHandler struct {
	service ItemService
}

type itemRequest struct {
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Unit     string `json:"unit"`
}

type listMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func NewItemHandler(service ItemService) *ItemHandler {
	return &ItemHandler{service: service}
}

func (h *ItemHandler) Create(w http.ResponseWriter, r *http.Request) error {
	var request itemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return validationError("", "Invalid JSON body")
	}

	if details := validateItemRequest(request); len(details) > 0 {
		return AppError{Status: http.StatusBadRequest, Message: "Invalid request body", Errors: details}
	}

	item, err := h.service.Create(r.Context(), service.CreateItemInput{
		SKU:      strings.TrimSpace(request.SKU),
		Name:     strings.TrimSpace(request.Name),
		Category: strings.TrimSpace(request.Category),
		Unit:     strings.TrimSpace(request.Unit),
	})
	if err != nil {
		return mapItemError(err)
	}

	writeJSON(w, http.StatusCreated, Envelope{Success: true, Message: "Item created successfully", Data: item})
	return nil
}

func (h *ItemHandler) List(w http.ResponseWriter, r *http.Request) error {
	page, err := positiveIntQuery(r, "page", 1)
	if err != nil {
		return validationError("page", "Must be a positive integer")
	}

	limit, err := positiveIntQuery(r, "limit", 10)
	if err != nil {
		return validationError("limit", "Must be a positive integer")
	}
	if limit > maxListLimit {
		return validationError("limit", "Must be less than or equal to 100")
	}

	items, total, err := h.service.List(r.Context(), service.ItemListFilter{
		Category: strings.TrimSpace(r.URL.Query().Get("category")),
		Page:     page,
		Limit:    limit,
	})
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, Envelope{
		Success: true,
		Message: "Items retrieved successfully",
		Data:    items,
		Meta:    listMeta{Page: page, Limit: limit, Total: total},
	})
	return nil
}

func (h *ItemHandler) Detail(w http.ResponseWriter, r *http.Request) error {
	id, err := itemID(r)
	if err != nil {
		return err
	}

	item, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		return mapItemError(err)
	}

	writeJSON(w, http.StatusOK, Envelope{Success: true, Message: "Item retrieved successfully", Data: item})
	return nil
}

func (h *ItemHandler) Update(w http.ResponseWriter, r *http.Request) error {
	id, err := itemID(r)
	if err != nil {
		return err
	}

	var request itemRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return validationError("", "Invalid JSON body")
	}

	if details := validateItemRequest(request); len(details) > 0 {
		return AppError{Status: http.StatusBadRequest, Message: "Invalid request body", Errors: details}
	}

	item, err := h.service.Update(r.Context(), id, service.UpdateItemInput{
		SKU:      strings.TrimSpace(request.SKU),
		Name:     strings.TrimSpace(request.Name),
		Category: strings.TrimSpace(request.Category),
		Unit:     strings.TrimSpace(request.Unit),
	})
	if err != nil {
		return mapItemError(err)
	}

	writeJSON(w, http.StatusOK, Envelope{Success: true, Message: "Item updated successfully", Data: item})
	return nil
}

func (h *ItemHandler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, err := itemID(r)
	if err != nil {
		return err
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		return mapItemError(err)
	}

	writeJSON(w, http.StatusOK, Envelope{Success: true, Message: "Item deleted successfully"})
	return nil
}

func validateItemRequest(request itemRequest) []ErrorDetail {
	fields := map[string]string{
		"sku":      request.SKU,
		"name":     request.Name,
		"category": request.Category,
		"unit":     request.Unit,
	}

	details := make([]ErrorDetail, 0)
	for field, value := range fields {
		if strings.TrimSpace(value) == "" {
			details = append(details, ErrorDetail{Field: field, Reason: "Required"})
		}
	}

	return details
}

func positiveIntQuery(r *http.Request, key string, fallback int) (int, error) {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, errors.New("invalid positive integer")
	}

	return parsed, nil
}

func itemID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, validationError("id", "Must be a positive integer")
	}

	return id, nil
}

func validationError(field string, reason string) AppError {
	return AppError{
		Status:  http.StatusBadRequest,
		Message: "Invalid request",
		Errors:  []ErrorDetail{{Field: field, Reason: reason}},
	}
}

func mapItemError(err error) error {
	switch {
	case errors.Is(err, service.ErrDuplicateSKU):
		return AppError{
			Status:  http.StatusConflict,
			Message: "SKU already exists",
			Errors:  []ErrorDetail{{Field: "sku", Reason: "Duplicate entry"}},
		}
	case errors.Is(err, service.ErrItemNotFound):
		return AppError{Status: http.StatusNotFound, Message: "Item not found"}
	default:
		return err
	}
}
