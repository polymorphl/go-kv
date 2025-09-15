package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestGeoadd(t *testing.T) {
	// Clear memory before each test
	clearMemory()

	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
		verify   func() // Function to verify the result
	}{
		{
			name:   "geoadd valid coordinates",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "-122.4194"},
				{Typ: "bulk", Bulk: "37.7749"},
				{Typ: "bulk", Bulk: "San Francisco"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["places"]
				if !exists {
					t.Error("Key should exist after GEOADD")
				}
				if entry.SortedSet == nil {
					t.Error("SortedSet should be created")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
				score, exists := entry.SortedSet.GetScore("San Francisco")
				if !exists {
					t.Error("Member should exist")
				}
				if score != -122.4194 {
					t.Errorf("Expected score -122.4194, got %f", score)
				}
			},
		},
		{
			name:   "geoadd multiple coordinates",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "-74.0059"},
				{Typ: "bulk", Bulk: "40.7128"},
				{Typ: "bulk", Bulk: "New York"},
				{Typ: "bulk", Bulk: "-0.1276"},
				{Typ: "bulk", Bulk: "51.5074"},
				{Typ: "bulk", Bulk: "London"},
			},
			expected: shared.Value{Typ: "integer", Num: 2},
			verify: func() {
				entry, exists := server.Memory["places"]
				if !exists {
					t.Error("Key should exist after GEOADD")
				}
				if entry.SortedSet.Size != 2 {
					t.Errorf("Expected size 2, got %d", entry.SortedSet.Size)
				}
				// Verify both members exist
				if _, exists := entry.SortedSet.GetScore("New York"); !exists {
					t.Error("New York should exist")
				}
				if _, exists := entry.SortedSet.GetScore("London"); !exists {
					t.Error("London should exist")
				}
			},
		},
		{
			name:   "geoadd boundary longitude positive",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "180"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "Boundary"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["places"]
				if !exists {
					t.Error("Key should exist after GEOADD")
				}
				score, exists := entry.SortedSet.GetScore("Boundary")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 180.0 {
					t.Errorf("Expected score 180.0, got %f", score)
				}
			},
		},
		{
			name:   "geoadd boundary longitude negative",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "-180"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "Negative Boundary"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["places"]
				if !exists {
					t.Error("Key should exist after GEOADD")
				}
				score, exists := entry.SortedSet.GetScore("Negative Boundary")
				if !exists {
					t.Error("Member should exist")
				}
				if score != -180.0 {
					t.Errorf("Expected score -180.0, got %f", score)
				}
			},
		},
		{
			name:   "geoadd boundary latitude positive",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "85.05112878"},
				{Typ: "bulk", Bulk: "Max Lat"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["places"]
				if !exists {
					t.Error("Key should exist after GEOADD")
				}
				score, exists := entry.SortedSet.GetScore("Max Lat")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 0.0 {
					t.Errorf("Expected score 0.0, got %f", score)
				}
			},
		},
		{
			name:   "geoadd boundary latitude negative",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-85.05112878"},
				{Typ: "bulk", Bulk: "Min Lat"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["places"]
				if !exists {
					t.Error("Key should exist after GEOADD")
				}
				score, exists := entry.SortedSet.GetScore("Min Lat")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 0.0 {
					t.Errorf("Expected score 0.0, got %f", score)
				}
			},
		},
		{
			name:   "geoadd invalid longitude too high",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "181"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "Invalid"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR invalid longitude value"},
			verify:   func() {},
		},
		{
			name:   "geoadd invalid longitude too low",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "-181"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "Invalid"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR invalid longitude value"},
			verify:   func() {},
		},
		{
			name:   "geoadd invalid latitude too high",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "90"},
				{Typ: "bulk", Bulk: "Invalid"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR invalid latitude value"},
			verify:   func() {},
		},
		{
			name:   "geoadd invalid latitude too low",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-90"},
				{Typ: "bulk", Bulk: "Invalid"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR invalid latitude value"},
			verify:   func() {},
		},
		{
			name:   "geoadd invalid longitude format",
			connID: "test-conn-11",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "invalid"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "Invalid"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR invalid longitude argument"},
			verify:   func() {},
		},
		{
			name:   "geoadd invalid latitude format",
			connID: "test-conn-12",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "invalid"},
				{Typ: "bulk", Bulk: "Invalid"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR invalid latitude argument"},
			verify:   func() {},
		},
		{
			name:   "geoadd wrong number of arguments",
			connID: "test-conn-13",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "0"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'geoadd' command"},
			verify:   func() {},
		},
		{
			name:   "geoadd insufficient arguments",
			connID: "test-conn-14",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'geoadd' command"},
			verify:   func() {},
		},
		{
			name:   "geoadd update existing member",
			connID: "test-conn-15",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "existing"},
			},
			expected: shared.Value{Typ: "integer", Num: 0}, // 0 because it's an update, not new
			verify: func() {
				entry, exists := server.Memory["places"]
				if !exists {
					t.Error("Key should exist after GEOADD")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
				score, exists := entry.SortedSet.GetScore("existing")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 0.0 {
					t.Errorf("Expected score 0.0, got %f", score)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up initial data for update test
			if tt.name == "geoadd update existing member" {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["places"].SortedSet.Add("existing", 1.0)
			}

			result := Geoadd(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Geoadd() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "integer" && result.Num != tt.expected.Num {
				t.Errorf("Geoadd() num = %v, expected %v", result.Num, tt.expected.Num)
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Geoadd() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkGeoadd(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "-122.4194"},
		{Typ: "bulk", Bulk: "37.7749"},
		{Typ: "bulk", Bulk: "San Francisco"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Geoadd(connID, args)
	}
}

func BenchmarkGeoaddMultiple(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "-122.4194"},
		{Typ: "bulk", Bulk: "37.7749"},
		{Typ: "bulk", Bulk: "San Francisco"},
		{Typ: "bulk", Bulk: "-74.0059"},
		{Typ: "bulk", Bulk: "40.7128"},
		{Typ: "bulk", Bulk: "New York"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Geoadd(connID, args)
	}
}

func BenchmarkGeoaddBoundary(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "180"},
		{Typ: "bulk", Bulk: "85.05112878"},
		{Typ: "bulk", Bulk: "Boundary"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Geoadd(connID, args)
	}
}
