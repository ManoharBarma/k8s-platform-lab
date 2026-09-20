// Command document-service manages document metadata and issues presigned S3
// URLs for direct upload/download.
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
	pkgs3 "banking-platform/pkg/s3"
	"banking-platform/pkg/telemetry"

	"banking-platform/services/document-service/internal/config"
	"banking-platform/services/document-service/internal/handler"
	"banking-platform/services/document-service/internal/repository"
	"banking-platform/services/document-service/internal/service"
	"banking-platform/services/document-service/migrations"
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

	// Object storage (optional).
	var storage *pkgs3.Client
	if cfg.S3.Enabled {
		storage, err = pkgs3.New(ctx, pkgs3.Config{
			Region: cfg.S3.Region, Bucket: cfg.S3.Bucket, Endpoint: cfg.S3.Endpoint,
			AccessKey: cfg.S3.AccessKey, SecretKey: cfg.S3.SecretKey, UsePathStyle: cfg.S3.UsePathStyle,
		})
		if err != nil {
			return err
		}
		log.Info("connected to object storage", zap.String("bucket", cfg.S3.Bucket))
	} else {
		log.Warn("S3 disabled; presigned upload/download URLs will be unavailable")
	}

	svc := service.New(repo, storage, log)
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
	log.Info("document-service started", zap.Int("http_port", cfg.HTTP.Port))

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
	log.Info("document-service stopped cleanly")
	return nil
}
