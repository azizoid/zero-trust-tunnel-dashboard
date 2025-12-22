package scanner

import (
	"fmt"
	"strings"
	"testing"
)

func BenchmarkParsePortRange(b *testing.B) {
	tests := []string{
		"3000-9000",
		"8080",
		"1000-2000",
		"",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, test := range tests {
			parsePortRange(test)
		}
	}
}

func BenchmarkDeduplicatePorts(b *testing.B) {
	ports := []int{3000, 8080, 9090, 3000, 8080, 3001, 8080, 9090, 3002}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deduplicatePorts(ports)
	}
}

func BenchmarkParseSSOutput(b *testing.B) {
	output := `State      Recv-Q Send-Q Local Address:Port               Peer Address:Port              
LISTEN     0      128          *:22                       *:*                  
LISTEN     0      128          *:3000                     *:*                  
LISTEN     0      128          *:8080                     *:*                  
LISTEN     0      128          *:9090                     *:*                  
LISTEN     0      128          *:3306                     *:*                  
LISTEN     0      128          *:5432                     *:*                  
`
	portRange := "3000-9000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseSSOutput(output, portRange)
	}
}

func BenchmarkParseNetstatOutput(b *testing.B) {
	output := `Active Internet connections (only servers)
Proto Recv-Q Send-Q Local Address           Foreign Address         State      
tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN     
tcp        0      0 0.0.0.0:3000            0.0.0.0:*               LISTEN     
tcp        0      0 0.0.0.0:8080            0.0.0.0:*               LISTEN     
tcp        0      0 0.0.0.0:9090            0.0.0.0:*               LISTEN     
tcp        0      0 0.0.0.0:3306            0.0.0.0:*               LISTEN     
tcp        0      0 0.0.0.0:5432            0.0.0.0:*               LISTEN     
`
	portRange := "3000-9000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseNetstatOutput(output, portRange)
	}
}

func BenchmarkParseSSOutputLarge(b *testing.B) {
	// Simulate output with many ports
	var lines []string
	for i := 3000; i < 9000; i += 10 {
		lines = append(lines, fmt.Sprintf("LISTEN     0      128          *:%d                     *:*", i))
	}
	output := strings.Join(lines, "\n")
	portRange := "3000-9000"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseSSOutput(output, portRange)
	}
}

