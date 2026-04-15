package filter

import "testing"

func TestFilterSetAddGetAll(t *testing.T) {
	fs := NewFilterSet()
	id1 := fs.Add(Filter{Field: "level", Pattern: "error", Mode: Include, Enabled: true})
	id2 := fs.Add(Filter{Field: "any", Pattern: "timeout", Mode: Exclude, Enabled: false})
	if id1 == id2 {
		t.Fatal("IDs should be unique")
	}
	all := fs.GetAll()
	if len(all) != 2 {
		t.Fatalf("GetAll() len = %d, want 2", len(all))
	}
}

func TestFilterSetRemove(t *testing.T) {
	fs := NewFilterSet()
	id := fs.Add(Filter{Field: "msg", Pattern: "test", Mode: Include, Enabled: true})
	if !fs.Remove(id) {
		t.Error("Remove should return true")
	}
	if fs.Remove(id) {
		t.Error("Remove again should return false")
	}
	if len(fs.GetAll()) != 0 {
		t.Error("GetAll should be empty")
	}
}

func TestFilterSetEnableDisable(t *testing.T) {
	fs := NewFilterSet()
	id := fs.Add(Filter{Field: "level", Pattern: "info", Mode: Include, Enabled: true})
	if !fs.Disable(id) {
		t.Error("Disable should return true")
	}
	if fs.GetAll()[0].Enabled {
		t.Error("should be disabled")
	}
	// Retained even when disabled
	if len(fs.GetAll()) != 1 {
		t.Error("disabled filter should still be in GetAll")
	}
	if !fs.Enable(id) {
		t.Error("Enable should return true")
	}
	if !fs.GetAll()[0].Enabled {
		t.Error("should be re-enabled")
	}
}

func TestFilterSetGetEnabled(t *testing.T) {
	fs := NewFilterSet()
	fs.Add(Filter{Field: "a", Enabled: true})
	id2 := fs.Add(Filter{Field: "b", Enabled: true})
	fs.Add(Filter{Field: "c", Enabled: false})

	if len(fs.GetEnabled()) != 2 {
		t.Fatalf("GetEnabled len = %d, want 2", len(fs.GetEnabled()))
	}
	fs.Disable(id2)
	if len(fs.GetEnabled()) != 1 {
		t.Fatalf("GetEnabled len = %d, want 1", len(fs.GetEnabled()))
	}
}

func TestModeTypedEnum(t *testing.T) {
	f := Filter{Mode: Exclude}
	if f.Mode != Exclude {
		t.Error("Mode should be Exclude")
	}
	if Include.String() != "include" || Exclude.String() != "exclude" {
		t.Error("Mode.String() wrong")
	}
}
