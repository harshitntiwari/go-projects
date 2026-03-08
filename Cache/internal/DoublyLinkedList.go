package internal

type DoublyLinkedList[K comparable, V any] struct {
	head *Node[K, V]
	tail *Node[K, V]
}

// NewList allocates and returns a new [DoublyLinkedList]
func NewList[K comparable, V any]() *DoublyLinkedList[K, V]{
	var zeroK K
	var zeroV V

	head := NewNode(zeroK, zeroV)
	tail := NewNode(zeroK, zeroV)

	head.next = tail
	tail.prev = head

	return &DoublyLinkedList[K, V]{
		head: head,
		tail: tail,
	}
}

func (dll *DoublyLinkedList[K, V]) addFirst(node *Node[K, V]) {
	node.next = dll.head.next
	node.prev = dll.head

	dll.head.next.prev = node
	dll.head.next = node
}

func (dll *DoublyLinkedList[K, V]) moveToFront(node *Node[K, V]) {
	dll.remove(node)
	dll.addFirst(node)
}

func (dll *DoublyLinkedList[K, V]) remove(node *Node[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (dll *DoublyLinkedList[K, V]) removeLast() *Node[K, V] {
	if dll.tail.prev == dll.head { 
		return nil 
	}

	nodeToRemove := dll.tail.prev
	dll.remove(nodeToRemove)
	return nodeToRemove
}