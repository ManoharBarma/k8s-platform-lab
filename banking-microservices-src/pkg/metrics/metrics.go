// Package metrics exposes the Prometheus scrape endpoint. Application metrics
// are registered via the default Prometheus registry (see httpserver
// middleware and per-service instrumentation).
package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Mount exposes GET /metrics for Prometheus scraping.
func Mount(r gin.IRouter) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
