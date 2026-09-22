package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"listenly-backend/internal/auth/dto"
	"listenly-backend/internal/auth/service"
	"listenly-backend/pkg/response"
)

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "invalid request data", err.Error())
	}

	result, err := h.service.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			return response.Error(c, http.StatusConflict, "email already exists", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "internal server error", nil)
	}
	return response.Success(c, http.StatusCreated, "User registered successfully", result)
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "invalid request data", err.Error())
	}

	result, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return response.Error(c, http.StatusUnauthorized, "invalid credentials", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "internal server error", nil)
	}
	return response.Success(c, http.StatusOK, "User logged in successfully", result)
}

func (h *AuthHandler) Refresh(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.Error(c, http.StatusUnprocessableEntity, "invalid request data", err.Error())
	}

	result, err := h.service.Refresh(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			return response.Error(c, http.StatusUnauthorized, "invalid token", nil)
		}
		return response.Error(c, http.StatusInternalServerError, "internal server error", nil)
	}
	return response.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	var req dto.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	if err := h.service.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return response.Error(c, http.StatusInternalServerError, "internal server error", nil)
	}
	return c.NoContent(http.StatusNoContent)
}
