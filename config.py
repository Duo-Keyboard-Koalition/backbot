import json
import os
from pathlib import Path

DEFAULT_CONFIG = {
    "gateway": {
        "mode": "local",
        "port": 18789,
        "host": "127.0.0.1"
    }
}

def get_config_path() -> Path:
    home = Path.home()
    config_dir = home / ".openclaw"
    config_dir.mkdir(exist_ok=True)
    return config_dir / "openclaw.json"

def load_config() -> dict:
    path = get_config_path()
    if path.exists():
        try:
            with open(path, "r") as f:
                config = json.load(f)
                # Merge with defaults
                merged = DEFAULT_CONFIG.copy()
                for k, v in config.items():
                    if isinstance(v, dict) and k in merged:
                        merged[k].update(v)
                    else:
                        merged[k] = v
                return merged
        except Exception:
            return DEFAULT_CONFIG
    else:
        # Create default config
        save_config(DEFAULT_CONFIG)
        return DEFAULT_CONFIG

def save_config(config: dict):
    path = get_config_path()
    with open(path, "w") as f:
        json.dump(config, f, indent=2)
