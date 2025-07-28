package models

// Request represents the core domain model for HTTP requests
type Request struct {
	Args    map[string][]string
	Headers map[string]string
	Origin  string
	URL     string
}

// NewRequest creates a new Request instance
func NewRequest(args map[string][]string, headers map[string]string, origin, url string) *Request {
	return &Request{
		Args:    args,
		Headers: headers,
		Origin:  origin,
		URL:     url,
	}
}
