package memory

import "sync"

type Store interface {
	Get(key string) (string, bool)
	Set(key, value string)
	Namespace(ns string) Store
}

var _ Store = (*MapStore)(nil)

type MapStore struct {
	store  map[string]string
	mu     *sync.RWMutex
	prefix string
}

func NewMapStore() *MapStore {
	return &MapStore{
		store: make(map[string]string),
		mu:    &sync.RWMutex{},
	}
}

func (m *MapStore) key(k string) string { return m.prefix + k }

func (m *MapStore) Get(key string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.store[m.key(key)]
	return val, ok
}

func (m *MapStore) Set(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[m.key(key)] = value
}

func (m *MapStore) Namespace(ns string) Store {
	return &MapStore{
		store:  m.store,
		mu:     m.mu,
		prefix: m.prefix + ns + ":",
	}
}
