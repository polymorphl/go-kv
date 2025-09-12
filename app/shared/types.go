package shared

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/protocol"
)

// Type aliases for backward compatibility
type Value = protocol.Value
type Resp = protocol.Resp
type Writer = protocol.Writer

// StreamEntry represents a single entry in a Redis stream
type StreamEntry struct {
	ID   string            // Stream ID (e.g., "1526985054069-0")
	Data map[string]string // Field-value pairs
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

// MemoryEntry represents a value stored in the in-memory database.
// It can hold either a string value, an array of strings, or a linked list, with optional expiration.
type MemoryEntry struct {
	Value     string        // String value (used when Array is empty)
	Array     []string      // Array of strings (used for list operations - kept for compatibility)
	List      *LinkedList   // Linked list (used for optimized list operations)
	Stream    []StreamEntry // Stream entries (used for stream operations)
	SortedSet *SortedSet    // Sorted set (used for sorted set operations)
	Expires   int64         // Unix timestamp in milliseconds, 0 means no expiry
}

// QueuedCommand represents a command that is queued in a transaction.
type QueuedCommand struct {
	Command string
	Args    []protocol.Value
}

// Transaction represents a transaction that is being executed.
type Transaction struct {
	Commands []QueuedCommand
}

// ListNode represents a single node in a doubly linked list
type ListNode struct {
	Value string
	Prev  *ListNode
	Next  *ListNode
}

// LinkedList represents a doubly linked list with head, tail, and size tracking
type LinkedList struct {
	Head *ListNode
	Tail *ListNode
	Size int
}

// CommandHandler represents a function that handles a Redis command
type CommandHandler func(string, []protocol.Value) protocol.Value

// State represents the server state including replication information
type State struct {
	Role             string
	ReplicaOf        string
	MasterReplID     string
	MasterReplOffset int64
	Replicas         map[string]net.Conn // Map of replica connection IDs to their connections
	ConfigDir        string              // Directory where Redis stores its data
	ConfigDbfilename string              // Database filename
}

// NewLinkedList creates a new empty linked list
func NewLinkedList() *LinkedList {
	return &LinkedList{Size: 0}
}

// AddToHead adds a value to the head of the linked list (for LPUSH)
func (ll *LinkedList) AddToHead(value string) {
	newNode := &ListNode{Value: value, Next: ll.Head}

	if ll.Head != nil {
		ll.Head.Prev = newNode
	} else {
		ll.Tail = newNode
	}

	ll.Head = newNode
	ll.Size++
}

// AddToTail adds a value to the tail of the linked list (for RPUSH)
func (ll *LinkedList) AddToTail(value string) {
	newNode := &ListNode{Value: value, Prev: ll.Tail}

	if ll.Tail != nil {
		ll.Tail.Next = newNode
	} else {
		ll.Head = newNode
	}

	ll.Tail = newNode
	ll.Size++
}

// ToArray converts the linked list to a slice (for compatibility)
func (ll *LinkedList) ToArray() []string {
	if ll.Size == 0 {
		return []string{}
	}

	result := make([]string, ll.Size)
	current := ll.Head
	for i := 0; i < ll.Size; i++ {
		result[i] = current.Value
		current = current.Next
	}
	return result
}

// RemoveFromHead removes and returns the value at the head of the linked list (for LPOP/BLPOP)
func (ll *LinkedList) RemoveFromHead() string {
	if ll.Size == 0 {
		return ""
	}

	value := ll.Head.Value
	ll.Head = ll.Head.Next

	if ll.Head != nil {
		ll.Head.Prev = nil
	} else {
		ll.Tail = nil
	}

	ll.Size--
	return value
}

// FromArray creates a linked list from a slice
func FromArray(arr []string) *LinkedList {
	ll := NewLinkedList()
	for _, value := range arr {
		ll.AddToTail(value)
	}
	return ll
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

// GetRank returns the rank (0-based index) of a member in the sorted set
// Members are sorted by score in ascending order, then by member name lexicographically
func (ss *SortedSet) GetRank(member string) (int, bool) {
	_, exists := ss.Members[member]
	if !exists {
		return 0, false
	}

	// Create a slice of members with their scores for sorting
	type memberScore struct {
		member string
		score  float64
	}

	members := make([]memberScore, 0, len(ss.Members))
	for m, s := range ss.Members {
		members = append(members, memberScore{m, s})
	}

	// Sort by score (ascending), then by member name (lexicographically)
	for i := 0; i < len(members)-1; i++ {
		for j := i + 1; j < len(members); j++ {
			if members[i].score > members[j].score ||
				(members[i].score == members[j].score && members[i].member > members[j].member) {
				members[i], members[j] = members[j], members[i]
			}
		}
	}

	// Find the rank of the target member
	for i, ms := range members {
		if ms.member == member {
			return i, true
		}
	}

	return 0, false
}

// GetSortedMembers returns all members of the sorted set in sorted order
// Members are sorted by score (ascending), then by member name (lexicographically)
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

	// Sort by score (ascending), then by member name (lexicographically)
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
	ss.Size--
	return true
}
