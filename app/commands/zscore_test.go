package commands

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestZscore(t *testing.T) {
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
			name:   "zscore existing member",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "member1"},
			},
			expected: shared.Value{Typ: "bulk", Bulk: "1"},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				score, exists := entry.SortedSet.GetScore("member1")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 1.0 {
					t.Errorf("Expected score 1.0, got %f", score)
				}
			},
		},
		{
			name:   "zscore member with high precision",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "precision_member"},
			},
			expected: shared.Value{Typ: "bulk", Bulk: "19.6089680148389"},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				score, exists := entry.SortedSet.GetScore("precision_member")
				if !exists {
					t.Error("Member should exist")
				}
				expectedScore := 19.608968014838933
				if score != expectedScore {
					t.Errorf("Expected score %f, got %f", expectedScore, score)
				}
			},
		},
		{
			name:   "zscore member with negative score",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "negative_member"},
			},
			expected: shared.Value{Typ: "bulk", Bulk: "-1.5"},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				score, exists := entry.SortedSet.GetScore("negative_member")
				if !exists {
					t.Error("Member should exist")
				}
				if score != -1.5 {
					t.Errorf("Expected score -1.5, got %f", score)
				}
			},
		},
		{
			name:   "zscore member with zero score",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "zero_member"},
			},
			expected: shared.Value{Typ: "bulk", Bulk: "0"},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist")
				}
				score, exists := entry.SortedSet.GetScore("zero_member")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 0.0 {
					t.Errorf("Expected score 0.0, got %f", score)
				}
			},
		},
		{
			name:   "zscore non-existent member",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "nonexistent"},
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "zscore non-existent key",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "nonexistent"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "zscore key with wrong type",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "wrongtype"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "zscore unicode member",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicode"},
				{Typ: "bulk", Bulk: "成员1"},
			},
			expected: shared.Value{Typ: "bulk", Bulk: "1"},
			verify: func() {
				entry, exists := server.Memory["unicode"]
				if !exists {
					t.Error("Key should exist")
				}
				score, exists := entry.SortedSet.GetScore("成员1")
				if !exists {
					t.Error("Unicode member should exist")
				}
				if score != 1.0 {
					t.Errorf("Expected score 1.0, got %f", score)
				}
			},
		},
		{
			name:   "zscore empty set",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "empty"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "null", Str: ""},
			verify:   func() {},
		},
		{
			name:   "zscore wrong number of arguments",
			connID: "test-conn-10",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zscore' command"},
			verify:   func() {},
		},
		{
			name:   "zscore too many arguments",
			connID: "test-conn-11",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "member1"},
				{Typ: "bulk", Bulk: "member2"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zscore' command"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up test data
			switch tt.name {
			case "zscore existing member", "zscore member with high precision", "zscore member with negative score", "zscore member with zero score", "zscore non-existent member":
				server.Memory["myzset"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["myzset"].SortedSet.Add("member1", 1.0)
				server.Memory["myzset"].SortedSet.Add("precision_member", 19.608968014838933)
				server.Memory["myzset"].SortedSet.Add("negative_member", -1.5)
				server.Memory["myzset"].SortedSet.Add("zero_member", 0.0)
			case "zscore key with wrong type":
				server.Memory["wrongtype"] = shared.MemoryEntry{
					Value:   "string value",
					Expires: 0,
				}
			case "zscore unicode member":
				server.Memory["unicode"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["unicode"].SortedSet.Add("成员1", 1.0)
			case "zscore empty set":
				server.Memory["empty"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
			}

			result := Zscore(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Zscore() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "bulk" && result.Bulk != tt.expected.Bulk {
				t.Errorf("Zscore() bulk = %v, expected %v", result.Bulk, tt.expected.Bulk)
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Zscore() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkZscore(b *testing.B) {
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
		Zscore(connID, args)
	}
}

func BenchmarkZscoreHighPrecision(b *testing.B) {
	clearMemory()

	// Set up test data
	server.Memory["benchkey"] = shared.MemoryEntry{
		SortedSet: shared.NewSortedSet(),
		Expires:   0,
	}
	server.Memory["benchkey"].SortedSet.Add("precision_member", 19.608968014838933)

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "precision_member"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zscore(connID, args)
	}
}

func BenchmarkZscoreLargeSet(b *testing.B) {
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
		Zscore(connID, args)
	}
}

func BenchmarkZscoreNonExistent(b *testing.B) {
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
		Zscore(connID, args)
	}
}
