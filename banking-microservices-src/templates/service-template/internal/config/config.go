// Package config holds this service's configuration. It embeds the shared
// platform Base config (HTTP, DB, Redis, Kafka, telemetry, logging) and adds
// any service-specific fields below.
package config

import platform "banking-platform/pkg/config"

// Config is the fully-resolved configuration for the service.
type Config struct {
	platform.Base

	// Service-specific configuration goes here, e.g.:
	// FeatureXEnabled bool `env:"FEATURE_X_ENABLED" envDefault:"false"`
}
