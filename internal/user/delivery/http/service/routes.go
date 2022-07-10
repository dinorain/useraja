package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *userHandlersHTTP) MapRoutes() {
	h.group.POST("", h.Register())
	h.group.Any("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "OK")
	})
}
