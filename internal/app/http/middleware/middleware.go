package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"

	"market-order-service/internal/platform/logger"
	"market-order-service/internal/platform/metrics"
)

func GinRecover(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				log.ErrorContext(c.Request.Context(), "panic recovered",
					"error", fmt.Sprintf("%v", rec),
					"stack", string(debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError,
					gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

func GinLogger(log *slog.Logger, m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		span := trace.SpanFromContext(c.Request.Context())
		tid := span.SpanContext().TraceID().String()
		ctx := logger.WithTraceID(c.Request.Context(), tid)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		dur := time.Since(start)
		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		logger.FromContext(ctx, log).Info("request",
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", dur.Milliseconds(),
		)
		if m != nil {
			m.HTTPRequests.WithLabelValues(method, path, statusStr).Inc()
			m.HTTPDuration.WithLabelValues(method, path).Observe(dur.Seconds())
		}
	}
}

func GinRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		ctx := logger.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Request-ID", rid)
		c.Next()
	}
}

func GinMaxBodySize(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
