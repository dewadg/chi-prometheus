package chi_prometheus

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

//go:generate mockgen -source middleware.go -destination mocks.go -package chi_prometheus
//go:generate mockgen -destination prometheus_mocks.go -package chi_prometheus github.com/prometheus/client_golang/prometheus Counter,Observer

type Config struct {
	ServiceName  string
	ServiceLabel string
	MetricPrefix string
}

func New(cfg Config) func(http.Handler) http.Handler {
	if cfg.ServiceLabel == "" {
		cfg.ServiceLabel = "service"
	}
	if cfg.MetricPrefix == "" {
		cfg.MetricPrefix = "http"
	}

	m := &middleware{
		serviceName: cfg.ServiceName,
		requestCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: cfg.MetricPrefix + "_requests_total",
			Help: "Total HTTP requests received",
		}, []string{cfg.ServiceLabel, "code", "method", "path"}),
		latencyHistogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: cfg.MetricPrefix + "_request_duration_seconds",
			Help: "Duration of HTTP requests performed",
		}, []string{cfg.ServiceLabel, "code", "method", "path"}),
	}

	return m.Handle
}

type counter interface {
	WithLabelValues(lvs ...string) prometheus.Counter
}

type histogram interface {
	WithLabelValues(lvs ...string) prometheus.Observer
}

type middleware struct {
	serviceName      string
	requestCounter   counter
	latencyHistogram histogram
}

func (m *middleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := &responseWriter{
			ResponseWriter: w,
		}

		start := time.Now()
		next.ServeHTTP(ww, r)
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(ww.statusCode)

		path := r.URL.Path
		if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
			path = routeCtx.RoutePattern()
		}

		m.requestCounter.WithLabelValues(m.serviceName, statusCode, r.Method, path).Inc()
		m.latencyHistogram.WithLabelValues(m.serviceName, statusCode, r.Method, path).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
