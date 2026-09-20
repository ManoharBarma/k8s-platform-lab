// Package handler contains account-service transport layers (HTTP + gRPC).
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/account-service/internal/domain"
	"banking-platform/services/account-service/internal/service"
)

// HTTP is the REST handler for account-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts account routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/accounts", authMW)
	{
		g.POST("", h.open)
		g.GET("", h.listMine)
		g.GET("/:id", h.get)
		g.GET("/:id/balance", h.balance)
	}
}

type openRequest struct {
	Type     string `json:"type" binding:"required"`
	Currency string `json:"currency" binding:"required"`
}

func (h *HTTP) open(c *gin.Context) {
	var req openRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	acc, err := h.svc.Open(c.Request.Context(), auth.Subject(c), domain.Type(req.Type), req.Currency)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, acc)
}

func (h *HTTP) listMine(c *gin.Context) {
	items, err := h.svc.ListByCustomer(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// ownedAccount fetches an account and enforces ownership (staff may access any).
func (h *HTTP) ownedAccount(c *gin.Context) (domain.Account, bool) {
	acc, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return domain.Account{}, false
	}
	if acc.CustomerID != auth.Subject(c) && !isStaff(c) {
		httpserver.Error(c, apierror.ErrForbidden("not your account"))
		return domain.Account{}, false
	}
	return acc, true
}

func (h *HTTP) get(c *gin.Context) {
	acc, ok := h.ownedAccount(c)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, acc)
}

func (h *HTTP) balance(c *gin.Context) {
	acc, ok := h.ownedAccount(c)
	if !ok {
		return
	}
	bal, currency, err := h.svc.Balance(c.Request.Context(), acc.ID)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"account_id": acc.ID, "balance": bal, "currency": currency})
}

func isStaff(c *gin.Context) bool {
	for _, r := range auth.Roles(c) {
		if r == "teller" || r == "admin" {
			return true
		}
	}
	return false
}
