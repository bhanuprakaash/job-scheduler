package worker

import (
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

type registryEntry struct {
	handler Handler
	limiter *rate.Limiter
}

type Registry struct {
	mu      sync.RWMutex
	entries map[string]registryEntry
}

func NewRegistry() *Registry {
	return &Registry{
		entries: make(map[string]registryEntry),
	}
}

func (r *Registry) Register(jobType string, handler Handler, eventsPerSecond int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var limiter *rate.Limiter
	if eventsPerSecond > 0 {
		limiter = rate.NewLimiter(rate.Limit(eventsPerSecond), eventsPerSecond)
	} else {
		limiter = rate.NewLimiter(rate.Inf, 0)
	}

	r.entries[jobType] = registryEntry{
		handler: handler,
		limiter: limiter,
	}
}

func (r *Registry) Get(jobType string) (Handler, *rate.Limiter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.entries[jobType]
	if !exists {
		return nil, nil, fmt.Errorf("no handler registered for job type: %s", jobType)
	}

	return entry.handler, entry.limiter, nil
}

func (r *Registry) Has(jobType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// _, exists := r.handlers[jobType]
	_, exists := r.entries[jobType]
	return exists
}
