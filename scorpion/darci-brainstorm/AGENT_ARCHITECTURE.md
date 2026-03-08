# DarCI Agent Architecture

**Comprehensive architecture specification for the DarCI project management agent**

---

## 🎯 System Overview

DarCI is an autonomous AI agent built on the Scorpion framework, designed for:
- **Project Management** - Task tracking, prioritization, dependency management
- **Engineering Automation** - Build monitoring, test execution, code analysis
- **Documentation** - Engineering notebooks, status reports, feature matrices
- **Communication** - Multi-channel notifications, status broadcasts

---

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Interfaces                           │
├─────────────────────────────────────────────────────────────────┤
│  Telegram │ Discord │ Slack │ CLI │ Web Dashboard │ API         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    DarCI Agent Core                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Agent Loop (Scorpion ADK)                    │  │
│  │  ┌────────────┐  ┌────────────┐  ┌────────────┐         │  │
│  │  │  Perceive  │→│   Think    │→│    Act     │         │  │
│  │  └────────────┘  └────────────┘  └────────────┘         │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   Skills     │  │   Memory     │  │   Tools      │         │
│  │  (Plugins)   │  │   (Context)  │  │  (Actions)   │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Sub-Agent Layer                               │
├─────────────────────────────────────────────────────────────────┤
│  ┌────────────┐  ┌────────────┐  ┌────────────┐  ┌──────────┐ │
│  │  Project   │  │  Build     │  │  Code      │  │  Scribe  │ │
│  │  Manager   │  │  Engineer  │  │  Analyst   │  │  Agent   │ │
│  └────────────┘  └────────────┘  └────────────┘  └──────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    External Systems                              │
├─────────────────────────────────────────────────────────────────┤
│  GitHub │ Go CI │ Python CI │ MCP Servers │ Cloud Services     │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🧠 Agent Core Components

### 1. **Agent Loop** (Scorpion ADK)

The core agent loop follows the Perceive → Think → Act pattern:

```go
// Conceptual flow
type DarCIAgent struct {
    perception  *PerceptionEngine
    cognition   *CognitionEngine
    action      *ActionEngine
    memory      *MemorySystem
}

func (a *DarCIAgent) Run() {
    for {
        // Perceive: Gather input from channels, files, systems
        input := a.perception.Gather()
        
        // Think: Process with LLM, consult memory, plan
        plan := a.cognition.Process(input)
        
        // Act: Execute tools, update state, notify
        result := a.action.Execute(plan)
        
        // Remember: Store in memory for future context
        a.memory.Store(input, plan, result)
    }
}
```

---

### 2. **Skills System**

DarCI extends Scorpion's skill system with specialized capabilities:

```
darci-skills/
├── project-manager/
│   ├── SKILL.md
│   ├── task_tracker.py
│   ├── priority_engine.py
│   └── dependency_resolver.py
├── build-engineer/
│   ├── SKILL.md
│   ├── build_monitor.py
│   ├── test_runner.py
│   └── linter.py
├── code-analyst/
│   ├── SKILL.md
│   ├── diff_analyzer.py
│   ├── import_checker.py
│   └── feature_comparator.py
├── scribe/
│   ├── SKILL.md
│   ├── notebook_generator.py
│   ├── report_writer.py
│   └── template_engine.py
└── communicator/
    ├── SKILL.md
    ├── notification_router.py
    ├── message_formatter.py
    └── channel_bridge.py
```

**Skill Registration:**
```yaml
# SKILL.md frontmatter
---
name: project_manager
description: "Task tracking and project coordination"
version: 1.0.0
author: DarCI Team
tools:
  - task_create
  - task_update
  - task_query
  - dependency_graph
requires:
  bins: ["jq"]
  files: ["~/.scorpion/darci/tasks.json"]
---
```

---

### 3. **Memory System**

DarCI maintains multiple memory layers:

```yaml
Memory Layers:
  
  short_term:
    type: "conversation_buffer"
    capacity: "last_50_messages"
    purpose: "Context for current session"
  
  working:
    type: "task_context"
    storage: "~/.scorpion/darci/context.json"
    purpose: "Active tasks, current focus, recent actions"
  
  long_term:
    type: "vector_store"
    storage: "~/.scorpion/darci/memory/"
    purpose: "Historical sessions, patterns, learnings"
  
  semantic:
    type: "knowledge_graph"
    storage: "~/.scorpion/darci/knowledge.json"
    purpose: "Project structure, dependencies, relationships"
```

**Memory Schema:**
```json
{
  "session": {
    "id": "S20260307-001",
    "started_at": "2026-03-07T10:00:00Z",
    "focus": "Telegram migration",
    "tasks_active": ["T002", "T003"]
  },
  "context": {
    "workspace": "/path/to/scorpion",
    "current_build_status": "success",
    "last_error": null,
    "active_channels": ["telegram", "cli"]
  },
  "knowledge": {
    "project_structure": {
      "python_dir": "scorpion-python",
      "go_dir": "scorpion-go",
      "docs_dir": "darci-brainstorm"
    },
    "feature_parity": {
      "telegram": {
        "python_features": 14,
        "go_features": 0,
        "completion": "0%"
      }
    }
  }
}
```

---

### 4. **Tool System**

DarCI tools are implemented as Scorpion tool definitions:

```python
# Example: task_create tool
from scorpion.adk import Tool, tool

@tool
async def task_create(
    title: str,
    description: str = None,
    priority: str = "P2",
    dependencies: list = None,
    labels: list = None
) -> dict:
    """Create a new task in DarCI's tracking system."""
    
    task_id = generate_task_id()
    task = {
        "id": task_id,
        "title": title,
        "description": description,
        "priority": priority,
        "status": "pending",
        "dependencies": dependencies or [],
        "labels": labels or [],
        "created_at": datetime.utcnow().isoformat()
    }
    
    await save_task(task)
    await log_action("task_create", task_id)
    
    return {
        "task_id": task_id,
        "status": "created",
        "message": f"Task '{title}' created with ID {task_id}"
    }
```

---

## 🤖 Sub-Agent Specializations

### Project Manager Agent

```yaml
Role: Project coordination and task management
Responsibilities:
  - Create and prioritize tasks
  - Track dependencies and blockers
  - Generate status reports
  - Estimate timelines
  
Tools:
  - task_create, task_update, task_query
  - dependency_graph
  - status_report_generate
  
Prompts:
  system: |
    You are DarCI Project Manager.
    Focus on: prioritization, dependencies, timelines.
    Always consider: critical path, blockers, resource allocation.
```

### Build Engineer Agent

```yaml
Role: Build monitoring and CI/CD
Responsibilities:
  - Watch for file changes
  - Trigger builds and tests
  - Fix common build errors
  - Report build status
  
Tools:
  - build_run, test_run, lint_run
  - build_watch
  - git_status, git_commit
  
Prompts:
  system: |
    You are DarCI Build Engineer.
    Focus on: build success, test coverage, code quality.
    Always consider: build time, error patterns, fixes.
```

### Code Analyst Agent

```yaml
Role: Code analysis and comparison
Responsibilities:
  - Compare Python ↔ Go implementations
  - Identify feature gaps
  - Analyze code quality
  - Suggest improvements
  
Tools:
  - code_compare, code_analyze
  - import_analyzer
  - feature_matrix_generate
  
Prompts:
  system: |
    You are DarCI Code Analyst.
    Focus on: feature parity, code quality, patterns.
    Always consider: consistency, best practices, gaps.
```

### Scribe Agent

```yaml
Role: Documentation and notebooks
Responsibilities:
  - Generate engineering notebooks
  - Write status reports
  - Maintain documentation
  - Create feature matrices
  
Tools:
  - notebook_create, notebook_update
  - status_report_generate
  - feature_matrix_generate
  
Prompts:
  system: |
    You are DarCI Scribe.
    Focus on: clear documentation, accurate records.
    Always consider: completeness, clarity, traceability.
```

---

## 📡 Communication Channels

### Channel Integration

DarCI supports multiple communication channels via Scorpion:

```yaml
channels:
  telegram:
    enabled: true
    use_cases: ["notifications", "status_updates", "chat"]
    features: ["typing_indicators", "reactions", "media"]
  
  discord:
    enabled: true
    use_cases: ["notifications", "embeds", "threads"]
    features: ["rich_embeds", "threads", "reactions"]
  
  slack:
    enabled: true
    use_cases: ["notifications", "blocks", "threads"]
    features: ["block_kit", "threads", "reactions"]
  
  cli:
    enabled: true
    use_cases: ["direct_interaction", "debugging"]
    features: ["interactive", "logs", "colors"]
  
  webhook:
    enabled: false
    use_cases: ["external_integrations", "ci_cd"]
    features: ["json_payloads", "signatures"]
```

### Notification Routing

```python
class NotificationRouter:
    """Route notifications to appropriate channels."""
    
    async def send(self, notification: Notification):
        # Determine channels based on priority and type
        channels = self.route(notification)
        
        for channel in channels:
            formatted = self.format(notification, channel)
            await self.send_to_channel(channel, formatted)
    
    def route(self, notification: Notification) -> list:
        routing_rules = {
            "P0": ["telegram", "slack", "discord"],  # Critical: all channels
            "P1": ["telegram", "slack"],              # High: IM channels
            "P2": ["slack"],                          # Normal: Slack only
            "P3": [],                                 # Low: log only
        }
        
        if notification.type == "build_failure":
            return routing_rules.get(notification.priority, ["slack"])
        
        return ["slack"]  # Default
```

---

## 🗄️ Data Storage

### File Structure

```
~/.scorpion/darci/
├── config.yaml              # DarCI configuration
├── tasks.json               # Task registry
├── context.json             # Working context
├── knowledge.json           # Semantic knowledge
├── memory/                  # Long-term memory
│   ├── sessions/
│   ├── patterns/
│   └── learnings/
├── notebooks/               # Engineering notebooks
│   └── *.md
├── metrics/                 # Build/test metrics
│   └── *.json
└── artifacts/               # Build artifacts, logs
    └── *.log
```

### Task Store Schema

```json
{
  "version": "1.0",
  "tasks": {
    "T001": {
      "id": "T001",
      "title": "Fix import paths",
      "description": "Update imports to internal/*",
      "priority": "P0",
      "status": "completed",
      "created_at": "2026-03-07T10:00:00Z",
      "updated_at": "2026-03-07T10:30:00Z",
      "completed_at": "2026-03-07T10:30:00Z",
      "dependencies": [],
      "labels": ["build", "scorpion-go"],
      "assignee": "build-engineer",
      "artifacts": {
        "notebook": "2026-03-07_import_fix.md",
        "commit": "abc123"
      }
    }
  },
  "indexes": {
    "by_status": {
      "pending": ["T005"],
      "in_progress": ["T002"],
      "completed": ["T001", "T003", "T004"]
    },
    "by_priority": {
      "P0": ["T001", "T002"],
      "P1": ["T003"],
      "P2": ["T004", "T005"]
    }
  }
}
```

---

## 🔄 Workflows

### Feature Development Workflow

```
┌─────────────┐
│ User Request│
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│ 1. Parse Request                │
│    - Extract feature details    │
│    - Identify requirements      │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│ 2. Create Tasks                 │
│    - Break into atomic tasks    │
│    - Set priorities             │
│    - Define dependencies        │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│ 3. Generate Notebook Entry      │
│    - Document feature spec      │
│    - Create tracking matrix     │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│ 4. Assign to Sub-Agent          │
│    - Route to appropriate agent │
│    - Provide context            │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│ 5. Monitor Progress             │
│    - Track task completion      │
│    - Update status              │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│ 6. Validate & Complete          │
│    - Run tests                  │
│    - Update notebook            │
│    - Notify user                │
└─────────────────────────────────┘
```

### Build Monitoring Workflow

```
File Change Detected
        │
        ▼
┌───────────────────┐
│ Debounce (1s)     │
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ Identify Target   │
│ (Go or Python)    │
└───────┬───────────┘
        │
        ▼
┌───────────────────┐
│ Run Build         │
│ Capture Output    │
└───────┬───────────┘
        │
        ├─────────────┐
        │             │
        ▼             ▼
┌───────────┐   ┌───────────┐
│ Success   │   │ Failure   │
└─────┬─────┘   └─────┬─────┘
      │               │
      ▼               ▼
┌───────────┐   ┌───────────┐
│ Run Tests │   │ Parse     │
└─────┬─────┘   │ Errors    │
      │         └─────┬─────┘
      │               │
      ▼               ▼
┌───────────┐   ┌───────────┐
│ Update    │   │ Create    │
│ Metrics   │   │ Fix Task  │
└─────┬─────┘   └─────┬─────┘
      │               │
      └───────┬───────┘
              │
              ▼
      ┌───────────────┐
      │ Notify Status │
      └───────────────┘
```

---

## 🔐 Security Model

```yaml
Security Boundaries:
  
  workspace:
    restrict_to_workspace: true
    allowed_paths:
      - "~/repos/sentinelai/scorpion/**"
    denied_paths:
      - "/etc/**"
      - "/root/**"
      - "~/.ssh/**"
  
  file_operations:
    read: ["workspace/**", "darci-brainstorm/**"]
    write: ["workspace/**", "darci-brainstorm/**"]
    execute: ["go", "python", "git", "docker"]
  
  network:
    allowed_hosts:
      - "github.com"
      - "api.telegram.org"
      - "discord.com"
      - "slack.com"
      - "api.github.com"
    denied_hosts: ["*"]  # Default deny
  
  secrets:
    handling: "environment_variables_only"
    never_log: ["API_KEY", "TOKEN", "SECRET", "PASSWORD"]
    redaction: "automatic"
```

---

## 📊 Monitoring & Observability

### Metrics Collected

```yaml
metrics:
  agent:
    - response_time_ms
    - tasks_created_total
    - tasks_completed_total
    - errors_total
  
  build:
    - build_duration_ms
    - build_success_rate
    - test_pass_rate
    - coverage_percent
  
  code:
    - features_implemented
    - feature_parity_percent
    - code_quality_score
  
  communication:
    - notifications_sent_total
    - channel_usage
    - response_time_by_channel
```

### Dashboards

DarCI can generate status dashboards:

```markdown
# DarCI Status Dashboard

**Generated:** 2026-03-07 15:30:00

## Project Health

| Metric | Value | Trend |
|--------|-------|-------|
| Tasks Completed | 12/50 | ↑ +3 today |
| Build Success | 95% | → stable |
| Feature Parity | 21% | ↑ +2% today |

## Active Tasks

| ID | Title | Priority | Status | Assignee |
|----|-------|----------|--------|----------|
| T002 | Long polling | P0 | in_progress | build-engineer |
| T003 | Command handlers | P1 | pending | build-engineer |

## Recent Builds

| Time | Target | Result | Duration |
|------|--------|--------|----------|
| 15:28 | scorpion-go | ✅ | 2.3s |
| 15:25 | scorpion-python | ✅ | 1.8s |

## Blockers

None 🎉
```

---

## 🚀 Deployment

### Local Development

```bash
# Install dependencies
cd scorpion-python
pip install -e .

# Configure DarCI
scorpion onboard

# Edit config
vim ~/.scorpion/config.json

# Run DarCI
scorpion agent
```

### Docker Deployment

```dockerfile
FROM python:3.11-slim

WORKDIR /app
COPY . .

RUN pip install -e .

ENV DARCI_MODE=project_manager
ENV SCORPION_WORKSPACE=/workspace

VOLUME /workspace
VOLUME /root/.scorpion

CMD ["scorpion", "gateway"]
```

---

## 📈 Future Enhancements

### Phase 1: Foundation ✅
- [x] Core agent architecture
- [x] Task tracking system
- [x] Engineering notebooks
- [ ] Go Telegram implementation

### Phase 2: Intelligence
- [ ] Predictive task estimation (ML-based)
- [ ] Automatic error pattern recognition
- [ ] Smart dependency detection

### Phase 3: Collaboration
- [ ] Multi-agent coordination
- [ ] Team dashboard (web UI)
- [ ] Integration with Jira/Linear

### Phase 4: Autonomy
- [ ] Self-healing builds
- [ ] Automatic PR creation
- [ ] Proactive task suggestions

---

*Architecture version: 1.0*
*Last updated: 2026-03-07*
*Author: DarCI Team*
