package agent

import (
	"context"
	"testing"
)

// fakeAgent is a minimal Agent for exercising the registry.
type fakeAgent struct{ meta Meta }

func (f fakeAgent) Meta() Meta                                             { return f.meta }
func (f fakeAgent) BuildBrief(context.Context, Input, Deps) (Brief, error) { return Brief{}, nil }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()

	if _, ok := r.Get("nope"); ok {
		t.Fatal("Get on empty registry should return ok=false")
	}

	r.Register(fakeAgent{meta: Meta{Name: "alpha"}})
	got, ok := r.Get("alpha")
	if !ok {
		t.Fatal("expected to find registered agent")
	}
	if got.Meta().Name != "alpha" {
		t.Errorf("Name = %q; want alpha", got.Meta().Name)
	}
}

func TestRegistry_OverrideLastWins(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeAgent{meta: Meta{Name: "dup", Title: "first"}})
	r.Register(fakeAgent{meta: Meta{Name: "dup", Title: "second"}})

	got, _ := r.Get("dup")
	if got.Meta().Title != "second" {
		t.Errorf("Title = %q; want second (last registration wins)", got.Meta().Title)
	}
}

func TestRegistry_ListSorted(t *testing.T) {
	r := NewRegistry()
	r.Register(fakeAgent{meta: Meta{Name: "charlie"}})
	r.Register(fakeAgent{meta: Meta{Name: "alpha"}})
	r.Register(fakeAgent{meta: Meta{Name: "bravo"}})

	metas := r.List()
	if len(metas) != 3 {
		t.Fatalf("List len = %d; want 3", len(metas))
	}
	want := []string{"alpha", "bravo", "charlie"}
	for i, m := range metas {
		if m.Name != want[i] {
			t.Errorf("List[%d].Name = %q; want %q", i, m.Name, want[i])
		}
	}
}
