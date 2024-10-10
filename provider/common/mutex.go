package common

import (
	"sync"
)

type mutexKV struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

func (m *mutexKV) Lock(key string) {
	m.get(key).Lock()
}

func (m *mutexKV) Unlock(key string) {
	m.get(key).Unlock()
}

func NewMutexKV() *mutexKV {
	return &mutexKV{
		store: make(map[string]*sync.Mutex),
	}
}

func (m *mutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}
