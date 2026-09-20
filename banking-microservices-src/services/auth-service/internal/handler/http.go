// Package handler contains auth-service transport layers (HTTP + gRPC).
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/httpserver"
	"banking-platform/services/auth-service/internal/service"
)

// HTTP is the REST handler for auth-service.
type HTTP struct {
	svc *service.Service
}

// NewHTTP builds the REST handler.
func NewHTTP(svc *service.Service) *HTTP { return &HTTP{svc: svc} }

// Register mounts auth routes. authMW protects the identity endpoints.
func (h *HTTP) Register(r gin.IRouter, authMW gin.HandlerFunc) {
	g := r.Group("/api/v1/auth")
	{
		g.POST("/register", h.register)
		g.POST("/login", h.login)
		g.POST("/refresh", h.refresh)
		g.POST("/logout", h.logout)
		g.GET("/me", authMW, h.me)
	}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	FullName string `json:"full_name"`
}

func (h *HTTP) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	user, tokens, err := h.svc.Register(c.Request.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": user, "tokens": tokens})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *HTTP) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	tokens, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *HTTP) refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	tokens, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, tokens)
}

func (h *HTTP) logout(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpserver.Error(c, apierror.ErrBadRequest(err.Error()))
		return
	}
	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		httpserver.Error(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "logged_out"})
}

func (h *HTTP) me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"subject": auth.Subject(c),
		"email":   auth.Email(c),
		"roles":   auth.Roles(c),
	})
}
