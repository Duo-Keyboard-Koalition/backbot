"""Task management tools for DarCI."""
from typing import Any

from darci.agent.tools.base import Tool

from orchestrator.state.store import TaskStore


class TaskCreateTool(Tool):
    def __init__(self, store: TaskStore):
        self._store = store

    @property
    def name(self) -> str:
        return "task_create"

    @property
    def description(self) -> str:
        return (
            "Create a new tracked task in DarCI. Returns the task ID and summary. "
            "Use this before assigning or monitoring any work."
        )

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "title": {"type": "string", "description": "Short task title"},
                "description": {"type": "string", "description": "Detailed description"},
                "priority": {
                    "type": "string",
                    "enum": ["P0", "P1", "P2", "P3"],
                    "description": "P0=critical, P1=high, P2=normal, P3=low",
                },
                "labels": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Optional tags (e.g. ['sentinel', 'tailbridge'])",
                },
                "dependencies": {
                    "type": "array",
                    "items": {"type": "string"},
                    "description": "Task IDs this task depends on",
                },
            },
            "required": ["title"],
        }

    async def execute(self, title: str, description: str = "", priority: str = "P2",
                      labels: list = None, dependencies: list = None, **kwargs) -> str:
        task = self._store.create_new(
            title=title,
            description=description,
            priority=priority,
            labels=labels or [],
            dependencies=dependencies or [],
        )
        return (
            f"Task created: {task.id}\n"
            f"Title: {task.title}\n"
            f"Priority: {task.priority} | Status: {task.status}\n"
            f"Driver: {task.darci.driver} | Approver: {task.darci.approver}\n"
            f"Labels: {task.labels or 'none'}"
        )


class TaskUpdateTool(Tool):
    def __init__(self, store: TaskStore):
        self._store = store

    @property
    def name(self) -> str:
        return "task_update"

    @property
    def description(self) -> str:
        return "Update a task's status, priority, or description."

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "task_id": {"type": "string", "description": "Task ID (e.g. T001)"},
                "status": {
                    "type": "string",
                    "enum": ["pending", "in_progress", "at_risk", "blocked", "completed"],
                },
                "priority": {"type": "string", "enum": ["P0", "P1", "P2", "P3"]},
                "description": {"type": "string"},
            },
            "required": ["task_id"],
        }

    async def execute(self, task_id: str, status: str = None, priority: str = None,
                      description: str = None, **kwargs) -> str:
        fields = {}
        if status:
            fields["status"] = status
        if priority:
            fields["priority"] = priority
        if description:
            fields["description"] = description
        if not fields:
            return f"Error: no fields to update for {task_id}"
        task = self._store.update(task_id, **fields)
        if not task:
            return f"Error: task {task_id} not found"
        return f"Updated {task_id}: {', '.join(f'{k}={v}' for k, v in fields.items())}"


class TaskQueryTool(Tool):
    def __init__(self, store: TaskStore):
        self._store = store

    @property
    def name(self) -> str:
        return "task_query"

    @property
    def description(self) -> str:
        return "Query tasks by status, priority, or label. Returns a markdown table."

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "status": {"type": "string", "enum": ["pending", "in_progress", "at_risk", "blocked", "completed"]},
                "priority": {"type": "string", "enum": ["P0", "P1", "P2", "P3"]},
                "label": {"type": "string", "description": "Filter by label"},
            },
        }

    async def execute(self, status: str = None, priority: str = None, label: str = None, **kwargs) -> str:
        tasks = self._store.query(status=status, priority=priority, label=label)
        if not tasks:
            return "No tasks found matching the query."
        lines = ["| ID | Title | Priority | Status | Responsible |",
                 "|---|---|---|---|---|"]
        for t in tasks:
            responsible = t.darci.responsible or "unassigned"
            lines.append(f"| {t.id} | {t.title} | {t.priority} | {t.status} | {responsible} |")
        return "\n".join(lines)


class StatusReportTool(Tool):
    def __init__(self, store: TaskStore):
        self._store = store

    @property
    def name(self) -> str:
        return "status_report"

    @property
    def description(self) -> str:
        return "Generate a full DARCI project status board grouped by status."

    @property
    def parameters(self) -> dict[str, Any]:
        return {"type": "object", "properties": {}}

    async def execute(self, **kwargs) -> str:
        tasks = self._store.all()
        if not tasks:
            return "No tasks yet. Use task_create to get started."

        groups: dict[str, list] = {
            "at_risk": [], "blocked": [], "in_progress": [],
            "pending": [], "completed": [],
        }
        for t in tasks:
            groups.setdefault(t.status, []).append(t)

        lines = ["# DarCI Project Status\n"]
        icons = {"at_risk": "⚠️", "blocked": "🛑", "in_progress": "🔄", "pending": "⏳", "completed": "✅"}
        order = ["at_risk", "blocked", "in_progress", "pending", "completed"]

        total = len(tasks)
        done = len(groups["completed"])
        lines.append(f"**Progress:** {done}/{total} tasks complete\n")

        for status in order:
            group = groups.get(status, [])
            if not group:
                continue
            icon = icons.get(status, "")
            lines.append(f"\n## {icon} {status.replace('_', ' ').title()} ({len(group)})\n")
            lines.append("| ID | Title | Priority | Responsible | Risk |")
            lines.append("|---|---|---|---|---|")
            for t in group:
                responsible = t.darci.responsible or "unassigned"
                risk = f"{t.sentinel_snapshot.risk_score:.2f}" if t.sentinel_snapshot.risk_score > 0 else "-"
                lines.append(f"| {t.id} | {t.title} | {t.priority} | {responsible} | {risk} |")

        return "\n".join(lines)


class AssignTaskTool(Tool):
    def __init__(self, store: TaskStore, send_tool):
        self._store = store
        self._send = send_tool

    @property
    def name(self) -> str:
        return "assign_task"

    @property
    def description(self) -> str:
        return (
            "Assign a task to a Responsible agent on the tailnet. "
            "Sets darci.responsible and sends a darci_directive via tailbridge."
        )

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "task_id": {"type": "string", "description": "Task ID to assign (e.g. T001)"},
                "node_name": {"type": "string", "description": "Tailnet node name of the Responsible agent"},
                "goal_description": {"type": "string", "description": "What the agent should do"},
            },
            "required": ["task_id", "node_name", "goal_description"],
        }

    async def execute(self, task_id: str, node_name: str, goal_description: str, **kwargs) -> str:
        task = self._store.get(task_id)
        if not task:
            return f"Error: task {task_id} not found"

        self._store.update(task_id, status="in_progress",
                           darci={"responsible": node_name})
        self._store.set_agent_assignment(node_name, task_id, "responsible")

        # Send directive via tailbridge
        directive_result = await self._send.execute(
            dest_node=node_name,
            message_type="darci_directive",
            payload={
                "task_id": task_id,
                "task_title": task.title,
                "goal": goal_description,
                "priority": task.priority,
            },
        )

        return (
            f"Task {task_id} assigned to {node_name}.\n"
            f"Status: in_progress | Responsible: {node_name}\n"
            f"Directive sent: {directive_result}"
        )
