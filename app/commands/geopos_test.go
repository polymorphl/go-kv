package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestGeopos(t *testing.T) {
	// Clear memory before each test
	clearMemory()

	tests := []struct {
		name     string
		connID   string
		args     []shared.Value
		expected shared.Value
		setup    func() // Function to set up test data
	}{
		{
			name:   "geopos single existing member",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "London"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{
						Typ: "array",
						Array: []shared.Value{
							{Typ: "bulk", Bulk: "-0.0884968042373657"},
							{Typ: "bulk", Bulk: "51.5064768740388"},
						},
					},
				},
			},
			setup: func() {
				// Add London with coordinates -0.0884948, 51.506479 -> score: 2163557758834106
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["places"].SortedSet.Add("London", 2163557758834106.0)
			},
		},
		{
			name:   "geopos multiple existing members",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "London"},
				{Typ: "bulk", Bulk: "Munich"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{
						Typ: "array",
						Array: []shared.Value{
							{Typ: "bulk", Bulk: "-0.0884968042373657"},
							{Typ: "bulk", Bulk: "51.5064768740388"},
						},
					},
					{
						Typ: "array",
						Array: []shared.Value{
							{Typ: "bulk", Bulk: "11.5030342340469"},
							{Typ: "bulk", Bulk: "48.1642695949692"},
						},
					},
				},
			},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["places"].SortedSet.Add("London", 2163557758834106.0)
				server.Memory["places"].SortedSet.Add("Munich", 3672376881541190.0)
			},
		},
		{
			name:   "geopos non-existent member",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "NonExistent"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "null_array", Str: ""},
				},
			},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["places"].SortedSet.Add("London", 2163557758834106.0)
			},
		},
		{
			name:   "geopos non-existent key",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "London"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "null_array", Str: ""},
				},
			},
			setup: func() {},
		},
		{
			name:   "geopos mixed existing and non-existing",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "London"},
				{Typ: "bulk", Bulk: "NonExistent"},
				{Typ: "bulk", Bulk: "Munich"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{
						Typ: "array",
						Array: []shared.Value{
							{Typ: "bulk", Bulk: "-0.0884968042373657"},
							{Typ: "bulk", Bulk: "51.5064768740388"},
						},
					},
					{Typ: "null_array", Str: ""},
					{
						Typ: "array",
						Array: []shared.Value{
							{Typ: "bulk", Bulk: "11.5030342340469"},
							{Typ: "bulk", Bulk: "48.1642695949692"},
						},
					},
				},
			},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["places"].SortedSet.Add("London", 2163557758834106.0)
				server.Memory["places"].SortedSet.Add("Munich", 3672376881541190.0)
			},
		},
		{
			name:   "geopos wrong number of arguments",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'geopos' command"},
			setup:    func() {},
		},
		{
			name:   "geopos key with wrong type",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "London"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "null_array", Str: ""},
				},
			},
			setup: func() {
				// Set up a key that doesn't hold a sorted set
				server.Memory["places"] = shared.MemoryEntry{
					Value:     "not a sorted set",
					SortedSet: nil,
					Expires:   0,
				}
			},
		},
		{
			name:   "geopos boundary coordinates",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "places"},
				{Typ: "bulk", Bulk: "Boundary"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{
						Typ: "array",
						Array: []shared.Value{
							{Typ: "bulk", Bulk: "180"},
							{Typ: "bulk", Bulk: "0"},
						},
					},
				},
			},
			setup: func() {
				server.Memory["places"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				// Boundary longitude positive: 180, 0 -> score: 10133099161583616
				server.Memory["places"].SortedSet.Add("Boundary", 10133099161583616.0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()
			tt.setup()

			result := Geopos(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Geopos() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Geopos() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			if result.Typ == "array" {
				if len(result.Array) != len(tt.expected.Array) {
					t.Errorf("Geopos() array length = %v, expected %v", len(result.Array), len(tt.expected.Array))
				}

				for i, item := range result.Array {
					if i >= len(tt.expected.Array) {
						break
					}
					expectedItem := tt.expected.Array[i]

					if item.Typ != expectedItem.Typ {
						t.Errorf("Geopos() array[%d] type = %v, expected %v", i, item.Typ, expectedItem.Typ)
					}

					if item.Typ == "array" && expectedItem.Typ == "array" {
						if len(item.Array) != len(expectedItem.Array) {
							t.Errorf("Geopos() array[%d] length = %v, expected %v", i, len(item.Array), len(expectedItem.Array))
						}
						for j, coord := range item.Array {
							if j >= len(expectedItem.Array) {
								break
							}
							expectedCoord := expectedItem.Array[j]
							if coord.Typ != expectedCoord.Typ {
								t.Errorf("Geopos() array[%d][%d] type = %v, expected %v", i, j, coord.Typ, expectedCoord.Typ)
							}
							if coord.Bulk != expectedCoord.Bulk {
								t.Errorf("Geopos() array[%d][%d] bulk = %v, expected %v", i, j, coord.Bulk, expectedCoord.Bulk)
							}
						}
					}
				}
			}
		})
	}
}

func BenchmarkGeopos(b *testing.B) {
	clearMemory()

	// Set up test data
	server.Memory["benchkey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}
	server.Memory["benchkey"].SortedSet.Add("London", 2163557758834106.0)
	server.Memory["benchkey"].SortedSet.Add("Munich", 3672376881541190.0)

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "London"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Geopos(connID, args)
	}
}

func BenchmarkGeoposMultiple(b *testing.B) {
	clearMemory()

	// Set up test data
	server.Memory["benchkey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}
	server.Memory["benchkey"].SortedSet.Add("London", 2163557758834106.0)
	server.Memory["benchkey"].SortedSet.Add("Munich", 3672376881541190.0)

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "London"},
		{Typ: "bulk", Bulk: "Munich"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Geopos(connID, args)
	}
}
