package dashboard

import (
	"fmt"
	"sort"
	"strings"

	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/detector"
)

// serviceIcons maps service types to their icon representations
var serviceIcons = map[string]string{
	"grafana":     "",
	"prometheus":  "",
	"kubernetes":  "",
	"jenkins":     "",
	"jupyter":     "",
	"web":         "",
	"api":         "",
	"application": "",
	"postgres":    "",
	"mysql":       "",
	"mongodb":     "",
	"redis":       "",
	"unknown":     "",
}

// getServiceIcon returns the icon for a service type
func getServiceIcon(serviceType string) string {
	if icon, ok := serviceIcons[strings.ToLower(serviceType)]; ok {
		return icon
	}
	return ""
}

// ServiceAccess represents the accessibility level of a service
type ServiceAccess int

const (
	AccessInternal ServiceAccess = iota
	AccessProxied
	AccessAccessible
)

// ServiceView represents a service with resolved access state and computed properties
type ServiceView struct {
	Name        string
	Type        string
	Description string
	URL         string
	Port        int
	LocalPort   int
	Icon        string
	Domain      string
	Network     string
	Access      ServiceAccess
	AccessClass string
}

// Stats represents dashboard statistics
type Stats struct {
	TotalServices int
	Accessible    int
	Proxied       int
	Internal      int
}

// ViewModel contains all data needed to render the dashboard
type ViewModel struct {
	Services []ServiceView
	Groups   map[string][]ServiceView
	Stats    Stats
	Networks []string
}

// resolveAccess determines the access level of a service based on its properties
func resolveAccess(svc detector.Service, hasURL bool) ServiceAccess {
	if hasURL {
		return AccessAccessible
	}
	if svc.Domain != "" {
		return AccessProxied
	}
	return AccessInternal
}

// accessClass returns the CSS class name for a given access level
func accessClass(access ServiceAccess) string {
	switch access {
	case AccessAccessible:
		return "accessible"
	case AccessProxied:
		return "proxied"
	default:
		return "no-access"
	}
}

// isDashboardInternalService determines if a service should be hidden from the dashboard
func isDashboardInternalService(svc detector.Service) bool {
	return svc.Type == "nginx"
}

// normalizeNetworkName normalizes network names by sorting comma-separated values
func normalizeNetworkName(network string) string {
	if network == "" {
		return ""
	}

	parts := strings.Split(network, ",")
	networks := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			networks = append(networks, trimmed)
		}
	}

	if len(networks) == 0 {
		return ""
	}

	sort.Strings(networks)
	return strings.Join(networks, ", ")
}

// buildViewModel creates a ViewModel from services and port mappings
func buildViewModel(services []detector.Service, localPorts map[int]int, tunnelStartPort int) ViewModel {
	var views []ServiceView
	networkSet := make(map[string]bool)

	for _, svc := range services {
		if isDashboardInternalService(svc) {
			continue
		}

		view := buildServiceView(svc, localPorts, tunnelStartPort)
		views = append(views, view)

		if view.Network != "" {
			networkSet[view.Network] = true
		}
	}

	stats := computeStats(views)
	groups := groupByNetwork(views)
	networks := make([]string, 0, len(networkSet))
	for net := range networkSet {
		networks = append(networks, net)
	}
	sort.Strings(networks)

	return ViewModel{
		Services: views,
		Groups:   groups,
		Stats:    stats,
		Networks: networks,
	}
}

// buildServiceView creates a ServiceView from a detector.Service
func buildServiceView(svc detector.Service, localPorts map[int]int, tunnelStartPort int) ServiceView {
	icon := getServiceIcon(svc.Type)
	localPort := 0
	serviceURL := svc.URL
	hasDomain := svc.Domain != ""
	isProxied := strings.Contains(svc.Description, "Nginx Proxy") || hasDomain

	if svc.Port > 0 && !isProxied {
		if lp, exists := localPorts[svc.Port]; exists {
			localPort = lp
			if localPort == tunnelStartPort {
				serviceURL = ""
				localPort = 0
			} else if serviceURL == "" {
				serviceURL = fmt.Sprintf("http://localhost:%d", localPort)
			}
		} else {
			serviceURL = ""
		}
	}

	if isProxied {
		localPort = 0
	}

	if strings.Contains(serviceURL, fmt.Sprintf(":%d", tunnelStartPort)) || localPort == tunnelStartPort {
		serviceURL = ""
		localPort = 0
	}

	normalizedNetwork := normalizeNetworkName(svc.Network)
	hasURL := serviceURL != ""
	access := resolveAccess(svc, hasURL)

	return ServiceView{
		Name:        svc.Name,
		Type:        svc.Type,
		Description: svc.Description,
		URL:         serviceURL,
		Port:        svc.Port,
		LocalPort:   localPort,
		Icon:        icon,
		Domain:      svc.Domain,
		Network:     normalizedNetwork,
		Access:      access,
		AccessClass: accessClass(access),
	}
}

// computeStats calculates statistics from service views
func computeStats(views []ServiceView) Stats {
	stats := Stats{
		TotalServices: len(views),
	}

	for _, view := range views {
		switch view.Access {
		case AccessAccessible:
			stats.Accessible++
		case AccessProxied:
			stats.Proxied++
		default:
			stats.Internal++
		}
	}

	return stats
}

// groupByNetwork groups services by their network
func groupByNetwork(views []ServiceView) map[string][]ServiceView {
	grouped := make(map[string][]ServiceView)
	for _, view := range views {
		network := view.Network
		if network == "" {
			network = "default"
		}
		grouped[network] = append(grouped[network], view)
	}
	return grouped
}
