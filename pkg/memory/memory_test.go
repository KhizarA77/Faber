package memory

import (
	"fmt"
	"testing"
)

func TestMapStore_GetSet(t *testing.T) {
	s := NewMapStore()

	if _, ok := s.Get("missing"); ok {
		t.Fatal("expected missing key to return ok=false")
	}

	s.Set("k", "v")
	got, ok := s.Get("k")
	if !ok || got != "v" {
		t.Fatalf("Get(k) = %q, %v; want \"v\", true", got, ok)
	}
}

func TestMapStore_NamespaceIsolation(t *testing.T) {
	root := NewMapStore()
	a := root.Namespace("a")
	b := root.Namespace("b")

	a.Set("key", "from-a")
	b.Set("key", "from-b")

	if got, _ := a.Get("key"); got != "from-a" {
		t.Errorf("a.Get(key) = %q; want from-a", got)
	}
	if got, _ := b.Get("key"); got != "from-b" {
		t.Errorf("b.Get(key) = %q; want from-b", got)
	}
	// The root view must not see a namespaced key under its bare name.
	if _, ok := root.Get("key"); ok {
		t.Error("root unexpectedly sees a namespaced key")
	}
}

func TestMapStore_ConcurrentAccess(t *testing.T) {
	s := NewMapStore()
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			s.Set(fmt.Sprintf("k%d", i), "v")
			_, _ = s.Get("k0")
			done <- struct{}{}
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}
