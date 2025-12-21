package dashboard

import (
	"fmt"
	"html/template"
	"strings"
	"zero-trust-dashboard/pkg/detector"
)

type Generator struct {
	services []detector.Service
}

func NewGenerator(services []detector.Service) *Generator {
	return &Generator{
		services: services,
	}
}

func (g *Generator) GenerateHTML(localPorts map[int]int, tunnelStartPort int) (string, error) {
	tmpl := `{{define "contains"}}{{if .}}{{.}}{{end}}{{end}}
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Zero-Trust Tunnel Dashboard</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        .header {
            background: white;
            padding: 30px;
            border-radius: 15px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
            margin-bottom: 30px;
            text-align: center;
        }
        .header h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 32px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .header p {
            color: #666;
            margin-bottom: 20px;
        }
        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .stat-card {
            background: white;
            padding: 25px;
            border-radius: 15px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
            text-align: center;
            transition: transform 0.2s, box-shadow 0.2s;
            position: relative;
            overflow: hidden;
        }
        .stat-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 4px;
            background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
        }
        .stat-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 10px 25px rgba(0,0,0,0.2);
        }
        .stat-value {
            font-size: 36px;
            font-weight: 700;
            color: #667eea;
            margin-bottom: 5px;
        }
        .stat-label {
            font-size: 14px;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .controls {
            display: flex;
            gap: 15px;
            margin-bottom: 20px;
            flex-wrap: wrap;
            justify-content: center;
        }
        .search-box {
            flex: 1;
            min-width: 250px;
            padding: 12px 20px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
            transition: border-color 0.2s;
        }
        .search-box:focus {
            outline: none;
            border-color: #667eea;
        }
        .btn {
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.2s;
        }
        .btn-primary {
            background: #667eea;
            color: white;
        }
        .btn-primary:hover {
            background: #5568d3;
            transform: translateY(-2px);
        }
        .btn-success {
            background: #4CAF50;
            color: white;
        }
        .btn-success:hover {
            background: #45a049;
            transform: translateY(-2px);
        }
        .btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
        }
        .services-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
            gap: 20px;
        }
        .service-card {
            background: white;
            border-radius: 15px;
            padding: 25px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
            transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
            position: relative;
            overflow: hidden;
        }
        .service-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 3px;
            background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
            transform: scaleX(0);
            transition: transform 0.3s;
        }
        .service-card:hover {
            transform: translateY(-8px) scale(1.02);
            box-shadow: 0 15px 35px rgba(0,0,0,0.2);
        }
        .service-card:hover::before {
            transform: scaleX(1);
        }
        .service-header {
            display: flex;
            align-items: center;
            margin-bottom: 15px;
        }
        .service-icon {
            width: 50px;
            height: 50px;
            border-radius: 12px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-right: 15px;
            font-weight: bold;
            color: white;
            font-size: 22px;
            box-shadow: 0 4px 10px rgba(0,0,0,0.2);
        }
        .service-icon.grafana { background: linear-gradient(135deg, #F46800 0%, #FF8C42 100%); }
        .service-icon.prometheus { background: linear-gradient(135deg, #E6522C 0%, #FF6B4A 100%); }
        .service-icon.kubernetes { background: linear-gradient(135deg, #326CE5 0%, #4A90E2 100%); }
        .service-icon.jenkins { background: linear-gradient(135deg, #D24939 0%, #E85D4F 100%); }
        .service-icon.jupyter { background: linear-gradient(135deg, #F37626 0%, #FF8C42 100%); }
        .service-icon.web { background: linear-gradient(135deg, #4CAF50 0%, #66BB6A 100%); }
        .service-icon.api { background: linear-gradient(135deg, #2196F3 0%, #42A5F5 100%); }
        .service-icon.application { background: linear-gradient(135deg, #9C27B0 0%, #BA68C8 100%); }
        .service-icon.unknown { background: linear-gradient(135deg, #9E9E9E 0%, #BDBDBD 100%); }
        .service-card.no-access {
            opacity: 0.7;
            border: 2px dashed #ccc;
        }
        .service-card.proxied {
            border: 2px solid #FF9800;
            background: linear-gradient(135deg, #FFF8E1 0%, #FFF3C4 100%);
        }
        .service-name {
            font-size: 22px;
            font-weight: 600;
            color: #333;
            margin-bottom: 3px;
        }
        .service-type {
            font-size: 11px;
            color: #999;
            text-transform: uppercase;
            letter-spacing: 1px;
            font-weight: 600;
        }
        .service-description {
            color: #666;
            margin-bottom: 15px;
            line-height: 1.6;
            font-size: 14px;
        }
        .service-link {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 12px 24px;
            border-radius: 8px;
            text-decoration: none;
            font-weight: 600;
            transition: all 0.2s;
            box-shadow: 0 4px 10px rgba(102, 126, 234, 0.3);
        }
        .service-link:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 15px rgba(102, 126, 234, 0.4);
        }
        .service-info {
            padding: 12px 20px;
            border-radius: 8px;
            background: linear-gradient(135deg, #f5f5f5 0%, #e8e8e8 100%);
            color: #666;
            font-size: 14px;
            border-left: 4px solid #FF9800;
        }
        .port-info {
            font-size: 12px;
            color: #999;
            margin-top: 12px;
            padding-top: 12px;
            border-top: 1px solid #f0f0f0;
        }
        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 11px;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.5px;
            margin-top: 8px;
        }
        .status-accessible {
            background: #E8F5E9;
            color: #2E7D32;
        }
        .status-proxied {
            background: #FFF3E0;
            color: #E65100;
        }
        .status-internal {
            background: #F3E5F5;
            color: #7B1FA2;
        }
        .network-header {
            background: rgba(255, 255, 255, 0.95);
            padding: 20px 30px;
            border-radius: 15px;
            margin-bottom: 20px;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
            display: flex;
            align-items: center;
            justify-content: space-between;
        }
        .network-header h2 {
            color: #333;
            font-size: 24px;
            font-weight: 700;
            margin: 0;
        }
        .network-badge {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 6px 16px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
        }
        .empty-state {
            background: white;
            padding: 80px;
            border-radius: 15px;
            text-align: center;
            box-shadow: 0 5px 15px rgba(0,0,0,0.1);
        }
        .empty-state h2 {
            color: #333;
            margin-bottom: 10px;
            font-size: 28px;
        }
        .empty-state p {
            color: #666;
            font-size: 16px;
        }
        .scan-result {
            margin-top: 15px;
            padding: 12px;
            border-radius: 8px;
            font-size: 14px;
            display: none;
        }
        .scan-result.success {
            background: #E8F5E9;
            color: #2E7D32;
            border-left: 4px solid #4CAF50;
        }
        .scan-result.error {
            background: #FFEBEE;
            color: #C62828;
            border-left: 4px solid #f44336;
        }
        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .service-card {
            animation: fadeIn 0.5s ease-out;
        }
        .stat-card {
            animation: fadeIn 0.5s ease-out;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Zero-Trust Tunnel Dashboard</h1>
            <p>Access your remote services securely through SSH tunnels</p>
        </div>
        
        {{if .Services}}
        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value">{{.TotalServices}}</div>
                <div class="stat-label">Total Services</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" style="color: #4CAF50;">{{.Accessible}}</div>
                <div class="stat-label">Accessible</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" style="color: #FF9800;">{{.Proxied}}</div>
                <div class="stat-label">Proxied</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" style="color: #9C27B0;">{{.Internal}}</div>
                <div class="stat-label">Internal Only</div>
            </div>
        </div>
        
        <div class="controls">
            <input type="text" id="searchBox" class="search-box" placeholder="Search services..." onkeyup="filterServices()">
            <button class="btn btn-success" id="scanBtn" onclick="scanPorts()">Scan Open Ports</button>
            <button class="btn btn-primary" onclick="location.reload()">Refresh Dashboard</button>
        </div>
        <div id="scanResult" class="scan-result"></div>
        <script>
        function scanPorts() {
            const btn = document.getElementById('scanBtn');
            const result = document.getElementById('scanResult');
            const portRange = prompt('Enter port range to scan (e.g., 3000-9000):', '3000-9000');
            
            if (!portRange) return;
            
            btn.disabled = true;
            btn.textContent = 'Scanning...';
            result.style.display = 'block';
            result.className = 'scan-result';
            result.innerHTML = 'Scanning ports...';
            
            fetch('/api/scan?range=' + encodeURIComponent(portRange), {
                method: 'POST'
            })
            .then(response => {
                if (!response.ok) {
                    throw new Error('HTTP error: ' + response.status);
                }
                return response.json();
            })
            .then(data => {
                if (data.error) {
                    result.className = 'scan-result error';
                    result.innerHTML = 'Error: ' + data.error;
                } else if (data.ports && Array.isArray(data.ports)) {
                    result.className = 'scan-result success';
                    const portsList = data.ports.join(', ');
                    const count = data.count || data.ports.length;
                    result.innerHTML = 'Found ' + count + ' open port(s): ' + portsList;
                } else {
                    result.className = 'scan-result error';
                    result.innerHTML = 'Error: Invalid response from server';
                }
                btn.disabled = false;
                btn.textContent = 'Scan Open Ports';
            })
            .catch(error => {
                result.className = 'scan-result error';
                result.innerHTML = 'Error: ' + error.message;
                btn.disabled = false;
                btn.textContent = 'Scan Open Ports';
            });
        }
        
        function filterServices() {
            const searchBox = document.getElementById('searchBox');
            const filter = searchBox.value.toLowerCase();
            const cards = document.querySelectorAll('.service-card');
            
            cards.forEach(card => {
                const name = card.querySelector('.service-name').textContent.toLowerCase();
                const description = card.querySelector('.service-description').textContent.toLowerCase();
                const type = card.querySelector('.service-type').textContent.toLowerCase();
                
                if (name.includes(filter) || description.includes(filter) || type.includes(filter)) {
                    card.style.display = '';
                } else {
                    card.style.display = 'none';
                }
            });
        }
        </script>
        
        {{$grouped := groupByNetwork .Services}}
        {{range $network, $services := $grouped}}
        {{if ne $network "default"}}
        <div style="margin-bottom: 40px;">
            <div class="network-header">
                <h2>Network: {{$network}}</h2>
                <span class="network-badge">{{len $services}} service(s)</span>
            </div>
            <div class="services-grid">
                {{range $services}}
                <div class="service-card{{if not .URL}}{{if .Port}} proxied{{else}} no-access{{end}}{{end}}" data-service-name="{{.Name}}">
                    <div class="service-header">
                        <div class="service-icon {{.Type}}">{{.Icon}}</div>
                        <div>
                            <div class="service-name">{{.Name}}</div>
                            <div class="service-type">{{.Type}}</div>
                        </div>
                    </div>
                    <div class="service-description">{{.Description}}</div>
                    {{if .URL}}
                    <a href="{{.URL}}" target="_blank" class="service-link">Open Service →</a>
                    <span class="status-badge status-accessible">Accessible</span>
                    {{else}}
                    <div class="service-info">
                        {{if contains .Description "Nginx Proxy"}}
                        Access via Nginx Proxy Manager
                        {{else if .Domain}}
                        Domain: {{.Domain}} (configured in Nginx)
                        {{else}}
                        No direct access (internal network only)
                        {{end}}
                    </div>
                    {{if .Domain}}
                    <span class="status-badge status-proxied">Proxied</span>
                    {{else}}
                    <span class="status-badge status-internal">Internal</span>
                    {{end}}
                    {{end}}
                    {{if .Network}}
                    <div class="port-info" style="color: #2196F3; font-weight: 500; margin-bottom: 5px;">Network: {{.Network}}</div>
                    {{end}}
                    {{if .Domain}}
                    <div class="port-info" style="color: #4CAF50; font-weight: 500;">Domain: {{.Domain}}</div>
                    {{else if .Port}}
                    {{if .LocalPort}}
                    <div class="port-info">Port: {{.Port}} → Local: {{.LocalPort}}</div>
                    {{else}}
                    <div class="port-info">Port: {{.Port}}</div>
                    {{end}}
                    {{else if (contains .Description "Nginx Proxy")}}
                    <div class="port-info">Accessible via Nginx Proxy Manager</div>
                    {{else}}
                    <div class="port-info">No exposed ports</div>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
        {{end}}
        {{if index $grouped "default"}}
        <div style="margin-bottom: 40px;">
            <div class="network-header">
                <h2>Other Services</h2>
                <span class="network-badge">{{len (index $grouped "default")}} service(s)</span>
            </div>
            <div class="services-grid">
                {{range index $grouped "default"}}
                <div class="service-card{{if not .URL}}{{if .Port}} proxied{{else}} no-access{{end}}{{end}}" data-service-name="{{.Name}}">
                    <div class="service-header">
                        <div class="service-icon {{.Type}}">{{.Icon}}</div>
                        <div>
                            <div class="service-name">{{.Name}}</div>
                            <div class="service-type">{{.Type}}</div>
                        </div>
                    </div>
                    <div class="service-description">{{.Description}}</div>
                    {{if .URL}}
                    <a href="{{.URL}}" target="_blank" class="service-link">Open Service →</a>
                    <span class="status-badge status-accessible">Accessible</span>
                    {{else}}
                    <div class="service-info">
                        {{if contains .Description "Nginx Proxy"}}
                        Access via Nginx Proxy Manager
                        {{else if .Domain}}
                        Domain: {{.Domain}} (configured in Nginx)
                        {{else}}
                        No direct access (internal network only)
                        {{end}}
                    </div>
                    {{if .Domain}}
                    <span class="status-badge status-proxied">Proxied</span>
                    {{else}}
                    <span class="status-badge status-internal">Internal</span>
                    {{end}}
                    {{end}}
                    {{if .Network}}
                    <div class="port-info" style="color: #2196F3; font-weight: 500; margin-bottom: 5px;">Network: {{.Network}}</div>
                    {{end}}
                    {{if .Domain}}
                    <div class="port-info" style="color: #4CAF50; font-weight: 500;">Domain: {{.Domain}}</div>
                    {{else if .Port}}
                    {{if .LocalPort}}
                    <div class="port-info">Port: {{.Port}} → Local: {{.LocalPort}}</div>
                    {{else}}
                    <div class="port-info">Port: {{.Port}}</div>
                    {{end}}
                    {{else if (contains .Description "Nginx Proxy")}}
                    <div class="port-info">Accessible via Nginx Proxy Manager</div>
                    {{else}}
                    <div class="port-info">No exposed ports</div>
                    {{end}}
                </div>
                {{end}}
            </div>
        </div>
        {{end}}
        {{else}}
        <div class="empty-state">
            <h2>No services detected</h2>
            <p>No services were found on the scanned ports. Make sure the tunnel is established and ports are accessible.</p>
        </div>
        {{end}}
    </div>
</body>
</html>`

	type ServiceData struct {
		Name        string
		Type        string
		Description string
		URL         string
		Port        int
		LocalPort   int
		Icon        string
		Domain      string
		Network     string
	}

	type TemplateData struct {
		Services      []ServiceData
		TotalServices int
		Accessible    int
		Proxied       int
		Internal      int
		Networks      []string
	}

		var serviceData []ServiceData
	for _, svc := range g.services {
		if svc.Type == "nginx" || strings.Contains(strings.ToLower(svc.Name), "nginx") {
			continue
		}
		
		icon := getServiceIcon(svc.Type)
		localPort := 0
		serviceURL := svc.URL
		isProxied := strings.Contains(svc.Description, "Nginx Proxy") || svc.Domain != ""
		
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
		
		serviceData = append(serviceData, ServiceData{
			Name:        svc.Name,
			Type:        svc.Type,
			Description: svc.Description,
			URL:         serviceURL,
			Port:        svc.Port,
			LocalPort:   localPort,
			Icon:        icon,
			Domain:      svc.Domain,
			Network:     svc.Network,
		})
	}

	totalServices := len(serviceData)
	accessible := 0
	proxied := 0
	internal := 0
	networkSet := make(map[string]bool)
	
	for _, svc := range serviceData {
		if svc.URL != "" {
			accessible++
		} else if svc.Domain != "" || strings.Contains(svc.Description, "Nginx Proxy") {
			proxied++
		} else {
			internal++
		}
		if svc.Network != "" {
			networkSet[svc.Network] = true
		}
	}
	
	networks := make([]string, 0, len(networkSet))
	for net := range networkSet {
		networks = append(networks, net)
	}

	funcMap := template.FuncMap{
		"contains": func(s, substr string) bool {
			return strings.Contains(s, substr)
		},
		"groupByNetwork": func(services []ServiceData) map[string][]ServiceData {
			grouped := make(map[string][]ServiceData)
			for _, svc := range services {
				network := svc.Network
				if network == "" {
					network = "default"
				}
				grouped[network] = append(grouped[network], svc)
			}
			return grouped
		},
	}
	t, err := template.New("dashboard").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	err = t.Execute(&buf, TemplateData{
		Services:      serviceData,
		TotalServices: totalServices,
		Accessible:    accessible,
		Proxied:       proxied,
		Internal:      internal,
		Networks:      networks,
	})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (g *Generator) GenerateCLI(localPorts map[int]int) string {
	if len(g.services) == 0 {
		return "No services detected.\n"
	}

	var sb strings.Builder
	sb.WriteString("\nZero-Trust Tunnel Dashboard\n")
	sb.WriteString("===========================================================\n\n")

	for _, svc := range g.services {
		if svc.Port == 0 {
			sb.WriteString(fmt.Sprintf("%s\n", svc.Name))
			sb.WriteString(fmt.Sprintf("   Type: %s\n", svc.Type))
			sb.WriteString(fmt.Sprintf("   Description: %s\n", svc.Description))
			sb.WriteString("   Status: No exposed ports (internal network only)\n")
			sb.WriteString("\n")
			continue
		}
		
		if svc.Port == 0 && strings.Contains(svc.Description, "Nginx Proxy") {
			sb.WriteString(fmt.Sprintf("%s\n", svc.Name))
			sb.WriteString(fmt.Sprintf("   Type: %s\n", svc.Type))
			sb.WriteString(fmt.Sprintf("   Description: %s\n", svc.Description))
			if svc.URL != "" {
				sb.WriteString(fmt.Sprintf("   URL: %s\n", svc.URL))
			}
			sb.WriteString("\n")
			continue
		}

		localPort, hasLocal := localPorts[svc.Port]
		if !hasLocal {
			continue
		}

		sb.WriteString(fmt.Sprintf("%s\n", svc.Name))
		sb.WriteString(fmt.Sprintf("   Type: %s\n", svc.Type))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", svc.Description))
		sb.WriteString(fmt.Sprintf("   Remote Port: %d\n", svc.Port))
		sb.WriteString(fmt.Sprintf("   Local Port: %d\n", localPort))
		sb.WriteString(fmt.Sprintf("   URL: %s\n", svc.URL))
		sb.WriteString("\n")
	}

	return sb.String()
}

func getServiceIcon(serviceType string) string {
	icons := map[string]string{
		"grafana":     "",
		"prometheus":   "",
		"kubernetes":   "",
		"jenkins":      "",
		"jupyter":      "",
		"web":          "",
		"api":          "",
		"application":  "",
		"unknown":      "",
	}

	if icon, ok := icons[strings.ToLower(serviceType)]; ok {
		return icon
	}
	return ""
}

