// Command auth-service issues and validates platform JWTs (REST + gRPC).
package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	platformcfg "banking-platform/pkg/config"
	"banking-platform/pkg/auth"
	"banking-platform/pkg/grpcserver"
	"banking-platform/pkg/health"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/logger"
	"banking-platform/pkg/metrics"
	"banking-platform/pkg/postgres"
	pkgredis "banking-platform/pkg/redis"
	"banking-platform/pkg/telemetry"

	authv1 "banking-platform/proto/gen/auth/v1"
	"banking-platform/services/auth-service/internal/config"
	"banking-platform/services/auth-service/internal/handler"
	"banking-platform/services/auth-service/internal/repository"
	"banking-platform/services/auth-service/internal/service"
	"banking-platform/services/auth-service/internal/session"
	"banking-platform/services/auth-service/migrations"
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

	// Persistence.
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

	// Session store (refresh tokens).
	var sessions session.Store = session.NewMemoryStore()
	if cfg.Redis.Enabled {
		rc, err := pkgredis.New(ctx, pkgredis.Config{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
		if err != nil {
			return err
		}
		defer func() { _ = rc.Close() }()
		sessions = session.NewRedisStore(rc)
		checker.Register("redis", func(ctx context.Context) error { return rc.Ping(ctx).Err() })
		log.Info("connected to redis")
	}

	tokenMgr := auth.NewManager(cfg.Auth.JWTSecret, cfg.Auth.JWTIssuer, cfg.Auth.AccessTTL, cfg.Auth.RefreshTTL)
	svc := service.New(repo, sessions, tokenMgr, log)

	// HTTP server.
	srv := httpserver.New(httpserver.Config{
		Host: cfg.HTTP.Host, Port: cfg.HTTP.Port,
		ReadTimeout: cfg.HTTP.ReadTimeout, WriteTimeout: cfg.HTTP.WriteTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, cfg.ServiceName, log)
	metrics.Mount(srv.Router())
	checker.Mount(srv.Router())
	handler.NewHTTP(svc).Register(srv.Router(), tokenMgr.Middleware())
	checker.SetReady(true)

	// gRPC server (optional).
	var grpcSrv *grpcserver.Server
	if cfg.GRPC.Enabled {
		grpcSrv = grpcserver.New(grpcserver.Config{Host: cfg.GRPC.Host, Port: cfg.GRPC.Port}, log)
		authv1.RegisterAuthServiceServer(grpcSrv.Server(), handler.NewGRPC(svc))
	}

	// Serve.
	errCh := make(chan error, 2)
	go func() { errCh <- srv.Start() }()
	if grpcSrv != nil {
		go func() { errCh <- grpcSrv.Start() }()
	}
	log.Info("auth-service started", zap.Int("http_port", cfg.HTTP.Port), zap.Bool("grpc", cfg.GRPC.Enabled))

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	checker.SetReady(false)
	sc, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if grpcSrv != nil {
		_ = grpcSrv.Shutdown(sc)
	}
	if err := srv.Shutdown(sc); err != nil {
		return err
	}
	log.Info("auth-service stopped cleanly")
	return nil
}
