// Package handler contains the card-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/card-service/internal/service"
)

// HTTP is the REST handler for card-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts card routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/cards", authMW)
	{
		g.POST("", h.issue)
		g.GET("", h.list)
		g.POST("/:id/block", h.block)
		g.POST("/:id/unblock", h.unblock)
	}
}

type issueRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Type      string `json:"type" binding:"required"`
}

func (h *HTTP) issue(c *gin.Context) {
	var req issueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	card, err := h.svc.Issue(c.Request.Context(), auth.Subject(c), req.AccountID, req.Type)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, card)
}

func (h *HTTP) list(c *gin.Context) {
	items, err := h.svc.ListByOwner(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTP) block(c *gin.Context) {
	card, err := h.svc.SetStatus(c.Request.Context(), auth.Subject(c), c.Param("id"), "BLOCKED")
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *HTTP) unblock(c *gin.Context) {
	card, err := h.svc.SetStatus(c.Request.Context(), auth.Subject(c), c.Param("id"), "ACTIVE")
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, card)
}
