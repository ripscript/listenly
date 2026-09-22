package ctxutil

import "github.com/labstack/echo/v5"

func GetUserUUID(c *echo.Context) string {
	if v, ok := c.Get("user_uuid").(string); ok {
		return v
	}
	return ""
}

func GetEmail(c *echo.Context) string {
	if v, ok := c.Get("email").(string); ok {
		return v
	}
	return ""
}
