// Package handler contains the loan-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/loan-service/internal/service"
)

// HTTP is the REST handler for loan-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts loan routes. authMW authenticates; staffMW gates approval and
// disbursement.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/loans", authMW)
	{
		g.POST("", h.apply)
		g.GET("", h.listMine)
		g.GET("/:id", h.get)
		g.POST("/:id/approve", staffMW, h.approve)
		g.POST("/:id/disburse", staffMW, h.disburse)
	}
}

type applyRequest struct {
	AccountID    string  `json:"account_id" binding:"required"`
	Principal    string  `json:"principal" binding:"required"`
	Currency     string  `json:"currency"`
	AnnualRate   float64 `json:"annual_rate_pct"`
	TenureMonths int     `json:"tenure_months" binding:"required"`
}

func (h *HTTP) apply(c *gin.Context) {
	var req applyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	loan, err := h.svc.Apply(c.Request.Context(), service.ApplyInput{
		BorrowerID: auth.Subject(c), AccountID: req.AccountID, Principal: req.Principal,
		Currency: req.Currency, AnnualRate: req.AnnualRate, TenureMonths: req.TenureMonths,
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, loan)
}

func (h *HTTP) listMine(c *gin.Context) {
	items, err := h.svc.ListByBorrower(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTP) get(c *gin.Context) {
	loan, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, loan)
}

type approveRequest struct {
	Approve bool `json:"approve"`
}

func (h *HTTP) approve(c *gin.Context) {
	var req approveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	loan, err := h.svc.Approve(c.Request.Context(), c.Param("id"), req.Approve)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, loan)
}

func (h *HTTP) disburse(c *gin.Context) {
	loan, err := h.svc.Disburse(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, loan)
}
