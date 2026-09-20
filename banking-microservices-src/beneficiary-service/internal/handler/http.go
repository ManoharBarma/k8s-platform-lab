// Package handler contains the beneficiary-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/beneficiary-service/internal/domain"
	"banking-platform/services/beneficiary-service/internal/service"
)

// HTTP is the REST handler for beneficiary-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts beneficiary routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/beneficiaries", authMW)
	{
		g.POST("", h.add)
		g.GET("", h.list)
		g.GET("/:id", h.get)
		g.DELETE("/:id", h.delete)
	}
}

type addRequest struct {
	Name          string `json:"name" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	BankName      string `json:"bank_name"`
	RoutingCode   string `json:"routing_code"`
	Currency      string `json:"currency"`
}

func (h *HTTP) add(c *gin.Context) {
	var req addRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	b, err := h.svc.Add(c.Request.Context(), domain.Beneficiary{
		OwnerID: auth.Subject(c), Name: req.Name, AccountNumber: req.AccountNumber,
		BankName: req.BankName, RoutingCode: req.RoutingCode, Currency: req.Currency,
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *HTTP) list(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *HTTP) get(c *gin.Context) {
	b, err := h.svc.Get(c.Request.Context(), auth.Subject(c), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *HTTP) delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), auth.Subject(c), c.Param("id")); err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
