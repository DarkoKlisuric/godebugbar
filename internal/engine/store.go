package engine

import "sync"

const defaultStoreCapacity = 50

type Store struct {
	mu        sync.RWMutex
	capacity  int
	items     []Profile
	head      int
	count     int
	byID      map[string]int
	persister *Persister
}

func NewStore(capacity int) *Store {
	if capacity <= 0 {
		capacity = defaultStoreCapacity
	}

	return &Store{
		capacity: capacity,
		items:    make([]Profile, capacity),
		byID:     make(map[string]int, capacity),
	}
}

func (s *Store) SetPersister(p *Persister) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.persister = p
}

func (s *Store) Save(p Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.saveLocked(p)
	if s.persister != nil {
		s.persister.Schedule(s.listLocked())
	}
}

func (s *Store) Restore(profiles []Profile) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := len(profiles) - 1; i >= 0; i-- {
		s.saveLocked(profiles[i])
	}
}

func (s *Store) Get(id string) (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, ok := s.byID[id]
	if !ok {
		return Profile{}, false
	}

	return s.items[idx], true
}

func (s *Store) Latest() (Profile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.count == 0 {
		return Profile{}, false
	}

	return s.items[s.slot(s.count-1)], true
}

func (s *Store) List() []Profile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.listLocked()
}

func (s *Store) Summaries(since string) []ProfileSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]ProfileSummary, 0, s.count)
	for i := s.count - 1; i >= 0; i-- {
		p := &s.items[s.slot(i)]
		if since != "" && p.ID == since {
			break
		}
		out = append(out, p.Summary())
	}

	return out
}

func (s *Store) saveLocked(p Profile) {
	if old, ok := s.byID[p.ID]; ok {
		s.items[old] = p
		return
	}

	idx := s.slot(s.count)
	if s.count == s.capacity {
		delete(s.byID, s.items[s.head].ID)
		idx = s.head
		s.head = (s.head + 1) % s.capacity
	} else {
		s.count++
	}

	s.items[idx] = p
	s.byID[p.ID] = idx
}

func (s *Store) listLocked() []Profile {
	out := make([]Profile, 0, s.count)
	for i := s.count - 1; i >= 0; i-- {
		out = append(out, s.items[s.slot(i)])
	}

	return out
}

func (s *Store) slot(i int) int {
	return (s.head + i) % s.capacity
}
