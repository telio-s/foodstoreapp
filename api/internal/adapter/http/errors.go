package http

import (
	"errors"
	"net/http"

	"food-store-apis/internal/domain/apperror"
)

func writeError(w http.ResponseWriter, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		status := http.StatusInternalServerError
		switch appErr.Code {
		case apperror.CodeNotFound:
			status = http.StatusNotFound
		case apperror.CodeInvalid:
			status = http.StatusBadRequest
		}
		http.Error(w, appErr.Message, status)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}
