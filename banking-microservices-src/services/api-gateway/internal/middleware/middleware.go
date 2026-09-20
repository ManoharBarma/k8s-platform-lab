// Package middleware provides gateway-specific middleware: CORS and rate limiting.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// CORS allows the browser SPA to call the gateway.
func CORS(allowedOrigins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		switch {
		case allowedOrigins == "*":
			c.Header("Access-Control-Allow-Origin", "*")
		case origin != "" && originAllowed(origin, allowedOrigins):
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,Idempotency-Key,X-Request-ID")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func originAllowed(origin, allowed string) bool {
	for _, a := range splitComma(allowed) {
		if a == origin {
			return true
		}
	}
	return false
}

func splitComma(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ',' {
			out = append(out, cur)
			cur = ""
			continue
		}
		if r != ' ' {
			cur += string(r)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

// RateLimit applies a global token-bucket limit.
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": "rate_limited", "message": "too many requests"})
			return
		}
		c.Next()
	}
}
