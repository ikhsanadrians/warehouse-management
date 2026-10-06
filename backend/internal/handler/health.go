package handler

import (
	"net/http"

	"warehouse-management/internal/model"
)

type HealthHandler struct{}

func NewHealthHandler() HealthHandler {
	return HealthHandler{}
}

func (h HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Envelope{
		Success: true,
		Message: "Service is healthy",
		Data: model.HealthStatus{
			Status: "ok",
		},
	})
}
