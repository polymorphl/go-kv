package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestGeosearch(t *testing.T) {
	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
		setup    func()
	}{
		{
			name:   "geosearch Paris within 100m",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "2.2944692"},  // Paris longitude
				{Typ: "bulk", Bulk: "48.8584625"}, // Paris latitude
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "100"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "Paris"},
				},
			},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				// Paris: 2.2944692, 48.8584625 -> score: 3663832614298053
				server.Memory["places"].SortedSet.Add("Paris", 3663832614298053.0)
				// Munich: 11.5030378, 48.164271 -> score: 3672376881541190
				server.Memory["places"].SortedSet.Add("Munich", 3672376881541190.0)
			},
		},
		{
			name:   "geosearch Munich within 1000km",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "11.5030378"}, // Munich longitude
				{Typ: "bulk", Bulk: "48.164271"},  // Munich latitude
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "1000"},
				{Typ: "bulk", Bulk: "km"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "bulk", Bulk: "Munich"},
					{Typ: "bulk", Bulk: "Paris"},
				},
			},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				// Paris: 2.2944692, 48.8584625 -> score: 3663832614298053
				server.Memory["places"].SortedSet.Add("Paris", 3663832614298053.0)
				// Munich: 11.5030378, 48.164271 -> score: 3672376881541190
				server.Memory["places"].SortedSet.Add("Munich", 3672376881541190.0)
			},
		},
		{
			name:   "geosearch no results within small radius",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "0.0"}, // Center of Earth
				{Typ: "bulk", Bulk: "0.0"},
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "1"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ:   "array",
				Array: []shared.Value{},
			},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				// Paris: 2.2944692, 48.8584625 -> score: 3663832614298053
				server.Memory["places"].SortedSet.Add("Paris", 3663832614298053.0)
				// Munich: 11.5030378, 48.164271 -> score: 3672376881541190
				server.Memory["places"].SortedSet.Add("Munich", 3672376881541190.0)
			},
		},
		{
			name:   "geosearch non-existent key",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "2.2944692"},
				{Typ: "bulk", Bulk: "48.8584625"},
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "100"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ:   "array",
				Array: []shared.Value{},
			},
			setup: func() {
				// No setup - key doesn't exist
			},
		},
		{
			name:   "geosearch wrong number of arguments",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "2.2944692"},
			},
			expected: shared.Value{
				Typ: "error",
				Str: "ERR wrong number of arguments for 'geosearch' command",
			},
			setup: func() {
				// No setup needed for error case
			},
		},
		{
			name:   "geosearch invalid longitude",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "invalid"},
				{Typ: "bulk", Bulk: "48.8584625"},
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "100"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ: "error",
				Str: "ERR invalid longitude",
			},
			setup: func() {
				// No setup needed for error case
			},
		},
		{
			name:   "geosearch invalid latitude",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "2.2944692"},
				{Typ: "bulk", Bulk: "invalid"},
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "100"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ: "error",
				Str: "ERR invalid latitude",
			},
			setup: func() {
				// No setup needed for error case
			},
		},
		{
			name:   "geosearch invalid radius",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "2.2944692"},
				{Typ: "bulk", Bulk: "48.8584625"},
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "invalid"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ: "error",
				Str: "ERR invalid radius",
			},
			setup: func() {
				// No setup needed for error case
			},
		},
		{
			name:   "geosearch unsupported mode",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMMEMBER"},
				{Typ: "bulk", Bulk: "Paris"},
				{Typ: "bulk", Bulk: "BYRADIUS"},
				{Typ: "bulk", Bulk: "100"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ: "error",
				Str: "ERR only FROMLONLAT mode is supported",
			},
			setup: func() {
				// No setup needed for error case
			},
		},
		{
			name:   "geosearch unsupported search type",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "FROMLONLAT"},
				{Typ: "bulk", Bulk: "2.2944692"},
				{Typ: "bulk", Bulk: "48.8584625"},
				{Typ: "bulk", Bulk: "BYBOX"},
				{Typ: "bulk", Bulk: "100"},
				{Typ: "bulk", Bulk: "m"},
			},
			expected: shared.Value{
				Typ: "error",
				Str: "ERR only BYRADIUS mode is supported",
			},
			setup: func() {
				// No setup needed for error case
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear memory before each test
			server.Memory = make(map[string]shared.MemoryEntry)

			// Setup test data
			if tt.setup != nil {
				tt.setup()
			}

			result := Geosearch(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Geosearch() typ = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "error" {
				if result.Str != tt.expected.Str {
					t.Errorf("Geosearch() error = %v, expected %v", result.Str, tt.expected.Str)
				}
			} else if result.Typ == "array" {
				if len(result.Array) != len(tt.expected.Array) {
					t.Errorf("Geosearch() array length = %v, expected %v", len(result.Array), len(tt.expected.Array))
				} else {
					for i, item := range result.Array {
						if item.Typ != tt.expected.Array[i].Typ || item.Bulk != tt.expected.Array[i].Bulk {
							t.Errorf("Geosearch() array[%d] = %v, expected %v", i, item, tt.expected.Array[i])
						}
					}
				}
			}
		})
	}
}
