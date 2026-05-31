# Faber — Architecture

> **Faber** (Latin: *craftsman, maker*) is an agent-orchestration library written in Go.
> It gives any MCP-capable coding IDE — Claude Code, Codex, Cursor — the ability to
> launch specialized coding agents (code review, architecture, scalability) that follow
> good engineering discipline, most importantly treating *official documentation as the
> source of truth* when working with external APIs and libraries.

> **Status:** M0–M2 implemented (MCP server, agent registry, two built-in agents,
> docs-first policy). Orchestration and the Claude Code plugin generator are planned —
> see the [roadmap](#13-roadmap).

---

## 1. Vision & scope

Modern coding IDEs already have a capable model in the loop. What they lack is a
**reusable, IDE-agnostic library of specialists** and a **discipline layer** that enforces
good engineering practice — above all, grounding external-API code in current official
documentation rather than the model's training memory.

Faber provides exactly that, and no more than it needs to:

- A registry of **specialized coding agents** defined in Go.
- An **MCP server** so any MCP host can discover and launch those agents.
- A **docs-first policy** that mandates agents verify external APIs against official docs
  using the host's own tools.
- A **Claude Code plugin generator** for native ergonomics on Claude Code (planned).

### Non-goals (v0)

- Faber does **not** run LLM inference, hold API keys, or manage providers. The host
  IDE's own model does the thinking. (See §3, *Brief & delegate*.)
- Faber does **not** fetch, parse, or cache documentation. It delegates that to the host's
  existing web and codebase search tools. (See §4.)
- No daemon, background workers, or swarm consensus. Faber is a synchronous server that
  answers tool calls.

These are deliberate. The smaller the surface, the faster Faber ships and the easier it is
to trust in an open-source supply chain — its only external dependency is the Go MCP SDK.

---

## 2. The four foundational decisions

Every design choice below follows from four early decisions:

| Decision | Choice | Consequence |
|---|---|---|
| Who runs inference? | **Brief & delegate** — host runs it | No provider layer, no API keys, no agent loop in Go |
| Integration surface | **MCP server + Claude Code plugin** | One core protocol + generated native ergonomics |
| Docs strategy | **Pure delegation (policy)** | No fetcher/cache/URLs; the host reads docs with its own tools |
| Agent definition | **Go-coded agents** | Type-safe, compiler-checked, developer-extensible |

---

## 3. Core model — "Brief & delegate"

Faber agents do **not** call an LLM. An agent is a Go type whose job is to **assemble a
brief**: a persona, an ordered set of instructions, and the policies it must follow (notably
docs-first). The host IDE's model then *adopts* that brief and executes the work itself,
using its own tools.

```
Claude Code  ──calls──▶  faber_launch_agent(role="code-reviewer", task, libraries)
                              │
                              ▼
                    Agent.BuildBrief() ──▶ assembles persona + instructions + policies
                              │
        returns Brief{ SystemPrompt, Instructions, Tools, Policies }
                              │
   Host model adopts the persona and carries out the task with its own tools,
   honoring the docs-first policy (grep the codebase, then read official docs).
```

**Why this is the right v0:** it is free (uses the user's existing IDE subscription), needs
no secrets, and keeps Faber tiny. The trade-off — Faber cannot run a fully autonomous
multi-agent loop by itself — is acceptable now and is the explicit upgrade path in §11.

---

## 4. Documentation as the source of truth (pure delegation)

This is Faber's signature behavior. The key realization: **the host IDE already has
excellent search and fetch tools** (web search, web fetch, codebase grep/glob). Faber does
not duplicate them — it would do the host's job worse. Instead, Faber's job is purely to
**enforce the discipline**.

Every agent whose `Meta.DocsFirst` is true carries the `docs.Directive` policy in its brief.
The directive mandates a strict source-of-truth order:

1. **Codebase first.** Use the host's grep/glob to find existing usage and follow the
   project's established patterns.
2. **Then official docs.** Use the host's web search/fetch to read the official
   documentation for the libraries involved.
3. **Docs win conflicts.** When prior knowledge or assumptions disagree with the official
   docs, the docs are authoritative. Never rely on training memory for a verifiable API.

Faber performs **no network I/O** and ships **no documentation data** — the directive is a
single constant (`docs.Directive`), and the host does all the reading. "Docs" here means
documentation for *any* ecosystem (npm, PyPI, crates, Go, GitHub, …); the host's own search
locates it.

---

## 5. Integration surfaces

### 5a. MCP server (canonical, implemented)

Faber runs as a Model Context Protocol server over stdio, built on the official Go SDK
(`github.com/modelcontextprotocol/go-sdk`, v1.6.x). Any MCP host registers it:

```
claude mcp add faber -- faber mcp start
```

Tools appear to the host as `mcp__faber__list_agents`, `mcp__faber__launch_agent`, etc.

### 5b. Claude Code plugin (generated, planned)

`faber init --claude` will read the agent registry and **generate** native Claude Code
artifacts from the same Go definitions — a subagent file per agent (persona + docs-first
contract) and a plugin manifest. One source of truth (the Go registry) → two delivery
mechanisms, no drift.

---

## 6. Package layout

```
faber/
├── cmd/faber/                 # CLI entrypoint: `faber mcp start`
├── internal/mcp/              # MCP server: tool registration, stdio transport
├── pkg/agent/
│   ├── agent.go               # Agent interface; Meta, Input, Brief, Deps, Policy
│   ├── registry.go            # Register / Get / List (concurrency-safe)
│   └── builtin/               # code-reviewer, architect
├── pkg/docs/                  # docs-first policy: the Directive constant
├── pkg/orchestrator/          # multi-agent coordination (planned, M3)
├── pkg/memory/                # session-scoped shared store (MapStore)
└── go.mod                     # module github.com/KhizarA77/Faber
```

**Dependency direction:** `cmd → internal/mcp → pkg/*`. `pkg/agent` depends on `pkg/docs`
(for the directive) and `pkg/memory` (via the `Deps` seam). Agents are pure functions of
their input, so they unit-test without a network or a running server.

---

## 7. Core types (`pkg/agent`)

The entire agent contract is one small interface plus its data types.

```go
// Agent assembles a brief for the host to execute. It never calls an LLM.
type Agent interface {
    Meta() Meta
    BuildBrief(ctx context.Context, in Input, deps Deps) (Brief, error)
}

// Meta is the discoverable, routable description of an agent.
type Meta struct {
    Name, Title, Description string
    Tags      []string
    DocsFirst bool        // carry the docs-as-truth policy in the brief
}

// Input is what the host passes when launching an agent.
type Input struct {
    Task      string
    Context   map[string]string // diff, target files, constraints…
    Libraries []string          // external libs in play (hint to the host)
}

// Brief is what the host receives and then executes itself.
type Brief struct {
    SystemPrompt string   // persona + rules the host adopts
    Instructions string   // ordered steps to follow
    Tools        []string // host tools to favor (optional)
    Policies     []Policy  // machine-checkable contracts, e.g. docs_first
}

// Deps is the injected toolbox an Agent uses when building a brief — the seam
// for future subsystems (e.g. shared memory during orchestration).
type Deps struct {
    Mem memory.Store
}

type Policy struct {
    Name string `json:"name"` // e.g. "docs_first"
    Rule string `json:"rule"` // host-readable contract text
}
```

A built-in agent is trivial and fully type-checked:

```go
type CodeReviewer struct{}

var _ agent.Agent = CodeReviewer{} // compile-time interface proof

func (CodeReviewer) Meta() agent.Meta { /* … DocsFirst: true … */ }

func (CodeReviewer) BuildBrief(_ context.Context, _ agent.Input, _ agent.Deps) (agent.Brief, error) {
    return agent.Brief{
        SystemPrompt: reviewerPersona,
        Instructions: reviewerSteps,
        Policies:     []agent.Policy{{Name: "docs_first", Rule: docs.Directive}},
    }, nil
}
```

### Registry

```go
type Registry struct { /* RWMutex + name → Agent */ }

func (r *Registry) Register(a Agent)            // last write wins (override built-ins)
func (r *Registry) Get(name string) (Agent, bool)
func (r *Registry) List() []Meta               // sorted by Name for stable output
```

Built-ins are registered at startup. Third parties extend Faber by implementing `Agent` and
calling `Register`.

---

## 8. Docs policy (`pkg/docs`)

The whole package is a single exported constant plus its doc comment:

```go
// Directive is the standing instruction Faber attaches to docs-first work.
const Directive = "Ground all external API/library usage in the source of truth, " +
    "in this order: (1) FIRST use your codebase search tools … (2) THEN use your web " +
    "search/fetch tools to read the official documentation … (3) treat the official docs " +
    "as the ABSOLUTE source of truth …"
```

Agents reference `docs.Directive` so the wording lives in exactly one place and can never
drift between agents. Kept ASCII-only because it travels over the wire as protocol payload.

---

## 9. Memory (`pkg/memory`)

A minimal session-scoped store so steps in a future orchestration can pass context.

```go
type Store interface {
    Get(key string) (string, bool)
    Set(key, value string)
    Namespace(ns string) Store
}
```

`MapStore` implements it with an `RWMutex`; `Namespace` returns a key-prefixed view that
shares the same underlying map and lock. A vector-backed implementation can satisfy the same
interface later without touching agents.

---

## 10. Orchestrator (`pkg/orchestrator`, planned — M3)

In the brief-and-delegate model, "orchestration" means **composing multiple agent briefs
into one coordination plan** the host executes step by step (sequential / parallel /
pipeline), with shared state flowing through `pkg/memory`. Exposed as `faber_orchestrate`.

---

## 11. MCP tool surface

| Tool | Status | Returns |
|---|---|---|
| `faber_list_agents` | implemented | `[]Meta` for host routing |
| `faber_launch_agent` | implemented | `Brief` for `role` + `task` |
| `faber_orchestrate` | planned (M3) | a multi-step coordination plan |

Note: there is **no** `faber_consult_docs` tool — docs reading is delegated to the host's own
tools by policy, not performed by Faber.

---

## 12. Design principles

- **Tiny trusted core.** One protocol (MCP), one external dependency (the Go MCP SDK).
- **One source of truth, many surfaces.** The Go registry will generate the Claude Code
  plugin; nothing is authored twice.
- **Delegate, don't duplicate.** The host already has search/fetch/grep — Faber enforces
  *that they're used*, rather than reimplementing them.
- **Interfaces at the seams.** `memory.Store` and `Agent` are interfaces so implementations
  swap (RAG memory, a self-hosted runtime) without churn.
- **Compiler-checked specialists.** Go-coded agents catch mistakes at build time.

---

## 13. Roadmap

| Milestone | Deliverable | Status |
|---|---|---|
| **M0** | MCP server + `faber_list_agents` | ✅ done |
| **M1** | `Agent` interface, registry, `code-reviewer` + `architect`, `faber_launch_agent` | ✅ done |
| **M2** | Docs-first policy (`docs.Directive`) baked into briefs | ✅ done |
| **Tests** | Unit tests for registry, memory, and built-in briefs | ▶ in progress |
| **M3** | Orchestrator + `faber_orchestrate` | planned |
| **M4** | `faber init --claude` generates subagent files + plugin manifest | planned |
| **M5** | More built-in agents, examples, polish | planned |

### Future: optional self-contained runtime ("Model B")

A later `pkg/runtime` can implement a provider-agnostic agent loop so Faber can *execute*
briefs itself (true autonomous swarms, IDE-independent). Because agents already return
`Brief`s through a stable interface, this is additive — the same agents work in both modes.

---

## 14. Glossary

- **Brief** — the persona + instructions + policies an agent hands to the host.
- **Host** — the IDE/model that executes a brief (Claude Code, Codex, …).
- **Directive** — the single docs-first policy string agents embed in their briefs.
- **Docs-first** — the policy that official documentation overrides any prior model belief.
