# DarCI Implementation Checklist

**Step-by-step guide to implementing DarCI agent capabilities**

---

## Phase 1: Foundation Setup ✅

### 1.1 Environment Setup

- [ ] Install DarCI framework
  ```bash
  cd ../darci-python
  pip install -e .
  ```

- [ ] Initialize DarCI configuration
  ```bash
  darci onboard
  ```

- [ ] Add Gemini API key to `~/.darci/config.json`
  ```json
  {
    "providers": {
      "gemini": {
        "apiKey": "YOUR_GEMINI_API_KEY"
      }
    }
  }
  ```

- [ ] Create DarCI directory structure
  ```bash
  mkdir -p ~/.darci/darci/{memory,notebooks,metrics,artifacts}
  ```

- [ ] Create DarCI config file
  ```bash
  cat > ~/.darci/darci/config.yaml << EOF
  darci:
    mode: ["project_manager", "scribe"]
    workspace:
      root: "~/repos/sentinelai/darci"
    notifications:
      telegram:
        enabled: false
  EOF
  ```

### 1.2 Task Management System

- [ ] Create task store schema (`~/.darci/darci/tasks.json`)
  ```json
  {
    "version": "1.0",
    "tasks": {},
    "indexes": {
      "by_status": {},
      "by_priority": {}
    }
  }
  ```

- [ ] Implement `task_create` tool
  - [ ] Generate task ID
  - [ ] Validate priority (P0-P3)
  - [ ] Store in tasks.json
  - [ ] Return task object

- [ ] Implement `task_update` tool
  - [ ] Find task by ID
  - [ ] Update fields
  - [ ] Update timestamp
  - [ ] Persist changes

- [ ] Implement `task_query` tool
  - [ ] Filter by status
  - [ ] Filter by priority
  - [ ] Filter by labels
  - [ ] Sort results

- [ ] Implement `task_list` tool
  - [ ] Generate summary counts
  - [ ] List active tasks
  - [ ] Format output

### 1.3 Basic Agent Prompts

- [ ] Create system prompt for DarCI
  ```
  You are DarCI, Darcy's AI project management agent.
  Your role is to track tasks, monitor builds, and document progress.
  Always be concise, actionable, and thorough.
  ```

- [ ] Create task creation prompt template
- [ ] Create status report prompt template
- [ ] Create notebook generation prompt template

---

## Phase 2: Build Monitoring 🔨

### 2.1 Build Tools

- [ ] Implement `build_run` tool
  - [ ] Support darci-go target
  - [ ] Support darci-python target
  - [ ] Capture stdout/stderr
  - [ ] Parse success/failure
  - [ ] Return build result object

- [ ] Implement `test_run` tool
  - [ ] Run Go tests
  - [ ] Run Python tests
  - [ ] Capture pass/fail counts
  - [ ] Parse test failures
  - [ ] Optional: coverage report

- [ ] Implement `lint_run` tool
  - [ ] Run golangci-lint for Go
  - [ ] Run ruff for Python
  - [ ] Parse lint issues
  - [ ] Optional: auto-fix

### 2.2 Build Watching

- [ ] Implement `build_watch` tool
  - [ ] Watch file system
  - [ ] Debounce changes (1s)
  - [ ] Trigger build on change
  - [ ] Notify on failure

- [ ] Create file watcher for darci-go
  - [ ] Watch `*.go` files
  - [ ] Watch `go.mod`, `go.sum`
  - [ ] Ignore `vendor/`, `.git/`

- [ ] Create file watcher for darci-python
  - [ ] Watch `*.py` files
  - [ ] Watch `pyproject.toml`
  - [ ] Ignore `__pycache__/`, `.venv/`

### 2.3 Error Auto-Fix

- [ ] Implement import path fixer
  - [ ] Detect `darci-go/darci/*` imports
  - [ ] Replace with `darci-go/internal/*`
  - [ ] Run `go mod tidy`
  - [ ] Verify build

- [ ] Create common error pattern recognizer
  - [ ] Import errors
  - [ ] Syntax errors
  - [ ] Type errors
  - [ ] Missing dependencies

---

## Phase 3: Documentation 📝

### 3.1 Engineering Notebooks

- [ ] Implement `notebook_create` tool
  - [ ] Load template
  - [ ] Fill metadata (date, engineer, task)
  - [ ] Generate filename (YYYY-MM-DD_description.md)
  - [ ] Save to darci-brainstorm/

- [ ] Implement `notebook_update` tool
  - [ ] Find notebook by path
  - [ ] Append to section
  - [ ] Update timestamp

- [ ] Create notebook templates
  - [ ] Default template
  - [ ] Migration template
  - [ ] Build fix template
  - [ ] Feature implementation template

### 3.2 Status Reports

- [ ] Implement `status_report_generate` tool
  - [ ] Collect task statistics
  - [ ] Collect build metrics
  - [ ] Generate markdown report
  - [ ] Save to artifacts/

- [ ] Create report templates
  - [ ] Daily report
  - [ ] Weekly report
  - [ ] Feature completion report

### 3.3 Feature Matrices

- [ ] Implement `feature_matrix_generate` tool
  - [ ] Define feature list
  - [ ] Check Python implementation
  - [ ] Check Go implementation
  - [ ] Generate comparison table
  - [ ] Calculate completion %

- [ ] Create feature checklist for Telegram
  - [ ] Long polling
  - [ ] Command handlers
  - [ ] Message handling
  - [ ] Media download
  - [ ] Gemini audio analysis
  - [ ] Typing indicators
  - [ ] Message reactions
  - [ ] Media group buffering
  - [ ] User tracking
  - [ ] Markdown → HTML
  - [ ] Message splitting
  - [ ] Reply support
  - [ ] Proxy support
  - [ ] TTS/Voice reply

---

## Phase 4: Communication 💬

### 4.1 Channel Setup

- [ ] Configure Telegram channel
  - [ ] Get bot token from @BotFather
  - [ ] Add to config.json
  - [ ] Test with `/start` command

- [ ] Configure Discord channel (optional)
  - [ ] Create Discord application
  - [ ] Enable Message Content intent
  - [ ] Get bot token
  - [ ] Add to config.json

- [ ] Configure Slack channel (optional)
  - [ ] Create Slack application
  - [ ] Enable Socket Mode
  - [ ] Get tokens
  - [ ] Add to config.json

### 4.2 Notification Tools

- [ ] Implement `notify_telegram` tool
  - [ ] Format message (Markdown/HTML)
  - [ ] Send via Telegram API
  - [ ] Handle errors
  - [ ] Return message ID

- [ ] Implement `notify_discord` tool
  - [ ] Create embed
  - [ ] Send via Discord API
  - [ ] Handle errors

- [ ] Implement `notify_slack` tool
  - [ ] Create blocks
  - [ ] Send via Slack API
  - [ ] Handle errors

- [ ] Implement `notification_broadcast` tool
  - [ ] Route to multiple channels
  - [ ] Format per channel
  - [ ] Aggregate results

### 4.3 Notification Routing

- [ ] Create routing rules
  ```yaml
  P0: [telegram, slack, discord]  # Critical
  P1: [telegram, slack]            # High
  P2: [slack]                      # Normal
  P3: []                           # Low (log only)
  ```

- [ ] Implement priority-based routing
- [ ] Implement type-based routing
  - Build failures → all channels
  - Task completions → Slack only
  - Status updates → scheduled

---

## Phase 5: Code Analysis 🔍

### 5.1 Comparison Tools

- [ ] Implement `code_compare` tool
  - [ ] Load source file
  - [ ] Load target file
  - [ ] Compute diff
  - [ ] Calculate similarity score
  - [ ] Identify missing features

- [ ] Implement `code_analyze` tool
  - [ ] Parse code (AST for Go/Python)
  - [ ] Check code quality metrics
  - [ ] Identify issues
  - [ ] Suggest improvements

### 5.2 Import Analysis

- [ ] Implement `import_analyzer` tool
  - [ ] Parse Go imports
  - [ ] Parse Python imports
  - [ ] Detect incorrect paths
  - [ ] Suggest fixes
  - [ ] Optional: auto-fix

### 5.3 Feature Parity Tracking

- [ ] Create feature registry
  ```json
  {
    "telegram": {
      "features": [
        {"name": "long_polling", "python": true, "go": false},
        {"name": "command_handlers", "python": true, "go": false}
      ]
    }
  }
  ```

- [ ] Implement parity checker
  - [ ] Scan Python code for features
  - [ ] Scan Go code for features
  - [ ] Update registry
  - [ ] Generate report

---

## Phase 6: Git Integration 🔧

### 6.1 Git Tools

- [ ] Implement `git_status` tool
  - [ ] Run `git status`
  - [ ] Parse staged changes
  - [ ] Parse unstaged changes
  - [ ] Parse untracked files

- [ ] Implement `git_commit` tool
  - [ ] Stage files
  - [ ] Create commit message
  - [ ] Run `git commit`
  - [ ] Return commit hash

- [ ] Implement `git_diff` tool
  - [ ] Run `git diff`
  - [ ] Parse diff output
  - [ ] Count changes

- [ ] Implement `git_log` tool
  - [ ] Run `git log`
  - [ ] Parse commits
  - [ ] Format output

### 6.2 Auto-Commit Workflows

- [ ] Create commit message generator
  - [ ] Analyze changes
  - [ ] Generate conventional commit
  - [ ] Include task references

- [ ] Implement auto-commit on task complete
  - [ ] Stage related files
  - [ ] Create commit
  - [ ] Update task with commit hash

---

## Phase 7: MCP Integration 🔌

### 7.1 MCP Tools

- [ ] Implement `mcp_list_servers` tool
  - [ ] Read config.json
  - [ ] List configured servers
  - [ ] Check connection status

- [ ] Implement `mcp_list_tools` tool
  - [ ] Query MCP servers
  - [ ] List available tools
  - [ ] Cache tool definitions

- [ ] Implement `mcp_call_tool` tool
  - [ ] Validate tool exists
  - [ ] Call MCP server
  - [ ] Return result

### 7.2 External Integrations

- [ ] Add GitHub MCP server
  - [ ] Configure in config.json
  - [ ] Test tool calls
  - [ ] Use for issue/PR management

- [ ] Add filesystem MCP server
  - [ ] Configure workspace access
  - [ ] Test file operations

---

## Phase 8: Memory System 🧠

### 8.1 Memory Layers

- [ ] Implement short-term memory
  - [ ] Conversation buffer (last 50 messages)
  - [ ] Session context

- [ ] Implement working memory
  - [ ] Active tasks
  - [ ] Current focus
  - [ ] Recent actions

- [ ] Implement long-term memory
  - [ ] Session history
  - [ ] Pattern storage
  - [ ] Learnings

- [ ] Implement semantic memory
  - [ ] Project structure
  - [ ] Dependencies
  - [ ] Relationships

### 8.2 Memory Operations

- [ ] Implement `memory_store` operation
  - [ ] Encode experience
  - [ ] Store in appropriate layer
  - [ ] Index for retrieval

- [ ] Implement `memory_retrieve` operation
  - [ ] Query by context
  - [ ] Retrieve relevant memories
  - [ ] Rank by relevance

---

## Phase 9: Sub-Agent System 🤖

### 9.1 Sub-Agent Definitions

- [ ] Define Project Manager agent
  - [ ] System prompt
  - [ ] Tool access
  - [ ] Specialized prompts

- [ ] define Build Engineer agent
  - [ ] System prompt
  - [ ] Tool access
  - [ ] Error handling strategies

- [ ] Define Code Analyst agent
  - [ ] System prompt
  - [ ] Tool access
  - [ ] Analysis frameworks

- [ ] define Scribe agent
  - [ ] System prompt
  - [ ] Tool access
  - [ ] Documentation standards

### 9.2 Agent Coordination

- [ ] Implement task routing
  - [ ] Analyze task type
  - [ ] Route to appropriate agent
  - [ ] Provide context

- [ ] Implement agent handoff
  - [ ] Pass context between agents
  - [ ] Maintain conversation continuity

---

## Phase 10: Testing & Validation ✅

### 10.1 Unit Tests

- [ ] Test task management tools
  - [ ] task_create
  - [ ] task_update
  - [ ] task_query
  - [ ] task_list

- [ ] Test build tools
  - [ ] build_run
  - [ ] test_run
  - [ ] lint_run

- [ ] Test documentation tools
  - [ ] notebook_create
  - [ ] notebook_update
  - [ ] status_report_generate

### 10.2 Integration Tests

- [ ] Test end-to-end workflows
  - [ ] Feature development workflow
  - [ ] Build monitoring workflow
  - [ ] Migration tracking workflow

- [ ] Test channel integrations
  - [ ] Telegram notifications
  - [ ] Discord notifications
  - [ ] Slack notifications

### 10.3 User Acceptance Testing

- [ ] Test with real projects
  - [ ] Go Telegram migration
  - [ ] Python feature development
  - [ ] Documentation generation

- [ ] Gather feedback
  - [ ] Usability
  - [ ] Accuracy
  - [ ] Performance

---

## Phase 11: Deployment & Operations 🚀

### 11.1 Local Deployment

- [ ] Create startup script
  ```bash
  #!/bin/bash
  cd ~/repos/sentinelai/darci/darci-python
  source .venv/bin/activate
  darci gateway
  ```

- [ ] Create systemd service (Linux)
  ```ini
  [Unit]
  Description=DarCI Agent
  After=network.target

  [Service]
  ExecStart=/path/to/darci gateway
  Restart=always
  User=darcy

  [Install]
  WantedBy=multi-user.target
  ```

### 11.2 Docker Deployment

- [ ] Create Dockerfile
  ```dockerfile
  FROM python:3.11-slim
  WORKDIR /app
  COPY . .
  RUN pip install -e .
  CMD ["darci", "gateway"]
  ```

- [ ] Create docker-compose.yml
  ```yaml
  version: '3.8'
  services:
    darci:
      build: .
      volumes:
        - ~/.darci:/root/.darci
        - ~/repos:/workspace
      environment:
        - GEMINI_API_KEY=${GEMINI_API_KEY}
  ```

### 11.3 Monitoring

- [ ] Set up health checks
  - [ ] Agent responsiveness
  - [ ] Channel connectivity
  - [ ] Build system access

- [ ] Create monitoring dashboard
  - [ ] Task velocity
  - [ ] Build success rate
  - [ ] Response times

---

## Completion Criteria

### Phase 1-3 (MVP) ✅
- [ ] Can create and track tasks
- [ ] Can run builds and tests
- [ ] Can generate engineering notebooks
- [ ] Can report status

### Phase 4-6 (Advanced) 🔨
- [ ] Multi-channel notifications
- [ ] Code analysis and comparison
- [ ] Git integration
- [ ] Auto-fix common errors

### Phase 7-11 (Complete) 🎯
- [ ] MCP integration
- [ ] Memory system
- [ ] Sub-agent coordination
- [ ] Production deployment

---

## Progress Tracking

```markdown
## Current Phase: 1 (Foundation)

### Completed
- [x] Environment setup
- [x] Directory structure
- [ ] Task management tools

### In Progress
- [ ] task_create implementation

### Next Up
- [ ] task_update implementation

### Blockers
- None
```

---

*Checklist version: 1.0*
*Last updated: 2026-03-07*
*Update as you progress through implementation!*
