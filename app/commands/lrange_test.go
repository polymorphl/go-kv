package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestLrange(t *testing.T) {
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
			name:   "lrange basic range",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "2"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c", "d", "e"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "a"},
				{Typ: "string", Str: "b"},
				{Typ: "string", Str: "c"},
			}},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 5 {
					t.Errorf("Expected list length 5, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange all elements",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-1"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "a"},
				{Typ: "string", Str: "b"},
				{Typ: "string", Str: "c"},
			}},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange negative indices",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "-3"},
				{Typ: "bulk", Bulk: "-1"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c", "d", "e"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "c"},
				{Typ: "string", Str: "d"},
				{Typ: "string", Str: "e"},
			}},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 5 {
					t.Errorf("Expected list length 5, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange single element",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "1"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "b"},
			}},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange invalid range (start > stop)",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "3"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange out of bounds (start too large)",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "10"},
				{Typ: "bulk", Bulk: "15"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange non-existent key",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify:   func() {},
		},
		{
			name:   "lrange empty list",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["emptylist"] = shared.MemoryEntry{
					Array:   []string{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["emptylist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 0 {
					t.Errorf("Expected list length 0, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange wrong type (string key)",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "stringkey"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["stringkey"] = shared.MemoryEntry{
					Value:   "hello",
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"},
			verify: func() {
				// Verify the string value is unchanged
				entry, exists := shared.Memory["stringkey"]
				if !exists {
					t.Error("String should still exist after error")
				}
				if entry.Value != "hello" {
					t.Errorf("Expected string value 'hello', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "lrange invalid start index",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "invalid"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not an integer or out of range"},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after error")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lrange invalid stop index",
			connID: "test-conn-11",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "invalid"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not an integer or out of range"},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after error")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "wrong number of arguments",
			connID: "test-conn-12",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "0"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'lrange' command"},
			verify:   func() {},
		},
		{
			name:   "lrange with unicode elements",
			connID: "test-conn-13",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicodelist"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["unicodelist"] = shared.MemoryEntry{
					Array:   []string{"Hello 世界", "🌍", "测试"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "Hello 世界"},
				{Typ: "string", Str: "🌍"},
			}},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["unicodelist"]
				if !exists {
					t.Error("List should still exist after LRANGE")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Lrange(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Lrange() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Lrange() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if len(result.Array) != len(tt.expected.Array) {
				t.Errorf("Lrange() array length = %v, expected %v", len(result.Array), len(tt.expected.Array))
			}

			for i, expectedItem := range tt.expected.Array {
				if i < len(result.Array) {
					if result.Array[i].Str != expectedItem.Str {
						t.Errorf("Lrange() array[%d] = %v, expected %v", i, result.Array[i].Str, expectedItem.Str)
					}
				}
			}

			tt.verify()
		})
	}
}

func BenchmarkLrange(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "0"},
		{Typ: "bulk", Bulk: "4"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lrange(connID, args)
	}
}

func BenchmarkLrangeAll(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "0"},
		{Typ: "bulk", Bulk: "-1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lrange(connID, args)
	}
}

func BenchmarkLrangeNegativeIndices(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "-3"},
		{Typ: "bulk", Bulk: "-1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lrange(connID, args)
	}
}

func BenchmarkLrangeNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
		{Typ: "bulk", Bulk: "0"},
		{Typ: "bulk", Bulk: "1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lrange(connID, args)
	}
}

func BenchmarkLrangeLargeList(b *testing.B) {
	clearMemory()
	// Create a large list with 1000 elements
	list := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		list[i] = fmt.Sprintf("item-%d", i)
	}
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   list,
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "0"},
		{Typ: "bulk", Bulk: "99"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lrange(connID, args)
	}
}
