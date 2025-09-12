package commands

import (
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func TestSet(t *testing.T) {
	// Clear memory before each test
	clearMemory()

	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
		verify   func() // Function to verify the result
	}{
		{
			name:   "set basic key-value",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mykey"},
				{Typ: "bulk", Bulk: "Hello World"},
			},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				entry, exists := server.Memory["mykey"]
				if !exists {
					t.Error("Key should exist after SET")
				}
				if entry.Value != "Hello World" {
					t.Errorf("Expected value 'Hello World', got '%s'", entry.Value)
				}
				if entry.Expires != 0 {
					t.Error("Key should not have expiration")
				}
			},
		},
		{
			name:   "set key with expiration",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "expiringkey"},
				{Typ: "bulk", Bulk: "expires soon"},
				{Typ: "bulk", Bulk: "PX"},
				{Typ: "bulk", Bulk: "1000"},
			},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				entry, exists := server.Memory["expiringkey"]
				if !exists {
					t.Error("Key should exist after SET")
				}
				if entry.Value != "expires soon" {
					t.Errorf("Expected value 'expires soon', got '%s'", entry.Value)
				}
				if entry.Expires <= time.Now().UnixMilli() {
					t.Error("Key should have future expiration")
				}
			},
		},
		{
			name:   "set empty value",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptykey"},
				{Typ: "bulk", Bulk: ""},
			},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				entry, exists := server.Memory["emptykey"]
				if !exists {
					t.Error("Key should exist after SET")
				}
				if entry.Value != "" {
					t.Errorf("Expected empty value, got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "overwrite existing key",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "overwritekey"},
				{Typ: "bulk", Bulk: "new value"},
			},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				entry, exists := server.Memory["overwritekey"]
				if !exists {
					t.Error("Key should exist after SET")
				}
				if entry.Value != "new value" {
					t.Errorf("Expected value 'new value', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "set with invalid PX value",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "invalidkey"},
				{Typ: "bulk", Bulk: "value"},
				{Typ: "bulk", Bulk: "PX"},
				{Typ: "bulk", Bulk: "notanumber"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not an integer or out of range"},
			verify: func() {
				// Key should not exist after error
				if _, exists := server.Memory["invalidkey"]; exists {
					t.Error("Key should not exist after error")
				}
			},
		},
		{
			name:   "wrong number of arguments",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'set' command"},
			verify:   func() {},
		},
		{
			name:   "set unicode value",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicodekey"},
				{Typ: "bulk", Bulk: "Hello 世界 🌍"},
			},
			expected: shared.Value{Typ: "string", Str: "OK"},
			verify: func() {
				entry, exists := server.Memory["unicodekey"]
				if !exists {
					t.Error("Key should exist after SET")
				}
				if entry.Value != "Hello 世界 🌍" {
					t.Errorf("Expected unicode value, got '%s'", entry.Value)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up initial data for overwrite test
			if tt.name == "overwrite existing key" {
				server.Memory["overwritekey"] = shared.MemoryEntry{Value: "old value", Expires: 0}
			}

			result := Set(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Set() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Set() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkSet(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "Hello World"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set(connID, args)
	}
}

func BenchmarkSetWithExpiration(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "Hello World"},
		{Typ: "bulk", Bulk: "PX"},
		{Typ: "bulk", Bulk: "1000"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set(connID, args)
	}
}

func BenchmarkSetEmptyValue(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: ""},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set(connID, args)
	}
}

func BenchmarkSetUnicode(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "Hello 世界 🌍"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Set(connID, args)
	}
}
