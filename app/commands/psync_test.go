package commands

import (
	"net"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestPsync(t *testing.T) {
	// Reset store state for clean test
	shared.StoreState = shared.State{
		Role:             "master",
		MasterReplID:     "test-repl-id-123",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	}

	tests := []struct {
		name        string
		args        []shared.Value
		expectError bool
		expectedStr string
	}{
		{
			name: "PSYNC with valid arguments",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "?"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expectError: false,
			expectedStr: "FULLRESYNC test-repl-id-123 0",
		},
		{
			name: "PSYNC with specific repl ID and offset",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "some-repl-id"},
				{Typ: "bulk", Bulk: "100"},
			},
			expectError: false,
			expectedStr: "FULLRESYNC test-repl-id-123 0",
		},
		{
			name: "PSYNC with insufficient arguments",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "?"},
			},
			expectError: true,
		},
		{
			name:        "PSYNC with no arguments",
			args:        []shared.Value{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Psync("test-conn", tt.args)

			if tt.expectError {
				if result.Typ != "error" {
					t.Errorf("Expected error type, got %s", result.Typ)
				}
			} else {
				if result.Typ != "string" {
					t.Errorf("Expected string type, got %s", result.Typ)
				}
				if result.Str != tt.expectedStr {
					t.Errorf("Expected %q, got %q", tt.expectedStr, result.Str)
				}
			}
		})
	}
}

func TestPsyncWithDifferentMasterInfo(t *testing.T) {
	tests := []struct {
		name     string
		replID   string
		offset   int64
		expected string
	}{
		{
			name:     "Zero offset",
			replID:   "repl-001",
			offset:   0,
			expected: "FULLRESYNC repl-001 0",
		},
		{
			name:     "Non-zero offset",
			replID:   "repl-002",
			offset:   12345,
			expected: "FULLRESYNC repl-002 12345",
		},
		{
			name:     "Large offset",
			replID:   "repl-003",
			offset:   999999,
			expected: "FULLRESYNC repl-003 999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up store state
			shared.StoreState = shared.State{
				Role:             "master",
				MasterReplID:     tt.replID,
				MasterReplOffset: tt.offset,
				Replicas:         make(map[string]net.Conn),
			}

			args := []shared.Value{
				{Typ: "bulk", Bulk: "?"},
				{Typ: "bulk", Bulk: "-1"},
			}

			result := Psync("test-conn", args)

			if result.Typ != "string" {
				t.Errorf("Expected string type, got %s", result.Typ)
			}
			if result.Str != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Str)
			}
		})
	}
}

func TestGetRDBData(t *testing.T) {
	data, err := GetRDBData()
	if err != nil {
		t.Errorf("GetRDBData() returned error: %v", err)
	}

	if len(data) == 0 {
		t.Error("GetRDBData() returned empty data")
	}

	// The RDB data should start with "REDIS" magic string
	expectedStart := []byte("REDIS")
	if len(data) < len(expectedStart) {
		t.Error("RDB data too short")
	}

	for i, b := range expectedStart {
		if data[i] != b {
			t.Errorf("RDB data doesn't start with REDIS magic string")
			break
		}
	}
}

// BenchmarkPsync benchmarks the PSYNC command
func BenchmarkPsync(b *testing.B) {
	// Reset store state for clean benchmark
	shared.StoreState = shared.State{
		Role:             "master",
		MasterReplID:     "bench-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	}

	args := []shared.Value{
		{Typ: "bulk", Bulk: "?"},
		{Typ: "bulk", Bulk: "-1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Psync("bench-conn", args)
	}
}
