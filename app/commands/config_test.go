package commands

import (
	"net"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestConfigGet(t *testing.T) {
	// Reset store state for clean test
	server.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "test-repl-id",
		MasterReplOffset: 12345,
		Replicas:         make(map[string]net.Conn),
		ConfigDir:        "/tmp/redis-data",
		ConfigDbfilename: "rdbfile",
	})

	tests := []struct {
		name     string
		args     []shared.Value
		expected []string
	}{
		{
			name: "CONFIG GET single parameter",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "GET"},
				{Typ: "bulk", Bulk: "dir"},
			},
			expected: []string{"dir", "/tmp/redis-data"},
		},
		{
			name: "CONFIG GET multiple parameters",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "GET"},
				{Typ: "bulk", Bulk: "dir"},
				{Typ: "bulk", Bulk: "dbfilename"},
			},
			expected: []string{"dir", "/tmp/redis-data", "dbfilename", "rdbfile"},
		},
		{
			name: "CONFIG GET dbfilename parameter",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "GET"},
				{Typ: "bulk", Bulk: "dbfilename"},
			},
			expected: []string{"dbfilename", "rdbfile"},
		},
		{
			name: "CONFIG GET case insensitive",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "GET"},
				{Typ: "bulk", Bulk: "DIR"},
				{Typ: "bulk", Bulk: "Dbfilename"},
			},
			expected: []string{"DIR", "/tmp/redis-data", "Dbfilename", "rdbfile"},
		},
		{
			name: "CONFIG GET unknown parameter",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "GET"},
				{Typ: "bulk", Bulk: "unknown"},
			},
			expected: []string{"unknown", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Config("test-conn", tt.args)

			if result.Typ != "array" {
				t.Errorf("Expected array response, got %s", result.Typ)
				return
			}

			if len(result.Array) != len(tt.expected) {
				t.Errorf("Expected %d elements, got %d", len(tt.expected), len(result.Array))
				return
			}

			for i, expected := range tt.expected {
				if result.Array[i].Bulk != expected {
					t.Errorf("Expected %q at index %d, got %q", expected, i, result.Array[i].Bulk)
				}
			}
		})
	}
}

func TestConfigGetInvalidArgs(t *testing.T) {
	tests := []struct {
		name string
		args []shared.Value
	}{
		{
			name: "CONFIG without arguments",
			args: []shared.Value{},
		},
		{
			name: "CONFIG GET without parameters",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "GET"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Config("test-conn", tt.args)

			if result.Typ != "error" {
				t.Errorf("Expected error response, got %s", result.Typ)
			}

			if !contains(result.Str, "wrong number of arguments") {
				t.Errorf("Expected error message about wrong number of arguments, got '%s'", result.Str)
			}
		})
	}
}

func TestConfigUnknownSubcommand(t *testing.T) {
	args := []shared.Value{
		{Typ: "bulk", Bulk: "SET"},
		{Typ: "bulk", Bulk: "dir"},
		{Typ: "bulk", Bulk: "/new/path"},
	}

	result := Config("test-conn", args)

	if result.Typ != "error" {
		t.Errorf("Expected error response, got %s", result.Typ)
	}

	if !contains(result.Str, "unknown subcommand") {
		t.Errorf("Expected error message about unknown subcommand, got '%s'", result.Str)
	}
}

func TestConfigGetAllSupportedParameters(t *testing.T) {
	// Reset store state for clean test
	server.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "test-repl-id",
		MasterReplOffset: 12345,
		Replicas:         make(map[string]net.Conn),
		ConfigDir:        "/tmp/redis-data",
		ConfigDbfilename: "rdbfile",
	})

	// Test all supported configuration parameters
	supportedParams := []string{
		"dir", "dbfilename",
	}

	args := []shared.Value{{Typ: "bulk", Bulk: "GET"}}
	for _, param := range supportedParams {
		args = append(args, shared.Value{Typ: "bulk", Bulk: param})
	}

	result := Config("test-conn", args)

	if result.Typ != "array" {
		t.Errorf("Expected array response, got %s", result.Typ)
		return
	}

	// Should have 2 * len(supportedParams) elements (key-value pairs)
	expectedLen := 2 * len(supportedParams)
	if len(result.Array) != expectedLen {
		t.Errorf("Expected %d elements, got %d", expectedLen, len(result.Array))
		return
	}

	// Verify that all parameters are returned
	for i := 0; i < len(supportedParams); i++ {
		keyIndex := i * 2
		valueIndex := keyIndex + 1

		expectedKey := supportedParams[i]
		if result.Array[keyIndex].Bulk != expectedKey {
			t.Errorf("Expected key %q at index %d, got %q", expectedKey, keyIndex, result.Array[keyIndex].Bulk)
		}

		// Value should not be empty for supported parameters
		if result.Array[valueIndex].Bulk == "" {
			t.Errorf("Expected non-empty value for parameter %q", expectedKey)
		}
	}
}

// BenchmarkConfigGet benchmarks the CONFIG GET command
func BenchmarkConfigGet(b *testing.B) {
	// Reset store state for clean benchmark
	server.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "bench-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	})

	args := []shared.Value{
		{Typ: "bulk", Bulk: "GET"},
		{Typ: "bulk", Bulk: "dir"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Config("bench-conn", args)
	}
}

// BenchmarkConfigGetMultiple benchmarks the CONFIG GET command with multiple parameters
func BenchmarkConfigGetMultiple(b *testing.B) {
	// Reset store state for clean benchmark
	server.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "bench-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	})

	args := []shared.Value{
		{Typ: "bulk", Bulk: "GET"},
		{Typ: "bulk", Bulk: "dir"},
		{Typ: "bulk", Bulk: "port"},
		{Typ: "bulk", Bulk: "databases"},
		{Typ: "bulk", Bulk: "replication"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Config("bench-conn", args)
	}
}
