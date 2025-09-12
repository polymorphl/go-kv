package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestZrank(t *testing.T) {
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
			name:   "zrank existing member",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "member1"},
			},
			expected: shared.Value{Typ: "integer", Num: 0}, // First member has rank 0
			verify: func() {
				// Verify the member exists and has correct rank
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				rank, exists := entry.SortedSet.GetRank("member1")
				if !exists {
					t.Error("Member should exist")
				}
				if rank != 0 {
					t.Errorf("Expected rank 0, got %d", rank)
				}
			},
		},
		{
			name:   "zrank member with higher score",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "member3"},
			},
			expected: shared.Value{Typ: "integer", Num: 2}, // Third member has rank 2
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				rank, exists := entry.SortedSet.GetRank("member3")
				if !exists {
					t.Error("Member should exist")
				}
				if rank != 2 {
					t.Errorf("Expected rank 2, got %d", rank)
				}
			},
		},
		{
			name:   "zrank member with same score (alphabetical order)",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "grape"},
			},
			expected: shared.Value{Typ: "integer", Num: 0}, // grape comes before pineapple alphabetically
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				rank, exists := entry.SortedSet.GetRank("grape")
				if !exists {
					t.Error("Member should exist")
				}
				if rank != 0 {
					t.Errorf("Expected rank 0, got %d", rank)
				}
			},
		},
		{
			name:   "zrank non-existent member",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "zrank non-existent key",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "zrank key with wrong type",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "wrongtype"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "zrank unicode member",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicode"},
				{Typ: "bulk", Bulk: "成员1"},
			},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify: func() {
				entry, exists := server.Memory["unicode"]
				if !exists {
					t.Error("Key should exist")
				}
				rank, exists := entry.SortedSet.GetRank("成员1")
				if !exists {
					t.Error("Unicode member should exist")
				}
				if rank != 0 {
					t.Errorf("Expected rank 0, got %d", rank)
				}
			},
		},
		{
			name:   "zrank wrong number of arguments",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zrank' command"},
			verify:   func() {},
		},
		{
			name:   "zrank too many arguments",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "member1"},
				{Typ: "bulk", Bulk: "member2"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zrank' command"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up test data
			switch tt.name {
			case "zrank existing member", "zrank member with higher score":
				server.Memory["myzset"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["myzset"].SortedSet.Add("member1", 1.0)
				server.Memory["myzset"].SortedSet.Add("member2", 2.0)
				server.Memory["myzset"].SortedSet.Add("member3", 3.0)
			case "zrank member with same score (alphabetical order)":
				server.Memory["myzset"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["myzset"].SortedSet.Add("grape", 1.0)
				server.Memory["myzset"].SortedSet.Add("pineapple", 1.0)
			case "zrank unicode member":
				server.Memory["unicode"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["unicode"].SortedSet.Add("成员1", 1.0)
				server.Memory["unicode"].SortedSet.Add("成员2", 2.0)
			case "zrank key with wrong type":
				server.Memory["wrongtype"] = shared.MemoryEntry{
					Value:   "string value",
					Expires: 0,
				}
			}

			result := Zrank(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Zrank() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "integer" && result.Num != tt.expected.Num {
				t.Errorf("Zrank() num = %v, expected %v", result.Num, tt.expected.Num)
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Zrank() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkZrank(b *testing.B) {
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
		{Typ: "bulk", Bulk: "member2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zrank(connID, args)
	}
}

func BenchmarkZrankLargeSet(b *testing.B) {
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
		{Typ: "bulk", Bulk: "member500"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zrank(connID, args)
	}
}

func BenchmarkZrankNonExistent(b *testing.B) {
	clearMemory()

	// Set up test data
	server.Memory["benchkey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}
	server.Memory["benchkey"].SortedSet.Add("member1", 1.0)

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "nonexistent"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zrank(connID, args)
	}
}
