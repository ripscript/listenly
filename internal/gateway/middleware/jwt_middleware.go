package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"

	authv1 "listenly-backend/gen/go/auth/v1"
	"listenly-backend/pkg/jwt"
	"listenly-backend/pkg/response"
)

func JWTAuth(authClient authv1.AuthServiceClient, jwtManager *jwt.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return response.Error(c, http.StatusUnauthorized, "missing authorization header", nil)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return response.Error(c, http.StatusUnauthorized, "invalid authorization header format", nil)
			}
			tokenStr := strings.TrimSpace(parts[1])
			if tokenStr == "" {
				return response.Error(c, http.StatusUnauthorized, "empty token", nil)
			}

			// --- Jalur cepat: verifikasi JWT lokal ---
			if jwtManager != nil {
				claims, err := jwtManager.VerifyAccessToken(tokenStr)
				if err != nil {
					return response.Error(c, http.StatusUnauthorized, "invalid or expired token", nil)
				}
				c.Set("user_uuid", claims.UserUUID)
				c.Set("email", claims.Email)
				return next(c)
			}

			// --- Fallback: verifikasi via auth-service (perilaku lama) ---
			result, err := authClient.VerifyToken(c.Request().Context(), &authv1.VerifyTokenRequest{
				AccessToken: tokenStr,
			})
			if err != nil {
				return response.Error(c, http.StatusUnauthorized, "unauthorized", nil)
			}
			if !result.Valid {
				return response.Error(c, http.StatusUnauthorized, "invalid or expired token", nil)
			}

			c.Set("user_uuid", result.UserUuid)
			c.Set("email", result.Email)
			return next(c)
		}
	}
}
