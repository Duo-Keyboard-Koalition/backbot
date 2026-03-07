# Scorpion-Go ADK

Self-contained Go implementation of the Scorpion agent loop.

## Core value

Scorpion-Go ADK defaults to local, self-contained behavior:
- no external model dependency required to run
- workspace-scoped file tools with path escape protection
- single binary execution path (`go run ./cmd/scorpion-go`)

## Run

```bash
cd /workspaces/AuraFlow/scorpion-go
go run ./cmd/scorpion-go
```

Optional:
- `SCORPION_GO_WORKSPACE=/path/to/workspace` to limit file tools.

## Built-in commands

- `/time` calls `time_now`
- `/ls <path>` calls `list_dir`
- `/cat <path>` calls `read_file`

## Architecture

- `internal/adk/types.go`: model, tool, and state contracts
- `internal/adk/agent.go`: iterative model-tool loop
- `internal/adk/model_local.go`: self-contained rule model fallback
- `internal/adk/tools.go`: built-in tools and workspace safety
