// Package config holds document-service configuration.
package config

import platform "banking-platform/pkg/config"

// Config embeds the shared platform Base (which includes S3 settings).
type Config struct {
	platform.Base
}
