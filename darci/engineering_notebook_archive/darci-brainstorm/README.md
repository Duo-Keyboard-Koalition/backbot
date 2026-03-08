# DarCI Brainstorm & Documentation

**Central hub for DarCI (Darcy's AI Agent) project management strategy, tools, and architecture**

---

## 🚀 HACKATHON MODE

> **Activate via system prompt. Switch to high-velocity sprint behavior.**

| Document | Description | Best For |
|----------|-------------|----------|
| [HACKATHON_MODE.md](./HACKATHON_MODE.md) | 🔥 Mode specification & behavior changes | Understanding the mode |
| [HACKATHON_MODE_PROMPTS.md](./HACKATHON_MODE_PROMPTS.md) | ⚡ **Copy-paste system prompts** | **Activating hackathon mode** |

**Quick Activation:**
```
"DarCI, activate HACKATHON MODE - 48 hours"
```

---

## 📚 Documentation Index

### Core Documents

| Document | Description | Status |
|----------|-------------|--------|
| [DARCI_PROJECT_MANAGEMENT.md](./DARCI_PROJECT_MANAGEMENT.md) | Project management strategy, workflows, and configuration | ✅ Complete |
| [TOOL_SPECIFICATION.md](./TOOL_SPECIFICATION.md) | Comprehensive tool definitions and implementations | ✅ Complete |
| [TOOL_BRAINSTORM.md](./TOOL_BRAINSTORM.md) | **117 tools** across 15 categories | ✅ Complete |
| [TOOLS_SUMMARY.md](./TOOLS_SUMMARY.md) | Prioritization matrix & phases | ✅ Complete |
| [AGENT_ARCHITECTURE.md](./AGENT_ARCHITECTURE.md) | System architecture and sub-agent design | ✅ Complete |
| [ENGINEERING_NOTEBOOK_TEMPLATE.md](./ENGINEERING_NOTEBOOK_TEMPLATE.md) | Template for engineering session documentation | ✅ Complete |
| [IMPLEMENTATION_CHECKLIST.md](./IMPLEMENTATION_CHECKLIST.md) | 11-phase implementation roadmap | ✅ Complete |
| [QUICK_REFERENCE.md](./QUICK_REFERENCE.md) | Commands & workflows quick card | ✅ Complete |

### Engineering Notebooks

| Notebook | Date | Topic |
|----------|------|-------|
| [2026-03-07_telegram_migration_status.md](./2026-03-07_telegram_migration_status.md) | 2026-03-07 | Python → Go Telegram migration assessment |

---

## 🎯 What is DarCI?

**DarCI** (Darcy's AI Agent) is an autonomous AI agent for:

- 📋 **Project Management** - Task tracking, prioritization, dependencies
- 🔧 **Engineering Automation** - Build monitoring, test execution, code analysis
- 📝 **Documentation** - Engineering notebooks, status reports, feature matrices
- 💬 **Communication** - Multi-channel notifications (Telegram, Discord, Slack)

Built on the [DarCI](../darci-python/) framework - ultra-lightweight personal AI assistant.

---

## 🚀 Quick Start

### 1. Install DarCI

```bash
cd ../darci-python
pip install -e .
```

### 2. Configure DarCI

Create `~/.darci/config.json`:

```json
{
  "providers": {
    "gemini": {
      "apiKey": "YOUR_GEMINI_API_KEY"
    }
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "token": "YOUR_BOT_TOKEN",
      "allowFrom": ["YOUR_USER_ID"]
    }
  },
  "tools": {
    "restrictToWorkspace": true
  }
}
```

### 3. Initialize DarCI Workspace

```bash
mkdir -p ~/.darci/darci/{memory,notebooks,metrics,artifacts}
```

### 4. Run DarCI

```bash
# Chat mode
darci agent

# Gateway mode (for Telegram/Discord/Slack)
darci gateway

# CLI command
darci agent -m "Create tasks for implementing Telegram webhook support"
```

---

## 📖 How to Use This Directory

### For New Features

1. **Read** [Project Management Strategy](./DARCI_PROJECT_MANAGEMENT.md)
2. **Create** tasks using the defined workflow
3. **Document** in engineering notebook using [template](./ENGINEERING_NOTEBOOK_TEMPLATE.md)

### For Tool Development

1. **Read** [Tool Specification](./TOOL_SPECIFICATION.md)
2. **Implement** tool following the schema
3. **Register** in DarCI's skill system

### For Architecture Understanding

1. **Read** [Agent Architecture](./AGENT_ARCHITECTURE.md)
2. **Understand** sub-agent roles
3. **Extend** with new capabilities

---

## 🏗️ Directory Structure

```
darci-brainstorm/
├── README.md                           # This file
├── DARCI_PROJECT_MANAGEMENT.md         # PM strategy & workflows
├── TOOL_SPECIFICATION.md               # Tool definitions
├── AGENT_ARCHITECTURE.md               # System architecture
├── ENGINEERING_NOTEBOOK_TEMPLATE.md    # Notebook template
├── 2026-03-07_telegram_migration_status.md  # Example notebook
├── skills/                             # DarCI skills (future)
│   ├── project-manager/
│   ├── build-engineer/
│   ├── code-analyst/
│   └── scribe/
└── templates/                          # Reusable templates (future)
    ├── task_report.md
    ├── status_dashboard.md
    └── feature_matrix.md
```

---

## 🛠️ Key Concepts

### Task Management

DarCI uses a priority-based task system:

```
P0 - Critical (immediate action required)
P1 - High (next session)
P2 - Normal (this week)
P3 - Low (when time permits)
```

### Engineering Notebooks

Every development session is documented:

```markdown
# Engineering Notebook - {{TITLE}}

**Date:** {{DATE}}
**Engineer:** DarCI Agent
**Task:** {{PRIMARY_TASK}}

---

## Objective
## Session Log
## Implementation Details
## Test Results
## Next Steps
```

### Feature Parity Tracking

For Python → Go migration:

| Feature | Python | Go | Status |
|---------|--------|----|--------|
| Long polling | ✅ | ❌ | Not started |
| Command handlers | ✅ | ❌ | In progress |

---

## 📊 Current Projects

### 1. Go Telegram Migration

**Status:** 0/14 features complete (0%)

**Goal:** Migrate Python Telegram bot to Go with full feature parity.

**Tasks:**
- [ ] T001: Fix import paths
- [ ] T002: Implement long polling
- [ ] T003: Add command handlers
- [ ] T004: Implement message handling
- [ ] T005: Add media download

**Notebook:** [2026-03-07_telegram_migration_status.md](./2026-03-07_telegram_migration_status.md)

---

## 🤖 Sub-Agents

DarCI operates in different modes:

| Agent | Role | Tools |
|-------|------|-------|
| **Project Manager** | Task coordination | task_create, task_update, dependency_graph |
| **Build Engineer** | Build monitoring | build_run, test_run, lint_run |
| **Code Analyst** | Code comparison | code_compare, code_analyze, feature_matrix |
| **Scribe** | Documentation | notebook_create, status_report_generate |

---

## 🔧 Configuration

### DarCI Config (`~/.darci/darci/config.yaml`)

```yaml
darci:
  mode: ["project_manager", "scribe"]

  workspace:
    root: "~/repos/sentinelai/darci"
    python_dir: "darci-python"
    go_dir: "darci-go"
  
  monitoring:
    build_interval: 60s
    status_polling: 30s
  
  notifications:
    telegram:
      enabled: true
      channel_id: "@darci_status"
    on_task_complete: true
    on_build_failure: true
```

---

## 📈 Metrics

DarCI tracks:

| Metric | Current | Target |
|--------|---------|--------|
| Task Velocity | 0/day | ≥ 5/day |
| Build Success Rate | 100% | ≥ 95% |
| Feature Parity (Go) | 0% | 100% |
| Response Time | < 30s | < 30s |
| Documentation Coverage | 1/1 | 1:1 |

---

## 🎭 Example Interactions

### Create Tasks

**User:** "Create tasks for implementing webhook support"

**DarCI:**
```
📋 Created feature: Telegram Webhook

Tasks:
  [P0] T004: Add webhook handler struct
  [P0] T005: Implement webhook endpoint
  [P1] T006: Add certificate management
  [P1] T007: Implement secret validation

📝 Notebook created: 2026-03-07_webhook_implementation.md
```

### Status Query

**User:** "What's the Go migration status?"

**DarCI:**
```
📊 Go Migration Status: 0/14 features (0%)

✅ Completed: None
🔄 In Progress: None
⏳ Pending: 14 tasks

📈 Velocity: 0 tasks/day
🎯 ETA: TBD
```

### Build Request

**User:** "Build darci-go and fix any errors"

**DarCI:**
```
🔧 Building darci-go...

❌ Build failed with 3 errors:
  - Import path errors in agent/context.go
  - Import path errors in channels/base.go
  - Import path errors in cli/commands.go

🔧 Fixing import paths...
✅ Fixed 3 import paths

🔧 Rebuilding...
✅ Build successful!

📝 Notebook updated: 2026-03-07_build_fix.md
```

---

## 🚧 Roadmap

### Phase 1: Foundation (Current)
- [x] Project management strategy
- [x] Tool specification
- [x] Agent architecture
- [x] Engineering notebook template
- [ ] Go Telegram implementation

### Phase 2: Automation (Q2 2026)
- [ ] Auto-fix common build errors
- [ ] Automated test generation
- [ ] CI/CD integration

### Phase 3: Intelligence (Q3 2026)
- [ ] Predictive task estimation
- [ ] Error pattern recognition
- [ ] Smart dependency detection

### Phase 4: Collaboration (Q4 2026)
- [ ] Multi-agent coordination
- [ ] Team dashboard (web UI)
- [ ] Jira/Linear integration

---

## 🔐 Security

```yaml
Security Model:
  - Workspace isolation: enabled
  - Restrict to workspace: true
  - No secrets in logs: enforced
  - API keys: environment variables only
  - Network: allowlist only
```

---

## 📚 Related Resources

- [DarCI Python](../darci-python/) - Main Python implementation
- [DarCI Go](../darci-go/) - Go implementation
- [DarCI README](../darci-python/README.md) - Framework documentation
- [Engineering Notebook](./2026-03-07_telegram_migration_status.md) - Example session

---

## 💡 Contributing

### Adding New Tools

1. Define tool in [TOOL_SPECIFICATION.md](./TOOL_SPECIFICATION.md)
2. Implement in DarCI skills
3. Test with agent
4. Document usage

### Adding Engineering Notebooks

1. Copy [template](./ENGINEERING_NOTEBOOK_TEMPLATE.md)
2. Fill in session details
3. Save as `YYYY-MM-DD_description.md`
4. Update index in this README

---

## 🎓 Learning Resources

### For New Users

1. Start with [Project Management Strategy](./DARCI_PROJECT_MANAGEMENT.md)
2. Review [Tool Specification](./TOOL_SPECIFICATION.md)
3. Read example [engineering notebook](./2026-03-07_telegram_migration_status.md)

### For Developers

1. Study [Agent Architecture](./AGENT_ARCHITECTURE.md)
2. Implement tools following [specification](./TOOL_SPECIFICATION.md)
3. Test with DarCI framework

---

## 📞 Communication

| Channel | Purpose | Link |
|---------|---------|------|
| Telegram | Real-time notifications | @BotFather |
| Discord | Community discussions | [Invite](https://discord.gg/MnCvHqpUGB) |
| Slack | Team collaboration | Workspace invite |
| CLI | Direct interaction | `darci agent` |

---

## 🏆 Success Criteria

DarCI is successful when:

- ✅ Tasks are tracked and completed efficiently
- ✅ Builds are monitored and failures fixed automatically
- ✅ Documentation is comprehensive and up-to-date
- ✅ Feature parity between Python and Go is maintained
- ✅ Team is informed of status and changes

---

## 📝 License

Same as DarCI project - MIT License

---

*Last updated: 2026-03-07*
*Version: 1.0*
*Author: DarCI Team*
