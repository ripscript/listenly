package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"google.golang.org/grpc/status"

	authv1 "listenly-backend/gen/go/auth/v1"
	"listenly-backend/internal/gateway/dto"
	"listenly-backend/pkg/ctxutil"
	"listenly-backend/pkg/response"
)

type AuthHandler struct {
	authClient authv1.AuthServiceClient
}

func NewAuthHandler(authClient authv1.AuthServiceClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

func (h *AuthHandler) Register(c *echo.Context) error {
	var req dto.RegisterHTTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.ValidationErrorResponse(c, err)
	}

	result, err := h.authClient.Register(c.Request().Context(), &authv1.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
		FullName: &req.FullName,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusBadRequest, st.Message(), nil)
	}

	return response.Success(c, http.StatusCreated, "User registered successfully", result)
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var req dto.LoginHTTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.ValidationErrorResponse(c, err)
	}

	result, err := h.authClient.Login(c.Request().Context(), &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusUnauthorized, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "User logged in successfully", result)
}

func (h *AuthHandler) Refresh(c *echo.Context) error {
	var req dto.RefreshHTTPRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return response.ValidationErrorResponse(c, err)
	}

	result, err := h.authClient.Refresh(c.Request().Context(), &authv1.RefreshRequest{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusUnauthorized, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "Token refreshed successfully", result)
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return response.Error(c, http.StatusUnauthorized, "Authorization header is missing", nil)
	}

	result, err := h.authClient.Logout(c.Request().Context(), &authv1.LogoutRequest{
		RefreshToken: authHeader,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusUnauthorized, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "User logged out successfully", result)
}

func (h *AuthHandler) LogoutAll(c *echo.Context) error {
	userUUID := ctxutil.GetUserUUID(c)

	result, err := h.authClient.LogoutAll(c.Request().Context(), &authv1.LogoutAllRequest{
		UserUuid: userUUID,
	})
	if err != nil {
		st, _ := status.FromError(err)
		return response.Error(c, http.StatusInternalServerError, st.Message(), nil)
	}

	return response.Success(c, http.StatusOK, "logged out from all devices", result)
}

func (h *AuthHandler) Verify(c *echo.Context) error {
	userUUID := ctxutil.GetUserUUID(c)
	email := ctxutil.GetEmail(c)

	return response.Success(c, http.StatusOK, "authenticated", map[string]string{
		"user_uuid": userUUID,
		"email":     email,
	})
}
