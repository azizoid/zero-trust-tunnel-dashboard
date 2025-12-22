package detector

import (
	"fmt"
	"os/exec"
	"strings"
)

func QueryNPMDatabase(nginxContainerName, containerName string, containerPort int, server, user, keyPath string, useHostAlias bool, hostAlias string, insecure bool) ([]string, error) {

	domains, err := queryNPMWithSQLite3(nginxContainerName, containerName, containerPort, server, user, keyPath, useHostAlias, hostAlias, insecure)
	if err == nil && len(domains) > 0 {
		return domains, nil
	}

	domains, err = getNginxDomainsFromConfig(nginxContainerName, containerName, containerPort, server, user, keyPath, useHostAlias, hostAlias, insecure)
	if err == nil && len(domains) > 0 {
		return domains, nil
	}

	domains, err = queryNPMFromHost(nginxContainerName, containerName, containerPort, server, user, keyPath, useHostAlias, hostAlias, insecure)
	if err == nil && len(domains) > 0 {
		return domains, nil
	}

	return nil, nil
}

func queryNPMWithSQLite3(nginxContainerName, containerName string, containerPort int, server, user, keyPath string, useHostAlias bool, hostAlias string, insecure bool) ([]string, error) {
	var cmd *exec.Cmd
	dbPath := "/data/database.sqlite"

	query := fmt.Sprintf("SELECT domain_names FROM proxy_host WHERE (forward_host LIKE '%%%s%%' OR forward_host = '%s') AND forward_port = %d", containerName, containerName, containerPort)

	// Helper to add insecure flags
	addInsecureFlags := func(args []string) []string {
		if insecure {
			return append(args,
				"-o", "StrictHostKeyChecking=no",
				"-o", "UserKnownHostsFile=/dev/null",
			)
		}
		return args
	}

	if useHostAlias {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR", hostAlias,
			fmt.Sprintf("docker exec %s sqlite3 %s %q 2>/dev/null || echo ''", nginxContainerName, dbPath, query))
		cmd = exec.Command("ssh", args...)
	} else {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR")
		if keyPath != "" {
			args = append(args, "-i", keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", user, server),
			fmt.Sprintf("docker exec %s sqlite3 %s %q 2>/dev/null || echo ''", nginxContainerName, dbPath, query))
		cmd = exec.Command("ssh", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var domains []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, "exec failed") || strings.Contains(line, "executable file not found") {
			continue
		}
		domainList := strings.Split(line, ",")
		for _, domain := range domainList {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				domains = append(domains, domain)
			}
		}
	}

	if len(domains) > 0 {
		return domains, nil
	}
	return nil, fmt.Errorf("no domains found")
}

func queryNPMFromHost(nginxContainerName, containerName string, containerPort int, server, user, keyPath string, useHostAlias bool, hostAlias string, insecure bool) ([]string, error) {
	var cmd *exec.Cmd

	// Helper to add insecure flags
	addInsecureFlags := func(args []string) []string {
		if insecure {
			return append(args,
				"-o", "StrictHostKeyChecking=no",
				"-o", "UserKnownHostsFile=/dev/null",
			)
		}
		return args
	}

	if useHostAlias {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR", hostAlias,
			fmt.Sprintf("docker inspect %s --format '{{range .Mounts}}{{if eq .Destination \"/data\"}}{{.Source}}{{end}}{{end}}' 2>/dev/null", nginxContainerName))
		cmd = exec.Command("ssh", args...)
	} else {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR")
		if keyPath != "" {
			args = append(args, "-i", keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", user, server),
			fmt.Sprintf("docker inspect %s --format '{{range .Mounts}}{{if eq .Destination \"/data\"}}{{.Source}}{{end}}{{end}}' 2>/dev/null", nginxContainerName))
		cmd = exec.Command("ssh", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	mountPath := strings.TrimSpace(string(output))
	if mountPath == "" {
		return nil, fmt.Errorf("could not find mount path")
	}

	dbPath := fmt.Sprintf("%s/database.sqlite", mountPath)
	query := fmt.Sprintf("SELECT domain_names FROM proxy_host WHERE (forward_host LIKE '%%%s%%' OR forward_host = '%s') AND forward_port = %d", containerName, containerName, containerPort)

	if useHostAlias {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR", hostAlias,
			fmt.Sprintf("sqlite3 %s %q 2>/dev/null || echo ''", dbPath, query))
		cmd = exec.Command("ssh", args...)
	} else {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR")
		if keyPath != "" {
			args = append(args, "-i", keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", user, server),
			fmt.Sprintf("sqlite3 %s %q 2>/dev/null || echo ''", dbPath, query))
		cmd = exec.Command("ssh", args...)
	}

	output, err = cmd.Output()
	if err != nil {
		return nil, err
	}

	var domains []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		domainList := strings.Split(line, ",")
		for _, domain := range domainList {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				domains = append(domains, domain)
			}
		}
	}

	if len(domains) > 0 {
		return domains, nil
	}
	return nil, fmt.Errorf("no domains found")
}

func getNginxDomainsFromConfig(nginxContainerName, containerName string, containerPort int, server, user, keyPath string, useHostAlias bool, hostAlias string, insecure bool) ([]string, error) {
	var cmd *exec.Cmd

	// Helper to add insecure flags
	addInsecureFlags := func(args []string) []string {
		if insecure {
			return append(args,
				"-o", "StrictHostKeyChecking=no",
				"-o", "UserKnownHostsFile=/dev/null",
			)
		}
		return args
	}

	if useHostAlias {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR", hostAlias,
			fmt.Sprintf("docker exec %s find /data/nginx/proxy_host -name '*.conf' -exec grep -l '%s:%d' {} \\; 2>/dev/null | head -1 | xargs grep -oP 'server_name\\s+\\K[^;]+' 2>/dev/null || echo ''", nginxContainerName, containerName, containerPort))
		cmd = exec.Command("ssh", args...)
	} else {
		args := []string{}
		args = addInsecureFlags(args)
		args = append(args, "-o", "LogLevel=ERROR")
		if keyPath != "" {
			args = append(args, "-i", keyPath)
		}
		args = append(args, fmt.Sprintf("%s@%s", user, server),
			fmt.Sprintf("docker exec %s find /data/nginx/proxy_host -name '*.conf' -exec grep -l '%s:%d' {} \\; 2>/dev/null | head -1 | xargs grep -oP 'server_name\\s+\\K[^;]+' 2>/dev/null || echo ''", nginxContainerName, containerName, containerPort))
		cmd = exec.Command("ssh", args...)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var domains []string
	line := strings.TrimSpace(string(output))
	if line != "" {
		domainList := strings.Fields(line)
		domains = append(domains, domainList...)
	}

	if len(domains) > 0 {
		return domains, nil
	}
	return nil, fmt.Errorf("no domains found")
}
