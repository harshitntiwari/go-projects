package internal

type EvictionPolicy[K comparable, V any] interface {
	onInsert(node *Node[K, V])
	onGet(node *Node[K, V])
	onRemove(node *Node[K, V])
	evict() *Node[K, V]
}

type LRUEvictionPolicy[K comparable, V any] struct {
	dll *DoublyLinkedList[K, V]
}

func NewLRUEvictionPolicy[K comparable, V any]() *LRUEvictionPolicy[K, V] {
	return &LRUEvictionPolicy[K, V]{
		dll: NewList[K, V](),
	}
}

func (p *LRUEvictionPolicy[K, V]) onInsert(node *Node[K, V]) {
	p.dll.addFirst(node)
}

func (p *LRUEvictionPolicy[K, V]) onGet(node *Node[K, V]) {
	p.dll.moveToFront(node)
}

func (p *LRUEvictionPolicy[K, V]) onRemove(node *Node[K, V]) {
	p.dll.remove(node)
}

func (p *LRUEvictionPolicy[K, V]) evict() *Node[K, V] {
	return p.dll.removeLast()
}