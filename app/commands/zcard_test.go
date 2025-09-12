package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestZcard(t *testing.T) {
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
			name:   "zcard non-empty set",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
			},
			expected: shared.Value{Typ: "integer", Num: 3},
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
			name:   "zcard empty set",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "empty"},
			},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify: func() {
				entry, exists := server.Memory["empty"]
				if !exists {
					t.Error("Key should exist")
				}
				if entry.SortedSet.Size != 0 {
					t.Errorf("Expected size 0, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zcard single member",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "single"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["single"]
				if !exists {
					t.Error("Key should exist")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zcard large set",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "large"},
			},
			expected: shared.Value{Typ: "integer", Num: 100},
			verify: func() {
				entry, exists := server.Memory["large"]
				if !exists {
					t.Error("Key should exist")
				}
				if entry.SortedSet.Size != 100 {
					t.Errorf("Expected size 100, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zcard non-existent key",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify:   func() {},
		},
		{
			name:   "zcard key with wrong type",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "wrongtype"},
			},
			expected: shared.Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"},
			verify:   func() {},
		},
		{
			name:   "zcard unicode members",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicode"},
			},
			expected: shared.Value{Typ: "integer", Num: 2},
			verify: func() {
				entry, exists := server.Memory["unicode"]
				if !exists {
					t.Error("Key should exist")
				}
				if entry.SortedSet.Size != 2 {
					t.Errorf("Expected size 2, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zcard after adding members",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "dynamic"},
			},
			expected: shared.Value{Typ: "integer", Num: 2},
			verify: func() {
				entry, exists := server.Memory["dynamic"]
				if !exists {
					t.Error("Key should exist")
				}
				if entry.SortedSet.Size != 2 {
					t.Errorf("Expected size 2, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zcard after removing members",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "reduced"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["reduced"]
				if !exists {
					t.Error("Key should exist")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zcard wrong number of arguments",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "extra"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zcard' command"},
			verify:   func() {},
		},
		{
			name:     "zcard no arguments",
			connID:   "test-conn-11",
			args:     []shared.Value{},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zcard' command"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up test data
			switch tt.name {
			case "zcard non-empty set":
				server.Memory["myzset"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["myzset"].SortedSet.Add("member1", 1.0)
				server.Memory["myzset"].SortedSet.Add("member2", 2.0)
				server.Memory["myzset"].SortedSet.Add("member3", 3.0)
			case "zcard empty set":
				server.Memory["empty"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
			case "zcard single member":
				server.Memory["single"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["single"].SortedSet.Add("member1", 1.0)
			case "zcard large set":
				server.Memory["large"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				for i := 0; i < 100; i++ {
					server.Memory["large"].SortedSet.Add(fmt.Sprintf("member%d", i), float64(i))
				}
			case "zcard key with wrong type":
				server.Memory["wrongtype"] = shared.MemoryEntry{
					Value:   "string value",
					Expires: 0,
				}
			case "zcard unicode members":
				server.Memory["unicode"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["unicode"].SortedSet.Add("成员1", 1.0)
				server.Memory["unicode"].SortedSet.Add("成员2", 2.0)
			case "zcard after adding members":
				server.Memory["dynamic"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["dynamic"].SortedSet.Add("member1", 1.0)
				server.Memory["dynamic"].SortedSet.Add("member2", 2.0)
			case "zcard after removing members":
				server.Memory["reduced"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["reduced"].SortedSet.Add("member1", 1.0)
				server.Memory["reduced"].SortedSet.Add("member2", 2.0)
				server.Memory["reduced"].SortedSet.Add("member3", 3.0)
				// Remove one member
				server.Memory["reduced"].SortedSet.Remove("member2")
				server.Memory["reduced"].SortedSet.Remove("member3")
			}

			result := Zcard(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Zcard() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "integer" && result.Num != tt.expected.Num {
				t.Errorf("Zcard() num = %v, expected %v", result.Num, tt.expected.Num)
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Zcard() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkZcard(b *testing.B) {
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
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zcard(connID, args)
	}
}

func BenchmarkZcardLargeSet(b *testing.B) {
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
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zcard(connID, args)
	}
}

func BenchmarkZcardEmptySet(b *testing.B) {
	clearMemory()

	// Set up empty test data
	server.Memory["emptykey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "emptykey"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zcard(connID, args)
	}
}

func BenchmarkZcardNonExistent(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "nonexistent"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zcard(connID, args)
	}
}
