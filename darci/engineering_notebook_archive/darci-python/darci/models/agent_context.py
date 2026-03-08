from dataclasses import dataclass, field
from datetime import datetime, timezone
from typing import Optional


@dataclass
class AgentContextSnapshot:
    node_name: str
    ip: str
    online: bool
    current_goal: Optional[str] = None
    task_id: Optional[str] = None
    sentinel_url: Optional[str] = None    # ws://ip:port/ws/run
    darci_url: Optional[str] = None    # http://ip:18790
    risk_score: float = 0.0
    failure_types: list = field(default_factory=list)
    darci_role: str = "responsible"
    last_seen: str = field(default_factory=lambda: datetime.now(timezone.utc).isoformat())
