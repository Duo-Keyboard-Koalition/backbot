# Scorpion-Go ADK

Self-contained Go implementation of the Scorpion agent loop.

## Status

**Build:** ✅ Passing (`go build ./...`)
**Executable:** ✅ Working (`scorpion.exe`)
**Last Updated:** 2026-03-07

## Core value

Scorpion-Go ADK defaults to local, self-contained behavior:
- no external model dependency required to run
- workspace-scoped file tools with path escape protection
- single binary execution path (`go run ./cmd/scorpion-go`)

## Run

```bash
cd scorpion-go
go run ./cmd/scorpion-go
```

Or build the executable:

```bash
go build -o scorpion.exe ./cmd/scorpion-go
./scorpion.exe help
```

Optional:
- `SCORPION_GO_WORKSPACE=/path/to/workspace` to limit file tools.

## Built-in commands

- `/time` calls `time_now`
- `/ls <path>` calls `list_dir`
- `/cat <path>` calls `read_file`
- `/write <path>` calls `write_file`
- `/exec <command>` calls `exec`
- `/search <query>` calls `web_search`
- `/fetch <url>` calls `web_fetch`

## Architecture

- `internal/adk/types.go`: model, tool, and state contracts
- `internal/adk/agent.go`: iterative model-tool loop
- `internal/adk/model_local.go`: self-contained rule model fallback
- `internal/adk/tools.go`: built-in tools and workspace safety
- `scorpion/agent/`: agent, subagent, skills, and tools management
- `scorpion/session/`: session management for user conversations
- `scorpion/cli/`: command-line interface

## Project Structure

```
scorpion-go/
├── cmd/scorpion-go/     # Main executable
├── internal/adk/        # Core ADK types and implementations
├── scorpion/
│   ├── agent/          # Agent and subagent management
│   ├── session/        # Session management
│   ├── tools/          # Tool registry and base tools
│   └── cli/            # CLI commands
└── go.mod              # Go module definition
```

## Development

### Build

```bash
go build ./...
```

### Test

```bash
go test ./...
```

### Clean and Rebuild

```bash
go clean -cache && go build ./...
```

## Configuration

Configuration is stored in `~/.scorpion-go/config.json`.

## Engineering Notes

For detailed implementation notes and history, see:

| Topic | Document |
|-------|----------|
| Compilation Fix Guide | [`../engineering_notebook/go/2026-03-07_go_compilation_fix.md`](../engineering_notebook/go/2026-03-07_go_compilation_fix.md) |
| TUI Test Results | [`../engineering_notebook/shared/2026-03-07_tui_agent_test_results.md`](../engineering_notebook/shared/2026-03-07_tui_agent_test_results.md) |
| Telegram Status | [`../engineering_notebook/shared/2026-03-07_telegram_migration_status.md`](../engineering_notebook/shared/2026-03-07_telegram_migration_status.md) |

## Project Index

Main engineering notebook: [`../engineering_notebook/README.md`](../engineering_notebook/README.md)
