package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func TestRpush(t *testing.T) {
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
			name:   "rpush single value to new list",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "newlist"},
				{Typ: "bulk", Bulk: "first"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				_, exists := server.Memory["newlist"]
				if !exists {
					t.Error("List should exist after RPUSH")
				}
				array := getListAsArray("newlist")
				if len(array) != 1 {
					t.Errorf("Expected 1 element, got %d", len(array))
				}
				if array[0] != "first" {
					t.Errorf("Expected 'first', got '%s'", array[0])
				}
			},
		},
		{
			name:   "rpush multiple values to new list",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "newlist"},
				{Typ: "bulk", Bulk: "first"},
				{Typ: "bulk", Bulk: "second"},
				{Typ: "bulk", Bulk: "third"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "integer", Num: 3},
			verify: func() {
				_, exists := server.Memory["newlist"]
				if !exists {
					t.Error("List should exist after RPUSH")
				}
				array := getListAsArray("newlist")
				if len(array) != 3 {
					t.Errorf("Expected 3 elements, got %d", len(array))
				}
				// Values should be in the same order as pushed
				expected := []string{"first", "second", "third"}
				for i, val := range expected {
					if array[i] != val {
						t.Errorf("Expected '%s' at position %d, got '%s'", val, i, array[i])
					}
				}
			},
		},
		{
			name:   "rpush to existing list",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "existinglist"},
				{Typ: "bulk", Bulk: "newlast"},
			},
			setup: func() {
				server.Memory["existinglist"] = shared.MemoryEntry{
					Value:   "",
					Array:   []string{"old1", "old2"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "integer", Num: 3},
			verify: func() {
				_, exists := server.Memory["existinglist"]
				if !exists {
					t.Error("List should exist after RPUSH")
				}
				array := getListAsArray("existinglist")
				if len(array) != 3 {
					t.Errorf("Expected 3 elements, got %d", len(array))
				}
				expected := []string{"old1", "old2", "newlast"}
				for i, val := range expected {
					if array[i] != val {
						t.Errorf("Expected '%s' at position %d, got '%s'", val, i, array[i])
					}
				}
			},
		},
		{
			name:   "rpush to empty list",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
				{Typ: "bulk", Bulk: "first"},
			},
			setup: func() {
				server.Memory["emptylist"] = shared.MemoryEntry{
					Value:   "",
					Array:   []string{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				_, exists := server.Memory["emptylist"]
				if !exists {
					t.Error("List should exist after RPUSH")
				}
				array := getListAsArray("emptylist")
				if len(array) != 1 {
					t.Errorf("Expected 1 element, got %d", len(array))
				}
				if array[0] != "first" {
					t.Errorf("Expected 'first', got '%s'", array[0])
				}
			},
		},
		{
			name:   "rpush converts string to list",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "stringkey"},
				{Typ: "bulk", Bulk: "first"},
			},
			setup: func() {
				server.Memory["stringkey"] = shared.MemoryEntry{Value: "old string", Expires: 0}
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				_, exists := server.Memory["stringkey"]
				if !exists {
					t.Error("Key should exist after RPUSH")
				}
				array := getListAsArray("stringkey")
				if len(array) != 1 {
					t.Errorf("Expected 1 element, got %d", len(array))
				}
				if array[0] != "first" {
					t.Errorf("Expected 'first', got '%s'", array[0])
				}
				// Original string value should be cleared
				entry := server.Memory["stringkey"]
				if entry.Value != "" {
					t.Errorf("Expected empty string value, got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "rpush empty string",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
				{Typ: "bulk", Bulk: ""},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				_, exists := server.Memory["emptylist"]
				if !exists {
					t.Error("List should exist after RPUSH")
				}
				array := getListAsArray("emptylist")
				if len(array) != 1 {
					t.Errorf("Expected 1 element, got %d", len(array))
				}
				if array[0] != "" {
					t.Errorf("Expected empty string, got '%s'", array[0])
				}
			},
		},
		{
			name:   "rpush unicode values",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicodelist"},
				{Typ: "bulk", Bulk: "Hello 世界"},
				{Typ: "bulk", Bulk: "🌍"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "integer", Num: 2},
			verify: func() {
				_, exists := server.Memory["unicodelist"]
				if !exists {
					t.Error("List should exist after RPUSH")
				}
				array := getListAsArray("unicodelist")
				if len(array) != 2 {
					t.Errorf("Expected 2 elements, got %d", len(array))
				}
				expected := []string{"Hello 世界", "🌍"}
				for i, val := range expected {
					if array[i] != val {
						t.Errorf("Expected '%s' at position %d, got '%s'", val, i, array[i])
					}
				}
			},
		},
		{
			name:   "wrong number of arguments",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'rpush' command"},
			verify:   func() {},
		},
		{
			name:   "rpush large number of values",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "largelist"},
				{Typ: "bulk", Bulk: "val1"},
				{Typ: "bulk", Bulk: "val2"},
				{Typ: "bulk", Bulk: "val3"},
				{Typ: "bulk", Bulk: "val4"},
				{Typ: "bulk", Bulk: "val5"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "integer", Num: 5},
			verify: func() {
				_, exists := server.Memory["largelist"]
				if !exists {
					t.Error("List should exist after RPUSH")
				}
				array := getListAsArray("largelist")
				if len(array) != 5 {
					t.Errorf("Expected 5 elements, got %d", len(array))
				}
				// Values should be in the same order as pushed
				expected := []string{"val1", "val2", "val3", "val4", "val5"}
				for i, val := range expected {
					if array[i] != val {
						t.Errorf("Expected '%s' at position %d, got '%s'", val, i, array[i])
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Rpush(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Rpush() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Rpush() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if result.Num != tt.expected.Num {
				t.Errorf("Rpush() number = %v, expected %v", result.Num, tt.expected.Num)
			}

			tt.verify()
		})
	}
}

func BenchmarkRpush(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "value"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Rpush(connID, args)
	}
}

func BenchmarkRpushMultiple(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "val1"},
		{Typ: "bulk", Bulk: "val2"},
		{Typ: "bulk", Bulk: "val3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Rpush(connID, args)
	}
}

func BenchmarkRpushToExisting(b *testing.B) {
	clearMemory()
	server.Memory["benchlist"] = shared.MemoryEntry{
		Value:   "",
		Array:   []string{"existing1", "existing2"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: "newvalue"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Rpush(connID, args)
	}
}

func BenchmarkRpushEmpty(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
		{Typ: "bulk", Bulk: ""},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Rpush(connID, args)
	}
}
