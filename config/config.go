package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

// Config represents the main configuration structure for the load balancer
type Config struct {
	// Servers contains the configuration for backend servers
	Servers []ServerConfig `yaml:"servers"`
	// Listeners defines the load balancer listening ports and modes
	Listeners []ListenerConfig `yaml:"listeners"`
	// HealthCheckInterval specifies how often to check backend server health
	HealthCheckInterval string `yaml:"health_check_interval"`
}

// ServerConfig defines the configuration for a single backend server
type ServerConfig struct {
	// Host is the address of the backend server
	Host string `yaml:"host"`
	// MaxConnections limits the number of simultaneous connections
	MaxConnections int `yaml:"max_connections"`
	// Mode specifies the server operating mode
	Mode string `yaml:"mode"`
}

// ListenerConfig defines how the load balancer should listen for incoming connections
type ListenerConfig struct {
	// ListenAddr is the address and port to listen on
	ListenAddr string `yaml:"listen_addr"`
	// Mode specifies the listener operating mode
	Mode string `yaml:"mode"`
	// Algorithm defines which load balancing algorithm to use
	Algorithm string `yaml:"algorithm"`
}

// LoadConfig reads and parses the YAML configuration file
// filename: path to the YML configuration file
// returns: pointer to Config struct and error if any occurred
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	return &cfg, err
}
