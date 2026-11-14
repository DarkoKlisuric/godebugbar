package engine

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const persistDelay = 500 * time.Millisecond

type Persister struct {
	path    string
	mu      sync.Mutex
	pending []Profile
	timer   *time.Timer
}

func NewPersister(path string) *Persister {
	return &Persister{path: path}
}

func (p *Persister) Load() ([]Profile, error) {
	data, err := os.ReadFile(p.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var profiles []Profile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (p *Persister) Schedule(profiles []Profile) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pending = profiles
	if p.timer == nil {
		p.timer = time.AfterFunc(persistDelay, p.flush)
	}
}

func (p *Persister) Flush() error {
	p.mu.Lock()
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	profiles := p.pending
	p.pending = nil
	p.mu.Unlock()

	if profiles == nil {
		return nil
	}

	return p.write(profiles)
}

func (p *Persister) flush() {
	p.mu.Lock()
	profiles := p.pending
	p.pending = nil
	p.timer = nil
	p.mu.Unlock()

	if profiles != nil {
		_ = p.write(profiles)
	}
}

func (p *Persister) write(profiles []Profile) error {
	data, err := json.Marshal(profiles)
	if err != nil {
		return err
	}

	dir := filepath.Dir(p.path)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".godebugbar-*.json")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), p.path)
}
