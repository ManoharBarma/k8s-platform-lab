// Package handler contains the statement-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/statement-service/internal/service"
)

// HTTP is the REST handler for statement-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts statement routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/statements", authMW)
	{
		g.GET("", h.list)
	}
}

func (h *HTTP) list(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		httpserver.Error(c, apierror.ErrBadRequest("account_id is required"))
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	entries, err := h.svc.ListByAccount(c.Request.Context(), accountID, limit)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"entries": entries})
}
