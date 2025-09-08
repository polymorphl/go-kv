package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestLpop(t *testing.T) {
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
			name:   "lpop single element",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "string", Str: "a"},
			verify: func() {
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LPOP")
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
			name:   "lpop multiple elements",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "3"},
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
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LPOP")
				}
				if len(entry.Array) != 2 {
					t.Errorf("Expected list length 2, got %d", len(entry.Array))
				}
				if entry.Array[0] != "d" {
					t.Errorf("Expected first element 'd', got '%s'", entry.Array[0])
				}
			},
		},
		{
			name:   "lpop count 0",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "0"},
			},
			setup: func() {
				shared.Memory["mylist"] = shared.MemoryEntry{
					Array:   []string{"a", "b", "c"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LPOP")
				}
				if len(entry.Array) != 3 {
					t.Errorf("Expected list length 3, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lpop count greater than list length",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "10"},
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
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LPOP")
				}
				if len(entry.Array) != 0 {
					t.Errorf("Expected list length 0, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lpop non-existent key",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "lpop empty list",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
			},
			setup: func() {
				shared.Memory["emptylist"] = shared.MemoryEntry{
					Array:   []string{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify: func() {
				entry, exists := shared.Memory["emptylist"]
				if !exists {
					t.Error("List should still exist after LPOP")
				}
				if len(entry.Array) != 0 {
					t.Errorf("Expected list length 0, got %d", len(entry.Array))
				}
			},
		},
		{
			name:   "lpop wrong type (string key)",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "stringkey"},
			},
			setup: func() {
				shared.Memory["stringkey"] = shared.MemoryEntry{
					Value:   "hello",
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify: func() {
				// Verify the string value is unchanged
				entry, exists := shared.Memory["stringkey"]
				if !exists {
					t.Error("String should still exist after LPOP")
				}
				if entry.Value != "hello" {
					t.Errorf("Expected string value 'hello', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "lpop invalid count (negative)",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
				{Typ: "bulk", Bulk: "-1"},
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
			name:   "lpop invalid count (non-integer)",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mylist"},
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
			name:     "wrong number of arguments",
			connID:   "test-conn-10",
			args:     []shared.Value{},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'lpop' command"},
			verify:   func() {},
		},
		{
			name:   "lpop with unicode elements",
			connID: "test-conn-11",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicodelist"},
				{Typ: "bulk", Bulk: "2"},
			},
			setup: func() {
				shared.Memory["unicodelist"] = shared.MemoryEntry{
					Array:   []string{"Hello 世界", "🌍", "测试", "end"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{
				{Typ: "string", Str: "Hello 世界"},
				{Typ: "string", Str: "🌍"},
			}},
			verify: func() {
				entry, exists := shared.Memory["unicodelist"]
				if !exists {
					t.Error("List should still exist after LPOP")
				}
				if len(entry.Array) != 2 {
					t.Errorf("Expected list length 2, got %d", len(entry.Array))
				}
				if entry.Array[0] != "测试" {
					t.Errorf("Expected first element '测试', got '%s'", entry.Array[0])
				}
			},
		},
		{
			name:   "lpop single element with count 1",
			connID: "test-conn-12",
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
			expected: shared.Value{Typ: "string", Str: "a"},
			verify: func() {
				entry, exists := shared.Memory["mylist"]
				if !exists {
					t.Error("List should still exist after LPOP")
				}
				if len(entry.Array) != 2 {
					t.Errorf("Expected list length 2, got %d", len(entry.Array))
				}
				if entry.Array[0] != "b" {
					t.Errorf("Expected first element 'b', got '%s'", entry.Array[0])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Lpop(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Lpop() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Lpop() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if len(result.Array) != len(tt.expected.Array) {
				t.Errorf("Lpop() array length = %v, expected %v", len(result.Array), len(tt.expected.Array))
			}

			for i, expectedItem := range tt.expected.Array {
				if i < len(result.Array) {
					if result.Array[i].Str != expectedItem.Str {
						t.Errorf("Lpop() array[%d] = %v, expected %v", i, result.Array[i].Str, expectedItem.Str)
					}
				}
			}

			tt.verify()
		})
	}
}

func BenchmarkLpop(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lpop(connID, args)
	}
}

func BenchmarkLpopMultiple(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lpop(connID, args)
	}
}

func BenchmarkLpopNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lpop(connID, args)
	}
}

func BenchmarkLpopEmpty(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Array:   []string{},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Lpop(connID, args)
	}
}
