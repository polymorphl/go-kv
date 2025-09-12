package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestZrange(t *testing.T) {
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
			name:   "zrange all elements",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "member1"},
					{Typ: "string", Str: "member2"},
					{Typ: "string", Str: "member3"},
				},
			},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				if entry.SortedSet.Size != 3 {
					t.Errorf("Expected size 3, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zrange first element",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "0"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "member1"},
				},
			},
			verify: func() {},
		},
		{
			name:   "zrange middle elements",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "1"},
				{Typ: "bulk", Bulk: "2"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "member2"},
					{Typ: "string", Str: "member3"},
				},
			},
			verify: func() {},
		},
		{
			name:   "zrange last element",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "-1"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "member3"},
				},
			},
			verify: func() {},
		},
		{
			name:   "zrange last two elements",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "-2"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "member2"},
					{Typ: "string", Str: "member3"},
				},
			},
			verify: func() {},
		},
		{
			name:   "zrange with same score (alphabetical order)",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "same_score"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "grape"},
					{Typ: "string", Str: "pineapple"},
				},
			},
			verify: func() {},
		},
		{
			name:   "zrange empty set",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "empty"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ:   "array",
				Array: []shared.Value{},
			},
			verify: func() {},
		},
		{
			name:   "zrange non-existent key",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ:   "array",
				Array: []shared.Value{},
			},
			verify: func() {},
		},
		{
			name:   "zrange key with wrong type",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "wrongtype"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ:   "array",
				Array: []shared.Value{},
			},
			verify: func() {},
		},
		{
			name:   "zrange start > stop",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "2"},
				{Typ: "bulk", Bulk: "1"},
			},
			expected: shared.Value{
				Typ:   "array",
				Array: []shared.Value{},
			},
			verify: func() {},
		},
		{
			name:   "zrange start out of bounds",
			connID: "test-conn-11",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "10"},
				{Typ: "bulk", Bulk: "20"},
			},
			expected: shared.Value{
				Typ:   "array",
				Array: []shared.Value{},
			},
			verify: func() {},
		},
		{
			name:   "zrange stop out of bounds",
			connID: "test-conn-12",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "10"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "member1"},
					{Typ: "string", Str: "member2"},
					{Typ: "string", Str: "member3"},
				},
			},
			verify: func() {},
		},
		{
			name:   "zrange unicode members",
			connID: "test-conn-13",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicode"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "-1"},
			},
			expected: shared.Value{
				Typ: "array",
				Array: []shared.Value{
					{Typ: "string", Str: "成员1"},
					{Typ: "string", Str: "成员2"},
				},
			},
			verify: func() {},
		},
		{
			name:   "zrange wrong number of arguments",
			connID: "test-conn-14",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "0"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zrange' command"},
			verify:   func() {},
		},
		{
			name:   "zrange invalid start",
			connID: "test-conn-15",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "notanumber"},
				{Typ: "bulk", Bulk: "0"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not an integer or out of range"},
			verify:   func() {},
		},
		{
			name:   "zrange invalid stop",
			connID: "test-conn-16",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "0"},
				{Typ: "bulk", Bulk: "notanumber"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not an integer or out of range"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up test data
			switch tt.name {
			case "zrange all elements", "zrange first element", "zrange middle elements", "zrange last element", "zrange last two elements", "zrange start > stop", "zrange start out of bounds", "zrange stop out of bounds":
				server.Memory["myzset"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["myzset"].SortedSet.Add("member1", 1.0)
				server.Memory["myzset"].SortedSet.Add("member2", 2.0)
				server.Memory["myzset"].SortedSet.Add("member3", 3.0)
			case "zrange with same score (alphabetical order)":
				server.Memory["same_score"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["same_score"].SortedSet.Add("grape", 1.0)
				server.Memory["same_score"].SortedSet.Add("pineapple", 1.0)
			case "zrange empty set":
				server.Memory["empty"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
			case "zrange key with wrong type":
				server.Memory["wrongtype"] = shared.MemoryEntry{
					Value:   "string value",
					Expires: 0,
				}
			case "zrange unicode members":
				server.Memory["unicode"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["unicode"].SortedSet.Add("成员1", 1.0)
				server.Memory["unicode"].SortedSet.Add("成员2", 2.0)
			case "zrange invalid start", "zrange invalid stop":
				server.Memory["key"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["key"].SortedSet.Add("member1", 1.0)
			}

			result := Zrange(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Zrange() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "array" {
				if len(result.Array) != len(tt.expected.Array) {
					t.Errorf("Zrange() array length = %v, expected %v", len(result.Array), len(tt.expected.Array))
				}
				for i, val := range result.Array {
					if val.Str != tt.expected.Array[i].Str {
						t.Errorf("Zrange() array[%d] = %v, expected %v", i, val.Str, tt.expected.Array[i].Str)
					}
				}
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Zrange() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkZrange(b *testing.B) {
	clearMemory()

	// Set up test data
	server.Memory["benchkey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}
	server.Memory["benchkey"].SortedSet.Add("member1", 1.0)
	server.Memory["benchkey"].SortedSet.Add("member2", 2.0)
	server.Memory["benchkey"].SortedSet.Add("member3", 3.0)

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "0"},
		{Typ: "bulk", Bulk: "-1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zrange(connID, args)
	}
}

func BenchmarkZrangeLargeSet(b *testing.B) {
	clearMemory()

	// Set up large test data
	server.Memory["largekey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}
	for i := 0; i < 1000; i++ {
		server.Memory["largekey"].SortedSet.Add(fmt.Sprintf("member%d", i), float64(i))
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "largekey"},
		{Typ: "bulk", Bulk: "0"},
		{Typ: "bulk", Bulk: "99"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zrange(connID, args)
	}
}

func BenchmarkZrangeSingleElement(b *testing.B) {
	clearMemory()

	// Set up test data
	server.Memory["benchkey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}
	server.Memory["benchkey"].SortedSet.Add("member1", 1.0)
	server.Memory["benchkey"].SortedSet.Add("member2", 2.0)
	server.Memory["benchkey"].SortedSet.Add("member3", 3.0)

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "1"},
		{Typ: "bulk", Bulk: "1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zrange(connID, args)
	}
}
