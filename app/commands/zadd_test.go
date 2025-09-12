package commands

import (
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/shared"
)

func TestZadd(t *testing.T) {
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
			name:   "zadd single member",
			connID: "test-conn-1",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "1.5"},
				{Typ: "bulk", Bulk: "member1"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist after ZADD")
				}
				if entry.SortedSet == nil {
					t.Error("SortedSet should be created")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
				score, exists := entry.SortedSet.GetScore("member1")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 1.5 {
					t.Errorf("Expected score 1.5, got %f", score)
				}
			},
		},
		{
			name:   "zadd multiple members",
			connID: "test-conn-2",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "1.0"},
				{Typ: "bulk", Bulk: "member1"},
				{Typ: "bulk", Bulk: "2.0"},
				{Typ: "bulk", Bulk: "member2"},
				{Typ: "bulk", Bulk: "3.0"},
				{Typ: "bulk", Bulk: "member3"},
			},
			expected: shared.Value{Typ: "integer", Num: 3},
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist after ZADD")
				}
				if entry.SortedSet.Size != 3 {
					t.Errorf("Expected size 3, got %d", entry.SortedSet.Size)
				}
				// Verify all members exist with correct scores
				expectedScores := map[string]float64{
					"member1": 1.0,
					"member2": 2.0,
					"member3": 3.0,
				}
				for member, expectedScore := range expectedScores {
					score, exists := entry.SortedSet.GetScore(member)
					if !exists {
						t.Errorf("Member %s should exist", member)
					}
					if score != expectedScore {
						t.Errorf("Expected score %f for %s, got %f", expectedScore, member, score)
					}
				}
			},
		},
		{
			name:   "zadd update existing member",
			connID: "test-conn-3",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "myzset"},
				{Typ: "bulk", Bulk: "5.0"},
				{Typ: "bulk", Bulk: "existing"},
			},
			expected: shared.Value{Typ: "integer", Num: 0}, // 0 because it's an update, not new
			verify: func() {
				entry, exists := server.Memory["myzset"]
				if !exists {
					t.Error("Key should exist after ZADD")
				}
				if entry.SortedSet.Size != 1 {
					t.Errorf("Expected size 1, got %d", entry.SortedSet.Size)
				}
				score, exists := entry.SortedSet.GetScore("existing")
				if !exists {
					t.Error("Member should exist")
				}
				if score != 5.0 {
					t.Errorf("Expected score 5.0, got %f", score)
				}
			},
		},
		{
			name:   "zadd with high precision scores",
			connID: "test-conn-4",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "precision"},
				{Typ: "bulk", Bulk: "19.608968014838933"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["precision"]
				if !exists {
					t.Error("Key should exist after ZADD")
				}
				score, exists := entry.SortedSet.GetScore("member")
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
			name:   "zadd with negative scores",
			connID: "test-conn-5",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "negative"},
				{Typ: "bulk", Bulk: "-1.5"},
				{Typ: "bulk", Bulk: "negative_member"},
			},
			expected: shared.Value{Typ: "integer", Num: 1},
			verify: func() {
				entry, exists := server.Memory["negative"]
				if !exists {
					t.Error("Key should exist after ZADD")
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
			name:   "zadd with unicode members",
			connID: "test-conn-6",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "unicode"},
				{Typ: "bulk", Bulk: "1.0"},
				{Typ: "bulk", Bulk: "成员1"},
				{Typ: "bulk", Bulk: "2.0"},
				{Typ: "bulk", Bulk: "🌍"},
			},
			expected: shared.Value{Typ: "integer", Num: 2},
			verify: func() {
				entry, exists := server.Memory["unicode"]
				if !exists {
					t.Error("Key should exist after ZADD")
				}
				if entry.SortedSet.Size != 2 {
					t.Errorf("Expected size 2, got %d", entry.SortedSet.Size)
				}
				// Verify unicode members exist
				score1, exists1 := entry.SortedSet.GetScore("成员1")
				score2, exists2 := entry.SortedSet.GetScore("🌍")
				if !exists1 || !exists2 {
					t.Error("Unicode members should exist")
				}
				if score1 != 1.0 || score2 != 2.0 {
					t.Errorf("Expected scores 1.0 and 2.0, got %f and %f", score1, score2)
				}
			},
		},
		{
			name:   "zadd wrong number of arguments",
			connID: "test-conn-7",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "1.0"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zadd' command"},
			verify:   func() {},
		},
		{
			name:   "zadd odd number of score-member pairs",
			connID: "test-conn-8",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "1.0"},
				{Typ: "bulk", Bulk: "member1"},
				{Typ: "bulk", Bulk: "2.0"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR wrong number of arguments for 'zadd' command"},
			verify:   func() {},
		},
		{
			name:   "zadd invalid score",
			connID: "test-conn-9",
			args: []shared.Value{
				{Typ: "bulk", Bulk: "key"},
				{Typ: "bulk", Bulk: "notanumber"},
				{Typ: "bulk", Bulk: "member"},
			},
			expected: shared.Value{Typ: "error", Str: "ERR value is not a valid float"},
			verify:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearMemory()

			// Set up initial data for update test
			if tt.name == "zadd update existing member" {
				server.Memory["myzset"] = shared.MemoryEntry{
					SortedSet: shared.NewSortedSet(),
					Expires:   0,
				}
				server.Memory["myzset"].SortedSet.Add("existing", 1.0)
			}

			result := Zadd(tt.connID, tt.args)

			if result.Typ != tt.expected.Typ {
				t.Errorf("Zadd() type = %v, expected %v", result.Typ, tt.expected.Typ)
			}

			if result.Typ == "integer" && result.Num != tt.expected.Num {
				t.Errorf("Zadd() num = %v, expected %v", result.Num, tt.expected.Num)
			}

			if result.Typ == "error" && result.Str != tt.expected.Str {
				t.Errorf("Zadd() error = %v, expected %v", result.Str, tt.expected.Str)
			}

			tt.verify()
		})
	}
}

func BenchmarkZadd(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "1.5"},
		{Typ: "bulk", Bulk: "member"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zadd(connID, args)
	}
}

func BenchmarkZaddMultiple(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "1.0"},
		{Typ: "bulk", Bulk: "member1"},
		{Typ: "bulk", Bulk: "2.0"},
		{Typ: "bulk", Bulk: "member2"},
		{Typ: "bulk", Bulk: "3.0"},
		{Typ: "bulk", Bulk: "member3"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zadd(connID, args)
	}
}

func BenchmarkZaddHighPrecision(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "19.608968014838933"},
		{Typ: "bulk", Bulk: "member"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zadd(connID, args)
	}
}

func BenchmarkZaddUnicode(b *testing.B) {
	clearMemory()

	connID := "benchmark-conn"
	args := []shared.Value{
		{Typ: "bulk", Bulk: "benchkey"},
		{Typ: "bulk", Bulk: "1.0"},
		{Typ: "bulk", Bulk: "成员1"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Zadd(connID, args)
	}
}
