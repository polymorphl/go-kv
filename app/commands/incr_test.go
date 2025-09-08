package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestIncr(t *testing.T) {
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
			name:   "increment new key",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "newcounter"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := shared.Memory["newcounter"]
				if !exists {
					t.Error("Key should exist after INCR")
				}
				if entry.Value != "1" {
					t.Errorf("Expected value '1', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "increment existing integer",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "existingcounter"},
			},
			setup: func() {
				shared.Memory["existingcounter"] = shared.MemoryEntry{Value: "5", Expires: 0}
			},
			expected: shared.Value{Typ: "integer", Num: 6},
			verify: func() {
				entry, exists := shared.Memory["existingcounter"]
				if !exists {
					t.Error("Key should exist after INCR")
				}
				if entry.Value != "6" {
					t.Errorf("Expected value '6', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "increment zero",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "zerocounter"},
			},
			setup: func() {
				shared.Memory["zerocounter"] = shared.MemoryEntry{Value: "0", Expires: 0}
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := shared.Memory["zerocounter"]
				if !exists {
					t.Error("Key should exist after INCR")
				}
				if entry.Value != "1" {
					t.Errorf("Expected value '1', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "increment negative number",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "negativecounter"},
			},
			setup: func() {
				shared.Memory["negativecounter"] = shared.MemoryEntry{Value: "-5", Expires: 0}
			},
			expected: shared.Value{Typ: "integer", Num: -4},
			verify: func() {
				entry, exists := shared.Memory["negativecounter"]
				if !exists {
					t.Error("Key should exist after INCR")
				}
				if entry.Value != "-4" {
					t.Errorf("Expected value '-4', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "increment large number",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "largecounter"},
			},
			setup: func() {
				shared.Memory["largecounter"] = shared.MemoryEntry{Value: "999999", Expires: 0}
			},
			expected: shared.Value{Typ: "integer", Num: 1000000},
			verify: func() {
				entry, exists := shared.Memory["largecounter"]
				if !exists {
					t.Error("Key should exist after INCR")
				}
				if entry.Value != "1000000" {
					t.Errorf("Expected value '1000000', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "increment non-integer value",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "stringcounter"},
			},
			setup: func() {
				shared.Memory["stringcounter"] = shared.MemoryEntry{Value: "hello", Expires: 0}
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not an integer or out of range"},
			verify: func() {
				// Value should remain unchanged after error
				entry, exists := shared.Memory["stringcounter"]
				if !exists {
					t.Error("Key should still exist after error")
				}
				if entry.Value != "hello" {
					t.Errorf("Expected value 'hello', got '%s'", entry.Value)
				}
			},
		},
		{
			name:   "increment array value (wrong type)",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "arraycounter"},
			},
			setup: func() {
				shared.Memory["arraycounter"] = shared.MemoryEntry{
					Value:   "",
					Array:   []string{"item1", "item2"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not an integer or out of range"},
			verify: func() {
				// Array should remain unchanged
				entry, exists := shared.Memory["arraycounter"]
				if !exists {
					t.Error("Key should still exist after error")
				}
				if len(entry.Array) != 2 {
					t.Error("Array should remain unchanged")
				}
			},
		},
		{
			name:     "wrong number of arguments",
			connID:   "test-conn-8",
			args:     []shared.Value{},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'incr' command"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Incr(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Incr() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Incr() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if result.Num != tt.expected.Num {
				t.Errorf("Incr() number = %v, expected %v", result.Num, tt.expected.Num)
			}

			tt.verify()
		})
	}
}

func BenchmarkIncr(b *testing.B) {
	clearMemory()
	shared.Memory["benchcounter"] = shared.MemoryEntry{Value: "0", Expires: 0}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchcounter"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Incr(connID, args)
	}
}

func BenchmarkIncrNewKey(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "newcounter"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Incr(connID, args)
	}
}

func BenchmarkIncrLargeNumber(b *testing.B) {
	clearMemory()
	shared.Memory["largecounter"] = shared.MemoryEntry{Value: "999999", Expires: 0}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "largecounter"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Incr(connID, args)
	}
}

func BenchmarkIncrNegativeNumber(b *testing.B) {
	clearMemory()
	shared.Memory["negativecounter"] = shared.MemoryEntry{Value: "-1000", Expires: 0}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "negativecounter"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Incr(connID, args)
	}
}
