package internal

import "sync"

type Cache[K comparable, V any] struct {
	_map map[K]*Node[K, V]
	capacity int
	policy EvictionPolicy[K, V]
	mu sync.Mutex
}

// NewCache allocates and returns a new [Cache[K, V]]
func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{
		_map: map[K]*Node[K, V]{},
		capacity: capacity,
		policy: NewLRUEvictionPolicy[K, V](),
	}
}

func (cache *Cache[K, V]) Put(key K, value V) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	// if the key already exists
	if node, exists := cache._map[key]; exists {
		node.value = value
		cache._map[key] = node
		cache.policy.onGet(node)
		return
	}
	if(len(cache._map) == cache.capacity) {
		removedNode := cache.policy.evict()
		if removedNode != nil {
			delete(cache._map, removedNode.key)
		}
	}
	newNode := NewNode(key, value)
	cache.policy.onInsert(newNode)
	cache._map[key] = newNode
}

// returns the value and a bool flag indicating if the key exists in the cache
func (cache *Cache[K, V]) Get(key K) (V, bool){
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if node, exists := cache._map[key]; exists {
		cache.policy.onGet(node)
		return node.value, true
	}
	var zero V
	return zero, false
}

func (cache *Cache[K, V]) Remove(key K) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if node, exists := cache._map[key]; exists {
		cache.policy.onRemove(node)
		delete(cache._map, key)
	}
}