// Package handler contains the profile-service HTTP transport layer.
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/profile-service/internal/domain"
	"banking-platform/services/profile-service/internal/service"
)

// HTTP is the REST handler for profile-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts profile routes; all require authentication.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/profiles", authMW)
	{
		g.GET("/me", h.getMe)
		g.PUT("/me", h.upsertMe)
	}
}

func (h *HTTP) getMe(c *gin.Context) {
	p, err := h.svc.Get(c.Request.Context(), auth.Subject(c))
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

type upsertRequest struct {
	DateOfBirth  string `json:"date_of_birth"`
	Gender       string `json:"gender"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	PostalCode   string `json:"postal_code"`
}

func (h *HTTP) upsertMe(c *gin.Context) {
	var req upsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	p, err := h.svc.Upsert(c.Request.Context(), domain.Profile{
		UserID:       auth.Subject(c),
		DateOfBirth:  req.DateOfBirth,
		Gender:       req.Gender,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		Country:      req.Country,
		PostalCode:   req.PostalCode,
	})
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}
