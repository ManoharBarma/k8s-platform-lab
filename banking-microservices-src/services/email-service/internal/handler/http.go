// Package handler contains the email-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/email-service/internal/service"
)

// HTTP is the REST handler for email-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts email routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/email", authMW)
	{
		g.POST("/send", h.send)
		g.GET("/messages", h.list)
	}
}

type sendRequest struct {
	To      string `json:"to" binding:"required"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func (h *HTTP) send(c *gin.Context) {
	var req sendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	m, err := h.svc.Send(c.Request.Context(), req.To, req.Subject, req.Body)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *HTTP) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	items, err := h.svc.List(c.Request.Context(), limit)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"messages": items})
}
