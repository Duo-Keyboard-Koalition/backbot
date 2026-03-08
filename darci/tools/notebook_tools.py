"""Engineering notebook generation tools for DarCI."""
import re
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

from scorpion.agent.tools.base import Tool

from darci.config import DarciConfig


def _slug(text: str) -> str:
    """Convert text to a filename-safe slug."""
    s = text.lower().strip()
    s = re.sub(r"[^\w\s-]", "", s)
    s = re.sub(r"[\s_-]+", "-", s)
    return s[:50]


class NotebookCreateTool(Tool):
    def __init__(self, config: DarciConfig):
        self._config = config

    @property
    def name(self) -> str:
        return "notebook_create"

    @property
    def description(self) -> str:
        return (
            "Create a dated engineering notebook entry. "
            "Call this when tasks are created, when risk events occur, or when work completes."
        )

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "topic": {"type": "string", "description": "Short topic name (used in filename)"},
                "objective": {"type": "string", "description": "What this session aimed to achieve"},
                "key_findings": {"type": "string", "description": "What was discovered or happened"},
                "decisions": {"type": "string", "description": "Decisions made (DARCI role assignments, interventions)"},
                "next_steps": {"type": "string", "description": "What to do next"},
            },
            "required": ["topic", "objective"],
        }

    async def execute(self, topic: str, objective: str, key_findings: str = "",
                      decisions: str = "", next_steps: str = "", **kwargs) -> str:
        now = datetime.now(timezone.utc)
        date_str = now.strftime("%Y-%m-%d")
        time_str = now.strftime("%H:%M UTC")
        filename = f"{date_str}_{_slug(topic)}.md"

        notebook_dir = self._config.notebook_dir
        if not Path(notebook_dir).is_absolute():
            notebook_dir = Path.cwd() / notebook_dir
        notebook_dir = Path(notebook_dir)
        notebook_dir.mkdir(parents=True, exist_ok=True)

        filepath = notebook_dir / filename

        # If file exists, append a new dated section instead
        if filepath.exists():
            section = f"\n\n---\n\n## Update · {date_str} {time_str}\n\n"
            section += f"**Objective:** {objective}\n\n"
            if key_findings:
                section += f"**Key Findings:**\n{key_findings}\n\n"
            if decisions:
                section += f"**Decisions:**\n{decisions}\n\n"
            if next_steps:
                section += f"**Next Steps:**\n{next_steps}\n"
            filepath.write_text(filepath.read_text() + section)
            return f"Appended to existing notebook: {filepath}"

        content = f"# Engineering Notebook — {topic}\n\n"
        content += f"**Date:** {date_str}  \n"
        content += f"**Time:** {time_str}  \n"
        content += f"**Author:** DarCI Agent  \n\n"
        content += "---\n\n"
        content += f"## Objective\n\n{objective}\n\n"

        if key_findings:
            content += f"## Key Findings\n\n{key_findings}\n\n"
        if decisions:
            content += f"## Decisions\n\n{decisions}\n\n"
        if next_steps:
            content += f"## Next Steps\n\n{next_steps}\n"

        filepath.write_text(content)
        return f"Notebook created: {filepath}"


class NotebookAppendTool(Tool):
    def __init__(self, config: DarciConfig):
        self._config = config

    @property
    def name(self) -> str:
        return "notebook_append"

    @property
    def description(self) -> str:
        return "Append a new section to an existing engineering notebook file."

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "filepath": {"type": "string", "description": "Path to the notebook file"},
                "section": {"type": "string", "description": "Section header (e.g. '## Risk Alert')"},
                "content": {"type": "string", "description": "Content to append under the section"},
            },
            "required": ["filepath", "section", "content"],
        }

    async def execute(self, filepath: str, section: str, content: str, **kwargs) -> str:
        p = Path(filepath)
        if not p.exists():
            return f"Error: file not found: {filepath}"
        now = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M UTC")
        addition = f"\n\n{section}\n\n*{now}*\n\n{content}\n"
        p.write_text(p.read_text() + addition)
        return f"Appended '{section}' to {filepath}"
