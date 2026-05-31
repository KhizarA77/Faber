// Package orchestrator composes multiple agent briefs into a single coordination
// plan for the host to execute. Like the agents themselves, it runs no LLM - it
// assembles briefs and hands back an ordered plan.
package orchestrator

import (
	"context"
	"fmt"

	"github.com/KhizarA77/Faber/pkg/agent"
)

type Mode string

const (
	Sequential Mode = "sequential"
	Parallel   Mode = "parallel"
)

// Step routes one unit of work to a named agent.
type Step struct {
	Agent     string   // registry name, e.g. "architect"
	Task      string   // What this step should accomplish
	Libraries []string // external libs
}

// Plan is a multi-step request the host will execute.
type Plan struct {
	Mode  Mode
	Steps []Step
}

// StepBrief pairs a step with the assembled brief for its agent.
type StepBrief struct {
	Agent string
	Task  string
	Brief agent.Brief
}

// CompositeBrief is the fully-resolved plan returned to the host.
type CompositeBrief struct {
	Mode  Mode
	Steps []StepBrief
}

type Orchestrator struct {
	reg  *agent.Registry
	deps agent.Deps
}

// New builds an Orchestrator over the given registry and shared deps.
func New(reg *agent.Registry, deps agent.Deps) *Orchestrator {
	return &Orchestrator{reg: reg, deps: deps}
}

// Compose resolves every step's agent into a brief. It fails fast on an unknown
// agent or a brief error, so a bad plan never half-executes on the host.
func (o *Orchestrator) Compose(ctx context.Context, plan Plan) (CompositeBrief, error) {
	if len(plan.Steps) == 0 {
		return CompositeBrief{}, fmt.Errorf("orchestrate: plan has no steps")
	}
	mode := plan.Mode
	if mode == "" {
		mode = Sequential // sensible default
	}

	out := CompositeBrief{Mode: mode, Steps: make([]StepBrief, 0, len(plan.Steps))}
	for i, step := range plan.Steps {
		a, ok := o.reg.Get(step.Agent)
		if !ok {
			return CompositeBrief{}, fmt.Errorf("orchestrate: step %d: unknown agent %q", i, step.Agent)
		}
		brief, err := a.BuildBrief(ctx, agent.Input{Task: step.Task, Libraries: step.Libraries}, o.deps)
		if err != nil {
			return CompositeBrief{}, fmt.Errorf("orchestrate: step %d (%s): %w", i, step.Agent, err)
		}
		out.Steps = append(out.Steps, StepBrief{Agent: step.Agent, Task: step.Task, Brief: brief})
	}
	return out, nil
}
