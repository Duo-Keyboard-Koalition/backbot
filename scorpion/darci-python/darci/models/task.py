from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Optional


def _now() -> str:
    return datetime.now(timezone.utc).isoformat()


@dataclass
class DarciRoles:
    driver: str = "darci-agent"
    approver: str = "sentinel-backend"
    responsible: str = ""
    consulted: list = field(default_factory=list)
    informed: list = field(default_factory=lambda: ["engineering-notebook"])


@dataclass
class SentinelSnapshot:
    risk_score: float = 0.0
    failure_types: list = field(default_factory=list)
    step_count: int = 0
    intervention_type: Optional[str] = None
    captured_at: str = field(default_factory=_now)


@dataclass
class Task:
    id: str
    title: str
    description: str = ""
    priority: str = "P2"      # P0 | P1 | P2 | P3
    status: str = "pending"   # pending | in_progress | at_risk | blocked | completed
    darci: DarciRoles = field(default_factory=DarciRoles)
    sentinel_snapshot: SentinelSnapshot = field(default_factory=SentinelSnapshot)
    dependencies: list = field(default_factory=list)
    labels: list = field(default_factory=list)
    artifacts: dict = field(default_factory=dict)
    created_at: str = field(default_factory=_now)
    updated_at: str = field(default_factory=_now)
