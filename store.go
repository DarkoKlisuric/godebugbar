package godebugbar

import "sync"

const defaultStoreCapacity = 100

type Store struct {
	mu       sync.RWMutex
	capacity int
	items    []*Profile
}

func NewStore(capacity int) *Store {
	if capacity <= 0 {
		capacity = defaultStoreCapacity
	}

	return &Store{
		capacity: capacity,
		items:    make([]*Profile, 0, capacity),
	}
}

func (s *Store) Save(profile *Profile) {
	if profile == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.items) >= s.capacity {
		copy(s.items, s.items[1:])
		s.items[len(s.items)-1] = profile

		return
	}

	s.items = append(s.items, profile)
}

func (s *Store) Get(id string) (*Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := len(s.items) - 1; i >= 0; i-- {
		if s.items[i].ID == id {
			return s.items[i], true
		}
	}

	return nil, false
}

func (s *Store) List() []*Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Profile, len(s.items))
	for i := range s.items {
		result[i] = s.items[len(s.items)-1-i]
	}

	return result
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = s.items[:0]
}
