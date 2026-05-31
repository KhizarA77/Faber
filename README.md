# Faber

> **Faber** (Latin: *craftsman, maker*) is an agent-orchestration library in Go that
> gives any MCP-capable coding IDE — Claude Code, Codex, Cursor — a set of specialized
> coding agents, with **official documentation treated as the absolute source of truth.**

Faber runs as a [Model Context Protocol](https://modelcontextprotocol.io) server. Your
IDE connects to it and gains tools to discover and launch specialists for **code review**,
**architecture**, and other engineering disciplines — each one designed to verify external
API and library usage against the real docs instead of relying on stale model memory.

> ⚠️ **Status: early (M1).** The core — MCP server, agent registry, and two built-in
> agents — works today. The live docs engine, orchestration, and Claude Code plugin
> generation are on the [roadmap](ARCHITECTURE.md#13-roadmap).

## Why Faber

- **Specialized agents, not one generalist.** Launch a code reviewer, an architect, and
  more — each with a focused persona and policies.
- **Docs are the source of truth.** Agents marked *docs-first* pre-fetch authoritative
  documentation and inject it into their brief, so the IDE reasons over real docs.
- **IDE-agnostic.** One MCP server works across Claude Code, Codex, and any MCP host.
- **Tiny, trusted core.** Written in Go with a single external dependency (the official
  Go MCP SDK). No API keys required — your IDE's own model does the work.

## How it works

Faber uses a **brief & delegate** model. When the IDE launches an agent, Faber assembles a
*brief* — persona, ordered instructions, pre-fetched docs, and policies — and hands it back.
The IDE's own model adopts that brief and executes the task, calling `faber_consult_docs`
for anything it still needs to verify.

```
IDE ──▶ faber_launch_agent(role="code-reviewer", task, libraries)
            └─▶ Agent.BuildBrief() ──▶ pre-fetch authoritative docs
                    └─▶ returns Brief{ persona, instructions, docs, policies }
        IDE adopts the brief and does the work, docs already in context.
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for the full design.

## Quickstart

Requires Go 1.25+.

```bash
git clone https://github.com/KhizarA77/Faber.git
cd Faber
go build -o faber.exe ./cmd/faber
```

Register it with Claude Code:

```bash
claude mcp add faber -- "/absolute/path/to/faber.exe" mcp start
```

Then, in a Claude Code session, run `/mcp` to confirm `faber` is connected, or ask it to
*"list the faber agents"* and *"launch the code-reviewer on this diff."*

## Built-in agents

| Agent | Role |
|---|---|
| `code-reviewer` | Reviews diffs for correctness, security, and scalable design. |
| `architect` | Designs scalable, maintainable architecture and weighs trade-offs. |

Both are *docs-first*: external API usage is grounded in official documentation.

## MCP tools

| Tool | Purpose |
|---|---|
| `faber_list_agents` | Discover the available specialists. |
| `faber_launch_agent` | Launch an agent for a task; returns its brief. |

More tools (`faber_consult_docs`, `faber_orchestrate`) land with later milestones.

## Project layout

```
cmd/faber/        # CLI entrypoint: `faber mcp start`
internal/mcp/     # MCP server and tool registration
pkg/agent/        # Agent interface, registry, built-in agents
pkg/docs/         # Documentation-as-source-of-truth subsystem
pkg/orchestrator/ # Multi-agent coordination (planned)
pkg/memory/       # Session-scoped shared store
```

## Contributing

Contributions are welcome. New agents implement the `agent.Agent` interface and register
themselves — see [pkg/agent/builtin](pkg/agent/builtin) for examples. Please open an issue
to discuss larger changes first.

## License

[MIT](LICENSE) © 2026 Khizar
