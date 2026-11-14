package engine_test

import (
	"fmt"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestStoreRingBuffer(t *testing.T) {
	t.Parallel()

	store := engine.NewStore(2)
	if _, ok := store.Latest(); ok {
		t.Fatalf("expected empty store Latest to return false")
	}

	store.Save(engine.Profile{ID: "1", Path: "/one"})
	store.Save(engine.Profile{ID: "2", Path: "/two"})
	store.Save(engine.Profile{ID: "3", Path: "/three"})

	if _, ok := store.Get("1"); ok {
		t.Fatalf("expected oldest profile to be evicted")
	}

	p2, ok := store.Get("2")
	if !ok || p2.Path != "/two" {
		t.Fatalf("expected profile 2 to remain, got %+v", p2)
	}

	latest, ok := store.Latest()
	if !ok || latest.ID != "3" {
		t.Fatalf("expected latest profile to be 3, got %+v", latest)
	}

	list := store.List()
	if len(list) != 2 || list[0].ID != "3" || list[1].ID != "2" {
		t.Fatalf("expected newest-first list [3, 2], got %+v", list)
	}
}

func TestStoreDefaultCapacity(t *testing.T) {
	t.Parallel()

	store := engine.NewStore(0)
	for i := range 60 {
		store.Save(engine.Profile{ID: fmt.Sprintf("id-%d", i)})
	}

	if len(store.List()) != 50 {
		t.Fatalf("expected default capacity of 50, got %d", len(store.List()))
	}
}
