package ports

// MetricsService defines the port for handling metrics operations
type MetricsService interface {
	// IncrementRequestCount increments the total request counter
	IncrementRequestCount()
}
