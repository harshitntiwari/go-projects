package internal

type Node[K comparable, V any] struct {
	key K
	value V
	prev *Node[K, V]
	next *Node[K, V]
}

// NewNode allocates and returns a new [Node[K, V]]
func NewNode[K comparable, V any](key K, value V) *Node[K, V] {
	return &Node[K, V] {
		key: key,
		value: value,
		prev: nil,
		next: nil,
	}
}