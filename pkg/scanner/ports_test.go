package scanner

import (
	"testing"
)

func TestParsePortRange(t *testing.T) {
	tests := []struct {
		name      string
		portRange string
		wantMin   int
		wantMax   int
	}{
		{
			name:      "valid range",
			portRange: "3000-9000",
			wantMin:   3000,
			wantMax:   9000,
		},
		{
			name:      "single port",
			portRange: "8080",
			wantMin:   8080,
			wantMax:   8080,
		},
		{
			name:      "empty range",
			portRange: "",
			wantMin:   1,
			wantMax:   65535,
		},
		{
			name:      "invalid range",
			portRange: "invalid",
			wantMin:   1,
			wantMax:   65535,
		},
		{
			name:      "range with spaces",
			portRange: " 3000 - 9000 ",
			wantMin:   3000,
			wantMax:   9000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := parsePortRange(tt.portRange)
			if gotMin != tt.wantMin {
				t.Errorf("parsePortRange() min = %v, want %v", gotMin, tt.wantMin)
			}
			if gotMax != tt.wantMax {
				t.Errorf("parsePortRange() max = %v, want %v", gotMax, tt.wantMax)
			}
		})
	}
}

func TestDeduplicatePorts(t *testing.T) {
	tests := []struct {
		name  string
		ports []int
		want  []int
	}{
		{
			name:  "no duplicates",
			ports: []int{3000, 8080, 9000},
			want:  []int{3000, 8080, 9000},
		},
		{
			name:  "with duplicates",
			ports: []int{3000, 8080, 3000, 9000, 8080},
			want:  []int{3000, 8080, 9000},
		},
		{
			name:  "empty slice",
			ports: []int{},
			want:  []int{},
		},
		{
			name:  "all duplicates",
			ports: []int{3000, 3000, 3000},
			want:  []int{3000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deduplicatePorts(tt.ports)
			if len(got) != len(tt.want) {
				t.Errorf("deduplicatePorts() length = %v, want %v", len(got), len(tt.want))
				return
			}

			gotMap := make(map[int]bool)
			for _, p := range got {
				gotMap[p] = true
			}

			wantMap := make(map[int]bool)
			for _, p := range tt.want {
				wantMap[p] = true
			}

			for port := range gotMap {
				if !wantMap[port] {
					t.Errorf("deduplicatePorts() unexpected port %v", port)
				}
			}

			for port := range wantMap {
				if !gotMap[port] {
					t.Errorf("deduplicatePorts() missing port %v", port)
				}
			}
		})
	}
}

func TestParseSSOutput(t *testing.T) {
	output := `State      Recv-Q Send-Q Local Address:Port               Peer Address:Port
LISTEN     0      128          *:3000                     *:*
LISTEN     0      128          *:8080                     *:*
LISTEN     0      128          *:9000                     *:*
LISTEN     0      128          *:22                       *:*`

	ports, err := parseSSOutput(output, "3000-9000")
	if err != nil {
		t.Fatalf("parseSSOutput() error = %v", err)
	}

	expectedPorts := []int{3000, 8080, 9000}
	if len(ports) != len(expectedPorts) {
		t.Errorf("parseSSOutput() returned %d ports, want %d", len(ports), len(expectedPorts))
	}

	portMap := make(map[int]bool)
	for _, p := range ports {
		portMap[p] = true
	}

	for _, expectedPort := range expectedPorts {
		if !portMap[expectedPort] {
			t.Errorf("parseSSOutput() missing port %d", expectedPort)
		}
	}
}

func TestParseNetstatOutput(t *testing.T) {
	output := `Proto Recv-Q Send-Q Local Address           Foreign Address         State
tcp        0      0 0.0.0.0:3000            0.0.0.0:*               LISTEN
tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN
tcp        0      0 0.0.0.0:9000            0.0.0.0:*               LISTEN
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN`

	ports, err := parseNetstatOutput(output, "3000-9000")
	if err != nil {
		t.Fatalf("parseNetstatOutput() error = %v", err)
	}

	expectedPorts := []int{3000, 8080, 9000}
	if len(ports) != len(expectedPorts) {
		t.Errorf("parseNetstatOutput() returned %d ports, want %d", len(ports), len(expectedPorts))
	}

	portMap := make(map[int]bool)
	for _, p := range ports {
		portMap[p] = true
	}

	for _, expectedPort := range expectedPorts {
		if !portMap[expectedPort] {
			t.Errorf("parseNetstatOutput() missing port %d", expectedPort)
		}
	}
}
