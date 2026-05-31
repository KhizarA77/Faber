# Faber — Architecture

> **Faber** (Latin: *craftsman, maker*) is an agent‑orchestration library written in Go.
> It gives any MCP‑capable coding IDE — Claude Code, Codex, Cursor — the ability to
> launch specialized coding agents (code review, architecture, scalability, docs research)
> through a single, dependency‑light server.

---

## 1. Vision & scope

Modern coding IDEs already have a capable model in the loop. What they lack is a
**reusable, IDE‑agnostic library of specialists** and a **discipline layer** that forces
good engineering practice — most importantly, treating *official documentation as the
single source of truth* whenever external APIs or libraries are involved.

Faber provides exactly that, and nothing more than it needs to:

- A registry of **specialized coding agents** defined in Go.
- An **MCP server** so any MCP host can discover and launch those agents.
- A **docs‑as‑source‑of‑truth** subsystem that fetches and caches authoritative
  documentation and injects it into agent briefs.
- A **Claude Code plugin generator** for native, first‑class ergonomics on Claude Code.

### Non‑goals (for v0)

- Faber does **not** run LLM inference, hold API keys, or manage providers. The host
  IDE's own model does the thinking. (See §3, *Brief & delegate*.)
- No vector database, embeddings, or RAG in v0 — docs are fetched and cached, not
  semantically indexed. (Reserved for a later milestone.)
- No daemon, background workers, or swarm consensus. Faber is a synchronous server
  that answers tool calls.

These are deliberate. The smaller the surface, the faster Faber ships and the easier
it is to trust in an open‑source supply chain.

---

## 2. The four foundational decisions

Every design choice below follows from four early decisions:

| Decision | Choice | Consequence |
|---|---|---|
| Who runs inference? | **Brief & delegate** — host runs it | No provider layer, no API keys, no agent loop in Go |
| Integration surface | **MCP server + Claude Code plugin** | One core protocol + generated native ergonomics |
| Docs engine depth | **Fetch + cache** | Simple, fast, no embeddings infra |
| Agent definition | **Go‑coded agents** | Type‑safe, compiler‑checked, developer‑extensible |

---

## 3. Core model — "Brief & delegate"

Faber agents do **not** call an LLM. An agent is a Go type whose job is to **assemble a
brief**: a persona, an ordered set of instructions, the authoritative documentation it
pre‑fetched, and the Faber tools the host should use. The host IDE's model then *adopts*
that brief and executes the work itself.

```
Claude Code  ──calls──▶  faber_launch_agent(role="code-reviewer", task, libraries=["pgx"])
                              │
                              ▼
                    Agent.BuildBrief(ctx, in, deps)
                              │  └─ deps.Docs.PrefetchAll(ctx, libraries)  ← docs-as-truth
                              ▼
        returns Brief{ SystemPrompt, Instructions, DocPacks, Tools, Policies }
                              │
   Host model adopts the persona — real docs already in its context — and may call
   faber_consult_docs on demand for anything it still needs to verify.
```

**Why this is the right v0:** it is free (uses the user's existing IDE subscription),
needs no secrets, and keeps Faber tiny. The trade‑off — Faber cannot run a fully
autonomous multi‑agent loop by itself — is acceptable now and is the explicit upgrade
path in §11.

---

## 4. Documentation as the source of truth

This is Faber's signature behavior, and it is enforced by **construction**, not just by
asking the model nicely.

1. **Pre‑fetch & inject.** When an agent declares `DocsFirst: true`, `BuildBrief`
   resolves every library in `Input.Libraries` to its canonical docs, fetches the
   relevant pages, and embeds them in the brief as `DocPacks`. The host model therefore
   reasons over *fresh, real documentation already in context* rather than stale training
   memory.
2. **On‑demand verification.** The brief grants the host the `faber_consult_docs` tool
   and instructs it to call that tool before asserting any external API behavior it has
   not already seen in a `DocPack`.
3. **Policy contracts.** Each brief carries machine‑checkable `Policy` entries (e.g.
   `docs_first`). In the Claude Code plugin path these can later be wired to a **hook**
   that hard‑blocks unverified external‑API claims (post‑v0).

In the brief‑and‑delegate model we cannot *force* the host's model, so the strongest
available lever is **getting the real docs into its context window** — which is exactly
what step 1 does.

---

## 5. Integration surfaces

### 5a. MCP server (canonical)

Faber runs as a Model Context Protocol server over stdio (and later HTTP), built on the
official Go SDK (`github.com/modelcontextprotocol/go-sdk`). Any MCP host registers it:

```
claude mcp add faber -- faber mcp start
```

Tools then appear to the host as `mcp__faber__list_agents`, `mcp__faber__launch_agent`, etc.

### 5b. Claude Code plugin (generated)

`faber init --claude` reads the agent registry and **generates** native Claude Code
artifacts from the same Go definitions:

- `.claude/agents/<name>.md` — a subagent file per registered agent, with its persona,
  the docs‑first contract, and a pointer to `faber_consult_docs`.
- Slash commands / a plugin manifest for marketplace install.

One source of truth (the Go registry) → two delivery mechanisms. No drift.

---

## 6. Package layout

```
faber/
├── cmd/faber/                 # CLI entrypoint: `faber mcp start`, `faber init --claude`
├── internal/mcp/              # MCP server: tool registration, transports (stdio→http)
├── pkg/agent/
│   ├── agent.go               # Agent interface; Meta, Input, Brief, Deps, Policy types
│   ├── registry.go            # Register / Get / List
│   └── builtin/               # code-reviewer, architect, scalability, docs-researcher…
├── pkg/docs/                  # ★ resolver → fetcher → cache (fetch+cache in v0)
├── pkg/orchestrator/          # sequential / parallel composition of briefs
├── pkg/memory/                # KV shared store across a session (vector later)
├── plugin/                    # Claude Code plugin/file generator
└── go.mod                     # module github.com/KhizarA77/Faber
```

**Dependency direction:** `cmd → internal/mcp → pkg/*`. `pkg/agent` depends on
`pkg/docs` and `pkg/memory` only through the `Deps` struct (dependency injection), so
agents stay testable without a live network or server.

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
    Name        string   // stable id, e.g. "code-reviewer"
    Title       string   // human label, e.g. "Code Reviewer"
    Description string   // shown to the host for routing
    Tags        []string // "review", "architecture", …
    DocsFirst   bool     // enforce documentation-as-truth
}

// Input is what the host passes when launching an agent.
type Input struct {
    Task      string            // what the user wants done
    Context   map[string]string // diff, target files, constraints…
    Libraries []string          // external libs in play → triggers doc pre-fetch
}

// Brief is what the host receives and then executes itself.
type Brief struct {
    SystemPrompt string       // persona + rules the host adopts
    Instructions string       // ordered steps to follow
    DocPacks     []docs.Pack  // authoritative excerpts, already fetched
    Tools        []string     // faber tools the host should use
    Policies     []Policy     // machine-checkable contracts
}

// Deps is the injected toolbox an agent uses while building a brief.
type Deps struct {
    Docs docs.Service
    Mem  memory.Store
}

type Policy struct {
    Name string // e.g. "docs_first"
    Rule string // human/host-readable contract text
}
```

A built‑in agent is then trivial and fully type‑checked:

```go
// pkg/agent/builtin/codereviewer.go
type CodeReviewer struct{}

func (CodeReviewer) Meta() agent.Meta {
    return agent.Meta{
        Name: "code-reviewer", Title: "Code Reviewer",
        Description: "Reviews diffs for correctness, security, and scalability",
        Tags:        []string{"review", "quality"},
        DocsFirst:   true,
    }
}

func (CodeReviewer) BuildBrief(ctx context.Context, in agent.Input, d agent.Deps) (agent.Brief, error) {
    packs, err := d.Docs.PrefetchAll(ctx, in.Libraries) // docs-as-truth
    if err != nil {
        return agent.Brief{}, err
    }
    return agent.Brief{
        SystemPrompt: reviewerPersona,
        Instructions: reviewerSteps,
        DocPacks:     packs,
        Tools:        []string{"consult_docs", "read_file"},
        Policies:     []agent.Policy{{Name: "docs_first", Rule: docsFirstRule}},
    }, nil
}
```

### Registry

```go
type Registry struct { /* name → Agent */ }

func (r *Registry) Register(a Agent)
func (r *Registry) Get(name string) (Agent, bool)
func (r *Registry) List() []Meta
```

Built‑ins are registered at startup. Third parties extend Faber by implementing `Agent`
and calling `Register`.

---

## 8. Docs subsystem (`pkg/docs`)

Three stages behind one façade, `docs.Service`:

```go
type Service interface {
    Consult(ctx context.Context, lib, version, query string) (Pack, error)
    PrefetchAll(ctx context.Context, libs []string) ([]Pack, error)
}

type Resolver interface { Resolve(lib, version string) (Source, error) } // lib → canonical docs location
type Fetcher  interface { Fetch(ctx context.Context, src Source, query string) (Pack, error) }
type Cache    interface { Get(key string) (Pack, bool); Set(key string, p Pack) }

type Pack struct {
    Library, Version, URL string
    Excerpts  []Excerpt
    FetchedAt time.Time
}
```

- **Resolver (v0):** a built‑in map of ecosystem → docs base URL (Go → `pkg.go.dev`,
  Rust → `docs.rs`, plus official sites), with sensible fallbacks. Extensible.
- **Fetcher (v0):** HTTP GET, HTML→text, naive query‑relevant section selection.
- **Cache (v0):** on‑disk, keyed by `lib@version#query`, with a freshness TTL.

`PrefetchAll` is what makes `DocsFirst` real — it is called inside `BuildBrief` and its
output lands directly in `Brief.DocPacks`.

---

## 9. Orchestrator (`pkg/orchestrator`)

In the brief‑and‑delegate model, "orchestration" means **composing multiple agent briefs
into one coordination plan** the host executes step by step.

```go
type Mode int
const ( Sequential Mode = iota; Parallel; Pipeline )

type Step struct {
    Agent     string   // registry name
    Task      string
    DependsOn []string // step ids
}

type Plan struct { Steps []Step; Mode Mode }

func (o *Orchestrator) Compose(ctx context.Context, plan Plan, in agent.Input) (CompositeBrief, error)
```

`faber_orchestrate` takes a `Plan`, builds each step's `Brief`, and returns a single
`CompositeBrief` describing the order and hand‑offs. The host carries it out; shared
state flows through `pkg/memory`.

---

## 10. Memory (`pkg/memory`)

A minimal session‑scoped store so steps in an orchestration can pass context.

```go
type Store interface {
    Get(key string) (string, bool)
    Set(key, value string)
    Namespace(ns string) Store
}
```

v0 is an in‑memory KV with namespacing. A vector‑backed implementation can satisfy the
same interface later without touching agents.

---

## 11. MCP tool surface (v0)

| Tool | Input | Returns |
|---|---|---|
| `faber_list_agents` | — | `[]Meta` for host routing |
| `faber_launch_agent` | `role`, `task`, `context?`, `libraries?` | `Brief` |
| `faber_consult_docs` | `library`, `query`, `version?` | `Pack` |
| `faber_orchestrate` | `Plan` | `CompositeBrief` |

---

## 12. Design principles

- **Tiny trusted core.** One protocol (MCP), one external dep (the Go MCP SDK) in v0.
- **One source of truth, many surfaces.** The Go registry generates the Claude Code
  plugin; nothing is authored twice.
- **Injection over instruction.** Enforce docs‑first by putting real docs in context, not
  by hoping the model complies.
- **Interfaces at the seams.** `docs.Service`, `memory.Store`, and `Agent` are interfaces
  so v0 implementations can be swapped (RAG memory, self‑hosted runtime) without churn.
- **Compiler‑checked specialists.** Go‑coded agents catch mistakes at build time.

---

## 13. Roadmap

| Milestone | Deliverable |
|---|---|
| **M0** | MCP server + `faber_list_agents` — proves the IDE handshake end‑to‑end |
| **M1** | `Agent` interface, registry, `code-reviewer` + `architect` built‑ins, `faber_launch_agent` |
| **M2** | Docs subsystem (resolver→fetcher→cache) + `faber_consult_docs`, wired into `DocsFirst` |
| **M3** | Orchestrator + `faber_orchestrate` |
| **M4** | `faber init --claude` generates subagent files + plugin manifest |
| **M5** | More built‑ins, tests, README, examples |

### Future: optional self‑contained runtime ("Model B")

A later `pkg/runtime` can implement a provider‑agnostic agent loop so Faber can *execute*
briefs itself (true autonomous swarms, IDE‑independent). Because agents already return
`Brief`s through a stable interface, this is additive — the same agents work in both modes.

---

## 14. Glossary

- **Brief** — the persona + instructions + docs + tools an agent hands to the host.
- **Host** — the IDE/model that executes a brief (Claude Code, Codex, …).
- **DocPack** — a bundle of authoritative documentation excerpts for one library.
- **Docs‑first** — the policy that official documentation overrides any prior model belief.
