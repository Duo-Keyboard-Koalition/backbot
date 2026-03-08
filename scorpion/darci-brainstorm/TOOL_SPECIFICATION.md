# DarCI Tool Specification

This document defines the tools required for the DarCI agent to function as an effective project manager and engineering assistant.

---

## 🛠️ Core Tool Categories

1. **Task Management Tools** - Create, track, and manage tasks
2. **Build & Test Tools** - Monitor builds, run tests, capture results
3. **Documentation Tools** - Generate notebooks, reports, and summaries
4. **Code Analysis Tools** - Compare implementations, detect issues
5. **Communication Tools** - Send notifications, update channels
6. **File System Tools** - Read/write configs, logs, and artifacts

---

## 1. Task Management Tools

### `task_create`

Create a new task in the tracking system.

```yaml
name: task_create
description: Create a new task with metadata
parameters:
  title: string (required)
  description: string (optional)
  priority: enum [P0, P1, P2, P3] (default: P2)
  dependencies: array[string] (optional)
  labels: array[string] (optional)
  due_date: string ISO8601 (optional)
returns:
  task_id: string
  status: string
```

**Example:**
```json
{
  "name": "task_create",
  "arguments": {
    "title": "Implement Telegram long polling",
    "description": "Add long polling support to Go Telegram channel",
    "priority": "P0",
    "labels": ["telegram", "scorpion-go", "feature"],
    "dependencies": ["T001"]
  }
}
```

---

### `task_update`

Update an existing task's status or metadata.

```yaml
name: task_update
parameters:
  task_id: string (required)
  status: enum [pending, in_progress, blocked, completed] (optional)
  title: string (optional)
  description: string (optional)
  priority: enum [P0, P1, P2, P3] (optional)
  add_labels: array[string] (optional)
  remove_labels: array[string] (optional)
returns:
  task_id: string
  updated_fields: array[string]
```

---

### `task_query`

Query tasks by filters.

```yaml
name: task_query
parameters:
  status: enum [pending, in_progress, blocked, completed] (optional)
  priority: enum [P0, P1, P2, P3] (optional)
  labels: array[string] (optional)
  limit: int (default: 20)
  sort_by: enum [created_at, priority, due_date] (default: created_at)
returns:
  tasks: array[Task]
  total_count: int
```

---

### `task_list`

List all active tasks with summary.

```yaml
name: task_list
parameters:
  include_completed: boolean (default: false)
returns:
  summary:
    total: int
    pending: int
    in_progress: int
    blocked: int
    completed: int
  tasks: array[Task]
```

---

### `task_dependency_graph`

Generate dependency graph for tasks.

```yaml
name: task_dependency_graph
parameters:
  task_id: string (optional) - specific task, or all if omitted
returns:
  nodes: array[TaskNode]
  edges: array[DependencyEdge]
  critical_path: array[string]
```

---

## 2. Build & Test Tools

### `build_run`

Execute a build command.

```yaml
name: build_run
parameters:
  target: enum [scorpion-go, scorpion-python, all] (required)
  command: string (optional) - custom command
  timeout_ms: int (default: 120000)
  capture_output: boolean (default: true)
returns:
  success: boolean
  exit_code: int
  stdout: string
  stderr: string
  duration_ms: int
  artifacts: array[string]
```

**Example:**
```json
{
  "name": "build_run",
  "arguments": {
    "target": "scorpion-go",
    "command": "go build ./..."
  }
}
```

---

### `test_run`

Run tests for a target.

```yaml
name: test_run
parameters:
  target: enum [scorpion-go, scorpion-python, all] (required)
  test_pattern: string (optional) - e.g., "TestTelegram"
  coverage: boolean (default: false)
  timeout_ms: int (default: 300000)
returns:
  passed: int
  failed: int
  skipped: int
  coverage_percent: float (if enabled)
  failures: array[TestFailure]
```

---

### `build_watch`

Watch for file changes and trigger builds.

```yaml
name: build_watch
parameters:
  paths: array[string] (required)
  ignore_patterns: array[string] (optional)
  on_change: enum [build, test, notify] (default: build)
  debounce_ms: int (default: 1000)
returns:
  watch_id: string
  status: string
```

---

### `lint_run`

Run linters on code.

```yaml
name: lint_run
parameters:
  target: enum [scorpion-go, scorpion-python, all] (required)
  linter: enum [golangci-lint, ruff, mypy, all] (optional)
  fix: boolean (default: false)
returns:
  issues: array[LintIssue]
  fixable: int
  fixed: int
```

---

## 3. Documentation Tools

### `notebook_create`

Generate an engineering notebook entry.

```yaml
name: notebook_create
parameters:
  title: string (required)
  template: enum [default, migration, build_fix, feature] (default: default)
  sections: array[string] (optional)
  metadata: object (optional)
returns:
  notebook_path: string
  url: string (if published)
```

**Template: Migration**
```markdown
# Engineering Notebook - {{title}}

**Date:** {{date}}
**Engineer:** {{engineer}}
**Task:** {{task}}

---

## Objective

{{objective}}

---

## Current State

| Feature | Python Status | Go Status | Gap |
|---------|--------------|-----------|-----|
{{feature_table}}

---

## Build Status

```bash
{{build_command}}
```

**Result:** {{success|failed}}

---

## Next Steps

1. {{step1}}
2. {{step2}}

---

## Session Log

| Time | Action | Notes |
|------|--------|-------|
{{session_log}}
```

---

### `notebook_update`

Append to an existing notebook.

```yaml
name: notebook_update
parameters:
  notebook_path: string (required)
  section: string (required)
  content: string (required)
  append: boolean (default: true)
returns:
  notebook_path: string
  updated: boolean
```

---

### `status_report_generate`

Generate a status report.

```yaml
name: status_report_generate
parameters:
  period: enum [daily, weekly, custom] (required)
  start_date: string ISO8601 (optional)
  end_date: string ISO8601 (optional)
  include_metrics: boolean (default: true)
  format: enum [markdown, json, html] (default: markdown)
returns:
  report_path: string
  summary: object
```

---

### `feature_matrix_generate`

Generate a feature comparison matrix.

```yaml
name: feature_matrix_generate
parameters:
  source: string (required) - e.g., "Python Telegram"
  target: string (required) - e.g., "Go Telegram"
  categories: array[string] (optional)
returns:
  matrix: object
  completion_percent: float
  gaps: array[string]
```

---

## 4. Code Analysis Tools

### `code_compare`

Compare implementations between two codebases.

```yaml
name: code_compare
parameters:
  source_path: string (required)
  target_path: string (required)
  ignore_whitespace: boolean (default: true)
  ignore_comments: boolean (default: false)
returns:
  similarity_score: float
  differences: array[Diff]
  missing_features: array[string]
```

---

### `code_analyze`

Analyze code for issues.

```yaml
name: code_analyze
parameters:
  path: string (required)
  language: enum [go, python, auto] (default: auto)
  checks: array[string] (optional)
returns:
  issues: array[CodeIssue]
  suggestions: array[CodeSuggestion]
  metrics: object
```

---

### `import_analyzer`

Analyze and fix import paths.

```yaml
name: import_analyzer
parameters:
  path: string (required)
  fix: boolean (default: false)
  dry_run: boolean (default: true)
returns:
  imports: array[Import]
  issues: array[ImportIssue]
  fixes_applied: int (if fix=true)
```

---

## 5. Communication Tools

### `notify_telegram`

Send a notification via Telegram.

```yaml
name: notify_telegram
parameters:
  chat_id: string (required)
  message: string (required)
  parse_mode: enum [markdown, html, plain] (default: markdown)
  attachments: array[string] (optional)
returns:
  message_id: int
  success: boolean
```

---

### `notify_discord`

Send a notification via Discord.

```yaml
name: notify_discord
parameters:
  channel_id: string (required)
  message: string (required)
  embed: object (optional)
  attachments: array[string] (optional)
returns:
  message_id: string
  success: boolean
```

---

### `notify_slack`

Send a notification via Slack.

```yaml
name: notify_slack
parameters:
  channel: string (required)
  text: string (required)
  blocks: array[object] (optional)
  attachments: array[string] (optional)
returns:
  ts: string
  success: boolean
```

---

### `notification_broadcast`

Send to multiple channels.

```yaml
name: notification_broadcast
parameters:
  channels: array[enum [telegram, discord, slack, email]] (required)
  message: string (required)
  title: string (optional)
  priority: enum [low, normal, high, urgent] (default: normal)
returns:
  results: array[NotificationResult]
  success_count: int
  failure_count: int
```

---

## 6. File System Tools

### `file_read`

Read a file with workspace safety.

```yaml
name: file_read
parameters:
  path: string (required)
  max_lines: int (optional)
  offset: int (default: 0)
returns:
  content: string
  line_count: int
  truncated: boolean
```

---

### `file_write`

Write a file with workspace safety.

```yaml
name: file_write
parameters:
  path: string (required)
  content: string (required)
  create_dirs: boolean (default: true)
  overwrite: boolean (default: false)
returns:
  path: string
  bytes_written: int
  success: boolean
```

---

### `file_append`

Append to a file.

```yaml
name: file_append
parameters:
  path: string (required)
  content: string (required)
  add_newline: boolean (default: true)
returns:
  path: string
  bytes_appended: int
  success: boolean
```

---

### `file_search`

Search for files by pattern.

```yaml
name: file_search
parameters:
  pattern: string (required) - glob pattern
  path: string (default: workspace root)
  max_results: int (default: 50)
returns:
  files: array[string]
  count: int
```

---

### `config_read`

Read and validate a config file.

```yaml
name: config_read
parameters:
  config_type: enum [darci, scorpion, custom] (required)
  path: string (optional)
returns:
  config: object
  schema_valid: boolean
  errors: array[string]
```

---

### `config_write`

Write a config file with validation.

```yaml
name: config_write
parameters:
  config_type: enum [darci, scorpion, custom] (required)
  config: object (required)
  path: string (optional)
  backup: boolean (default: true)
returns:
  path: string
  success: boolean
  backup_path: string (if backup=true)
```

---

## 7. Git Tools

### `git_status`

Get git repository status.

```yaml
name: git_status
parameters:
  repo_path: string (default: workspace root)
  include_untracked: boolean (default: true)
returns:
  branch: string
  ahead: int
  behind: int
  staged: array[string]
  unstaged: array[string]
  untracked: array[string]
```

---

### `git_commit`

Create a commit.

```yaml
name: git_commit
parameters:
  message: string (required)
  files: array[string] (optional) - or commit all staged
  sign: boolean (default: false)
returns:
  commit_hash: string
  success: boolean
```

---

### `git_diff`

Get diff since last commit.

```yaml
name: git_diff
parameters:
  staged: boolean (default: false)
  path: string (optional)
returns:
  diff: string
  files_changed: int
  insertions: int
  deletions: int
```

---

### `git_log`

Get commit history.

```yaml
name: git_log
parameters:
  max_commits: int (default: 10)
  path: string (optional)
  format: enum [oneline, full, json] (default: oneline)
returns:
  commits: array[Commit]
```

---

## 8. MCP Integration Tools

### `mcp_list_servers`

List configured MCP servers.

```yaml
name: mcp_list_servers
returns:
  servers: array[MCPServer]
  active_count: int
```

---

### `mcp_list_tools`

List available MCP tools.

```yaml
name: mcp_list_tools
parameters:
  server: string (optional) - filter by server
returns:
  tools: array[MCPTool]
```

---

### `mcp_call_tool`

Call an MCP tool.

```yaml
name: mcp_call_tool
parameters:
  tool_name: string (required)
  arguments: object (required)
  server: string (optional)
returns:
  result: object
  success: boolean
  error: string (if failed)
```

---

## 📦 Tool Implementation Priority

### Phase 1 (Critical - Week 1)
- `task_create`, `task_update`, `task_query`, `task_list`
- `build_run`, `test_run`
- `notebook_create`, `notebook_update`
- `file_read`, `file_write`, `file_append`
- `notify_telegram`

### Phase 2 (High - Week 2)
- `build_watch`, `lint_run`
- `status_report_generate`, `feature_matrix_generate`
- `code_compare`, `code_analyze`
- `notify_discord`, `notify_slack`
- `git_status`, `git_commit`, `git_diff`

### Phase 3 (Medium - Week 3)
- `task_dependency_graph`
- `import_analyzer`
- `notification_broadcast`
- `config_read`, `config_write`
- `mcp_list_tools`, `mcp_call_tool`

### Phase 4 (Nice-to-have)
- Advanced analytics tools
- Visualization tools
- Integration with external PM tools

---

## 🔧 Tool Registration

Tools should be registered in DarCI's skill system:

```yaml
# ~/.scorpion/darci/skills/project_manager/SKILL.md
---
name: project_manager
description: "Project management and task tracking for DarCI"
metadata:
  tools:
    - task_create
    - task_update
    - task_query
    - task_list
    - build_run
    - notebook_create
    - notify_telegram
---

# Tool implementations...
```

---

*Last updated: 2026-03-07*
*Version: 1.0*
