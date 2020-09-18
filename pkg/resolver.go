package pkg

import "time"

// Resolver resolves an upstream IP address for a given upstream host
type Resolver interface {
	// Start any polling or connections required to resolve
	Start()

	// Get an IP address using an asynchronous operation
	Get(upstream string, got chan<- string, timeout time.Duration)
}
