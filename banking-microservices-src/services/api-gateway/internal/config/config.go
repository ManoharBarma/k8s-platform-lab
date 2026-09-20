// Package config holds api-gateway configuration.
package config

import platform "banking-platform/pkg/config"

// Config embeds the shared platform Base and adds gateway-specific settings.
type Config struct {
	platform.Base

	// CORS
	AllowedOrigins string `env:"CORS_ALLOWED_ORIGINS" envDefault:"*"`

	// Rate limiting (global token bucket).
	RateLimitRPS   float64 `env:"RATE_LIMIT_RPS" envDefault:"100"`
	RateLimitBurst int     `env:"RATE_LIMIT_BURST" envDefault:"200"`

	// Upstreams are addressed as <scheme>://<service-name>:<port> — matching the
	// service DNS names in docker-compose and Kubernetes.
	UpstreamScheme string `env:"UPSTREAM_SCHEME" envDefault:"http"`
	UpstreamPort   string `env:"UPSTREAM_PORT" envDefault:"8080"`
}
