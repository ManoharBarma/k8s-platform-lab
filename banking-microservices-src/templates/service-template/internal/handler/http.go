// Package handler contains the HTTP transport layer (Gin). Handlers translate
// HTTP <-> domain, delegating all logic to the service layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/httpserver"
	"banking-platform/templates/service-template/internal/service"
)

// Handler wires HTTP routes to the service.
type Handler struct {
	svc *service.Service
	log *zap.Logger
}

// New builds a Handler.
func New(svc *service.Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Register mounts the service's routes under /api/v1.
func (h *Handler) Register(r gin.IRouter) {
	g := r.Group("/api/v1/resources")
	{
		g.POST("", h.create)
		g.GET("", h.list)
		g.GET("/:id", h.get)
	}
}

type createRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	res, err := h.svc.Create(c.Request.Context(), req.Name)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *Handler) get(c *gin.Context) {
	res, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) list(c *gin.Context) {
	res, err := h.svc.List(c.Request.Context())
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": res})
}
