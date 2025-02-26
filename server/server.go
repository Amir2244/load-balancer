package server

import (
	"log"
	"sync"
)

type Server struct {
	Host           string
	MaxConnections int
	Mode           string
	activeConns    int
	healthy        bool
	mutex          sync.RWMutex
}

func NewServer(host string, maxConn int, mode string) *Server {
	return &Server{
		Host:           host,
		MaxConnections: maxConn,
		Mode:           mode,
		healthy:        false,
	}
}

func (s *Server) IncrementConnections() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.activeConns >= s.MaxConnections {
		log.Printf("Server %s reached max connections: %d", s.Host, s.MaxConnections)
		return false
	}
	s.activeConns++
	log.Printf("Server %s connections: %d/%d", s.Host, s.activeConns, s.MaxConnections)
	return true
}

func (s *Server) DecrementConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.activeConns > 0 {
		s.activeConns--
		log.Printf("Server %s connections: %d/%d", s.Host, s.activeConns, s.MaxConnections)
	}
}

func (s *Server) SetHealthy(status bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.healthy = status
}

func (s *Server) IsHealthy() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.healthy
}

func (s *Server) GetActiveConns() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.activeConns
}
