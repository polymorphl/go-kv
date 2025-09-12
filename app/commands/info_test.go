package commands

import (
	"net"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func TestInfo(t *testing.T) {
	// Reset store state for clean test
	server.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "test-repl-id",
		MasterReplOffset: 12345,
		Replicas:         make(map[string]net.Conn),
	})

	tests := []struct {
		name     string
		args     []shared.Value
		expected string
	}{
		{
			name:     "INFO without arguments",
			args:     []shared.Value{},
			expected: "role:master\r\nmaster_replid:test-repl-id\r\nmaster_repl_offset:12345\r\n",
		},
		{
			name: "INFO with replication section",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "replication"},
			},
			expected: "role:master\r\nmaster_replid:test-repl-id\r\nmaster_repl_offset:12345\r\n",
		},
		{
			name: "INFO with server section",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "server"},
			},
			expected: "role:master\r\nmaster_replid:test-repl-id\r\nmaster_repl_offset:12345\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Info("test-conn", tt.args)

			if result.Typ != "bulk" {
				t.Errorf("Expected bulk type, got %s", result.Typ)
			}

			if result.Bulk != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Bulk)
			}
		})
	}
}

func TestInfoWithDifferentRoles(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		replID   string
		offset   int64
		expected string
	}{
		{
			name:     "Master role",
			role:     "master",
			replID:   "master-123",
			offset:   1000,
			expected: "role:master\r\nmaster_replid:master-123\r\nmaster_repl_offset:1000\r\n",
		},
		{
			name:     "Slave role",
			role:     "slave",
			replID:   "slave-456",
			offset:   2000,
			expected: "role:slave\r\nmaster_replid:slave-456\r\nmaster_repl_offset:2000\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up store state
			server.SetStoreState(shared.State{
				Role:             tt.role,
				MasterReplID:     tt.replID,
				MasterReplOffset: tt.offset,
				Replicas:         make(map[string]net.Conn),
			})

			result := Info("test-conn", []shared.Value{})

			if result.Typ != "bulk" {
				t.Errorf("Expected bulk type, got %s", result.Typ)
			}

			if result.Bulk != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Bulk)
			}
		})
	}
}

// BenchmarkInfo benchmarks the INFO command
func BenchmarkInfo(b *testing.B) {
	// Reset store state for clean benchmark
	server.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "bench-repl-id",
		MasterReplOffset: 12345,
		Replicas:         make(map[string]net.Conn),
	})

	args := []shared.Value{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("bench-conn", args)
	}
}

// BenchmarkInfoWithSection benchmarks the INFO command with section argument
func BenchmarkInfoWithSection(b *testing.B) {
	// Reset store state for clean benchmark
	server.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "bench-repl-id",
		MasterReplOffset: 12345,
		Replicas:         make(map[string]net.Conn),
	})

	args := []shared.Value{
		{Typ: "bulk", Bulk: "replication"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("bench-conn", args)
	}
}
