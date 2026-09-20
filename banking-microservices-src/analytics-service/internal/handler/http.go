// Package handler contains the analytics-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/httpserver"
	"banking-platform/services/analytics-service/internal/service"
)

// HTTP is the REST handler for analytics-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts analytics routes. Reading the summary is restricted to staff.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/analytics", authMW, staffMW)
	{
		g.GET("/summary", h.summary)
	}
}

func (h *HTTP) summary(c *gin.Context) {
	metrics, err := h.svc.Summary(c.Request.Context())
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}
