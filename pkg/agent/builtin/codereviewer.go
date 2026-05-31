// Package builtin contains Faber's first-party agents.
package builtin

import (
	"context"

	"github.com/KhizarA77/Faber/pkg/agent"
)

const docsFirstRule = "When working with any external API or library, treat its " +
	"official documentation as the absolute source of truth. Verify every API " +
	"call, signature, and behavior against the provided DocPacks or by calling " +
	"consult_docs before asserting anything — never rely on prior assumptions."

// CodeReviewer reviews diffs for correctness, security and scalability.
type CodeReviewer struct{}

var _ agent.Agent = CodeReviewer{}

// Meta implements agent.Agent.
func (CodeReviewer) Meta() agent.Meta {
	return agent.Meta{
		Name:        "code-reviewer",
		Title:       "Code Reviewer",
		Description: "Reviews diffs for correctness, security, and scalable design, verifying external API usage against official docs.",
		Tags:        []string{"review", "quality"},
		DocsFirst:   true,
	}
}

// BuildBrief implements agent.Agent
func (CodeReviewer) BuildBrief(ctx context.Context, in agent.Input, deps agent.Deps) (agent.Brief, error) {
	brief := agent.Brief{
		SystemPrompt: "You are a meticulous senior code reviewer. You focus on " +
			"correctness, security, error handling, and whether the code will " +
			"scale. You are direct and specific, citing exact lines.",
		Instructions: "1. Read the diff/files in Input.Context.\n" +
			"2. For any external library or API, consult its docs before judging usage.\n" +
			"3. Report issues grouped by severity (blocker, major, minor) with fixes.\n" +
			"4. Call out anything that won't scale or violates best practices.",
		Tools:    []string{"consult_docs"},
		Policies: []agent.Policy{{Name: "docs_first", Rule: docsFirstRule}},
	}

	// Docs-as-source-of-truth: pre-fetch authoritative docs for libraries in play.
	// The nil guard lets this work in M1 before the real docs.Service exists.
	if deps.Docs != nil && len(in.Libraries) > 0 {
		packs, err := deps.Docs.PrefetchAll(ctx, in.Libraries)
		if err != nil {
			return agent.Brief{}, err
		}
		brief.DocPacks = packs
	}

	return brief, nil
}
