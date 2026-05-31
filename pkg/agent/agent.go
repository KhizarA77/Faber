// Package agent defines the core contract of Faber: an Agent assembles a Brief
// for the host IDE's model to execute. Agents never call an LLM themselves —
// they gather context (notably authoritative docs) and hand back instructions.
package agent

import (
	"context"

	"github.com/KhizarA77/Faber/pkg/docs"
	"github.com/KhizarA77/Faber/pkg/memory"
)

// Agent is a specialized coding role. It produces a Brief; the host executes it.
type Agent interface {
	Meta() Meta
	BuildBrief(ctx context.Context, in Input, deps Deps) (Brief, error)
}

// Meta is the discoverable, routable description of an agent.
type Meta struct {
	Name        string
	Title       string
	Description string
	Tags        []string
	DocsFirst   bool
}

// Input is what the host passes when launching an agent.
type Input struct {
	Task      string            // what the user wants done
	Context   map[string]string // diff, target files, constraints
	Libraries []string          // external libs being used
}

// Brief is what the host receives and then executes itself.
type Brief struct {
	SystemPrompt string
	Instructions string
	DocPacks     []docs.Pack
	Tools        []string
	Policies     []Policy
}

// Deps is the injected toolbox an Agent uses when building a brief.
type Deps struct {
	Docs docs.Service
	Mem  memory.Store
}

// Policy is a machine-checkable contract carried by a Brief.
type Policy struct {
	Name string // e.g. "docs_first"
	Rule string // host-readable contract text
}
