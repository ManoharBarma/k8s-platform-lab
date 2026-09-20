// Package config holds auth-service configuration.
package config

import platform "banking-platform/pkg/config"

// Config embeds the shared platform Base (HTTP, gRPC, Postgres, Redis, Auth...).
type Config struct {
	platform.Base
}
