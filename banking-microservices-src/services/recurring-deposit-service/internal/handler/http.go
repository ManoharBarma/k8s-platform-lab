// Package handler contains the recurring-deposit-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/recurring-deposit-service/internal/service"
)

// HTTP is the REST handler for recurring-deposit-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts recurring-deposit routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/recurring-deposits", authMW)
	{
		g.POST("", h.create)
		g.GET("", h.list)
		g.GET("/:id", h.get)
		g.POST("/:id/deposit", h.deposit)
	}
}

type createRequest struct {
	MonthlyAmount string  `json:"monthly_amount" binding:"required"`
	Currency      string  `json:"currency"`
	AnnualRatePct float64 `json:"annual_rate_pct"`
	TenureMonths  int     `json:"tenure_months" binding:"required"`
}

func (h *HTTP) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	d, err := h.svc.Create(c.Request.Context(), auth.Subject(c),
		req.MonthlyAmount, req.Currency, req.AnnualRatePct, req.TenureMonths)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, d)
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
	d, err := h.svc.GetOwned(c.Request.Context(), auth.Subject(c), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *HTTP) deposit(c *gin.Context) {
	d, err := h.svc.Deposit(c.Request.Context(), auth.Subject(c), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}
