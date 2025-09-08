package commands

import (
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestType(t *testing.T) {
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
			name:   "type of string key",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "stringkey"},
			},
			setup: func() {
				shared.Memory["stringkey"] = shared.MemoryEntry{Value: "hello world", Expires: 0}
			},
			expected: shared.Value{Typ: "string", Str: "string"},
		},
		{
			name:   "type of list key",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "listkey"},
			},
			setup: func() {
				shared.Memory["listkey"] = shared.MemoryEntry{
					Value:   "",
					Array:   []string{"item1", "item2", "item3"},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "string", Str: "list"},
		},
		{
			name:   "type of stream key",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streamkey"},
			},
			setup: func() {
				shared.Memory["streamkey"] = shared.MemoryEntry{
					Value: "",
					Stream: []shared.StreamEntry{
						{ID: "1234567890-0", Data: map[string]string{"field1": "value1", "field2": "value2"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "string", Str: "stream"},
		},
		{
			name:   "type of non-existent key",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "string", Str: "none"},
		},
		{
			name:   "type of empty string key",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptykey"},
			},
			setup: func() {
				shared.Memory["emptykey"] = shared.MemoryEntry{Value: "", Expires: 0}
			},
			expected: shared.Value{Typ: "string", Str: "none"},
		},
		{
			name:   "type of empty list key",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptylist"},
			},
			setup: func() {
				shared.Memory["emptylist"] = shared.MemoryEntry{
					Value:   "",
					Array:   []string{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "string", Str: "none"},
		},
		{
			name:   "type of empty stream key",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "emptystream"},
			},
			setup: func() {
				shared.Memory["emptystream"] = shared.MemoryEntry{
					Value:   "",
					Stream:  []shared.StreamEntry{},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "string", Str: "none"},
		},
		{
			name:   "type of expired key",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "expiredkey"},
			},
			setup: func() {
				shared.Memory["expiredkey"] = shared.MemoryEntry{
					Value:   "expired value",
					Expires: time.Now().UnixMilli() - 1000, // Expired 1 second ago
				}
			},
			expected: shared.Value{Typ: "string", Str: "string"},
		},
		{
			name:     "wrong number of arguments",
			connID:   "test-conn-9",
			args:     []shared.Value{},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'type' command"},
		},
		{
			name:   "type of unicode key",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicodekey"},
			},
			setup: func() {
				shared.Memory["unicodekey"] = shared.MemoryEntry{Value: "Hello 世界 🌍", Expires: 0}
			},
			expected: shared.Value{Typ: "string", Str: "string"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Type(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Type() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Type() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			if result.Num != tt.expected.Num {
				t.Errorf("Type() number = %v, expected %v", result.Num, tt.expected.Num)
			}
		})
	}
}

func BenchmarkType(b *testing.B) {
	clearMemory()
	shared.Memory["benchkey"] = shared.MemoryEntry{Value: "Hello World", Expires: 0}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Type(connID, args)
	}
}

func BenchmarkTypeList(b *testing.B) {
	clearMemory()
	shared.Memory["benchlist"] = shared.MemoryEntry{
		Value:   "",
		Array:   []string{"item1", "item2", "item3"},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchlist"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Type(connID, args)
	}
}

func BenchmarkTypeStream(b *testing.B) {
	clearMemory()
	shared.Memory["benchstream"] = shared.MemoryEntry{
		Value: "",
		Stream: []shared.StreamEntry{
			{ID: "1234567890-0", Data: map[string]string{"field1": "value1", "field2": "value2"}},
		},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Type(connID, args)
	}
}

func BenchmarkTypeNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Type(connID, args)
	}
}
