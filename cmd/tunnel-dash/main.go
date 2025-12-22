package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/app"
	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/version"
)

func main() {
	var (
		host            = flag.String("host", "", "SSH host alias from ~/.ssh/config (alternative to --server/--user)")
		serverAddr      = flag.String("server", "", "SSH server address (required if --host not set)")
		user            = flag.String("user", "", "SSH username (required if --host not set)")
		keyPath         = flag.String("key", "", "Path to SSH private key (optional, overrides SSH config)")
		scanPorts       = flag.String("scan-ports", "3000-9000", "Port range to scan (e.g., 3000-9000)")
		dashboardPort   = flag.Int("dashboard-port", 8080, "Port for the web dashboard")
		tunnelStartPort = flag.Int("tunnel-start-port", 9000, "Starting port for local tunnel ports")
		detectionMode   = flag.String("detection-mode", "both", "Service detection method: docker, direct, or both (default: both)")
		insecure        = flag.Bool("insecure", false, "Disable strict host key checking (WARNING: Man-in-the-Middle risk)")
		showVersion     = flag.Bool("version", false, "Show version information and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Info())
		os.Exit(0)
	}

	if *host == "" && (*serverAddr == "" || *user == "") {
		fmt.Fprintf(os.Stderr, "Error: either --host or both --server and --user are required\n")
		fmt.Fprintf(os.Stderr, "  Use --host to read from ~/.ssh/config\n")
		fmt.Fprintf(os.Stderr, "  Or use --server and --user for direct connection\n")
		flag.Usage()
		os.Exit(1)
	}

	config := app.Config{
		Host:            *host,
		ServerAddr:      *serverAddr,
		User:            *user,
		KeyPath:         *keyPath,
		ScanPorts:       *scanPorts,
		DashboardPort:   *dashboardPort,
		TunnelStartPort: *tunnelStartPort,
		DetectionMode:   *detectionMode,
		Insecure:        *insecure,
	}

	controller, err := app.NewController(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing controller: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		cancel()
	}()

	if err := controller.Run(ctx); err != nil {
		// Only report error if it wasn't a clean shutdown or a known non-error case (like no ports found)
		// Run returns nil on clean shutdown or no ports found case.
		// If context was canceled (Ctrl+C), Run returns nil (it handles the loop exit).
		fmt.Fprintf(os.Stderr, "Current status: %v\n", err)
	}
}
