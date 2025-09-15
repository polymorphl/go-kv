package shared

import (
	"sort"
	"sync"
)

// Object pool for member slices to reduce allocations
var (
	memberScorePool = sync.Pool{
		New: func() interface{} {
			return make([]memberScore, 0, 16)
		},
	}
)

// Helper functions for pool management
func getMemberScoreSlice() []memberScore {
	return memberScorePool.Get().([]memberScore)
}

func putMemberScoreSlice(s []memberScore) {
	s = s[:0] // Reset length but keep capacity
	memberScorePool.Put(s)
}

// SortedSetMember represents a member in a sorted set with its score
type SortedSetMember struct {
	Score  float64
	Member string
}

// SortedSet represents a Redis sorted set (ZSET)
type SortedSet struct {
	Members map[string]float64 // Map of member -> score for O(1) lookups
	Size    int                // Number of members
}

// NewSortedSet creates a new empty sorted set
func NewSortedSet() *SortedSet {
	return &SortedSet{
		Members: make(map[string]float64),
		Size:    0,
	}
}

// Add adds a member with a score to the sorted set
// Returns true if the member was added (new), false if it was updated (existing)
func (ss *SortedSet) Add(member string, score float64) bool {
	_, exists := ss.Members[member]
	ss.Members[member] = score
	if !exists {
		ss.Size++
		return true
	}
	return false
}

// GetScore returns the score of a member, or 0 and false if not found
func (ss *SortedSet) GetScore(member string) (float64, bool) {
	score, exists := ss.Members[member]
	return score, exists
}

// memberScore represents a member with its score for sorting
type memberScore struct {
	member string
	score  float64
}

// GetRank returns the rank (0-based index) of a member in the sorted set
// Members are sorted by score in ascending order, then by member name alphabetically
func (ss *SortedSet) GetRank(member string) (int, bool) {
	_, exists := ss.Members[member]
	if !exists {
		return 0, false
	}

	// Use object pool for member slice
	members := getMemberScoreSlice()
	defer putMemberScoreSlice(members)

	// Pre-allocate with exact capacity
	members = members[:0]
	if cap(members) < len(ss.Members) {
		members = make([]memberScore, 0, len(ss.Members))
	}

	// Create a slice of members with their scores for sorting
	for m, s := range ss.Members {
		members = append(members, memberScore{m, s})
	}

	// Use Go's efficient sort instead of bubble sort
	sort.Slice(members, func(i, j int) bool {
		if members[i].score != members[j].score {
			return members[i].score < members[j].score
		}
		return members[i].member < members[j].member
	})

	// Find the rank of the target member
	for i, ms := range members {
		if ms.member == member {
			return i, true
		}
	}

	return 0, false
}

// GetSortedMembers returns all members of the sorted set in sorted order
// Members are sorted by score (ascending), then by member name (alphabetically)
func (ss *SortedSet) GetSortedMembers() []string {
	// Create a slice of members with their scores for sorting
	type memberScore struct {
		member string
		score  float64
	}

	members := make([]memberScore, 0, len(ss.Members))
	for m, s := range ss.Members {
		members = append(members, memberScore{m, s})
	}

	// Sort by score (ascending), then by member name (alphabetically)
	for i := 0; i < len(members)-1; i++ {
		for j := i + 1; j < len(members); j++ {
			if members[i].score > members[j].score ||
				(members[i].score == members[j].score && members[i].member > members[j].member) {
				members[i], members[j] = members[j], members[i]
			}
		}
	}

	// Extract just the member names
	result := make([]string, len(members))
	for i, ms := range members {
		result[i] = ms.member
	}

	return result
}

// Remove removes a member from the sorted set
func (ss *SortedSet) Remove(member string) bool {
	_, exists := ss.Members[member]
	if !exists {
		return false
	}
	delete(ss.Members, member)
	ss.Size--
	return true
}
