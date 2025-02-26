package loadbalancer

import (
	"load-balancer/server"
	"sync"
	"time"
)

// LoadBalancer represents a load balancer that manages multiple servers and their health status
type LoadBalancer struct {
	servers []*server.Server //  servers managed by the load balancer
	mutex   sync.RWMutex     // for thread-safe operations on servers
}

// creates and returns a new LoadBalancer instance with the given servers
func New(servers []*server.Server) *LoadBalancer {
	return &LoadBalancer{
		servers: servers,
	}
}

// StartHealthChecks initiates periodic health checks for all servers at the specified interval
func (lb *LoadBalancer) StartHealthChecks(interval time.Duration) {
	for _, srv := range lb.servers {
		srv.StartHealthCheck(interval)
	}
}

// GetHealthyServers returns a list of servers that are currently healthy
// Thread-safe operation is ensured using a read lock
func (lb *LoadBalancer) GetHealthyServers() []*server.Server {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	var healthy []*server.Server
	for _, srv := range lb.servers {
		if srv.IsHealthy() {
			healthy = append(healthy, srv)
		}
	}
	return healthy
}
