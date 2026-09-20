// Package handler contains the search-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/httpserver"
	"banking-platform/services/search-service/internal/service"
)

// HTTP is the REST handler for search-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts search routes behind the authentication middleware.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/search", authMW)
	{
		g.GET("", h.search)
	}
}

func (h *HTTP) search(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	results, err := h.svc.Search(c.Request.Context(), c.Query("q"), limit)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
