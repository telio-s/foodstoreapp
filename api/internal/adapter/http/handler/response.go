package handler

import (
	"errors"
	"net/http"

	"food-store-apis/internal/adapter/http/dto"
	"food-store-apis/internal/port"

	"github.com/gin-gonic/gin"
)

func success(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, dto.Response{Code: 0, Msg: msg, Data: data})
}

// failure writes an error response. If msg is empty, the message is derived
// from err (the AppError message, or err.Error() as a fallback).
func failure(c *gin.Context, msg string, err error) {
	var appErr *port.AppError
	if errors.As(err, &appErr) {
		status := http.StatusInternalServerError
		switch appErr.Code {
		case port.CodeNotFound:
			status = http.StatusNotFound
		case port.CodeInvalid:
			status = http.StatusBadRequest
		}
		if msg == "" {
			msg = appErr.Message
		}
		code := 1
		if appErr.ErrorCode != 0 {
			code = appErr.ErrorCode
		}
		c.JSON(status, dto.Response{Code: code, Msg: msg, Error: appErr.Errors})
		return
	}

	if msg == "" {
		msg = err.Error()
	}
	c.JSON(http.StatusInternalServerError, dto.Response{Code: 1, Msg: msg})
}
