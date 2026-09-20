// Package handler contains the document-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/document-service/internal/service"
)

// HTTP is the REST handler for document-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts document routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/documents", authMW)
	{
		g.POST("", h.requestUpload)
		g.GET("", h.list)
		g.GET("/:id/download", h.download)
		g.DELETE("/:id", h.delete)
	}
}

type uploadRequest struct {
	Type        string `json:"type" binding:"required"`
	ContentType string `json:"content_type"`
}

func (h *HTTP) requestUpload(c *gin.Context) {
	var req uploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	ticket, err := h.svc.RequestUpload(c.Request.Context(), auth.Subject(c), req.Type, req.ContentType)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, ticket)
}

func (h *HTTP) list(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTP) download(c *gin.Context) {
	url, err := h.svc.DownloadURL(c.Request.Context(), auth.Subject(c), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"download_url": url})
}

func (h *HTTP) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), auth.Subject(c), c.Param("id")); err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
