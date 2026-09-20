// Package handler contains the kyc-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/kyc-service/internal/service"
)

// HTTP is the REST handler for kyc-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts KYC routes. authMW authenticates; staffMW restricts review.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/kyc", authMW)
	{
		g.POST("", h.submit)
		g.GET("/me", h.getMine)
		g.GET("/pending", staffMW, h.pending)
		g.POST("/:id/review", staffMW, h.review)
	}
}

type submitRequest struct {
	DocumentType string `json:"document_type" binding:"required"`
	DocumentID   string `json:"document_id" binding:"required"`
}

func (h *HTTP) submit(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	v, err := h.svc.Submit(c.Request.Context(), auth.Subject(c), req.DocumentType, req.DocumentID)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *HTTP) getMine(c *gin.Context) {
	v, err := h.svc.LatestForUser(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *HTTP) pending(c *gin.Context) {
	items, err := h.svc.Pending(c.Request.Context())
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type reviewRequest struct {
	Approve bool   `json:"approve"`
	Reason  string `json:"reason"`
}

func (h *HTTP) review(c *gin.Context) {
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	v, err := h.svc.Review(c.Request.Context(), c.Param("id"), req.Approve, req.Reason)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, v)
}
