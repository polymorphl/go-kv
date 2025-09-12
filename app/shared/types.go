package shared

import (
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

// MemoryEntry represents a value stored in the in-memory database.
// It can hold either a string value, an array of strings, or a linked list, with optional expiration.
type MemoryEntry struct {
	Value   string        // String value (used when Array is empty)
	Array   []string      // Array of strings (used for list operations - kept for compatibility)
	List    *LinkedList   // Linked list (used for optimized list operations)
	Stream  []StreamEntry // Stream entries (used for stream operations)
	Expires int64         // Unix timestamp in milliseconds, 0 means no expiry
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
