package engine_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/DarkoKlisuric/godebugbar/internal/engine"
)

func TestStoreRingBuffer(t *testing.T) {
	t.Parallel()

	store := engine.NewStore(2)
	if _, ok := store.Latest(); ok {
		t.Fatalf("expected empty store Latest to return false")
	}

	for i := 1; i <= 5; i++ {
		store.Save(engine.Profile{ID: fmt.Sprint(i), Path: fmt.Sprintf("/%d", i)})
	}

	cases := []struct {
		id   string
		want bool
	}{
		{id: "3", want: false},
		{id: "4", want: true},
		{id: "5", want: true},
	}
	for _, tc := range cases {
		if _, ok := store.Get(tc.id); ok != tc.want {
			t.Errorf("Get(%s) found = %v, want %v", tc.id, ok, tc.want)
		}
	}

	list := store.List()
	if len(list) != 2 || list[0].ID != "5" || list[1].ID != "4" {
		t.Fatalf("expected newest-first [5 4], got %+v", list)
	}
	if latest, _ := store.Latest(); latest.ID != "5" {
		t.Fatalf("Latest() = %s, want 5", latest.ID)
	}

	store.Save(engine.Profile{ID: "5", Path: "/updated"})
	if got, _ := store.Get("5"); got.Path != "/updated" || len(store.List()) != 2 {
		t.Fatalf("expected saving an existing ID to replace it in place")
	}
}

func TestStoreSummariesSince(t *testing.T) {
	t.Parallel()

	store := engine.NewStore(10)
	for i := 1; i <= 4; i++ {
		store.Save(engine.Profile{ID: fmt.Sprint(i)})
	}

	cases := []struct {
		since string
		want  int
	}{
		{since: "", want: 4},
		{since: "2", want: 2},
		{since: "4", want: 0},
		{since: "unknown", want: 4},
	}
	for _, tc := range cases {
		if got := store.Summaries(tc.since); len(got) != tc.want {
			t.Errorf("Summaries(%q) = %d items, want %d", tc.since, len(got), tc.want)
		}
	}
}

func TestStorePersistence(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "profiles.json")
	persister := engine.NewPersister(path)

	store := engine.NewStore(5)
	store.SetPersister(persister)
	store.Save(engine.Profile{ID: "a", Path: "/a"})
	store.Save(engine.Profile{ID: "b", Path: "/b"})
	if err := persister.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	loaded, err := engine.NewPersister(path).Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	restored := engine.NewStore(5)
	restored.Restore(loaded)
	if latest, ok := restored.Latest(); !ok || latest.ID != "b" || len(restored.List()) != 2 {
		t.Fatalf("expected both profiles restored with b newest, got %+v", restored.List())
	}

	if missing, err := engine.NewPersister(filepath.Join(t.TempDir(), "none.json")).Load(); err != nil || missing != nil {
		t.Fatalf("expected a missing file to load as empty, got %v %v", missing, err)
	}
}
