package builtin

import (
	"context"

	"github.com/KhizarA77/Faber/pkg/agent"
	"github.com/KhizarA77/Faber/pkg/docs"
)

// Architect designs scalable, maintainable structure before implementation.
type Architect struct{}

var _ agent.Agent = Architect{}

// Meta implements agent.Agent.
func (Architect) Meta() agent.Meta {
	return agent.Meta{
		Name:        "architect",
		Title:       "Software Architect",
		Description: "Designs scalable, maintainable architecture and weighs trade-offs before any code is written.",
		Tags:        []string{"architecture", "design"},
		DocsFirst:   true,
	}
}

// BuildBrief implements agent.Agent.
func (Architect) BuildBrief(_ context.Context, _ agent.Input, _ agent.Deps) (agent.Brief, error) {
	return agent.Brief{
		SystemPrompt: "You are a pragmatic software architect. You design for " +
			"clarity, scalability, and maintainability, and you justify every " +
			"decision with explicit trade-offs.",
		Instructions: "1. Restate the goal and constraints from the input.\n" +
			"2. Propose a structure (packages/components) with clear responsibilities.\n" +
			"3. For any framework/library, follow the docs-first policy below before " +
			"committing to it: check existing codebase usage, then the official docs.\n" +
			"4. List trade-offs and the recommended option with reasoning.",
		Policies: []agent.Policy{{Name: "docs_first", Rule: docs.Directive}},
	}, nil
}
