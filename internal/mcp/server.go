// Package mcp exposes the Faber agent registry to MCP hosts (Claude Code, Codex)
// as callable tools.
package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/KhizarA77/Faber/pkg/agent"
	"github.com/KhizarA77/Faber/pkg/orchestrator"
)

// Server wraps an MCP server with Faber's agent registry and shared deps.
type Server struct {
	reg  *agent.Registry
	deps agent.Deps
	orch *orchestrator.Orchestrator
	mcp  *mcpsdk.Server
}

// NewServer builds a Faber MCP server and registers its tools.
func NewServer(reg *agent.Registry, deps agent.Deps) *Server {
	s := &Server{
		reg:  reg,
		deps: deps,
		orch: orchestrator.New(reg, deps),
		mcp:  mcpsdk.NewServer(&mcpsdk.Implementation{Name: "faber", Version: "0.1.0"}, nil),
	}
	s.registerTools()
	return s
}

// Run serves over stdio until the client disconnects or ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	return s.mcp.Run(ctx, &mcpsdk.StdioTransport{})
}

func (s *Server) registerTools() {
	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "faber_list_agents",
		Description: "List the specialized Faber agents available to launch.",
	}, s.handleListAgents)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "faber_launch_agent",
		Description: "Launch a specialized Faber agent. Returns a brief — persona, " +
			"instructions, and policies — for you to execute.",
	}, s.handleLaunchAgent)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "faber_orchestrate",
		Description: "Compose a multi-step plan across specialized agents. Returns each " +
			"step's brief plus the execution mode (sequential or parallel) for you to carry out.",
	}, s.handleOrchestrate)

}

// --- faber_list_agents ---

type listAgentsInput struct{} // no arguments

type agentInfo struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	DocsFirst   bool     `json:"docsFirst"`
}

type listAgentsOutput struct {
	Agents []agentInfo `json:"agents"`
}

func (s *Server) handleListAgents(_ context.Context, _ *mcpsdk.CallToolRequest, _ listAgentsInput) (*mcpsdk.CallToolResult, listAgentsOutput, error) {
	metas := s.reg.List()
	out := listAgentsOutput{Agents: make([]agentInfo, 0, len(metas))}
	for _, m := range metas {
		out.Agents = append(out.Agents, agentInfo{
			Name: m.Name, Title: m.Title, Description: m.Description,
			Tags: m.Tags, DocsFirst: m.DocsFirst,
		})
	}
	return nil, out, nil
}

// --- faber_launch_agent ---

type launchAgentInput struct {
	Role      string            `json:"role" jsonschema:"the agent to launch, e.g. code-reviewer or architect"`
	Task      string            `json:"task" jsonschema:"what you want the agent to do"`
	Context   map[string]string `json:"context,omitempty" jsonschema:"optional extra context such as a diff or target files"`
	Libraries []string          `json:"libraries,omitempty" jsonschema:"external libraries in play, to pre-fetch their official docs"`
}

type launchAgentOutput struct {
	SystemPrompt string         `json:"systemPrompt"`
	Instructions string         `json:"instructions"`
	Tools        []string       `json:"tools,omitempty"`
	Policies     []agent.Policy `json:"policies,omitempty"`
}

func (s *Server) handleLaunchAgent(ctx context.Context, _ *mcpsdk.CallToolRequest, in launchAgentInput) (*mcpsdk.CallToolResult, launchAgentOutput, error) {
	a, ok := s.reg.Get(in.Role)
	if !ok {
		return nil, launchAgentOutput{}, fmt.Errorf("unknown agent %q", in.Role)
	}
	brief, err := a.BuildBrief(ctx, agent.Input{
		Task:      in.Task,
		Context:   in.Context,
		Libraries: in.Libraries,
	}, s.deps)
	if err != nil {
		return nil, launchAgentOutput{}, err
	}
	return nil, launchAgentOutput{
		SystemPrompt: brief.SystemPrompt,
		Instructions: brief.Instructions,
		Tools:        brief.Tools,
		Policies:     brief.Policies,
	}, nil
}

// --- faber_orchestrate ---

type orchestrateStepInput struct {
	Agent     string   `json:"agent" jsonschema:"the agent for this step, e.g. architect"`
	Task      string   `json:"task" jsonschema:"what this step should accomplish"`
	Libraries []string `json:"libraries,omitempty" jsonschema:"external libraries in play for this step"`
}

type orchestrateInput struct {
	Mode  string                 `json:"mode,omitempty" jsonschema:"how to run the steps: sequential (default) or parallel"`
	Steps []orchestrateStepInput `json:"steps" jsonschema:"the ordered steps to run"`
}

type orchestrateStepOutput struct {
	Agent string      `json:"agent"`
	Task  string      `json:"task"`
	Brief agent.Brief `json:"brief"`
}

type orchestrateOutput struct {
	Mode  string                  `json:"mode"`
	Steps []orchestrateStepOutput `json:"steps"`
}

func (s *Server) handleOrchestrate(ctx context.Context, _ *mcpsdk.CallToolRequest, in orchestrateInput) (*mcpsdk.CallToolResult, orchestrateOutput, error) {
	steps := make([]orchestrator.Step, 0, len(in.Steps))
	for _, st := range in.Steps {
		steps = append(steps, orchestrator.Step{Agent: st.Agent, Task: st.Task, Libraries: st.Libraries})
	}

	composite, err := s.orch.Compose(ctx, orchestrator.Plan{
		Mode:  orchestrator.Mode(in.Mode),
		Steps: steps,
	})
	if err != nil {
		return nil, orchestrateOutput{}, err
	}

	out := orchestrateOutput{Mode: string(composite.Mode), Steps: make([]orchestrateStepOutput, 0, len(composite.Steps))}
	for _, st := range composite.Steps {
		out.Steps = append(out.Steps, orchestrateStepOutput{Agent: st.Agent, Task: st.Task, Brief: st.Brief})
	}
	return nil, out, nil
}
