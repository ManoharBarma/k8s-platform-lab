// Package handler contains ledger-service transport layers (HTTP + gRPC).
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/ledger-service/internal/domain"
	"banking-platform/services/ledger-service/internal/service"
)

// HTTP is the REST handler for the ledger.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts ledger routes. These are internal APIs; protect at the gateway.
func (h *HTTP) Register(r gin.IRouter) {
	g := r.Group("/api/v1/ledger")
	{
		g.POST("/accounts", h.openAccount)
		g.GET("/accounts/:id/balance", h.balance)
		g.POST("/transactions", h.post)
	}
}

type openAccountRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Currency  string `json:"currency" binding:"required"`
}

func (h *HTTP) openAccount(c *gin.Context) {
	var req openAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.OpenAccount(c.Request.Context(), req.AccountID, req.Currency); err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"account_id": req.AccountID, "currency": req.Currency})
}

type entryRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Direction string `json:"direction" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
}

type postRequest struct {
	TransactionID  string         `json:"transaction_id"`
	IdempotencyKey string         `json:"idempotency_key" binding:"required"`
	Currency       string         `json:"currency" binding:"required"`
	Reference      string         `json:"reference"`
	Entries        []entryRequest `json:"entries" binding:"required"`
}

func (h *HTTP) post(c *gin.Context) {
	var req postRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	entries := make([]domain.Entry, 0, len(req.Entries))
	for _, e := range req.Entries {
		amt, err := decimal.NewFromString(e.Amount)
		if err != nil {
			httpserver.Error(c, apierror.ErrBadRequest("invalid amount: "+e.Amount))
			return
		}
		entries = append(entries, domain.Entry{
			AccountID: e.AccountID, Direction: domain.Direction(e.Direction), Amount: amt,
		})
	}
	id, dup, err := h.svc.Post(c.Request.Context(), domain.Transaction{
		ID: req.TransactionID, IdempotencyKey: req.IdempotencyKey,
		Currency: req.Currency, Reference: req.Reference, Entries: entries,
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"transaction_id": id, "duplicate": dup})
}

func (h *HTTP) balance(c *gin.Context) {
	bal, currency, err := h.svc.Balance(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"account_id": c.Param("id"),
		"balance":    bal.StringFixed(2),
		"currency":   currency,
	})
}
