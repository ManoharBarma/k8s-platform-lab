// Package handler contains the payment-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/httpserver/idempotency"
	"banking-platform/services/payment-service/internal/service"
)

// HTTP is the REST handler for payment-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts payment routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/payments", authMW)
	{
		g.POST("", h.pay)
		g.GET("/:id", h.get)
		g.GET("", h.list)
	}
}

type payRequest struct {
	PayerAccountID string `json:"payer_account_id" binding:"required"`
	BeneficiaryID  string `json:"beneficiary_id" binding:"required"`
	Amount         string `json:"amount" binding:"required"`
	Currency       string `json:"currency" binding:"required"`
	Reference      string `json:"reference"`
}

func (h *HTTP) pay(c *gin.Context) {
	var req payRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	p, err := h.svc.Pay(c.Request.Context(), service.PayInput{
		PayerAccountID: req.PayerAccountID,
		BeneficiaryID:  req.BeneficiaryID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Reference:      req.Reference,
		IdempotencyKey: idempotency.Key(c),
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *HTTP) get(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *HTTP) list(c *gin.Context) {
	payer := c.Query("payer_account_id")
	if payer == "" {
		httpserver.Error(c, apierror.ErrBadRequest("payer_account_id query param is required"))
		return
	}
	items, err := h.svc.ListByPayer(c.Request.Context(), payer)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
