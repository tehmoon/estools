package storage

import (
	"sync"
	"github.com/tehmoon/errors"
)

type MemoryStorage struct {
	m map[string][]byte
	sync *sync.RWMutex
}

func NewMemoryStorage() (storage *MemoryStorage) {
	return &MemoryStorage{
		sync: &sync.RWMutex{},
		m: make(map[string][]byte, 0),
	}
}

func (s MemoryStorage) Init() (err error) {
	return nil
}

func (s MemoryStorage) Store(id string, data []byte) (err error) {
	s.sync.Lock()
	defer s.sync.Unlock()

	s.m[id] = data
	return nil
}

func (s MemoryStorage) Get(id string) (data []byte, err error) {
	s.sync.RLock()
	defer s.sync.RUnlock()

	data, ok := s.m[id]
	if ! ok {
		return nil, errors.New("Data not found in memory storage")
	}

	return data, nil
}
