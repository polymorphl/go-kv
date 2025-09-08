package commands

import (
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestGet(t *testing.T) {
	// Clear memory before each test
	clearMemory()

	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		setup    func() // Function to set up test data
		expected shared.Value
	}{
		{
			name:   "get existing key",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mykey"},
			},
			setup: func() {
				shared.Memory["mykey"] = shared.MemoryEntry{Value: "Hello World", Expires: 0}
			},
			expected: shared.Value{Typ: "string", Str: "Hello World"},
		},
		{
			name:   "get non-existent key",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "null", Str: ""},
		},
		{
			name:   "get expired key",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "expiredkey"},
			},
			setup: func() {
				shared.Memory["expiredkey"] = shared.MemoryEntry{
					Value:   "expired value",
					Expires: time.Now().UnixMilli() - 1000, // Expired 1 second ago
				}
			},
			expected: shared.Value{Typ: "null", Str: ""},
		},
		{
			name:   "get key with array (wrong type)",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "arraykey"},
			},
			setup: func() {
				shared.Memory["arraykey"] = shared.MemoryEntry{
					Value:   "",
					Array:   []string{"item1", "item2"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"},
		},
		{
			name:   "get empty string value",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptykey"},
			},
			setup: func() {
				shared.Memory["emptykey"] = shared.MemoryEntry{Value: "", Expires: 0}
			},
			expected: shared.Value{Typ: "string", Str: ""},
		},
		{
			name:     "wrong number of arguments",
			connID:   "test-conn-6",
			args:     []shared.Value{},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'get' command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Get(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Get() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Get() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if result.Num != tt.expected.Num {
				t.Errorf("Get() number = %v, expected %v", result.Num, tt.expected.Num)
			}
		})
	}
}

func BenchmarkGet(b *testing.B) {
	clearMemory()
	shared.Memory["benchkey"] = shared.MemoryEntry{Value: "Hello World", Expires: 0}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Get(connID, args)
	}
}

func BenchmarkGetNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Get(connID, args)
	}
}

func BenchmarkGetExpired(b *testing.B) {
	clearMemory()
	shared.Memory["expiredkey"] = shared.MemoryEntry{
		Value:   "expired value",
		Expires: time.Now().UnixMilli() - 1000,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "expiredkey"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Get(connID, args)
	}
}
