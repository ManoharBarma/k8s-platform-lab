// Package grpcclient dials other services over gRPC with OpenTelemetry tracing
// so cross-service calls appear as connected spans in Tempo.
package grpcclient

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial creates a client connection to target (host:port). The connection is
// lazy (grpc.NewClient) and instrumented for tracing. In-cluster traffic is
// plaintext; TLS is terminated at the mesh/ingress layer.
func Dial(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
}
