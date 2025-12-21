package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"zero-trust-dashboard/pkg/detector"
	"zero-trust-dashboard/pkg/scanner"
)

// Server handles HTTP requests for the dashboard
type Server struct {
	port     int
	services []detector.Service
	html     string
	scanner  *scanner.Scanner
}

// NewServer creates a new HTTP server
func NewServer(port int, services []detector.Service) *Server {
	return &Server{
		port:     port,
		services: services,
	}
}

// SetScanner sets the port scanner for the server
func (s *Server) SetScanner(sc *scanner.Scanner) {
	s.scanner = sc
}

// SetHTML sets the HTML content for the dashboard
func (s *Server) SetHTML(html string) {
	s.html = html
}

// Start starts the HTTP server
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/services", s.handleServicesAPI)
	http.HandleFunc("/api/scan", s.handleScan)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Dashboard server starting on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// UpdateServices updates the list of services
func (s *Server) UpdateServices(services []detector.Service) {
	s.services = services
}

// handleDashboard serves the HTML dashboard
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

// handleServicesAPI returns services as JSON
func (s *Server) handleServicesAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.services)
}

// handleScan triggers a port scan and returns results
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

// handleHealth returns server health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"services": len(s.services),
	})
}

