package server

import (
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func (s *Server) StartHealthCheck(interval time.Duration) {
	logFile, err := os.OpenFile("health_checks.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}

	multiLogger := log.New(logFile, "", log.LstdFlags)

	ticker := time.NewTicker(interval)
	go func() {
		s.checkHealthWithLogger(multiLogger)
		for range ticker.C {
			s.checkHealthWithLogger(multiLogger)
		}
	}()
}

func (s *Server) checkHealthWithLogger(logger *log.Logger) {
	conn, err := net.DialTimeout("tcp", s.Host, 3*time.Second)
	status := err == nil
	if conn != nil {
		err := conn.Close()
		if err != nil {
			log.Printf("Error Occured %s", err)
		}
	}
	s.SetHealthy(status)

	logMessage := fmt.Sprintf("[Health Check] Server %s (%s) status: %v", s.Host, s.Mode, status)
	log.Println(logMessage)
	logger.Println(logMessage)
}
