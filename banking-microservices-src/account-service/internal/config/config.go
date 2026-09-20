// Package config holds account-service configuration.
package config

import platform "banking-platform/pkg/config"

// Config embeds the shared platform Base and adds downstream service addresses.
type Config struct {
	platform.Base

	// LedgerGRPCAddr is the address of ledger-service's gRPC endpoint.
	LedgerGRPCAddr string `env:"LEDGER_GRPC_ADDR" envDefault:"localhost:9090"`
}
