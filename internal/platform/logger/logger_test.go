package logger_test

import (
	"context"
	"testing"

	"market-order-service/internal/platform/logger"
)

func TestWithRequestID(t *testing.T) {
	ctx := logger.WithRequestID(context.Background(), "req-123")
	if got := logger.RequestIDFromContext(ctx); got != "req-123" {
		t.Fatalf("want req-123, got %s", got)
	}
}

func TestFromContextAddsFields(t *testing.T) {
	ctx := logger.WithRequestID(context.Background(), "req-abc")
	ctx = logger.WithTraceID(ctx, "trace-xyz")
	base := logger.New("info", "json")
	l := logger.FromContext(ctx, base)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}
