package orchestrator

import (
	"context"
	"testing"

	"github.com/KhizarA77/Faber/pkg/agent"
)

// fakeAgent echoes its task into the brief so we can assert wiring.
type fakeAgent struct{ name string }

func (f fakeAgent) Meta() agent.Meta { return agent.Meta{Name: f.name} }
func (f fakeAgent) BuildBrief(_ context.Context, in agent.Input, _ agent.Deps) (agent.Brief, error) {
	return agent.Brief{SystemPrompt: "I am " + f.name, Instructions: in.Task}, nil
}

func newTestOrchestrator(names ...string) *Orchestrator {
	reg := agent.NewRegistry()
	for _, n := range names {
		reg.Register(fakeAgent{name: n})
	}
	return New(reg, agent.Deps{})
}

func TestCompose_OrderAndDefaults(t *testing.T) {
	o := newTestOrchestrator("architect", "code-reviewer")
	plan := Plan{Steps: []Step{
		{Agent: "architect", Task: "design"},
		{Agent: "code-reviewer", Task: "review"},
	}}

	got, err := o.Compose(context.Background(), plan)
	if err != nil {
		t.Fatalf("Compose returned error: %v", err)
	}
	if got.Mode != Sequential {
		t.Errorf("Mode = %q; want sequential (default)", got.Mode)
	}
	if len(got.Steps) != 2 {
		t.Fatalf("len(Steps) = %d; want 2", len(got.Steps))
	}
	if got.Steps[0].Agent != "architect" || got.Steps[1].Agent != "code-reviewer" {
		t.Errorf("step order not preserved: %q, %q", got.Steps[0].Agent, got.Steps[1].Agent)
	}
	// The step's task must flow into the agent's brief.
	if got.Steps[0].Brief.Instructions != "design" {
		t.Errorf("Instructions = %q; want design", got.Steps[0].Brief.Instructions)
	}
}

func TestCompose_UnknownAgent(t *testing.T) {
	o := newTestOrchestrator("architect")
	_, err := o.Compose(context.Background(), Plan{Steps: []Step{{Agent: "ghost", Task: "x"}}})
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestCompose_EmptyPlan(t *testing.T) {
	o := newTestOrchestrator("architect")
	_, err := o.Compose(context.Background(), Plan{})
	if err == nil {
		t.Fatal("expected error for empty plan")
	}
}
