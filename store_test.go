package godebugbar

import "testing"

func TestStore(t *testing.T) {
	t.Parallel()

	store := NewStore(2)
	store.Save(nil)
	store.Save(&Profile{ID: "1", Path: "/one"})
	store.Save(&Profile{ID: "2", Path: "/two"})
	store.Save(&Profile{ID: "3", Path: "/three"})

	if _, ok := store.Get("1"); ok {
		t.Error("Get(1) = true, want evicted false")
	}

	got, ok := store.Get("3")
	if !ok || got.Path != "/three" {
		t.Errorf("Get(3) = %+v, %v, want /three, true", got, ok)
	}

	list := store.List()
	if len(list) != 2 || list[0].ID != "3" || list[1].ID != "2" {
		t.Errorf("List() = %+v, want newest first [3, 2]", list)
	}

	store.Clear()
	if len(store.List()) != 0 {
		t.Errorf("len(List()) after Clear = %d, want 0", len(store.List()))
	}
}
