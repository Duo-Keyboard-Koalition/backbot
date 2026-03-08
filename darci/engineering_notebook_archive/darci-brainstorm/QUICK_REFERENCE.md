# DarCI Quick Reference Card

**Essential commands and workflows for DarCI project management**

---

## 🚀 Quick Commands

### Chat with DarCI

```bash
# Interactive chat
darci agent

# Single command
darci agent -m "Create tasks for the Go Telegram migration"

# With logs visible
darci agent --logs

# Gateway mode (Telegram/Discord/Slack)
darci gateway
```

### Common Requests

```
"Create tasks for implementing webhook support"
"What's the status of the Go migration?"
"Build darci-go and fix any errors"
"Generate a status report for this week"
"Compare Python and Go Telegram implementations"
"Create an engineering notebook for today's session"
```

---

## 📋 Task Priority Guide

| Priority | Meaning | Response Time | Example |
|----------|---------|---------------|---------|
| **P0** | Critical | Immediate | Build broken, production issue |
| **P1** | High | Next session | Core feature, blocker |
| **P2** | Normal | This week | Standard feature |
| **P3** | Low | When time permits | Nice-to-have, enhancement |

---

## 🛠️ Essential Tools

### Task Management

```yaml
task_create:
  Use: "Create a new task"
  Example: "Create task: Implement long polling (P0, telegram)"

task_update:
  Use: "Update task status"
  Example: "Mark T002 as in_progress"

task_query:
  Use: "Find tasks"
  Example: "Show all P0 tasks"

task_list:
  Use: "List all active tasks"
  Example: "What tasks are in progress?"
```

### Build & Test

```yaml
build_run:
  Use: "Execute build"
  Example: "Build darci-go"

test_run:
  Use: "Run tests"
  Example: "Test Telegram channel"

lint_run:
  Use: "Run linters"
  Example: "Lint darci-python"
```

### Documentation

```yaml
notebook_create:
  Use: "Generate engineering notebook"
  Example: "Create notebook for webhook implementation"

status_report_generate:
  Use: "Generate status report"
  Example: "Weekly status report"

feature_matrix_generate:
  Use: "Compare implementations"
  Example: "Python vs Go Telegram features"
```

---

## 📊 Workflow Patterns

### Feature Development

```
1. Request → "Create tasks for {{FEATURE}}"
2. Review → Check created tasks and priorities
3. Assign → "Start working on T002"
4. Monitor → "What's the status?"
5. Complete → "Mark T002 as completed"
```

### Build Fix

```
1. Detect → Build failure notification
2. Analyze → "What are the build errors?"
3. Fix → "Fix the import path errors"
4. Verify → "Rebuild and test"
5. Document → "Create notebook entry"
```

### Migration Tracking

```
1. Compare → "Compare Python and Go implementations"
2. Matrix → Generate feature matrix
3. Identify → "What features are missing?"
4. Plan → "Create tasks for missing features"
5. Track → "Update completion percentage"
```

---

## 📝 Engineering Notebook Quick Start

### Create New Notebook

```markdown
# Engineering Notebook - {{TITLE}}

**Date:** {{DATE}}
**Engineer:** DarCI Agent
**Task:** {{TASK}}

---

## Objective
{{What we're doing}}

## Session Log
| Time | Action | Result | Notes |
|------|--------|--------|-------|
| 10:00 | Started | - | - |

## Changes Made
{{Description}}

## Build Status
✅ Success / ❌ Failed

## Next Steps
1. {{Step}}
```

### File Naming

```
YYYY-MM-DD_description.md
Examples:
  2026-03-07_telegram_migration_status.md
  2026-03-07_webhook_implementation.md
  2026-03-07_build_fix.md
```

---

## 🎯 Sub-Agent Modes

| Mode | Use When | Example Request |
|------|----------|-----------------|
| **Project Manager** | Planning features | "Create tasks for webhook support" |
| **Build Engineer** | Fixing builds | "Build and fix errors" |
| **Code Analyst** | Comparing code | "Compare Python vs Go" |
| **Scribe** | Documentation | "Create engineering notebook" |

---

## 🔧 Configuration Quick Reference

### Add Telegram Channel

```json
{
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allowFrom": ["YOUR_USER_ID"]
    }
  }
}
```

### Set Workspace

```bash
export DARCI_WORKSPACE=/path/to/workspace
```

### DarCI Config

```yaml
# ~/.darci/darci/config.yaml
darci:
  mode: ["project_manager", "scribe"]
  notifications:
    telegram:
      enabled: true
```

---

## 📈 Common Metrics

| Metric | Command | Target |
|--------|---------|--------|
| Task count | "List tasks" | ≥ 5 completed/day |
| Build status | "Build darci-go" | ≥ 95% success |
| Feature parity | "Generate feature matrix" | 100% |
| Response time | (automatic) | < 30s |

---

## ⚠️ Troubleshooting

### Build Fails

```
1. "What are the build errors?"
2. "Fix the errors"
3. "Rebuild"
4. If persists: "Create notebook documenting the issue"
```

### Task Stuck

```
1. "Why is T002 blocked?"
2. "Remove blockers"
3. "Update task status"
```

### Missing Feature

```
1. "What features are missing?"
2. "Create tasks for missing features"
3. "Prioritize tasks"
```

---

## 🎭 Example Sessions

### Session 1: Feature Planning

```
User: "Create tasks for Telegram webhook support"

DarCI:
  📋 Created 5 tasks:
    [P0] T004: Add webhook handler struct
    [P0] T005: Implement webhook endpoint
    [P1] T006: Add certificate management
    [P1] T007: Implement secret validation
    [P2] T008: Add fallback to long polling
  
  📝 Notebook: 2026-03-07_webhook_implementation.md
```

### Session 2: Build Fix

```
User: "Build darci-go"

DarCI:
  🔧 Building... ❌ Failed
  
  3 import errors found:
    - darci/adk → internal/adk
    - darci/bus → internal/bus
    - darci/agent → internal/agent
  
  🔧 Fixing... ✅ Fixed
  
  🔧 Rebuilding... ✅ Success
  
  📝 Notebook: 2026-03-07_build_fix.md
```

### Session 3: Status Check

```
User: "Go migration status"

DarCI:
  📊 Go Migration: 3/14 features (21%)
  
  ✅ Completed: 3
  🔄 In Progress: 1
  ⏳ Pending: 10
  
  📈 Velocity: 2 tasks/day
  🎯 ETA: 2026-03-15
```

---

## 📚 File Locations

```
~/.darci/
├── config.json              # Main config
├── darci/
│   ├── config.yaml          # DarCI config
│   ├── tasks.json           # Task registry
│   ├── context.json         # Working context
│   ├── notebooks/           # Engineering notebooks
│   └── memory/              # Long-term memory
└── workspace/               # Project workspace
```

---

## 🔐 Security Checklist

- [ ] `restrictToWorkspace: true` in config
- [ ] API keys in environment variables only
- [ ] No secrets in logs or notebooks
- [ ] Network allowlist configured
- [ ] File permissions set correctly

---

## 📞 Getting Help

```bash
# DarCI help
darci --help
darci agent --help

# Status
darci status

# Channel status
darci channels status
```

---

## 🎓 Learning Path

1. **Day 1:** Read [Project Management](./DARCI_PROJECT_MANAGEMENT.md)
2. **Day 2:** Review [Tool Specification](./TOOL_SPECIFICATION.md)
3. **Day 3:** Study [Agent Architecture](./AGENT_ARCHITECTURE.md)
4. **Day 4:** Practice with example notebook
5. **Day 5:** Start real project work

---

*Quick Reference v1.0*
*Last updated: 2026-03-07*
*Print this for quick access!*
