// Package handler contains the transaction-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/httpserver/idempotency"
	"banking-platform/services/transaction-service/internal/service"
)

// HTTP is the REST handler for transaction-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts transaction routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/transactions", authMW)
	{
		g.POST("/deposit", h.deposit)
		g.POST("/withdraw", h.withdraw)
		g.POST("/transfer", h.transfer)
		g.GET("/:id", h.get)
		g.GET("", h.listByAccount)
	}
}

type depositRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Currency  string `json:"currency" binding:"required"`
	Reference string `json:"reference"`
}

func (h *HTTP) deposit(c *gin.Context) {
	var req depositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	t, err := h.svc.Deposit(c.Request.Context(), req.AccountID, req.Amount, req.Currency,
		idempotency.Key(c), req.Reference)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

type withdrawRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	Currency  string `json:"currency" binding:"required"`
	Reference string `json:"reference"`
}

func (h *HTTP) withdraw(c *gin.Context) {
	var req withdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	t, err := h.svc.Withdraw(c.Request.Context(), req.AccountID, req.Amount, req.Currency,
		idempotency.Key(c), req.Reference)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

type transferRequest struct {
	FromAccountID string `json:"from_account_id" binding:"required"`
	ToAccountID   string `json:"to_account_id" binding:"required"`
	Amount        string `json:"amount" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	Reference     string `json:"reference"`
}

func (h *HTTP) transfer(c *gin.Context) {
	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	t, err := h.svc.Transfer(c.Request.Context(), req.FromAccountID, req.ToAccountID, req.Amount,
		req.Currency, idempotency.Key(c), req.Reference)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *HTTP) get(c *gin.Context) {
	t, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *HTTP) listByAccount(c *gin.Context) {
	accountID := c.Query("account_id")
	if accountID == "" {
		httpserver.Error(c, apierror.ErrBadRequest("account_id query param is required"))
		return
	}
	items, err := h.svc.ListByAccount(c.Request.Context(), accountID)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
