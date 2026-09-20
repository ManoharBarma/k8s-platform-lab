// Package handler contains the investment-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/investment-service/internal/service"
)

// HTTP is the REST handler for investment-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts investment routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/investments", authMW)
	{
		g.POST("/buy", h.buy)
		g.POST("/sell", h.sell)
		g.GET("/portfolio", h.portfolio)
	}
}

type buyRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	Units  string `json:"units" binding:"required"`
	Price  string `json:"price" binding:"required"`
}

func (h *HTTP) buy(c *gin.Context) {
	var req buyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	holding, err := h.svc.Buy(c.Request.Context(), auth.Subject(c), req.Symbol, req.Units, req.Price)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, holding)
}

type sellRequest struct {
	Symbol string `json:"symbol" binding:"required"`
	Units  string `json:"units" binding:"required"`
}

func (h *HTTP) sell(c *gin.Context) {
	var req sellRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	holding, err := h.svc.Sell(c.Request.Context(), auth.Subject(c), req.Symbol, req.Units)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, holding)
}

func (h *HTTP) portfolio(c *gin.Context) {
	holdings, err := h.svc.Portfolio(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"holdings": holdings})
}
