package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/authz-service/internal/service"
)

// HTTP is the REST handler for authz-service admin operations.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts authz routes. adminMW restricts management endpoints.
func (h *HTTP) Register(r gin.IRouter, adminMW ...gin.HandlerFunc) {
	g := r.Group("/api/v1/authz")
	{
		g.GET("/roles", h.listRoles)
		g.POST("/check", h.check)
		protected := g.Group("", adminMW...)
		protected.POST("/assignments", h.assignRole)
	}
}

func (h *HTTP) listRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

type checkRequest struct {
	Subject    string `json:"subject" binding:"required"`
	Permission string `json:"permission" binding:"required"`
}

func (h *HTTP) check(c *gin.Context) {
	var req checkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	allowed, err := h.svc.Check(c.Request.Context(), req.Subject, req.Permission)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"allowed": allowed})
}

type assignRequest struct {
	Subject string `json:"subject" binding:"required"`
	Role    string `json:"role" binding:"required"`
}

func (h *HTTP) assignRole(c *gin.Context) {
	var req assignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.AssignRole(c.Request.Context(), req.Subject, req.Role); err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "assigned"})
}
