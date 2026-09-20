// Package handler contains the support-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/support-service/internal/service"
)

// HTTP is the REST handler for support-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts support routes. All routes require authentication (authMW);
// staffMW additionally restricts cross-customer actions to teller/admin.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/support/tickets", authMW)
	{
		g.POST("", h.open)
		g.GET("", h.listMine)
		g.GET("/all", staffMW, h.listAll)
		g.GET("/:id", h.get)
		g.POST("/:id/status", staffMW, h.updateStatus)
	}
}

type openRequest struct {
	Subject string `json:"subject" binding:"required"`
	Body    string `json:"body" binding:"required"`
}

func (h *HTTP) open(c *gin.Context) {
	var req openRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	t, err := h.svc.Open(c.Request.Context(), auth.Subject(c), req.Subject, req.Body)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *HTTP) listMine(c *gin.Context) {
	items, err := h.svc.ListMine(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTP) get(c *gin.Context) {
	t, err := h.svc.GetOwned(c.Request.Context(), auth.Subject(c), c.Param("id"), isStaff(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *HTTP) listAll(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.svc.ListAll(c.Request.Context(), limit)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type statusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *HTTP) updateStatus(c *gin.Context) {
	var req statusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	t, err := h.svc.UpdateStatus(c.Request.Context(), c.Param("id"), req.Status)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func isStaff(c *gin.Context) bool {
	for _, r := range auth.Roles(c) {
		if r == "teller" || r == "admin" {
			return true
		}
	}
	return false
}
