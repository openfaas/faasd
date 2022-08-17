package bootstrap

import (
	"net/http"
	"strconv"
	"time"

	"github.com/openfaas/faas-provider/httputil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// httpMetrics is for recording R.E.D. metrics for system endpoint calls
// for HTTP status code, method, duration and path.
type httpMetrics struct {
	// RequestsTotal is a Prometheus counter vector partitioned by method and status.
	RequestsTotal *prometheus.CounterVec

	// RequestDurationHistogram is a Prometheus summary vector partitioned by method and status.
	RequestDurationHistogram *prometheus.HistogramVec
}

// newHttpMetrics initialises a new httpMetrics struct for
// recording R.E.D. metrics for system endpoint calls
func newHttpMetrics() *httpMetrics {
	return &httpMetrics{
		RequestsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "provider",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		}, []string{"code", "method", "path"}),
		RequestDurationHistogram: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "provider",
			Name:      "http_request_duration_seconds",
			Help:      "Seconds spent serving HTTP requests.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"code", "method", "path"}),
	}
}

func (hm *httpMetrics) InstrumentHandler(next http.Handler, pathOverride string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		ww := httputil.NewHttpWriteInterceptor(w)
		next.ServeHTTP(ww, r)
		duration := time.Since(start)

		path := r.URL.Path
		if len(pathOverride) > 0 {
			path = pathOverride
		}

		defer func() {
			hm.RequestsTotal.With(
				prometheus.Labels{"code": strconv.Itoa(ww.Status()),
					"method": r.Method,
					"path":   path,
				}).
				Inc()
		}()

		defer func() {
			hm.RequestDurationHistogram.With(
				prometheus.Labels{"code": strconv.Itoa(ww.Status()),
					"method": r.Method,
					"path":   path,
				}).
				Observe(duration.Seconds())
		}()
	}
}
