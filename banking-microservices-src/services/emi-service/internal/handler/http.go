// Package handler contains the emi-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/emi-service/internal/service"
)

// HTTP is the REST handler for emi-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts EMI routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/emi", authMW)
	{
		g.POST("/plans", h.createPlan)
		g.GET("/plans/:loanId", h.schedule)
		g.POST("/plans/:loanId/pay", h.payNext)
	}
}

type createPlanRequest struct {
	LoanID       string  `json:"loan_id" binding:"required"`
	Principal    string  `json:"principal" binding:"required"`
	AnnualRate   float64 `json:"annual_rate_pct"`
	TenureMonths int     `json:"tenure_months" binding:"required"`
}

func (h *HTTP) createPlan(c *gin.Context) {
	var req createPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	schedule, err := h.svc.CreatePlan(c.Request.Context(), service.CreatePlanInput{
		LoanID: req.LoanID, Principal: req.Principal, AnnualRate: req.AnnualRate, TenureMonths: req.TenureMonths,
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"installments": schedule})
}

func (h *HTTP) schedule(c *gin.Context) {
	items, err := h.svc.Schedule(c.Request.Context(), c.Param("loanId"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"installments": items})
}

func (h *HTTP) payNext(c *gin.Context) {
	ins, err := h.svc.PayNext(c.Request.Context(), c.Param("loanId"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, ins)
}
