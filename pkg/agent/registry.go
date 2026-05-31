package agent

import (
	"sort"
	"sync"
)

// Registry containing all registered agents.
type Registry struct {
	mu     sync.RWMutex
	agents map[string]Agent
}

// NewRegistry returns an Empty registry
func NewRegistry() *Registry {
	return &Registry{agents: make(map[string]Agent)}
}

// Register adds an agent to the registry. Registering the same name again overwrites
// the previous one. This is intentional so users can override a built-in agent.
func (r *Registry) Register(a Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[a.Meta().Name] = a
}

// Get returns the agent registered under the name.
func (r *Registry) Get(name string) (Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.agents[name]
	return a, ok
}

// List returns every agent's Meta, sorted by Name.
func (r *Registry) List() []Meta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metas := make([]Meta, 0, len(r.agents))
	for _, a := range r.agents {
		metas = append(metas, a.Meta())
	}
	sort.Slice(metas, func(i, j int) bool {
		return metas[i].Name < metas[j].Name
	})
	return metas
}
