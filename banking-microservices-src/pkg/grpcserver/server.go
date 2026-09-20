// Package grpcserver provides a gRPC server with OpenTelemetry stats handler,
// the standard gRPC health service, server reflection and graceful shutdown.
package grpcserver

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Config controls the gRPC server bind address.
type Config struct {
	Host string
	Port int
}

// Server wraps a *grpc.Server with health + tracing preconfigured.
type Server struct {
	cfg    Config
	server *grpc.Server
	health *health.Server
	log    *zap.Logger
}

// New builds a gRPC server. Register service implementations on Server().
func New(cfg Config, log *zap.Logger) *Server {
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	hs := health.NewServer()
	healthpb.RegisterHealthServer(s, hs)
	reflection.Register(s)
	return &Server{cfg: cfg, server: s, health: hs, log: log}
}

// Server exposes the underlying *grpc.Server for service registration.
func (s *Server) Server() *grpc.Server { return s.server }

// Start binds the listener and serves until shut down.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port))
	if err != nil {
		return fmt.Errorf("grpc listen: %w", err)
	}
	s.health.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	s.log.Info("grpc server listening", zap.String("addr", lis.Addr().String()))
	return s.server.Serve(lis)
}

// Shutdown gracefully stops the server, forcing a stop if ctx expires first.
func (s *Server) Shutdown(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()
	select {
	case <-ctx.Done():
		s.server.Stop()
	case <-stopped:
	}
	return nil
}
