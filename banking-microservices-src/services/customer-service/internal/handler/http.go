// Package handler contains the customer-service HTTP transport layer.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/customer-service/internal/service"
)

// HTTP is the REST handler for customer-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts customer routes. All routes require authentication (authMW);
// staffMW additionally restricts cross-customer reads to teller/admin.
func (h *HTTP) Register(r gin.IRouter, authMW, staffMW gin.HandlerFunc) {
	g := r.Group("/api/v1/customers", authMW)
	{
		g.POST("", h.create)
		g.GET("/me", h.getMe)
		g.PUT("/me", h.updateMe)
		g.GET("", staffMW, h.list)
		g.GET("/:id", staffMW, h.getByID)
	}
}

type createRequest struct {
	Email     string `json:"email" binding:"required,email"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func (h *HTTP) create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	cust, err := h.svc.Create(c.Request.Context(), service.CreateInput{
		UserID: auth.Subject(c), Email: req.Email,
		FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone,
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, cust)
}

func (h *HTTP) getMe(c *gin.Context) {
	cust, err := h.svc.GetByUserID(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, cust)
}

type updateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

func (h *HTTP) updateMe(c *gin.Context) {
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	cust, err := h.svc.UpdateByUserID(c.Request.Context(), auth.Subject(c), service.UpdateInput{
		FirstName: req.FirstName, LastName: req.LastName, Phone: req.Phone,
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, cust)
}

func (h *HTTP) getByID(c *gin.Context) {
	cust, err := h.svc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, cust)
}

func (h *HTTP) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	items, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
