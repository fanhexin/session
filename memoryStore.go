package session

import (
	"sync"
)

type memoryStore struct {
	sync.RWMutex
	items map[string]Session
}

func (ms *memoryStore) Has(key string) bool {
	ms.RLock()
	defer ms.RUnlock()
	_, e := ms.items[key]
	return e
}

func (ms *memoryStore) Get(key string) Session {
	ms.RLock()
	defer ms.RUnlock()
	return ms.items[key]
}

func (ms *memoryStore) Add(value Session) bool {
	if ms.Has(value.Id()) {
		return false
	}

	ms.Lock()
	defer ms.Unlock()
	ms.items[value.Id()] = value
	return true
}

func (ms *memoryStore) Len() int {
	ms.RLock()
	defer ms.RUnlock()
	return len(ms.items)
}

func (ms *memoryStore) Remove(key string) {
	ms.Lock()
	defer ms.Unlock()
	delete(ms.items, key)
}

/*
func (ms *memoryStore) RemoveAll(predicate func(value Session) bool) {
	ms.Lock()
	defer ms.Unlock()
	for k, v := range ms.items {
		if predicate(v) {
			delete(ms.items, k)
		}
	}
}
*/

func NewMemoryStore() Store {
	return &memoryStore{items: map[string]Session{}}
}
