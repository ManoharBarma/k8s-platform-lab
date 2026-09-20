// Package handler contains the fraud-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/httpserver"
	"banking-platform/services/fraud-service/internal/service"
)

// HTTP is the REST handler for fraud-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts fraud routes. Reading fraud alerts is restricted to staff.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/fraud", authMW, staffMW)
	{
		g.GET("/alerts", h.listAlerts)
	}
}

func (h *HTTP) listAlerts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	alerts, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}
