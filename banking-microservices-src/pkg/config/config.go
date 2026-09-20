// Package config provides environment-driven configuration shared by every
// microservice in the banking platform. Service-specific configuration should
// embed Base and add its own fields.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Base holds configuration common to every microservice.
type Base struct {
	ServiceName string `env:"SERVICE_NAME" envDefault:"banking-service"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	Version     string `env:"VERSION" envDefault:"0.1.0"`

	HTTP      HTTP
	GRPC      GRPC
	Postgres  Postgres
	Redis     Redis
	Kafka     Kafka
	Auth      Auth
	S3        S3
	Telemetry Telemetry
	Log       Log
}

// HTTP configures the REST/HTTP server.
type HTTP struct {
	Host            string        `env:"HTTP_HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"HTTP_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"15s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" envDefault:"20s"`
}

// GRPC configures the gRPC server (disabled by default; enable per service).
type GRPC struct {
	Enabled bool   `env:"GRPC_ENABLED" envDefault:"false"`
	Host    string `env:"GRPC_HOST" envDefault:"0.0.0.0"`
	Port    int    `env:"GRPC_PORT" envDefault:"9090"`
}

// Postgres configures the connection to the shared PostgreSQL database.
type Postgres struct {
	Enabled  bool   `env:"POSTGRES_ENABLED" envDefault:"true"`
	Host     string `env:"POSTGRES_HOST" envDefault:"localhost"`
	Port     int    `env:"POSTGRES_PORT" envDefault:"5432"`
	User     string `env:"POSTGRES_USER" envDefault:"banking"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"banking"`
	Database string `env:"POSTGRES_DB" envDefault:"banking"`
	SSLMode  string `env:"POSTGRES_SSLMODE" envDefault:"disable"`
	MaxConns int32  `env:"POSTGRES_MAX_CONNS" envDefault:"10"`
	MinConns int32  `env:"POSTGRES_MIN_CONNS" envDefault:"2"`
}

// DSN renders a libpq-style connection string.
func (p Postgres) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode)
}

// Redis configures the connection to Redis.
type Redis struct {
	Enabled  bool   `env:"REDIS_ENABLED" envDefault:"false"`
	Addr     string `env:"REDIS_ADDR" envDefault:"localhost:6379"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

// Kafka configures the connection to the Kafka broker(s).
type Kafka struct {
	Enabled bool     `env:"KAFKA_ENABLED" envDefault:"false"`
	Brokers []string `env:"KAFKA_BROKERS" envDefault:"localhost:9092" envSeparator:","`
	GroupID string   `env:"KAFKA_GROUP_ID" envDefault:""`
}

// Auth configures JWT issuance and validation. The secret is shared across
// services so any service can validate a token issued by auth-service.
type Auth struct {
	JWTSecret  string        `env:"JWT_SECRET" envDefault:"dev-secret-change-me"`
	JWTIssuer  string        `env:"JWT_ISSUER" envDefault:"banking-platform"`
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
}

// S3 configures object storage. Endpoint/UsePathStyle allow pointing at a
// MinIO instance locally instead of real Amazon S3.
type S3 struct {
	Enabled      bool   `env:"S3_ENABLED" envDefault:"false"`
	Region       string `env:"AWS_REGION" envDefault:"us-east-1"`
	Bucket       string `env:"S3_BUCKET" envDefault:"banking-documents"`
	Endpoint     string `env:"S3_ENDPOINT" envDefault:""`
	AccessKey    string `env:"AWS_ACCESS_KEY_ID" envDefault:""`
	SecretKey    string `env:"AWS_SECRET_ACCESS_KEY" envDefault:""`
	UsePathStyle bool   `env:"S3_USE_PATH_STYLE" envDefault:"false"`
}

// Telemetry configures OpenTelemetry export (traces + metrics) via OTLP/gRPC.
type Telemetry struct {
	Enabled          bool    `env:"OTEL_ENABLED" envDefault:"true"`
	OTLPEndpoint     string  `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4317"`
	Insecure         bool    `env:"OTEL_EXPORTER_OTLP_INSECURE" envDefault:"true"`
	TraceSampleRatio float64 `env:"OTEL_TRACE_SAMPLE_RATIO" envDefault:"1.0"`
}

// Log configures structured logging.
type Log struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// Load reads configuration from environment variables into cfg. It first loads
// a .env file if present (ignored when absent) so local development is easy.
func Load[T any](cfg *T) error {
	_ = godotenv.Load()
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	return nil
}
