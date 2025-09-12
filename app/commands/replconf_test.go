package commands

import (
	"net"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestReplconfGetack(t *testing.T) {
	// Test REPLCONF GETACK * command
	args := []shared.Value{
		{Typ: "bulk", Bulk: "GETACK"},
		{Typ: "bulk", Bulk: "*"},
	}

	result := Replconf("test-conn", args)

	// Should return an array response with REPLCONF ACK 0
	if result.Typ != "array" {
		t.Errorf("Expected array response, got %s", result.Typ)
	}

	if len(result.Array) != 3 {
		t.Errorf("Expected array with 3 elements, got %d", len(result.Array))
	}

	if result.Array[0].Bulk != "REPLCONF" {
		t.Errorf("Expected first element to be 'REPLCONF', got '%s'", result.Array[0].Bulk)
	}

	if result.Array[1].Bulk != "ACK" {
		t.Errorf("Expected second element to be 'ACK', got '%s'", result.Array[1].Bulk)
	}

	if result.Array[2].Bulk != "0" {
		t.Errorf("Expected third element to be '0', got '%s'", result.Array[2].Bulk)
	}
}

func TestReplconfGetackInvalidArgs(t *testing.T) {
	// Test REPLCONF GETACK with invalid arguments
	args := []shared.Value{
		{Typ: "bulk", Bulk: "GETACK"},
		// Missing "*" argument
	}

	result := Replconf("test-conn", args)

	// Should return an error
	if result.Typ != "error" {
		t.Errorf("Expected error response, got %s", result.Typ)
	}

	if !contains(result.Str, "wrong number of arguments") {
		t.Errorf("Expected error message about wrong number of arguments, got '%s'", result.Str)
	}
}

func TestReplconfListeningPort(t *testing.T) {
	// Test REPLCONF listening-port command (existing functionality)
	args := []shared.Value{
		{Typ: "bulk", Bulk: "listening-port"},
		{Typ: "bulk", Bulk: "6380"},
	}

	result := Replconf("test-conn", args)

	// Should return OK
	if result.Typ != "string" {
		t.Errorf("Expected string response, got %s", result.Typ)
	}

	if result.Str != "OK" {
		t.Errorf("Expected 'OK', got '%s'", result.Str)
	}
}

func TestReplconfInsufficientArgs(t *testing.T) {
	// Test REPLCONF with insufficient arguments
	args := []shared.Value{
		{Typ: "bulk", Bulk: "GETACK"},
		// Only one argument, need at least 2
	}

	result := Replconf("test-conn", args)

	// Should return an error
	if result.Typ != "error" {
		t.Errorf("Expected error response, got %s", result.Typ)
	}

	if !contains(result.Str, "wrong number of arguments") {
		t.Errorf("Expected error message about wrong number of arguments, got '%s'", result.Str)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && contains(s[1:], substr)
}

// BenchmarkReplconf benchmarks the REPLCONF command
func BenchmarkReplconf(b *testing.B) {
	// Reset store state for clean benchmark
	shared.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "bench-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	})

	args := []shared.Value{
		{Typ: "bulk", Bulk: "listening-port"},
		{Typ: "bulk", Bulk: "6379"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Replconf("bench-conn", args)
	}
}

// BenchmarkReplconfGetack benchmarks the REPLCONF GETACK command
func BenchmarkReplconfGetack(b *testing.B) {
	// Reset store state for clean benchmark
	shared.SetStoreState(shared.State{
		Role:             "master",
		MasterReplID:     "bench-repl-id",
		MasterReplOffset: 0,
		Replicas:         make(map[string]net.Conn),
	})

	args := []shared.Value{
		{Typ: "bulk", Bulk: "GETACK"},
		{Typ: "bulk", Bulk: "*"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Replconf("bench-conn", args)
	}
}
