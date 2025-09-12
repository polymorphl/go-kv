package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestZrem(t *testing.T) {
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
			name:   "zrem single existing member",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "member1"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should still exist")
				}
				if entry.SortedSet.Size != 2 {
					t.Errorf("Expected size 2, got %d", entry.SortedSet.Size)
				}
				// Verify member1 is removed
				if _, exists := entry.SortedSet.GetScore("member1"); exists {
					t.Error("member1 should be removed")
				}
				// Verify other members still exist
				if _, exists := entry.SortedSet.GetScore("member2"); !exists {
					t.Error("member2 should still exist")
				}
				if _, exists := entry.SortedSet.GetScore("member3"); !exists {
					t.Error("member3 should still exist")
				}
			},
		},
		{
			name:   "zrem multiple existing members",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "member2"},
				{Typ: "bulk", Bulk: "member3"},
			},
			expected: shared.Value{Typ: "integer", Num: 2},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should still exist")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
				// Verify member2 and member3 are removed
				if _, exists := entry.SortedSet.GetScore("member2"); exists {
					t.Error("member2 should be removed")
				}
				if _, exists := entry.SortedSet.GetScore("member3"); exists {
					t.Error("member3 should be removed")
				}
				// Verify member1 still exists
				if _, exists := entry.SortedSet.GetScore("member1"); !exists {
					t.Error("member1 should still exist")
				}
			},
		},
		{
			name:   "zrem non-existent member",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should still exist")
				}
				if entry.SortedSet.Size != 3 {
					t.Errorf("Expected size 3, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zrem mix of existing and non-existent members",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "member1"},
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "member2"},
			},
			expected: shared.Value{Typ: "integer", Num: 2}, // Only existing members removed
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should still exist")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
				// Verify member1 and member2 are removed
				if _, exists := entry.SortedSet.GetScore("member1"); exists {
					t.Error("member1 should be removed")
				}
				if _, exists := entry.SortedSet.GetScore("member2"); exists {
					t.Error("member2 should be removed")
				}
				// Verify member3 still exists
				if _, exists := entry.SortedSet.GetScore("member3"); !exists {
					t.Error("member3 should still exist")
				}
			},
		},
		{
			name:   "zrem all members",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "member1"},
				{Typ: "bulk", Bulk: "member2"},
				{Typ: "bulk", Bulk: "member3"},
			},
			expected: shared.Value{Typ: "integer", Num: 3},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should still exist")
				}
				if entry.SortedSet.Size != 0 {
					t.Errorf("Expected size 0, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zrem from non-existent key",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify:   func() {},
		},
		{
			name:   "zrem from key with wrong type",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "wrongtype"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "error", Str: "WRONGTYPE Operation against a key holding the wrong kind of value"},
			verify:   func() {},
		},
		{
			name:   "zrem unicode members",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicode"},
				{Typ: "bulk", Bulk: "成员1"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["unicode"]
				if !exists {
					t.Error("Key should still exist")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
				// Verify 成员1 is removed
				if _, exists := entry.SortedSet.GetScore("成员1"); exists {
					t.Error("成员1 should be removed")
				}
				// Verify 成员2 still exists
				if _, exists := entry.SortedSet.GetScore("成员2"); !exists {
					t.Error("成员2 should still exist")
				}
			},
		},
		{
			name:   "zrem from empty set",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "empty"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "integer", Num: 0},
			verify: func() {
				entry, exists := server.Memory["empty"]
				if !exists {
					t.Error("Key should still exist")
				}
				if entry.SortedSet.Size != 0 {
					t.Errorf("Expected size 0, got %d", entry.SortedSet.Size)
				}
			},
		},
		{
			name:   "zrem wrong number of arguments",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zrem' command"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up test data
			switch tt.name {
			case "zrem single existing member", "zrem multiple existing members", "zrem non-existent member", "zrem mix of existing and non-existent members", "zrem all members":
				server.Memory["myzset"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["myzset"].SortedSet.Add("member1", 1.0)
				server.Memory["myzset"].SortedSet.Add("member2", 2.0)
				server.Memory["myzset"].SortedSet.Add("member3", 3.0)
			case "zrem from key with wrong type":
				server.Memory["wrongtype"] = shared.MemoryEntry{
					Value:   "string value",
					Expires: 0,
				}
			case "zrem unicode members":
				server.Memory["unicode"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["unicode"].SortedSet.Add("成员1", 1.0)
				server.Memory["unicode"].SortedSet.Add("成员2", 2.0)
			case "zrem from empty set":
				server.Memory["empty"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
			}

			result := Zrem(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Zrem() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "integer" && result.Num != tt.expected.Num {
				t.Errorf("Zrem() num = %v, expected %v", result.Num, tt.expected.Num)
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Zrem() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkZrem(b *testing.B) {
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
		Zrem(connID, args)
		// Re-add the member for next iteration
		server.Memory["benchkey"].SortedSet.Add("member2", 2.0)
	}
}

func BenchmarkZremMultiple(b *testing.B) {
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
		{Typ: "bulk", Bulk: "member1"},
		{Typ: "bulk", Bulk: "member2"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zrem(connID, args)
		// Re-add the members for next iteration
		server.Memory["benchkey"].SortedSet.Add("member1", 1.0)
		server.Memory["benchkey"].SortedSet.Add("member2", 2.0)
	}
}

func BenchmarkZremLargeSet(b *testing.B) {
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
		Zrem(connID, args)
		// Re-add the member for next iteration
		server.Memory["largekey"].SortedSet.Add("member500", 500.0)
	}
}

func BenchmarkZremNonExistent(b *testing.B) {
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
		Zrem(connID, args)
	}
}
