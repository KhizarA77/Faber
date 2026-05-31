// Package builtin contains Faber's first-party agents.
package builtin

import (
	"context"

	"github.com/KhizarA77/Faber/pkg/agent"
	"github.com/KhizarA77/Faber/pkg/docs"
)

// CodeReviewer reviews diffs for correctness, security, and scalability.
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

// BuildBrief implements agent.Agent.
func (CodeReviewer) BuildBrief(_ context.Context, _ agent.Input, _ agent.Deps) (agent.Brief, error) {
	return agent.Brief{
		SystemPrompt: "You are a meticulous senior code reviewer. You focus on " +
			"correctness, security, error handling, and whether the code will " +
			"scale. You are direct and specific, citing exact lines.",
		Instructions: "1. Read the diff/files provided in the input context.\n" +
			"2. For any external library or API, follow the docs-first policy below: search the " +
			"codebase for existing usage, then verify against the official docs.\n" +
			"3. Report issues grouped by severity (blocker, major, minor) with concrete fixes.\n" +
			"4. Call out anything that won't scale or violates best practices.",
		Policies: []agent.Policy{{Name: "docs_first", Rule: docs.Directive}},
	}, nil
}
