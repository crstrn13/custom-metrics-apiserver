package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// MetricsHandler handles custom metrics endpoints
type MetricsHandler struct {
	requestCount uint64
}

// CustomMetric represents a Kubernetes custom metric
type CustomMetric struct {
	MetricName string `json:"metricName"`
	Value      int64  `json:"value"`
	Timestamp  string `json:"timestamp"`
}

// NewMetricsHandler creates a new MetricsHandler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// IncrementRequestCount increments the request counter
func (h *MetricsHandler) IncrementRequestCount() {
	atomic.AddUint64(&h.requestCount, 1)
}

// ServeHTTP implements http.Handler
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	value := atomic.LoadUint64(&h.requestCount)

	metric := CustomMetric{
		MetricName: "http_requests_total",
		Value:      int64(value),
		Timestamp:  r.Header.Get("Date"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metric)
}
