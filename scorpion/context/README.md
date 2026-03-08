# Multi-Agent Qwen Environment Context

## System Information
- **Operating System:** Windows (win32)
- **Working Directory:** `C:\Users\darcy\repos\sentinelai\scorpion`

## Important Notes for Agents
1. **Path Separators:** Use backslashes (`\`) for Windows paths, or forward slashes (`/`) which Windows also accepts
2. **CLI Commands:** Commands execute via `cmd.exe /c <command>`
3. **Case Sensitivity:** File paths are case-insensitive on Windows
4. **Environment Variables:** Use `%VAR%` syntax in commands, not `$VAR`

## ⚠️ Engineering Notebook Requirement

**ALL agents MUST populate the engineering notebook for every significant action.**

See: [`engineering_notebook/README.md`](../engineering_notebook/README.md)

### When to Update
- Starting/completing tasks
- Running tests or builds
- Making code changes
- Fixing bugs
- Configuration changes

**Rule of thumb:** If you spent more than 5 minutes on something, document it.

### File Naming Convention
```
engineering_notebook/YYYY-MM-DD_topic.md
```

## Project Structure
- `scorpion-python/` - Python implementation
- `scorpion-go/` - Go implementation
- `context/` - This folder (multi-agent Qwen environment context)
- `darci-brainstorm/` - DARCI project documentation
- `engineering_notebook/` - **Engineering documentation (POPULATE THIS!)**
