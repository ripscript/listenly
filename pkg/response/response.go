package response

import (
	"listenly-backend/pkg/apperror"

	"github.com/labstack/echo/v5"
)

type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func Success(c *echo.Context, code int, message string, data interface{}) error {
	return c.JSON(code, BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *echo.Context, code int, message string, errors interface{}) error {
	return c.JSON(code, BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

func ValidationErrorResponse(c *echo.Context, err error) error {
	if ve, ok := err.(*apperror.ValidationError); ok {
		return c.JSON(422, BaseResponse{
			Success: false,
			Message: "validation failed",
			Errors:  ve.Errors,
		})
	}
	return Error(c, 422, err.Error(), nil)
}
