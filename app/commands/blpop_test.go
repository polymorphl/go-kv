package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestBlpop(t *testing.T) {
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
			name:   "blpop immediate result",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "mylist"},
				{Typ: "string", Str: "a"},
			}},
			verify: func() {
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after BLPOP")
				}
				if len(entry.Array) != 2 {
					t.Errorf("Expected list length 2, got %d", len(entry.Array))
				}
				if entry.Array[0] != "b" {
					t.Errorf("Expected first element 'b', got '%s'", entry.Array[0])
				}
			},
		},
		{
			name:   "blpop multiple lists - first has elements",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "list1"},
				{Typ: "bulk", Bulk: "list2"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["list1"] = shared.MemoryEntry{
					Array:   []string{"first"},
					Expires: 0,
				}
				shared.Memory["list2"] = shared.MemoryEntry{
					Array:   []string{"second"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "list1"},
				{Typ: "string", Str: "first"},
			}},
			verify: func() {
				// Verify list1 is popped from
				entry1, exists := shared.Memory["list1"]
				if !exists {
					t.Error("List1 should still exist after BLPOP")
				}
				if len(entry1.Array) != 0 {
					t.Errorf("Expected list1 length 0, got %d", len(entry1.Array))
				}

				// Verify list2 is unchanged
				entry2, exists := shared.Memory["list2"]
				if !exists {
					t.Error("List2 should still exist after BLPOP")
				}
				if len(entry2.Array) != 1 {
					t.Errorf("Expected list2 length 1, got %d", len(entry2.Array))
				}
			},
		},
		{
			name:   "blpop timeout with empty lists",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
				{Typ: "bulk", Bulk: "0.1"}, // 100ms timeout
			},
			setup: func() {
				shared.Memory["emptylist"] = shared.MemoryEntry{
					Array:   []string{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "null_array", Str: ""},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["emptylist"]
				if !exists {
					t.Error("List should still exist after BLPOP timeout")
				}
				if len(entry.Array) != 0 {
					t.Errorf("Expected list length 0, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "blpop non-existent key",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "0.1"}, // 100ms timeout
			},
			setup:    func() {},
			expected: shared.Value{Typ: "null_array", Str: ""},
			verify:   func() {},
		},
		{
			name:   "blpop very short timeout",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
				{Typ: "bulk", Bulk: "0.001"}, // 1ms timeout
			},
			setup: func() {
				shared.Memory["emptylist"] = shared.MemoryEntry{
					Array:   []string{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "null_array", Str: ""},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["emptylist"]
				if !exists {
					t.Error("List should still exist after BLPOP")
				}
				if len(entry.Array) != 0 {
					t.Errorf("Expected list length 0, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "blpop fractional timeout",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "0.5"}, // 500ms timeout
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "mylist"},
				{Typ: "string", Str: "a"},
			}},
			verify: func() {
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after BLPOP")
				}
				if len(entry.Array) != 1 {
					t.Errorf("Expected list length 1, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "blpop invalid timeout (negative)",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "-1"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "ERR timeout is not a float or out of range"},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after error")
				}
				if len(entry.Array) != 2 {
					t.Errorf("Expected list length 2, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "blpop invalid timeout (non-numeric)",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "invalid"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "ERR timeout is not a float or out of range"},
			verify: func() {
				// Verify the list is unchanged
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after error")
				}
				if len(entry.Array) != 2 {
					t.Errorf("Expected list length 2, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "wrong number of arguments (too few)",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'blpop' command"},
			verify:   func() {},
		},
		{
			name:   "blpop with unicode elements",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicodelist"},
				{Typ: "bulk", Bulk: "1"},
			},
			setup: func() {
				shared.Memory["unicodelist"] = shared.MemoryEntry{
					Array:   []string{"Hello 世界", "🌍", "测试"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "unicodelist"},
				{Typ: "string", Str: "Hello 世界"},
			}},
			verify: func() {
				entry, exists := shared.Memory["unicodelist"]
				if !exists {
					t.Error("List should still exist after BLPOP")
				}
				if len(entry.Array) != 2 {
					t.Errorf("Expected list length 2, got %d", len(entry.Array))
				}
				if entry.Array[0] != "🌍" {
					t.Errorf("Expected first element '🌍', got '%s'", entry.Array[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Blpop(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Blpop() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Blpop() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if len(result.Array) != len(tt.expected.Array) {
				t.Errorf("Blpop() array length = %v, expected %v", len(result.Array), len(tt.expected.Array))
			}

			for i, expectedItem := range tt.expected.Array {
				if i < len(result.Array) {
					if result.Array[i].Str != expectedItem.Str {
						t.Errorf("Blpop() array[%d] = %v, expected %v", i, result.Array[i].Str, expectedItem.Str)
					}
				}
			}

			tt.verify()
		})
	}
}

func BenchmarkBlpop(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Blpop(connID, args)
	}
}

func BenchmarkBlpopMultipleLists(b *testing.B) {
	clearMemory()
	shared.Memory["list1"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c"},
		Expires: 0,
	}
	shared.Memory["list2"] = shared.MemoryEntry{
		Array:   []string{"d", "e", "f"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "list1"},
		{Typ: "bulk", Bulk: "list2"},
		{Typ: "bulk", Bulk: "1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Blpop(connID, args)
	}
}

func BenchmarkBlpopNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
		{Typ: "bulk", Bulk: "0.001"}, // Very short timeout for benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Blpop(connID, args)
	}
}

func BenchmarkBlpopEmpty(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "0.001"}, // Very short timeout for benchmark
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Blpop(connID, args)
	}
}
