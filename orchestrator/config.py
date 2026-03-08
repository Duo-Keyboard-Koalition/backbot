from dataclasses import dataclass, field
from pathlib import Path


@dataclass
class DarciConfig:
    bridge_local_url: str = "http://localhost:8080"   # taila2a local API
    sentinel_port: int = 8000                          # sentinel backend port
    state_dir: Path = field(default_factory=lambda: Path("~/.darci/darci").expanduser())
    notebook_dir: Path = field(default_factory=lambda: Path("darci/engineering_notebook"))
    discovery_interval_s: int = 30
