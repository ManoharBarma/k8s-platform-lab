// Package handler contains the audit-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/httpserver"
	"banking-platform/services/audit-service/internal/service"
)

// HTTP is the REST handler for audit-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts audit routes. Reading the audit log is restricted to staff.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/audit", authMW, staffMW)
	{
		g.GET("", h.list)
	}
}

func (h *HTTP) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	events, err := h.svc.List(c.Request.Context(), c.Query("topic"), limit)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
}
