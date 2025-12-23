package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/detector"
	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/scanner"
)

type Server struct {
	port     int
	services []detector.Service
	html     string
	// shutdownFunc is called to gracefully shut down the application
	shutdownFunc func()
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

func (s *Server) SetShutdownFunc(fn func()) {
	s.shutdownFunc = fn
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleDashboard)
	http.HandleFunc("/api/services", s.handleServicesAPI)
	http.HandleFunc("/api/scan", s.handleScan)
	http.HandleFunc("/api/shutdown", s.handleShutdown)
	http.HandleFunc("/health", s.handleHealth)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Dashboard server starting on http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// UpdateServices updates the list of services dynamically.
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
		_, _ = w.Write([]byte(s.html)) //nolint:errcheck // Ignore write error
	} else {
		_, _ = fmt.Fprintf(w, "<html><body><h1>Zero-Trust Tunnel Dashboard</h1><p>Dashboard is loading...</p></body></html>") //nolint:errcheck // Ignore write error
	}
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = fmt.Fprintf(w, `
		<html>
		<body style="font-family: sans-serif; text-align: center; padding-top: 50px;">
			<h1>Tunnel Stopped</h1>
			<p>You can close this window now.</p>
			<script>window.stop();</script>
		</body>
		</html>
	`) //nolint:errcheck

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	if s.shutdownFunc != nil {
		go func() {
			s.shutdownFunc()
		}()
	}
}

func (s *Server) handleServicesAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.services) //nolint:errcheck // Ignore encode error
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
		_ = json.NewEncoder(w).Encode(map[string]interface{}{ //nolint:errcheck // Ignore encode error
			"error": err.Error(),
		}) // Ignore encode error
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"ports": ports,
		"count": len(ports),
	}) // Ignore encode error
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":   "healthy",
		"services": len(s.services),
	}) // Ignore encode error
}
