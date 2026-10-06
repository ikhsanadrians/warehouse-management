package handler

import (
	"errors"
	"net/http"
)

type AppError struct {
	Status  int
	Message string
	Errors  []ErrorDetail
}

func (e AppError) Error() string {
	return e.Message
}

type AppHandler func(http.ResponseWriter, *http.Request) error

func (h AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h(w, r); err != nil {
		writeError(w, err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	var appErr AppError
	if !errors.As(err, &appErr) {
		appErr = AppError{
			Status:  http.StatusInternalServerError,
			Message: "Internal server error",
		}
	}

	if appErr.Status == 0 {
		appErr.Status = http.StatusInternalServerError
	}
	if appErr.Message == "" {
		appErr.Message = http.StatusText(appErr.Status)
	}

	writeJSON(w, appErr.Status, Envelope{
		Success: false,
		Message: appErr.Message,
		Errors:  appErr.Errors,
	})
}
