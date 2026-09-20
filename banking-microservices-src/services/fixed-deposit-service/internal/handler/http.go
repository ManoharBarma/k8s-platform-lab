// Package handler contains the fixed-deposit-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/fixed-deposit-service/internal/service"
)

// HTTP is the REST handler for fixed-deposit-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts fixed deposit routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/fixed-deposits", authMW)
	{
		g.POST("", h.create)
		g.GET("", h.list)
		g.GET("/:id", h.get)
		g.POST("/:id/close", h.close)
	}
}

type createRequest struct {
	Principal     string  `json:"principal" binding:"required"`
	Currency      string  `json:"currency"`
	AnnualRatePct float64 `json:"annual_rate_pct"`
	TenureMonths  int     `json:"tenure_months"`
}

func (h *HTTP) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	fd, err := h.svc.Create(c.Request.Context(), auth.Subject(c), req.Principal, req.Currency, req.AnnualRatePct, req.TenureMonths)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, fd)
}

func (h *HTTP) list(c *gin.Context) {
	items, err := h.svc.ListByOwner(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTP) get(c *gin.Context) {
	fd, err := h.svc.GetOwned(c.Request.Context(), auth.Subject(c), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, fd)
}

func (h *HTTP) close(c *gin.Context) {
	fd, err := h.svc.Close(c.Request.Context(), auth.Subject(c), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, fd)
}
