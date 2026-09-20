// Package httpserver provides a Gin-based HTTP server with production defaults:
// request IDs, structured request logging, panic recovery, Prometheus metrics
// and OpenTelemetry tracing wired in as middleware, plus graceful shutdown.
package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
)

// Config controls the HTTP server.
type Config struct {
	Host            string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// Server wraps a Gin engine and an http.Server.
type Server struct {
	cfg    Config
	engine *gin.Engine
	http   *http.Server
	log    *zap.Logger
}

// New builds a Server with the platform middleware stack pre-installed.
func New(cfg Config, serviceName string, log *zap.Logger) *Server {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(
		RequestID(),
		RecoveryMiddleware(log),
		otelgin.Middleware(serviceName),
		LoggingMiddleware(log),
		PrometheusMiddleware(serviceName),
	)
	return &Server{cfg: cfg, engine: e, log: log}
}

// Router returns the underlying Gin router for route registration.
func (s *Server) Router() gin.IRouter { return s.engine }

// Start blocks serving HTTP until the server is shut down.
func (s *Server) Start() error {
	s.http = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Handler:      s.engine,
		ReadTimeout:  s.cfg.ReadTimeout,
		WriteTimeout: s.cfg.WriteTimeout,
	}
	s.log.Info("http server listening", zap.String("addr", s.http.Addr))
	if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Shutdown gracefully drains in-flight requests within the configured timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.http == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()
	return s.http.Shutdown(ctx)
}

// Error writes a JSON error response, mapping *apierror.Error to its status.
func Error(c *gin.Context, err error) {
	var apiErr *apierror.Error
	if errors.As(err, &apiErr) {
		c.AbortWithStatusJSON(apiErr.Status, apiErr)
		return
	}
	c.AbortWithStatusJSON(http.StatusInternalServerError,
		apierror.ErrInternal("internal server error"))
}
