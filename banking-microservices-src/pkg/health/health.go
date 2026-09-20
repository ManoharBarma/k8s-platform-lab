// Package health exposes Kubernetes-style liveness, readiness and startup
// endpoints backed by pluggable dependency checks (database, cache, broker...).
package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Check reports whether a dependency is healthy.
type Check func(ctx context.Context) error

// Checker aggregates readiness checks and a ready flag toggled by the service
// once it has finished starting up.
type Checker struct {
	mu     sync.RWMutex
	checks map[string]Check
	ready  bool
}

// New returns an empty Checker.
func New() *Checker {
	return &Checker{checks: make(map[string]Check)}
}

// Register adds a named readiness check.
func (c *Checker) Register(name string, check Check) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.checks[name] = check
}

// SetReady marks the service ready (or not) to receive traffic.
func (c *Checker) SetReady(ready bool) {
	c.mu.Lock()
	c.ready = ready
	c.mu.Unlock()
}

// Mount registers /healthz (liveness), /readyz (readiness) and /startupz
// (startup) on the router.
func (c *Checker) Mount(r gin.IRouter) {
	r.GET("/healthz", c.liveness)
	r.GET("/readyz", c.readiness)
	r.GET("/startupz", c.readiness)
}

func (c *Checker) liveness(gc *gin.Context) {
	gc.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (c *Checker) readiness(gc *gin.Context) {
	c.mu.RLock()
	ready := c.ready
	checks := make(map[string]Check, len(c.checks))
	for k, v := range c.checks {
		checks[k] = v
	}
	c.mu.RUnlock()

	results := gin.H{}
	healthy := ready
	for name, check := range checks {
		ctx, cancel := context.WithTimeout(gc.Request.Context(), 3*time.Second)
		if err := check(ctx); err != nil {
			results[name] = err.Error()
			healthy = false
		} else {
			results[name] = "ok"
		}
		cancel()
	}

	code, status := http.StatusOK, "ready"
	if !healthy {
		code, status = http.StatusServiceUnavailable, "not_ready"
	}
	gc.JSON(code, gin.H{"status": status, "checks": results})
}
