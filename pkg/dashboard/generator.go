package dashboard

import (
	_ "embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/azizoid/zero-trust-tunnel-dashboard/pkg/detector"
)

//go:embed dashboard.html
var dashboardTemplate string

type Generator struct {
	services []detector.Service
}

func NewGenerator(services []detector.Service) *Generator {
	return &Generator{
		services: services,
	}
}

func (g *Generator) GenerateHTML(localPorts map[int]int, tunnelStartPort int) (string, error) {
	viewModel := buildViewModel(g.services, localPorts, tunnelStartPort)

	funcMap := template.FuncMap{
		"contains": strings.Contains,
	}

	t, err := template.New("dashboard").Funcs(funcMap).Parse(dashboardTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := t.Execute(&buf, viewModel); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

func (g *Generator) GenerateCLI(localPorts map[int]int, tunnelStartPort int) string {
	viewModel := buildViewModel(g.services, localPorts, tunnelStartPort)

	if len(viewModel.Services) == 0 {
		return "No services detected.\n"
	}

	var sb strings.Builder
	sb.WriteString("\nZero-Trust Tunnel Dashboard\n")
	sb.WriteString("===========================================================\n\n")

	for _, view := range viewModel.Services {
		sb.WriteString(fmt.Sprintf("%s\n", view.Name))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", view.Type))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", view.Description))

		switch view.Access {
		case AccessAccessible:
			if view.URL != "" {
				sb.WriteString(fmt.Sprintf("   URL: %s\n", view.URL))
			}
			if view.Port > 0 {
				sb.WriteString(fmt.Sprintf("   Remote Port: %d\n", view.Port))
			}
			if view.LocalPort > 0 {
				sb.WriteString(fmt.Sprintf("   Local Port: %d\n", view.LocalPort))
			}
		case AccessProxied:
			if view.Domain != "" {
				sb.WriteString(fmt.Sprintf("   Domain: %s (configured in Nginx)\n", view.Domain))
			}
			sb.WriteString("   Status: Proxied via Nginx\n")
		case AccessInternal:
			if view.Port > 0 {
				sb.WriteString(fmt.Sprintf("   Port: %d\n", view.Port))
			}
			sb.WriteString("   Status: Internal network only\n")
		}

		if view.Network != "" {
			sb.WriteString(fmt.Sprintf("   Network: %s\n", view.Network))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
