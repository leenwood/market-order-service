package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"market-order-service/internal/platform/metrics"
)

type Config struct {
	Target     string
	BaseURL    string
	MaxRetries int
	Metrics    *metrics.Metrics
}

type Response struct {
	StatusCode int
	Body       []byte
}

type cbState int

const (
	stateClosed cbState = iota
	stateOpen
	stateHalfOpen
)

type circuitBreaker struct {
	mu               sync.Mutex
	state            cbState
	consecutiveFails int
	openUntil        time.Time
	halfOpenProbes   int

	maxFailures   int
	openTimeout   time.Duration
	maxHalfProbes int
}

func newCircuitBreaker() *circuitBreaker {
	return &circuitBreaker{
		maxFailures:   5,
		openTimeout:   30 * time.Second,
		maxHalfProbes: 2,
	}
}

var errCircuitOpen = fmt.Errorf("circuit breaker open")

func (cb *circuitBreaker) allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case stateClosed:
		return true
	case stateOpen:
		if time.Now().After(cb.openUntil) {
			cb.state = stateHalfOpen
			cb.halfOpenProbes = 0
			return true
		}
		return false
	case stateHalfOpen:
		return cb.halfOpenProbes < cb.maxHalfProbes
	}
	return false
}

func (cb *circuitBreaker) onSuccess(logFn func(string)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == stateHalfOpen {
		cb.halfOpenProbes++
		if cb.halfOpenProbes >= cb.maxHalfProbes {
			cb.state = stateClosed
			cb.consecutiveFails = 0
			logFn("circuit breaker closed")
		}
	} else {
		cb.consecutiveFails = 0
	}
}

func (cb *circuitBreaker) onFailure(logFn func(string)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.consecutiveFails++
	if cb.state == stateHalfOpen || cb.consecutiveFails >= cb.maxFailures {
		cb.state = stateOpen
		cb.openUntil = time.Now().Add(cb.openTimeout)
		logFn("circuit breaker opened")
	}
}

type Client struct {
	cfg  Config
	http *http.Client
	cb   *circuitBreaker
}

func New(cfg Config) *Client {
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: 10 * time.Second},
		cb:   newCircuitBreaker(),
	}
}

func jitteredBackoff(attempt int) time.Duration {
	maxDur := float64(30 * time.Second)
	ceil := math.Min(maxDur, float64(100*time.Millisecond)*math.Pow(2, float64(attempt)))
	if ceil <= 0 {
		return 0
	}
	return time.Duration(rand.Int63n(int64(ceil)))
}

func isRetryable(code int) bool {
	return code == 429 || code == 502 || code == 503 || code == 504 || code >= 500
}

func (c *Client) Do(ctx context.Context, method, path string, body []byte) (*Response, error) {
	tracer := otel.Tracer("httpclient")
	ctx, span := tracer.Start(ctx, fmt.Sprintf("%s %s %s", c.cfg.Target, method, path))
	defer span.End()

	span.SetAttributes(
		attribute.String(string(semconv.HTTPRequestMethodKey), method),
		attribute.String("http.url", c.cfg.BaseURL+path),
	)

	var lastErr error
	for attempt := 0; attempt <= c.cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(jitteredBackoff(attempt - 1)):
			}
		}

		if !c.cb.allow() {
			return nil, errCircuitOpen
		}

		var reqBody io.Reader
		if len(body) > 0 {
			reqBody = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, reqBody)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		if len(body) > 0 {
			req.Header.Set("Content-Type", "application/json")
		}
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

		resp, err := c.http.Do(req)
		statusStr := "error"
		if resp != nil {
			statusStr = strconv.Itoa(resp.StatusCode)
		}
		if c.cfg.Metrics != nil {
			c.cfg.Metrics.HTTPOutbound.WithLabelValues(c.cfg.Target, method, statusStr).Inc()
		}

		if err != nil {
			c.cb.onFailure(func(string) {})
			lastErr = err
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if isRetryable(resp.StatusCode) {
			c.cb.onFailure(func(string) {})
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			continue
		}

		c.cb.onSuccess(func(string) {})
		span.SetAttributes(attribute.Int(string(semconv.HTTPResponseStatusCodeKey), resp.StatusCode))
		return &Response{StatusCode: resp.StatusCode, Body: respBody}, nil
	}
	return nil, fmt.Errorf("all retries exhausted: %w", lastErr)
}
