package handler

import (
	"strconv"

	"github.com/labstack/echo/v5"
)

func parsePagination(c *echo.Context) (page int32, pageSize int32) {
	page = parseInt32(c.QueryParam("page"))
	pageSize = parseInt32(c.QueryParam("page_size"))
	return page, pageSize
}

func parseInt32(s string) int32 {
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		return 0
	}
	return int32(n)
}
