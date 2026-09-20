// Package handler contains the reports-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/httpserver"
	"banking-platform/services/reports-service/internal/service"
)

// HTTP is the REST handler for reports-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts reports routes. Reading reports is restricted to staff.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/reports", authMW, staffMW)
	{
		g.GET("/daily", h.daily)
	}
}

func (h *HTTP) daily(c *gin.Context) {
	totals, err := h.svc.DailyList(c.Request.Context(), c.Query("date"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"totals": totals})
}
