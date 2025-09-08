package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestXadd(t *testing.T) {
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
			name:   "xadd to new stream with explicit ID",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "1-0"},
				{Typ: "bulk", Bulk: "message"},
				{Typ: "bulk", Bulk: "Hello"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "bulk", Bulk: "1-0"},
			verify: func() {
				entry, exists := shared.Memory["mystream"]
				if !exists {
					t.Error("Stream should exist after XADD")
				}
				if len(entry.Stream) != 1 {
					t.Errorf("Expected 1 stream entry, got %d", len(entry.Stream))
				}
				if entry.Stream[0].ID != "1-0" {
					t.Errorf("Expected ID '1-0', got '%s'", entry.Stream[0].ID)
				}
				if entry.Stream[0].Data["message"] != "Hello" {
					t.Errorf("Expected message 'Hello', got '%s'", entry.Stream[0].Data["message"])
				}
			},
		},
		{
			name:   "xadd with auto-generated timestamp",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "*"},
				{Typ: "bulk", Bulk: "temperature"},
				{Typ: "bulk", Bulk: "25"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "bulk", Bulk: "*"}, // Will be replaced with actual ID
			verify: func() {
				entry, exists := shared.Memory["mystream"]
				if !exists {
					t.Error("Stream should exist after XADD")
				}
				if len(entry.Stream) != 1 {
					t.Errorf("Expected 1 stream entry, got %d", len(entry.Stream))
				}
				// Check that ID is in timestamp-sequence format
				id := entry.Stream[0].ID
				if len(id) < 3 || id[len(id)-2:] != "-0" {
					t.Errorf("Expected ID in format 'timestamp-0', got '%s'", id)
				}
				if entry.Stream[0].Data["temperature"] != "25" {
					t.Errorf("Expected temperature '25', got '%s'", entry.Stream[0].Data["temperature"])
				}
			},
		},
		{
			name:   "xadd with multiple field-value pairs",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "2-0"},
				{Typ: "bulk", Bulk: "temperature"},
				{Typ: "bulk", Bulk: "30"},
				{Typ: "bulk", Bulk: "humidity"},
				{Typ: "bulk", Bulk: "60"},
				{Typ: "bulk", Bulk: "pressure"},
				{Typ: "bulk", Bulk: "1013"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "bulk", Bulk: "2-0"},
			verify: func() {
				entry, exists := shared.Memory["mystream"]
				if !exists {
					t.Error("Stream should exist after XADD")
				}
				if len(entry.Stream) != 1 {
					t.Errorf("Expected 1 stream entry, got %d", len(entry.Stream))
				}
				data := entry.Stream[0].Data
				if data["temperature"] != "30" {
					t.Errorf("Expected temperature '30', got '%s'", data["temperature"])
				}
				if data["humidity"] != "60" {
					t.Errorf("Expected humidity '60', got '%s'", data["humidity"])
				}
				if data["pressure"] != "1013" {
					t.Errorf("Expected pressure '1013', got '%s'", data["pressure"])
				}
			},
		},
		{
			name:   "xadd to existing stream",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "existingstream"},
				{Typ: "bulk", Bulk: "3-0"},
				{Typ: "bulk", Bulk: "message"},
				{Typ: "bulk", Bulk: "World"},
			},
			setup: func() {
				shared.Memory["existingstream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1-0", Data: map[string]string{"message": "Hello"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "bulk", Bulk: "3-0"},
			verify: func() {
				entry, exists := shared.Memory["existingstream"]
				if !exists {
					t.Error("Stream should exist after XADD")
				}
				if len(entry.Stream) != 2 {
					t.Errorf("Expected 2 stream entries, got %d", len(entry.Stream))
				}
				if entry.Stream[1].ID != "3-0" {
					t.Errorf("Expected ID '3-0', got '%s'", entry.Stream[1].ID)
				}
				if entry.Stream[1].Data["message"] != "World" {
					t.Errorf("Expected message 'World', got '%s'", entry.Stream[1].Data["message"])
				}
			},
		},
		{
			name:   "xadd with invalid ID (0-0)",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "0-0"},
				{Typ: "bulk", Bulk: "message"},
				{Typ: "bulk", Bulk: "Hello"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR The ID specified in XADD must be greater than 0-0"},
			verify: func() {
				// Stream should not exist after error
				if _, exists := shared.Memory["mystream"]; exists {
					t.Error("Stream should not exist after error")
				}
			},
		},
		{
			name:   "xadd with ID smaller than last entry",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "existingstream"},
				{Typ: "bulk", Bulk: "1-0"},
				{Typ: "bulk", Bulk: "message"},
				{Typ: "bulk", Bulk: "Old"},
			},
			setup: func() {
				shared.Memory["existingstream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "2-0", Data: map[string]string{"message": "New"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "error", Str: "ERR The ID specified in XADD is equal or smaller than the target stream top item"},
			verify: func() {
				// Stream should remain unchanged
				entry, exists := shared.Memory["existingstream"]
				if !exists {
					t.Error("Stream should still exist after error")
				}
				if len(entry.Stream) != 1 {
					t.Error("Stream should remain unchanged")
				}
			},
		},
		{
			name:   "xadd with auto-generated sequence",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "1000-*"},
				{Typ: "bulk", Bulk: "message"},
				{Typ: "bulk", Bulk: "Hello"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "bulk", Bulk: "1000-0"},
			verify: func() {
				entry, exists := shared.Memory["mystream"]
				if !exists {
					t.Error("Stream should exist after XADD")
				}
				if len(entry.Stream) != 1 {
					t.Errorf("Expected 1 stream entry, got %d", len(entry.Stream))
				}
				if entry.Stream[0].ID != "1000-0" {
					t.Errorf("Expected ID '1000-0', got '%s'", entry.Stream[0].ID)
				}
			},
		},
		{
			name:   "xadd with same timestamp increments sequence",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "existingstream"},
				{Typ: "bulk", Bulk: "1000-*"},
				{Typ: "bulk", Bulk: "message"},
				{Typ: "bulk", Bulk: "Second"},
			},
			setup: func() {
				shared.Memory["existingstream"] = shared.MemoryEntry{
					Stream: []shared.StreamEntry{
						{ID: "1000-0", Data: map[string]string{"message": "First"}},
					},
					Expires: 0,
				}
			},
			expected: shared.Value{Typ: "bulk", Bulk: "1000-1"},
			verify: func() {
				entry, exists := shared.Memory["existingstream"]
				if !exists {
					t.Error("Stream should exist after XADD")
				}
				if len(entry.Stream) != 2 {
					t.Errorf("Expected 2 stream entries, got %d", len(entry.Stream))
				}
				if entry.Stream[1].ID != "1000-1" {
					t.Errorf("Expected ID '1000-1', got '%s'", entry.Stream[1].ID)
				}
			},
		},
		{
			name:   "wrong number of arguments",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "1-0"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'xadd' command"},
			verify:   func() {},
		},
		{
			name:   "xadd with unicode field-value pairs",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "mystream"},
				{Typ: "bulk", Bulk: "4-0"},
				{Typ: "bulk", Bulk: "消息"},
				{Typ: "bulk", Bulk: "你好世界 🌍"},
			},
			setup:    func() {},
			expected: shared.Value{Typ: "bulk", Bulk: "4-0"},
			verify: func() {
				entry, exists := shared.Memory["mystream"]
				if !exists {
					t.Error("Stream should exist after XADD")
				}
				if len(entry.Stream) != 1 {
					t.Errorf("Expected 1 stream entry, got %d", len(entry.Stream))
				}
				if entry.Stream[0].Data["消息"] != "你好世界 🌍" {
					t.Errorf("Expected unicode value, got '%s'", entry.Stream[0].Data["消息"])
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Xadd(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Xadd() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Str != tt.expected.Str {
				t.Errorf("Xadd() string = %v, expected %v", result.Str, tt.expected.Str)
			}

			// For auto-generated IDs, just check format
			if tt.expected.Bulk == "*" && result.Typ == "bulk" {
				// Verify it's a valid timestamp-sequence format
				if len(result.Bulk) < 3 || result.Bulk[len(result.Bulk)-2:] != "-0" {
					t.Errorf("Expected auto-generated ID in format 'timestamp-0', got '%s'", result.Bulk)
				}
			} else if result.Bulk != tt.expected.Bulk {
				t.Errorf("Xadd() bulk = %v, expected %v", result.Bulk, tt.expected.Bulk)
			}

			tt.verify()
		})
	}
}

func BenchmarkXadd(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "*"},
		{Typ: "bulk", Bulk: "message"},
		{Typ: "bulk", Bulk: "Hello World"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xadd(connID, args)
	}
}

func BenchmarkXaddExplicitID(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "1-0"},
		{Typ: "bulk", Bulk: "message"},
		{Typ: "bulk", Bulk: "Hello World"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xadd(connID, args)
	}
}

func BenchmarkXaddMultipleFields(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "*"},
		{Typ: "bulk", Bulk: "temperature"},
		{Typ: "bulk", Bulk: "25"},
		{Typ: "bulk", Bulk: "humidity"},
		{Typ: "bulk", Bulk: "60"},
		{Typ: "bulk", Bulk: "pressure"},
		{Typ: "bulk", Bulk: "1013"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xadd(connID, args)
	}
}

func BenchmarkXaddToExistingStream(b *testing.B) {
	clearMemory()
	shared.Memory["benchstream"] = shared.MemoryEntry{
		Stream: []shared.StreamEntry{
			{ID: "1-0", Data: map[string]string{"message": "existing"}},
		},
		Expires: 0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchstream"},
		{Typ: "bulk", Bulk: "*"},
		{Typ: "bulk", Bulk: "message"},
		{Typ: "bulk", Bulk: "new"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Xadd(connID, args)
	}
}
