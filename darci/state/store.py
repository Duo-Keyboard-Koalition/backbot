"""JSON file-backed task store."""
import json
import dataclasses
from datetime import datetime, timezone
from pathlib import Path

from darci.config import DarciConfig
from darci.models.task import Task, DarciRoles, SentinelSnapshot


def _now() -> str:
    return datetime.now(timezone.utc).isoformat()


class TaskStore:
    def __init__(self, config: DarciConfig):
        self._dir = config.state_dir
        self._dir.mkdir(parents=True, exist_ok=True)
        self._tasks_path = self._dir / "tasks.json"
        self._context_path = self._dir / "context.json"
        self._counter_path = self._dir / "counter.txt"
        self._ensure_defaults()

    def _ensure_defaults(self):
        if not self._tasks_path.exists():
            self._tasks_path.write_text(json.dumps({}, indent=2))
        if not self._context_path.exists():
            self._context_path.write_text(json.dumps({
                "agent_assignments": {},
                "darci_state": {"active_monitors": [], "last_discovery": None}
            }, indent=2))
        if not self._counter_path.exists():
            self._counter_path.write_text("0")

    def _next_id(self) -> str:
        n = int(self._counter_path.read_text().strip()) + 1
        self._counter_path.write_text(str(n))
        return f"T{n:03d}"

    # --- Tasks ---

    def _read_tasks(self) -> dict:
        return json.loads(self._tasks_path.read_text())

    def _write_tasks(self, tasks: dict):
        self._tasks_path.write_text(json.dumps(tasks, indent=2))

    def create(self, task: Task) -> Task:
        tasks = self._read_tasks()
        tasks[task.id] = dataclasses.asdict(task)
        self._write_tasks(tasks)
        return task

    def create_new(self, title: str, description: str = "", priority: str = "P2",
                   labels: list = None, dependencies: list = None) -> Task:
        task = Task(
            id=self._next_id(),
            title=title,
            description=description,
            priority=priority,
            labels=labels or [],
            dependencies=dependencies or [],
        )
        return self.create(task)

    def update(self, task_id: str, **fields) -> Task | None:
        tasks = self._read_tasks()
        if task_id not in tasks:
            return None
        # Handle nested darci and sentinel_snapshot updates
        for key, val in fields.items():
            if key in ("darci", "sentinel_snapshot") and isinstance(val, dict):
                tasks[task_id].setdefault(key, {}).update(val)
            else:
                tasks[task_id][key] = val
        tasks[task_id]["updated_at"] = _now()
        self._write_tasks(tasks)
        return self._from_dict(tasks[task_id])

    def get(self, task_id: str) -> Task | None:
        tasks = self._read_tasks()
        return self._from_dict(tasks[task_id]) if task_id in tasks else None

    def query(self, status: str = None, priority: str = None, label: str = None) -> list[Task]:
        tasks = self._read_tasks()
        results = [self._from_dict(t) for t in tasks.values()]
        if status:
            results = [t for t in results if t.status == status]
        if priority:
            results = [t for t in results if t.priority == priority]
        if label:
            results = [t for t in results if label in t.labels]
        return sorted(results, key=lambda t: (t.priority, t.created_at))

    def all(self) -> list[Task]:
        return self.query()

    def _from_dict(self, d: dict) -> Task:
        d = dict(d)
        darci_data = d.pop("darci", {})
        snap_data = d.pop("sentinel_snapshot", {})
        d["darci"] = DarciRoles(**darci_data) if darci_data else DarciRoles()
        d["sentinel_snapshot"] = SentinelSnapshot(**snap_data) if snap_data else SentinelSnapshot()
        return Task(**d)

    # --- Context ---

    def get_context(self) -> dict:
        return json.loads(self._context_path.read_text())

    def update_context(self, **kwargs):
        ctx = self.get_context()
        ctx.update(kwargs)
        self._context_path.write_text(json.dumps(ctx, indent=2))

    def set_agent_assignment(self, node_name: str, task_id: str, role: str,
                              risk_score: float = 0.0, status: str = "pending"):
        ctx = self.get_context()
        ctx.setdefault("agent_assignments", {})[node_name] = {
            "task_id": task_id,
            "darci_role": role,
            "risk_score": risk_score,
            "status": status,
        }
        self._context_path.write_text(json.dumps(ctx, indent=2))
