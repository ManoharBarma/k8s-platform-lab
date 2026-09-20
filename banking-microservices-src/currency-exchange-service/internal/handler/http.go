// Package handler contains the currency-exchange-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/currency-exchange-service/internal/service"
)

// HTTP is the REST handler for currency-exchange-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts fx routes. Listing rates is public; conversion requires
// authentication via authMW.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/fx")
	{
		g.GET("/rates", h.rates)
		protected := g.Group("", authMW)
		protected.POST("/convert", h.convert)
	}
}

func (h *HTTP) rates(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"rates": items})
}

type convertRequest struct {
	From   string `json:"from" binding:"required"`
	To     string `json:"to" binding:"required"`
	Amount string `json:"amount" binding:"required"`
}

func (h *HTTP) convert(c *gin.Context) {
	var req convertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	result, rate, err := h.svc.Convert(c.Request.Context(), req.From, req.To, req.Amount)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"from":   req.From,
		"to":     req.To,
		"amount": req.Amount,
		"result": result,
		"rate":   rate,
	})
}
