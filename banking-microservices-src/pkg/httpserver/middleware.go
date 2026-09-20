package httpserver

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"banking-platform/pkg/apierror"
	"banking-platform/pkg/logger"
)

// RequestIDHeader is the header used to carry a correlation ID.
const RequestIDHeader = "X-Request-ID"

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests processed, labeled by status.",
	}, []string{"service", "method", "path", "status"})

	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "path"})
)

// RequestID ensures every request carries a correlation ID, echoed back to the
// caller and stored in the context for logging.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if id == "" {
			id = uuid.NewString()
		}
		c.Set("request_id", id)
		c.Writer.Header().Set(RequestIDHeader, id)
		c.Next()
	}
}

// LoggingMiddleware logs one structured line per request, correlated with the
// active trace.
func LoggingMiddleware(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		logger.WithTrace(c.Request.Context(), l).Info("http_request",
			zap.String("request_id", c.GetString("request_id")),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// RecoveryMiddleware converts panics into 500 responses and logs them with the
// trace context.
func RecoveryMiddleware(l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithTrace(c.Request.Context(), l).Error("panic recovered",
					zap.Any("panic", r),
					zap.String("path", c.Request.URL.Path),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					apierror.ErrInternal("internal server error"))
			}
		}()
		c.Next()
	}
}

// PrometheusMiddleware records request counts and latencies.
func PrometheusMiddleware(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}
		status := strconv.Itoa(c.Writer.Status())
		httpRequests.WithLabelValues(service, c.Request.Method, path, status).Inc()
		httpDuration.WithLabelValues(service, c.Request.Method, path).Observe(time.Since(start).Seconds())
	}
}
