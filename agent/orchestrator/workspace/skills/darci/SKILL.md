---
name: darci
description: DARCI project management — task tracking, agent discovery, risk monitoring, notebook generation
always: true
---

# DarCI Tools Reference

## Task Management

- `task_create(title, description?, priority?, labels?, dependencies?)` → creates task, returns ID
- `task_update(task_id, status?, priority?, description?)` → updates task fields
- `task_query(status?, priority?, label?)` → returns markdown table of matching tasks
- `assign_task(task_id, node_name, goal_description)` → assigns to Responsible agent + sends darci_directive
- `status_report()` → full DARCI project board grouped by status

## Agent Communication (via tailbridge/taila2a)

- `discover_agents()` → lists all online tailnet agents with services
- `send_darci_message(dest_node, message_type, payload)` → sends envelope via taila2a
  - `message_type`: `"darci_directive"` | `"darci_status_request"`
  - `payload`: dict with task_id, goal, priority, etc.

## Risk Monitoring (Sentinel)

- `monitor_agent(node_name, sentinel_url, goal, task_id, api_key?)` → starts background WS monitor
  - Auto-alerts when risk_score >= 0.5 (Approver signal → send corrective directive)
  - Auto-marks task "blocked" on HALT intervention (Approver veto)
  - sentinel_url format: `ws://ip:port/ws/run`

## Engineering Notebook

- `notebook_create(topic, objective, key_findings?, decisions?, next_steps?)` → creates dated entry
- `notebook_append(filepath, section, content)` → appends section to existing file
- Notebooks saved to: `darci/engineering_notebook/YYYY-MM-DD_{topic}.md`

## DARCI Roles Quick Reference

| Role | Who | Trigger |
|---|---|---|
| Driver | DarCI (you) | Owns tasks, assigns work |
| Approver | Sentinel backend | risk ≥ 0.5 or HALT |
| Responsible | openclaw / nanobot / sclaw | Receives darci_directive |
| Consulted | Spawned subagents | Analysis tasks |
| Informed | Engineering notebook | Every state change |
