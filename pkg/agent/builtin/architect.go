package builtin

import (
	"context"

	"github.com/KhizarA77/Faber/pkg/agent"
)

type Architect struct{}

var _ agent.Agent = Architect{}

func (Architect) Meta() agent.Meta {
	return agent.Meta{
		Name:        "architect",
		Title:       "Software Architect",
		Description: "Designs scalable, maintainable architecture and weighs trade-offs before any code is written.",
		Tags:        []string{"architecture", "design"},
		DocsFirst:   true,
	}
}

func (Architect) BuildBrief(ctx context.Context, in agent.Input, deps agent.Deps) (agent.Brief, error) {
	brief := agent.Brief{
		SystemPrompt: "You are a pragmatic software architect. You design for " +
			"clarity, scalability, and maintainability, and you justify every " +
			"decision with explicit trade-offs.",
		Instructions: "1. Restate the goal and constraints from Input.\n" +
			"2. Propose a structure (packages/components) with responsibilities.\n" +
			"3. For any framework/library, ground choices in its official docs.\n" +
			"4. List trade-offs and the recommended option with reasoning.",
		Tools:    []string{"consult_docs"},
		Policies: []agent.Policy{{Name: "docs_first", Rule: docsFirstRule}},
	}

	if deps.Docs != nil && len(in.Libraries) > 0 {
		packs, err := deps.Docs.PrefetchAll(ctx, in.Libraries)
		if err != nil {
			return agent.Brief{}, err
		}
		brief.DocPacks = packs
	}

	return brief, nil
}
