package main

import (
	"io"
	"load-balancer/config"
	"load-balancer/loadbalancer"
	"load-balancer/server"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	cfg, err := config.LoadConfig("config.yml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	var servers []*server.Server
	for _, srv := range cfg.Servers {
		servers = append(servers, server.NewServer(srv.Host, srv.MaxConnections, srv.Mode))
	}

	lb := loadbalancer.New(servers)

	interval, err := time.ParseDuration(cfg.HealthCheckInterval)
	if err != nil {
		log.Fatalf("Invalid health check interval: %v", err)
	}
	lb.StartHealthChecks(interval)

	for _, listenerCfg := range cfg.Listeners {
		go startListener(listenerCfg, lb)
	}

	select {}
}

func startListener(cfg config.ListenerConfig, lb *loadbalancer.LoadBalancer) {
	listener, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to start listener on %s: %v", cfg.ListenAddr, err)
	}

	log.Printf("Started %s listener on %s", cfg.Mode, cfg.ListenAddr)

	switch strings.ToLower(cfg.Mode) {
	case "http":
		serveHTTP(listener, lb, cfg)
	case "tcp":
		serveTCP(listener, lb, cfg)
	default:
		log.Fatalf("Unknown mode: %s", cfg.Mode)
	}
}

func serveHTTP(listener net.Listener, lb *loadbalancer.LoadBalancer, cfg config.ListenerConfig) {
	serv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			selectedServer := selectServer(lb, cfg.Algorithm, r.RemoteAddr)
			if selectedServer == nil {
				http.Error(w, "No healthy servers available", http.StatusServiceUnavailable)
				return
			}

			if !selectedServer.IncrementConnections() {
				http.Error(w, "Server at capacity", http.StatusServiceUnavailable)
				return
			}
			defer selectedServer.DecrementConnections()

			backendURL, err := url.Parse("http://" + selectedServer.Host)
			if err != nil {
				http.Error(w, "Invalid backend URL", http.StatusInternalServerError)
				return
			}

			proxy := httputil.NewSingleHostReverseProxy(backendURL)
			proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("Proxy error: %v", err)
				http.Error(w, "Backend serv error", http.StatusBadGateway)
			}
			proxy.ServeHTTP(w, r)
		}),
	}
	log.Fatal(serv.Serve(listener))
}

func serveTCP(listener net.Listener, lb *loadbalancer.LoadBalancer, cfg config.ListenerConfig) {
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		go handleTCPConnection(client, lb, cfg)
	}
}

func handleTCPConnection(client net.Conn, lb *loadbalancer.LoadBalancer, cfg config.ListenerConfig) {
	defer func(client net.Conn) {
		err := client.Close()
		if err != nil {
			log.Printf("Failed to close connection: %v", err)
		}
	}(client)

	selectedServer := selectServer(lb, cfg.Algorithm, client.RemoteAddr().String())
	if selectedServer == nil {
		log.Printf("No healthy servers available for TCP connection")
		return
	}

	if !selectedServer.IncrementConnections() {
		log.Printf("Connection rejected: server %s at capacity", selectedServer.Host)
		return
	}
	defer selectedServer.DecrementConnections()

	backend, err := net.DialTimeout("tcp", selectedServer.Host, 5*time.Second)
	if err != nil {
		log.Printf("Failed to connect to backend: %v", err)
		return
	}
	defer func(backend net.Conn) {
		err := backend.Close()
		if err != nil {
			log.Printf("Failed to close backend connection: %v", err)
		}
	}(backend)

	done := make(chan struct{})
	go func() {
		_, err := io.Copy(client, backend)
		if err != nil {
			return
		}
		done <- struct{}{}
	}()
	go func() {
		_, err := io.Copy(backend, client)
		if err != nil {
			return
		}
		done <- struct{}{}
	}()

	<-done
}

func selectServer(lb *loadbalancer.LoadBalancer, algorithm string, remoteAddr string) *server.Server {
	switch strings.ToLower(algorithm) {
	case "round_robin":
		return lb.RoundRobin()
	case "least_connections":
		return lb.LeastConnections()
	case "consistent_hash":
		return lb.ConsistentHash(remoteAddr)
	default:
		log.Printf("Unknown algorithm %s, using round robin", algorithm)
		return lb.RoundRobin()
	}
}
