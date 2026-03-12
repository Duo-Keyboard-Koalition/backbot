import json
import os
from pathlib import Path
from dotenv import load_dotenv

# Load environment variables once for the entire application
load_dotenv()

DEFAULT_CONFIG = {
    "gateway": {
        "mode": "local",
        "port": 18789,
        "host": "127.0.0.1",
        "model": "000ADI/Qwen2.5-0.5B-Instruct-Gensyn-Swarm-webbed_eager_cassowary",
        "llm_provider": "together",
        "url": "http://localhost:18789"
    },
    "api": {
        "base_url": "https://api.backboard.ai"
    },
    "workspace": {
        "root": "~"
    },
    "agent": {
        "name": "Sentinel",
        "instructions": "You are a helpful AI assistant with tool calling capabilities. You have tools to access the local file system (list_files, read_file, write_file) - all paths are relative to the project root. Use these tools to help the user explore the codebase and perform tasks as requested."
    }
}

def get_config_dir() -> Path:
    config_dir = Path.home() / ".backclaw"
    config_dir.mkdir(parents=True, exist_ok=True)
    return config_dir

def get_config_path() -> Path:
    return get_config_dir() / "config.json"

def load_config() -> dict:
    path = get_config_path()
    if path.exists():
        try:
            with open(path, "r") as f:
                return json.load(f)
        except Exception:
            return DEFAULT_CONFIG
    else:
        # Initial creation
        save_config(DEFAULT_CONFIG)
        return DEFAULT_CONFIG

def save_config(config: dict):
    path = get_config_path()
    with open(path, "w") as f:
        json.dump(config, f, indent=2)
