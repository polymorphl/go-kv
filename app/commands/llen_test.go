package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func TestLlen(t *testing.T) {
	// Clear memory before each test
	clearMemory()

	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		setup    func() // Function to set up test data
		expected shared.Value
		verify   func() // Function to verify the result
	}{
		{
			name:   "llen existing list",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
			},
			setup: func() {
				server.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c", "d", "e"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "integer", Num: 5},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := server.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LLEN")
				}
				if len(entry.Array) != 5 {
					t.Errorf("Expected list length 5, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "llen empty list",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
			},
			setup: func() {
				server.Memory["emptylist"] = shared.MemoryEntry{
					Array:   []string{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := server.Memory["emptylist"]
				if !exists {
					t.Error("List should still exist after LLEN")
				}
				if len(entry.Array) != 0 {
					t.Errorf("Expected list length 0, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "llen non-existent key",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify:   func() {},
		},
		{
			name:   "llen single element list",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "singlelist"},
			},
			setup: func() {
				server.Memory["singlelist"] = shared.MemoryEntry{
					Array:   []string{"only"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := server.Memory["singlelist"]
				if !exists {
					t.Error("List should still exist after LLEN")
				}
				if len(entry.Array) != 1 {
					t.Errorf("Expected list length 1, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "llen large list",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "largelist"},
			},
			setup: func() {
				// Create a list with 1000 elements
				list := make([]string, 1000)
				for i := 0; i < 1000; i++ {
					list[i] = fmt.Sprintf("item-%d", i)
				}
				server.Memory["largelist"] = shared.MemoryEntry{
					Array:   list,
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "integer", Num: 1000},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := server.Memory["largelist"]
				if !exists {
					t.Error("List should still exist after LLEN")
				}
				if len(entry.Array) != 1000 {
					t.Errorf("Expected list length 1000, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "llen with unicode elements",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicodelist"},
			},
			setup: func() {
				server.Memory["unicodelist"] = shared.MemoryEntry{
					Array:   []string{"Hello 世界", "🌍", "测试", "end"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "integer", Num: 4},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := server.Memory["unicodelist"]
				if !exists {
					t.Error("List should still exist after LLEN")
				}
				if len(entry.Array) != 4 {
					t.Errorf("Expected list length 4, got %d", len(entry.Array))
				}
			},
		},
		{
			name:     "wrong number of arguments",
			connID:   "test-conn-7",
			args:     []shared.Value{},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'llen' command"},
			verify:   func() {},
		},
		{
			name:   "wrong number of arguments (too many)",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "extra"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'llen' command"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Llen(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Llen() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Llen() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if result.Num != tt.expected.Num {
				t.Errorf("Llen() number = %v, expected %v", result.Num, tt.expected.Num)
			}

			tt.verify()
		})
	}
}

func BenchmarkLlen(b *testing.B) {
	clearMemory()
	server.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Llen(connID, args)
	}
}

func BenchmarkLlenEmpty(b *testing.B) {
	clearMemory()
	server.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Llen(connID, args)
	}
}

func BenchmarkLlenNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Llen(connID, args)
	}
}

func BenchmarkLlenLarge(b *testing.B) {
	clearMemory()
	// Create a large list with 1000 elements
	list := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		list[i] = fmt.Sprintf("item-%d", i)
	}
	server.Memory["benchlist"] = shared.MemoryEntry{
		Array:   list,
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Llen(connID, args)
	}
}
