package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestXread(t *testing.T) {
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
			name:   "xread single stream from beginning",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "0-0"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
						{ID: "2-0", Data: map[string]string{"message": "World"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get entries newer than 0-0
				result := Xread("test-conn-1", []shared.Value{
					{Typ: "bulk", Bulk: "streams"},
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "0-0"},
				})
				if len(result.Array) != 1 {
					t.Errorf("Expected 1 stream result, got %d", len(result.Array))
				}
				if len(result.Array[0].Array[1].Array) != 2 {
					t.Errorf("Expected 2 entries, got %d", len(result.Array[0].Array[1].Array))
				}
			},
		},
		{
			name:   "xread single stream from specific ID",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "1-0"},
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
				// Verify we get entries newer than 1-0
				result := Xread("test-conn-2", []shared.Value{
					{Typ: "bulk", Bulk: "streams"},
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "1-0"},
				})
				if len(result.Array) != 1 {
					t.Errorf("Expected 1 stream result, got %d", len(result.Array))
				}
				if len(result.Array[0].Array[1].Array) != 2 {
					t.Errorf("Expected 2 entries, got %d", len(result.Array[0].Array[1].Array))
				}
			},
		},
		{
			name:   "xread multiple streams",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "stream1"},
				{Typ: "bulk", Bulk: "stream2"},
				{Typ: "bulk", Bulk: "0-0"},
				{Typ: "bulk", Bulk: "0-0"},
			},
			setup: func() {
				shared.Memory["stream1"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
					},
					Expires: 0,
				}
				shared.Memory["stream2"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "2-0", Data: map[string]string{"message": "World"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get results from both streams
				result := Xread("test-conn-3", []shared.Value{
					{Typ: "bulk", Bulk: "streams"},
					{Typ: "bulk", Bulk: "stream1"},
					{Typ: "bulk", Bulk: "stream2"},
					{Typ: "bulk", Bulk: "0-0"},
					{Typ: "bulk", Bulk: "0-0"},
				})
				if len(result.Array) != 2 {
					t.Errorf("Expected 2 stream results, got %d", len(result.Array))
				}
			},
		},
		{
			name:   "xread with $ (last entry)",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "$"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
						{ID: "2-0", Data: map[string]string{"message": "World"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify: func() {
				// Verify we get no entries (since $ means "after last entry")
				result := Xread("test-conn-4", []shared.Value{
					{Typ: "bulk", Bulk: "streams"},
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "$"},
				})
				if len(result.Array) != 0 {
					t.Errorf("Expected 0 entries with $, got %d", len(result.Array))
				}
			},
		},
		{
			name:   "xread non-existent stream",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "0-0"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "array", Array: []shared.Value{}},
			verify:   func() {},
		},
		{
			name:   "xread empty stream",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "emptystream"},
				{Typ: "bulk", Bulk: "0-0"},
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
			name:   "xread with blocking (no timeout)",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "BLOCK"},
				{Typ: "bulk", Bulk: "1000"},
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "0-0"},
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
				// Verify we get immediate results (no blocking needed)
				result := Xread("test-conn-7", []shared.Value{
					{Typ: "bulk", Bulk: "BLOCK"},
					{Typ: "bulk", Bulk: "1000"},
					{Typ: "bulk", Bulk: "streams"},
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "0-0"},
				})
				if len(result.Array) != 1 {
					t.Errorf("Expected 1 stream result, got %d", len(result.Array))
				}
			},
		},
		{
			name:   "xread with blocking and $",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "BLOCK"},
				{Typ: "bulk", Bulk: "1000"},
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "$"},
			},
			setup: func() {
				shared.Memory["mystream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "null_array"},
			verify: func() {
				// Verify we get null_array when blocking times out
				result := Xread("test-conn-8", []shared.Value{
					{Typ: "bulk", Bulk: "BLOCK"},
					{Typ: "bulk", Bulk: "1000"},
					{Typ: "bulk", Bulk: "streams"},
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "$"},
				})
				if result.Typ != "null_array" {
					t.Errorf("Expected null_array when blocking times out, got %s", result.Typ)
				}
			},
		},
		{
			name:   "wrong number of arguments",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "mystream"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'xread' command"},
			verify:   func() {},
		},
		{
			name:   "xread with unicode data",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "streams"},
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "0-0"},
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
				result := Xread("test-conn-10", []shared.Value{
					{Typ: "bulk", Bulk: "streams"},
					{Typ: "bulk", Bulk: "mystream"},
					{Typ: "bulk", Bulk: "0-0"},
				})
				if len(result.Array) != 1 {
					t.Errorf("Expected 1 stream result, got %d", len(result.Array))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Xread(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Xread() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Xread() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkXread(b *testing.B) {
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
		{Typ: "bulk", Bulk: "streams"},
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "0-0"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xread(connID, args)
	}
}

func BenchmarkXreadMultipleStreams(b *testing.B) {
	clearMemory()
	shared.Memory["stream1"] = shared.MemoryEntry{
		Stream: []shared.StreamEntry{
			{ID: "1-0", Data: map[string]string{"message": "Hello"}},
			{ID: "2-0", Data: map[string]string{"message": "World"}},
		},
		Expires: 0,
	}
	shared.Memory["stream2"] = shared.MemoryEntry{
		Stream: []shared.StreamEntry{
			{ID: "1-0", Data: map[string]string{"message": "Test"}},
			{ID: "2-0", Data: map[string]string{"message": "Extra"}},
		},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "streams"},
		{Typ: "bulk", Bulk: "stream1"},
		{Typ: "bulk", Bulk: "stream2"},
		{Typ: "bulk", Bulk: "0-0"},
		{Typ: "bulk", Bulk: "0-0"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xread(connID, args)
	}
}

func BenchmarkXreadWithDollar(b *testing.B) {
	clearMemory()
	shared.Memory["benchstream"] = shared.MemoryEntry{
		Stream: []shared.StreamEntry{
			{ID: "1-0", Data: map[string]string{"message": "Hello"}},
			{ID: "2-0", Data: map[string]string{"message": "World"}},
		},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "streams"},
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "$"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xread(connID, args)
	}
}

func BenchmarkXreadNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "streams"},
		{Typ: "bulk", Bulk: "nonexistent"},
		{Typ: "bulk", Bulk: "0-0"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xread(connID, args)
	}
}

func BenchmarkXreadWithBlocking(b *testing.B) {
	clearMemory()
	shared.Memory["benchstream"] = shared.MemoryEntry{
		Stream: []shared.StreamEntry{
			{ID: "1-0", Data: map[string]string{"message": "Hello"}},
		},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "BLOCK"},
		{Typ: "bulk", Bulk: "1000"},
		{Typ: "bulk", Bulk: "streams"},
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "0-0"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xread(connID, args)
	}
}
