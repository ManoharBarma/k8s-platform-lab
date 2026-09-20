// Package handler contains the admin-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/admin-service/internal/service"
)

// HTTP is the REST handler for admin-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts admin settings routes. Reading settings is public so other
// services can fetch feature flags; updating requires authentication (authMW)
// and staff privileges (staffMW).
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/admin/settings")
	{
		g.GET("", h.list)
		g.GET("/:key", h.get)
		protected := g.Group("", authMW, staffMW)
		protected.PUT("/:key", h.set)
	}
}

func (h *HTTP) list(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"settings": items})
}

func (h *HTTP) get(c *gin.Context) {
	s, err := h.svc.Get(c.Request.Context(), c.Param("key"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}

type setRequest struct {
	Value string `json:"value"`
}

func (h *HTTP) set(c *gin.Context) {
	var req setRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	s, err := h.svc.Set(c.Request.Context(), c.Param("key"), req.Value)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, s)
}
