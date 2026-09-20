// Command api-gateway is the single entry point for the banking platform. It
// terminates client requests, applies CORS + rate limiting, validates JWTs for
// protected resources, propagates identity, and reverse-proxies to the owning
// microservice.
package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"banking-platform/pkg/auth"
	platformcfg "banking-platform/pkg/config"
	"banking-platform/pkg/health"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/logger"
	"banking-platform/pkg/metrics"
	"banking-platform/pkg/telemetry"

	"banking-platform/services/api-gateway/internal/config"
	"banking-platform/services/api-gateway/internal/middleware"
	"banking-platform/services/api-gateway/internal/proxy"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	var cfg config.Config
	if err := platformcfg.Load(&cfg); err != nil {
		return err
	}

	log, err := logger.New(cfg.Log.Level, cfg.Log.Format, cfg.ServiceName, cfg.Version, cfg.Environment)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tp, err := telemetry.Setup(ctx, telemetry.Config{
		Enabled: cfg.Telemetry.Enabled, ServiceName: cfg.ServiceName, ServiceVersion: cfg.Version,
		Environment: cfg.Environment, OTLPEndpoint: cfg.Telemetry.OTLPEndpoint,
		Insecure: cfg.Telemetry.Insecure, TraceSampleRatio: cfg.Telemetry.TraceSampleRatio,
	})
	if err != nil {
		log.Warn("telemetry setup failed; continuing", zap.Error(err))
		tp = &telemetry.Provider{}
	}
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tokenMgr := auth.NewManager(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	gw, err := proxy.New(cfg.UpstreamScheme, cfg.UpstreamPort, tokenMgr, log)
	if err != nil {
		return err
	}

	srv := httpserver.New(httpserver.Config{
		Host: cfg.HTTP.Host, Port: cfg.HTTP.Port,
		ReadTimeout: cfg.HTTP.ReadTimeout, WriteTimeout: cfg.HTTP.WriteTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, cfg.ServiceName, log)

	srv.Router().Use(
		middleware.CORS(cfg.AllowedOrigins),
		middleware.RateLimit(cfg.RateLimitRPS, cfg.RateLimitBurst),
	)

	checker := health.New()
	metrics.Mount(srv.Router())
	checker.Mount(srv.Router())
	checker.SetReady(true)

	// Everything not matched above is proxied to a backend service.
	srv.Router().(*gin.Engine).NoRoute(gw.Handler())

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	log.Info("api-gateway started", zap.Int("http_port", cfg.HTTP.Port))

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	checker.SetReady(false)
	if err := srv.Shutdown(context.Background()); err != nil {
		return err
	}
	log.Info("api-gateway stopped cleanly")
	return nil
}
