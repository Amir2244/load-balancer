package loadbalancer

import (
	"hash/fnv"
	"load-balancer/server"
	"log"
	"sort"
	"sync/atomic"
)

var currentIndex uint32

const (
	virtualNodes = 20
)

func (lb *LoadBalancer) LeastConnections() *server.Server {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	var selected *server.Server
	//  maximum positive integer value possible on the system
	minConns := int(^uint(0) >> 1)

	for _, srv := range lb.servers {
		if !srv.IsHealthy() {
			continue
		}

		active := srv.GetActiveConns()
		if active < minConns {
			minConns = active
			selected = srv
		}
	}
	return selected
}
func (lb *LoadBalancer) RoundRobin() *server.Server {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	healthy := lb.GetHealthyServers()
	if len(healthy) == 0 {
		return nil
	}

	index := atomic.AddUint32(&currentIndex, 1)
	return healthy[index%uint32(len(healthy))]
}

func (lb *LoadBalancer) ConsistentHash(key string) *server.Server {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if len(lb.servers) == 0 {
		return nil
	}

	hash := fnv.New32a()
	hash.Write([]byte(key))
	keyHash := hash.Sum32()

	var ring []uint32
	hashMap := make(map[uint32]*server.Server)

	for _, srv := range lb.servers {
		if !srv.IsHealthy() {
			continue
		}

		for i := 0; i < virtualNodes; i++ {
			hash := fnv.New32a()
			_, err := hash.Write([]byte(srv.Host + string(rune(i))))
			if err != nil {
				log.Printf("Error Occurred %s", err)
				return nil
			}
			h := hash.Sum32()
			ring = append(ring, h)
			hashMap[h] = srv
		}
	}

	if len(ring) == 0 {
		return nil
	}

	sort.Slice(ring, func(i, j int) bool {
		return ring[i] < ring[j]
	})

	for _, h := range ring {
		if h >= keyHash {
			return hashMap[h]
		}
	}

	return hashMap[ring[0]]
}
