package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestGeodist(t *testing.T) {
	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
		setup    func()
	}{
		{
			name:   "geodist Munich to Paris",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "Munich"},
				{Typ: "bulk", Bulk: "Paris"},
			},
			expected: shared.Value{Typ: "bulk", Bulk: "682477.7582"},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				// Munich: 11.5030378, 48.164271 -> score: 3672376881541190
				// Paris: 2.2944692, 48.8584625 -> score: 3663832614298053
				server.Memory["places"].SortedSet.Add("Munich", 3672376881541190.0)
				server.Memory["places"].SortedSet.Add("Paris", 3663832614298053.0)
			},
		},
		{
			name:   "geodist non-existent member",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "NonExistent"},
				{Typ: "bulk", Bulk: "Paris"},
			},
			expected: shared.Value{Typ: "null_bulk", Str: ""},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["places"].SortedSet.Add("Paris", 3663832614298053.0)
			},
		},
		{
			name:   "geodist non-existent key",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "Munich"},
				{Typ: "bulk", Bulk: "Paris"},
			},
			expected: shared.Value{Typ: "null_bulk", Str: ""},
			setup:    func() {},
		},
		{
			name:   "geodist wrong number of arguments",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "Munich"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'geodist' command"},
			setup:    func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear memory
			server.Memory = make(map[string]shared.MemoryEntry)

			// Setup test data
			tt.setup()

			// Execute command
			result := Geodist(tt.connID, tt.args)

			// Check result type
			if result.Typ != tt.expected.Typ {
				t.Errorf("Geodist() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			// Check result content
			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Geodist() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			if result.Typ == "bulk" && result.Bulk != tt.expected.Bulk {
				t.Errorf("Geodist() bulk = %v, expected %v", result.Bulk, tt.expected.Bulk)
			}

			if result.Typ == "null_bulk" && result.Str != tt.expected.Str {
				t.Errorf("Geodist() null_bulk = %v, expected %v", result.Str, tt.expected.Str)
			}
		})
	}
}
