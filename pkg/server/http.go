package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/azizoid/zero-trust-dashnoard/pkg/detector"
	"github.com/azizoid/zero-trust-dashnoard/pkg/scanner"
)

type Server struct {
	port     int
	services []detector.Service
	html     string
	scanner  *scanner.Scanner
}

func NewServer(port int, services []detector.Service) *Server {
	return &Server{
		port:     port,
		services: services,
	}
}

func (s *Server) SetScanner(sc *scanner.Scanner) {
	s.scanner = sc
}

func (s *Server) SetHTML(html string) {
	s.html = html
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/services", s.handleServicesAPI)
	http.HandleFunc("/api/scan", s.handleScan)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Dashboard server starting on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) UpdateServices(services []detector.Service) {
	s.services = services
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if s.html != "" {
		w.Write([]byte(s.html))
	} else {
		fmt.Fprintf(w, "<html><body><h1>Zero-Trust Tunnel Dashboard</h1><p>Dashboard is loading...</p></body></html>")
	}
}

func (s *Server) handleServicesAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.services)
}

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.scanner == nil {
		http.Error(w, "Scanner not available", http.StatusServiceUnavailable)
		return
	}

	portRange := r.URL.Query().Get("range")
	if portRange == "" {
		portRange = "3000-9000"
	}

	ports, err := s.scanner.ScanPorts(portRange)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ports": ports,
		"count": len(ports),
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"services": len(s.services),
	})
}

