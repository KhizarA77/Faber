package builtin

import (
	"context"
	"strings"
	"testing"

	"github.com/KhizarA77/Faber/pkg/agent"
	"github.com/KhizarA77/Faber/pkg/docs"
)

func TestBuiltins_BriefAndDocsFirst(t *testing.T) {
	cases := []struct {
		name  string
		agent agent.Agent
	}{
		{"code-reviewer", CodeReviewer{}},
		{"architect", Architect{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.agent.Meta()
			if m.Name != tc.name {
				t.Errorf("Meta().Name = %q; want %q", m.Name, tc.name)
			}
			if !m.DocsFirst {
				t.Error("expected DocsFirst = true")
			}

			brief, err := tc.agent.BuildBrief(context.Background(), agent.Input{}, agent.Deps{})
			if err != nil {
				t.Fatalf("BuildBrief returned error: %v", err)
			}
			if strings.TrimSpace(brief.SystemPrompt) == "" {
				t.Error("SystemPrompt is empty")
			}
			if strings.TrimSpace(brief.Instructions) == "" {
				t.Error("Instructions are empty")
			}

			// The docs-first policy must be present and carry the canonical directive.
			var found bool
			for _, p := range brief.Policies {
				if p.Name == "docs_first" {
					found = true
					if p.Rule != docs.Directive {
						t.Error("docs_first rule does not match docs.Directive")
					}
				}
			}
			if !found {
				t.Error("brief is missing the docs_first policy")
			}
		})
	}
}
