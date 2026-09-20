// Command sms-service is a mock SMS provider.
package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"banking-platform/pkg/auth"
	platformcfg "banking-platform/pkg/config"
	"banking-platform/pkg/health"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/logger"
	"banking-platform/pkg/metrics"
	"banking-platform/pkg/postgres"
	"banking-platform/pkg/telemetry"

	"banking-platform/services/sms-service/internal/config"
	"banking-platform/services/sms-service/internal/handler"
	"banking-platform/services/sms-service/internal/repository"
	"banking-platform/services/sms-service/internal/service"
	"banking-platform/services/sms-service/migrations"
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
	defer func() {
		sc, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(sc)
	}()

	checker := health.New()

	var repo repository.Repository = repository.NewInMemory()
	if cfg.Postgres.Enabled {
		pool, err := postgres.New(ctx, postgres.Config{
			DSN: cfg.Postgres.DSN(), MaxConns: cfg.Postgres.MaxConns, MinConns: cfg.Postgres.MinConns,
		})
		if err != nil {
			return err
		}
		defer pool.Close()
		if err := postgres.Migrate(cfg.Postgres.DSN(), migrations.FS, ".", cfg.ServiceName); err != nil {
			return err
		}
		repo = repository.NewPostgres(pool)
		checker.Register("postgres", func(ctx context.Context) error { return pool.Ping(ctx) })
		log.Info("connected to postgres")
	}

	svc := service.New(repo, log)
	tokenMgr := auth.NewManager(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)

	srv := httpserver.New(httpserver.Config{
		Host: cfg.HTTP.Host, Port: cfg.HTTP.Port,
		ReadTimeout: cfg.HTTP.ReadTimeout, WriteTimeout: cfg.HTTP.WriteTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, cfg.ServiceName, log)
	metrics.Mount(srv.Router())
	checker.Mount(srv.Router())
	handler.NewHTTP(svc).Register(srv.Router(), tokenMgr.Middleware())
	checker.SetReady(true)

	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	log.Info("sms-service started", zap.Int("http_port", cfg.HTTP.Port))

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	checker.SetReady(false)
	sc, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(sc); err != nil {
		return err
	}
	log.Info("sms-service stopped cleanly")
	return nil
}
