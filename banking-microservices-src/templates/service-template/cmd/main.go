// Command service-template is the golden-path entrypoint every banking
// microservice is scaffolded from. It wires the shared platform kit
// (config, logging, telemetry, HTTP server, health, metrics, persistence)
// into a runnable service with graceful startup and shutdown.
package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	platformcfg "banking-platform/pkg/config"
	"banking-platform/pkg/health"
	"banking-platform/pkg/httpserver"
	"banking-platform/pkg/logger"
	"banking-platform/pkg/metrics"
	"banking-platform/pkg/postgres"
	"banking-platform/pkg/telemetry"

	"banking-platform/templates/service-template/internal/config"
	"banking-platform/templates/service-template/internal/handler"
	"banking-platform/templates/service-template/internal/repository"
	"banking-platform/templates/service-template/internal/service"
	"banking-platform/templates/service-template/migrations"
)

func main() {
	if err := run(); err != nil {
		// The logger may not be up yet; fail loudly.
		panic(err)
	}
}

func run() error {
	// 1. Configuration.
	var cfg config.Config
	if err := platformcfg.Load(&cfg); err != nil {
		return err
	}

	// 2. Structured logging.
	log, err := logger.New(cfg.Log.Level, cfg.Log.Format, cfg.ServiceName, cfg.Version, cfg.Environment)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 3. Telemetry (traces + metrics via OTLP). Non-fatal if the collector
	// is unavailable — the service still runs.
	tp, err := telemetry.Setup(ctx, telemetry.Config{
		Enabled:          cfg.Telemetry.Enabled,
		ServiceName:      cfg.ServiceName,
		ServiceVersion:   cfg.Version,
		Environment:      cfg.Environment,
		OTLPEndpoint:     cfg.Telemetry.OTLPEndpoint,
		Insecure:         cfg.Telemetry.Insecure,
		TraceSampleRatio: cfg.Telemetry.TraceSampleRatio,
	})
	if err != nil {
		log.Warn("telemetry setup failed; continuing without it", zap.Error(err))
		tp = &telemetry.Provider{}
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(shutdownCtx)
	}()

	// 4. Persistence. Falls back to in-memory when Postgres is disabled.
	checker := health.New()
	var repo service.Repository = repository.NewInMemory()
	if cfg.Postgres.Enabled {
		pool, err := postgres.New(ctx, postgres.Config{
			DSN:      cfg.Postgres.DSN(),
			MaxConns: cfg.Postgres.MaxConns,
			MinConns: cfg.Postgres.MinConns,
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
	} else {
		log.Info("postgres disabled; using in-memory repository")
	}

	// 5. Application wiring.
	svc := service.New(repo, log)

	srv := httpserver.New(httpserver.Config{
		Host:            cfg.HTTP.Host,
		Port:            cfg.HTTP.Port,
		ReadTimeout:     cfg.HTTP.ReadTimeout,
		WriteTimeout:    cfg.HTTP.WriteTimeout,
		ShutdownTimeout: cfg.HTTP.ShutdownTimeout,
	}, cfg.ServiceName, log)

	metrics.Mount(srv.Router())
	checker.Mount(srv.Router())
	handler.New(svc, log).Register(srv.Router())
	checker.SetReady(true)

	// 6. Serve until signalled.
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Start() }()
	log.Info("service started", zap.String("service", cfg.ServiceName), zap.Int("port", cfg.HTTP.Port))

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	// 7. Graceful shutdown.
	checker.SetReady(false)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Info("service stopped cleanly")
	return nil
}
