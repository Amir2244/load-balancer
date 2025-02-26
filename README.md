# Go Load Balancer

A high-performance load balancer written in Go that supports both TCP and HTTP protocols, similar to HAProxy. This load balancer provides multiple algorithms for traffic distribution and real-time health monitoring of backend servers.

## Features

### Multiple Protocols
- HTTP Load Balancing
- TCP Load Balancing
- Protocol-specific handling and optimizations

### Load Balancing Algorithms
- Round Robin
- Least Connections
- Consistent Hashing

### Health Monitoring
- Configurable health check intervals
- Automatic server health tracking
- Health status logging to file and console

### Connection Management
- Maximum connection limits per server
- Active connection tracking
- Graceful connection handling

## Configuration
The load balancer is configured using YML. Example configuration:


health_check_interval: "10s"

listeners:
  - listen_addr: ":8000"
    mode: "http"
    algorithm: "round_robin"
  - listen_addr: ":9000"
    mode: "tcp"
    algorithm: "least_connections"

servers:
  - host: "localhost:8082"
    max_connections: 100
    mode: "tcp"  mode: "tcp"