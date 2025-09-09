package shared

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

// FromArray creates a linked list from a slice
func FromArray(arr []string) *LinkedList {
	ll := NewLinkedList()
	for _, value := range arr {
		ll.AddToTail(value)
	}
	return ll
}
