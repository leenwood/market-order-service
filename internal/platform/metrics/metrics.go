// internal/platform/metrics/metrics.go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Metrics struct {
	Registry     *prometheus.Registry
	HTTPRequests *prometheus.CounterVec
	HTTPDuration *prometheus.HistogramVec
	HTTPOutbound *prometheus.CounterVec
}

func New() *Metrics {
	r := prometheus.NewRegistry()
	r.MustRegister(collectors.NewGoCollector())
	r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	httpReqs := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_http_requests_total",
		Help: "Total inbound HTTP requests.",
	}, []string{"method", "path", "status"})

	httpDur := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "platform_http_request_duration_seconds",
		Help:    "Inbound HTTP request duration in seconds.",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"method", "path"})

	httpOut := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "platform_http_outbound_requests_total",
		Help: "Total outbound HTTP requests.",
	}, []string{"target", "method", "status"})

	r.MustRegister(httpReqs, httpDur, httpOut)

	return &Metrics{
		Registry:     r,
		HTTPRequests: httpReqs,
		HTTPDuration: httpDur,
		HTTPOutbound: httpOut,
	}
}
