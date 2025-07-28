package metrics

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

// Provider handles custom metrics
type Provider struct {
	requestCount uint64
}

// NewProvider creates a new metrics provider
func NewProvider() *Provider {
	return &Provider{}
}

// IncrementRequestCount increments the request counter
func (p *Provider) IncrementRequestCount() {
	atomic.AddUint64(&p.requestCount, 1)
}

// ServeHTTP implements http.Handler
func (p *Provider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	value := atomic.LoadUint64(&p.requestCount)

	// Format response according to custom.metrics.k8s.io API
	response := map[string]interface{}{
		"kind":       "MetricValueList",
		"apiVersion": "custom.metrics.k8s.io/v1beta1",
		"metadata": map[string]interface{}{
			"selfLink": r.URL.Path,
		},
		"items": []map[string]interface{}{
			{
				"describedObject": map[string]string{
					"kind":      "Pod",
					"name":      "mock-backend",
					"namespace": "kuadrant",
				},
				"metricName": "http_requests_total",
				"value":      value,
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
