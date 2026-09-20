// Package config holds loan-service configuration.
package config

import platform "banking-platform/pkg/config"

// Config embeds the shared platform Base and adds downstream addresses.
type Config struct {
	platform.Base

	LedgerGRPCAddr string `env:"LEDGER_GRPC_ADDR" envDefault:"localhost:9090"`
}
