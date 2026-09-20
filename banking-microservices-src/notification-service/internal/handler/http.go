// Package handler contains the notification-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/notification-service/internal/service"
)

// HTTP is the REST handler for notification-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts notification routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/notifications", authMW)
	{
		g.GET("", h.list)
		g.POST("/:id/read", h.markRead)
	}
}

func (h *HTTP) list(c *gin.Context) {
	recipient := c.Query("recipient")
	if recipient == "" {
		recipient = auth.Subject(c)
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	notifications, err := h.svc.ListByRecipient(c.Request.Context(), recipient, limit)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *HTTP) markRead(c *gin.Context) {
	if err := h.svc.MarkRead(c.Request.Context(), c.Param("id")); err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "read"})
}
