package engine

import "sync"

const defaultStoreCapacity = 50

type Store struct {
	mu       sync.RWMutex
	capacity int
	items    []Profile
	byID     map[string]Profile
}

func NewStore(capacity int) *Store {
	if capacity <= 0 {
		capacity = defaultStoreCapacity
	}
	return &Store{
		capacity: capacity,
		items:    make([]Profile, 0, capacity),
		byID:     make(map[string]Profile, capacity),
	}
}

func (s *Store) Save(p Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.items) >= s.capacity {
		oldest := s.items[0]
		delete(s.byID, oldest.ID)
		s.items = s.items[1:]
	}

	s.items = append(s.items, p)
	s.byID[p.ID] = p
}

func (s *Store) Get(id string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.byID[id]
	return p, ok
}

func (s *Store) Latest() (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.items) == 0 {
		return Profile{}, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Store) List() []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Profile, len(s.items))
	for i := range s.items {
		out[len(s.items)-1-i] = s.items[i]
	}
	return out
}
