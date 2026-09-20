// Package handler contains the wallet-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/httpserver/idempotency"
	"banking-platform/services/wallet-service/internal/service"
)

// HTTP is the REST handler for wallet-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts wallet routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/wallet", authMW)
	{
		g.POST("/topup", h.topup)
		g.POST("/spend", h.spend)
		g.GET("/balance", h.balance)
		g.GET("/transactions", h.history)
	}
}

type topupRequest struct {
	Amount   string `json:"amount" binding:"required"`
	Currency string `json:"currency"`
}

func (h *HTTP) topup(c *gin.Context) {
	var req topupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	w, err := h.svc.TopUp(c.Request.Context(), auth.Subject(c), req.Currency, req.Amount, idempotency.Key(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

type spendRequest struct {
	Amount string `json:"amount" binding:"required"`
}

func (h *HTTP) spend(c *gin.Context) {
	var req spendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	w, err := h.svc.Spend(c.Request.Context(), auth.Subject(c), req.Amount, idempotency.Key(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *HTTP) balance(c *gin.Context) {
	w, err := h.svc.Balance(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, w)
}

func (h *HTTP) history(c *gin.Context) {
	items, err := h.svc.History(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
