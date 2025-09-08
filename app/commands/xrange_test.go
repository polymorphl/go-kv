package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestXrange(t *testing.T) {
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
			name:   "xrange all entries",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "-"},
				{Typ: "bulk", Bulk: "+"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
						{ID: "2-0", Data: map[string]string{"message": "World"}},
						{ID: "3-0", Data: map[string]string{"message": "Test"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get 3 entries
				result := Xrange("test-conn-1", []shared.Value{
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "-"},
					{Typ: "bulk", Bulk: "+"},
				})
				if len(result.Array) != 3 {
					t.Errorf("Expected 3 entries, got %d", len(result.Array))
				}
			},
		},
		{
			name:   "xrange specific ID range",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "2-0"},
				{Typ: "bulk", Bulk: "3-0"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
						{ID: "2-0", Data: map[string]string{"message": "World"}},
						{ID: "3-0", Data: map[string]string{"message": "Test"}},
						{ID: "4-0", Data: map[string]string{"message": "Extra"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get entries 2-0 and 3-0
				result := Xrange("test-conn-2", []shared.Value{
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "2-0"},
					{Typ: "bulk", Bulk: "3-0"},
				})
				if len(result.Array) != 2 {
					t.Errorf("Expected 2 entries, got %d", len(result.Array))
				}
				if result.Array[0].Array[0].Bulk != "2-0" {
					t.Errorf("Expected first entry ID '2-0', got '%s'", result.Array[0].Array[0].Bulk)
				}
				if result.Array[1].Array[0].Bulk != "3-0" {
					t.Errorf("Expected second entry ID '3-0', got '%s'", result.Array[1].Array[0].Bulk)
				}
			},
		},
		{
			name:   "xrange from specific ID to end",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "2-0"},
				{Typ: "bulk", Bulk: "+"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
						{ID: "2-0", Data: map[string]string{"message": "World"}},
						{ID: "3-0", Data: map[string]string{"message": "Test"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get entries from 2-0 onwards
				result := Xrange("test-conn-3", []shared.Value{
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "2-0"},
					{Typ: "bulk", Bulk: "+"},
				})
				if len(result.Array) != 2 {
					t.Errorf("Expected 2 entries, got %d", len(result.Array))
				}
			},
		},
		{
			name:   "xrange from beginning to specific ID",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "-"},
				{Typ: "bulk", Bulk: "2-0"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
						{ID: "2-0", Data: map[string]string{"message": "World"}},
						{ID: "3-0", Data: map[string]string{"message": "Test"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get entries up to 2-0
				result := Xrange("test-conn-4", []shared.Value{
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "-"},
					{Typ: "bulk", Bulk: "2-0"},
				})
				if len(result.Array) != 2 {
					t.Errorf("Expected 2 entries, got %d", len(result.Array))
				}
			},
		},
		{
			name:   "xrange non-existent stream",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "-"},
				{Typ: "bulk", Bulk: "+"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify:   func() {},
		},
		{
			name:   "xrange empty stream",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptystream"},
				{Typ: "bulk", Bulk: "-"},
				{Typ: "bulk", Bulk: "+"},
			},
			setup: func() {
				shared.Memory["emptystream"] = shared.MemoryEntry{
					Stream:  []shared.StreamEntry{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify:   func() {},
		},
		{
			name:   "xrange single entry",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "1-0"},
				{Typ: "bulk", Bulk: "1-0"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get exactly one entry
				result := Xrange("test-conn-7", []shared.Value{
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "1-0"},
					{Typ: "bulk", Bulk: "1-0"},
				})
				if len(result.Array) != 1 {
					t.Errorf("Expected 1 entry, got %d", len(result.Array))
				}
			},
		},
		{
			name:   "xrange with multiple field-value pairs",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "-"},
				{Typ: "bulk", Bulk: "+"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{
							"temperature": "25",
							"humidity":    "60",
							"pressure":    "1013",
						}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get the entry with multiple fields
				result := Xrange("test-conn-8", []shared.Value{
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "-"},
					{Typ: "bulk", Bulk: "+"},
				})
				if len(result.Array) != 1 {
					t.Errorf("Expected 1 entry, got %d", len(result.Array))
				}
				entry := result.Array[0]
				if len(entry.Array) != 2 {
					t.Errorf("Expected entry with ID and data, got %d parts", len(entry.Array))
				}
			},
		},
		{
			name:   "wrong number of arguments",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "-"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'xrange' command"},
			verify:   func() {},
		},
		{
			name:   "xrange with unicode data",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "-"},
				{Typ: "bulk", Bulk: "+"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{
							"消息": "你好世界 🌍",
							"温度": "25°C",
						}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get the unicode entry
				result := Xrange("test-conn-10", []shared.Value{
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "-"},
					{Typ: "bulk", Bulk: "+"},
				})
				if len(result.Array) != 1 {
					t.Errorf("Expected 1 entry, got %d", len(result.Array))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Xrange(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Xrange() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Xrange() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkXrange(b *testing.B) {
	clearMemory()
	shared.Memory["benchstream"] = shared.MemoryEntry{
		Stream: []shared.StreamEntry{
			{ID: "1-0", Data: map[string]string{"message": "Hello"}},
			{ID: "2-0", Data: map[string]string{"message": "World"}},
			{ID: "3-0", Data: map[string]string{"message": "Test"}},
			{ID: "4-0", Data: map[string]string{"message": "Extra"}},
			{ID: "5-0", Data: map[string]string{"message": "More"}},
		},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "-"},
		{Typ: "bulk", Bulk: "+"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xrange(connID, args)
	}
}

func BenchmarkXrangeSpecificRange(b *testing.B) {
	clearMemory()
	shared.Memory["benchstream"] = shared.MemoryEntry{
		Stream: []shared.StreamEntry{
			{ID: "1-0", Data: map[string]string{"message": "Hello"}},
			{ID: "2-0", Data: map[string]string{"message": "World"}},
			{ID: "3-0", Data: map[string]string{"message": "Test"}},
			{ID: "4-0", Data: map[string]string{"message": "Extra"}},
			{ID: "5-0", Data: map[string]string{"message": "More"}},
		},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "2-0"},
		{Typ: "bulk", Bulk: "4-0"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xrange(connID, args)
	}
}

func BenchmarkXrangeNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
		{Typ: "bulk", Bulk: "-"},
		{Typ: "bulk", Bulk: "+"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xrange(connID, args)
	}
}

func BenchmarkXrangeLargeStream(b *testing.B) {
	clearMemory()
	// Create a large stream with 100 entries
	stream := make([]shared.StreamEntry, 100)
	for i := 0; i < 100; i++ {
		stream[i] = shared.StreamEntry{
			ID: fmt.Sprintf("%d-0", i+1),
			Data: map[string]string{
				"message": fmt.Sprintf("Entry %d", i+1),
				"value":   fmt.Sprintf("%d", i+1),
			},
		}
	}
	shared.Memory["benchstream"] = shared.MemoryEntry{
		Stream:  stream,
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "-"},
		{Typ: "bulk", Bulk: "+"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xrange(connID, args)
	}
}
