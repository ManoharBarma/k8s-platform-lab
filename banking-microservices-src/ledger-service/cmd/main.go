// Command ledger-service is the double-entry bookkeeping engine and source of
// truth for money (REST + gRPC).
package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	platformcfg "banking-platform/pkg/config"
	"banking-platform/pkg/grpcserver"
	"banking-platform/pkg/health"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/logger"
	"banking-platform/pkg/metrics"
	"banking-platform/pkg/postgres"
	"banking-platform/pkg/telemetry"

	ledgerv1 "banking-platform/proto/gen/ledger/v1"
	"banking-platform/services/ledger-service/internal/config"
	"banking-platform/services/ledger-service/internal/handler"
	"banking-platform/services/ledger-service/internal/repository"
	"banking-platform/services/ledger-service/internal/service"
	"banking-platform/services/ledger-service/migrations"
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

	srv := httpserver.New(httpserver.Config{
		Host: cfg.HTTP.Host, Port: cfg.HTTP.Port,
		ReadTimeout: cfg.HTTP.ReadTimeout, WriteTimeout: cfg.HTTP.WriteTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, cfg.ServiceName, log)
	metrics.Mount(srv.Router())
	checker.Mount(srv.Router())
	handler.NewHTTP(svc).Register(srv.Router())
	checker.SetReady(true)

	var grpcSrv *grpcserver.Server
	if cfg.GRPC.Enabled {
		grpcSrv = grpcserver.New(grpcserver.Config{Host: cfg.GRPC.Host, Port: cfg.GRPC.Port}, log)
		ledgerv1.RegisterLedgerServiceServer(grpcSrv.Server(), handler.NewGRPC(svc))
	}

	errCh := make(chan error, 2)
	go func() { errCh <- srv.Start() }()
	if grpcSrv != nil {
		go func() { errCh <- grpcSrv.Start() }()
	}
	log.Info("ledger-service started", zap.Int("http_port", cfg.HTTP.Port), zap.Bool("grpc", cfg.GRPC.Enabled))

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
	log.Info("ledger-service stopped cleanly")
	return nil
}
