package registry

import (
	"sort"
	"sync"
	"time"
)

type ServiceStatus struct {
	Addr  string
	Alive bool
}

type ServiceRegistry struct {
	mu       sync.RWMutex
	services map[string]time.Time
	ttl      time.Duration
	idx      int
}

func NewServiceRegistry(ttl time.Duration) *ServiceRegistry {
	return &ServiceRegistry{
		services: make(map[string]time.Time),
		ttl:      ttl,
	}
}

func (r *ServiceRegistry) Register(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services[addr] = time.Now()
}

func (r *ServiceRegistry) Heartbeat(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.services[addr]; exists {
		r.services[addr] = time.Now()
	}
}

func (r *ServiceRegistry) Status() []ServiceStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	out := []ServiceStatus{}
	for addr, last := range r.services {
		out = append(out, ServiceStatus{
			Addr:  addr,
			Alive: now.Sub(last) <= r.ttl,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Addr < out[j].Addr
	})
	return out
}

func (r *ServiceRegistry) NextInstance() string {
	all := r.Status()
	validAddrs := []string{}
	for _, s := range all {
		if s.Alive {
			validAddrs = append(validAddrs, s.Addr)
		}
	}

	count := len(validAddrs)
	if count == 0 {
		return ""
	}

	if r.idx >= count {
		r.idx = 0
	}

	addr := validAddrs[r.idx]
	r.idx = (r.idx + 1) % count
	return addr
}
