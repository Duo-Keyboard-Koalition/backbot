"""Tailbridge (taila2a) integration tools for DarCI."""
from datetime import datetime, timezone
from typing import Any

import httpx
from darci.agent.tools.base import Tool

from darci.config import DarciConfig
from darci.state.store import TaskStore


class DiscoverAgentsTool(Tool):
    def __init__(self, config: DarciConfig, store: TaskStore):
        self._config = config
        self._store = store

    @property
    def name(self) -> str:
        return "discover_agents"

    @property
    def description(self) -> str:
        return (
            "Discover all agents currently online on the Tailscale tailnet via taila2a. "
            "Updates the internal agent registry. Call this before assigning tasks."
        )

    @property
    def parameters(self) -> dict[str, Any]:
        return {"type": "object", "properties": {}}

    async def execute(self, **kwargs) -> str:
        url = f"{self._config.bridge_local_url}/agents"
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                resp = await client.get(url)
                resp.raise_for_status()
                data = resp.json()
        except httpx.ConnectError:
            return f"Error: taila2a bridge not reachable at {url}. Is it running?"
        except httpx.HTTPError as e:
            return f"Error: {e}"

        agents = data.get("agents", [])
        now = datetime.now(timezone.utc).isoformat()
        self._store.update_context(**{
            "darci_state": {
                "last_discovery": now,
                "active_monitors": self._store.get_context()
                    .get("darci_state", {}).get("active_monitors", [])
            }
        })

        if not agents:
            return "No agents discovered on the tailnet."

        lines = [f"Discovered {len(agents)} agent(s):\n",
                 "| Name | IP | Online | Services |",
                 "|---|---|---|---|"]
        for agent in agents:
            name = agent.get("name", "?")
            ip = agent.get("ip", "?")
            online = "✅" if agent.get("online") else "❌"
            gateways = agent.get("gateways", [])
            services = ", ".join(f"{g['service']}:{g['port']}" for g in gateways) or "none"
            lines.append(f"| {name} | {ip} | {online} | {services} |")

        return "\n".join(lines)


class SendDarciMessageTool(Tool):
    def __init__(self, config: DarciConfig):
        self._config = config

    @property
    def name(self) -> str:
        return "send_darci_message"

    @property
    def description(self) -> str:
        return (
            "Send a DARCI message to a worker agent (openclaw, nanobot, sclaw) via tailbridge. "
            "Use message_type='darci_directive' to assign or correct work, "
            "or 'darci_status_request' to ask what the agent is doing."
        )

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "dest_node": {"type": "string", "description": "Tailnet node name of the target agent"},
                "message_type": {
                    "type": "string",
                    "enum": ["darci_directive", "darci_status_request"],
                },
                "payload": {
                    "type": "object",
                    "description": "Message payload (task_id, goal, priority, etc.)",
                },
            },
            "required": ["dest_node", "message_type", "payload"],
        }

    async def execute(self, dest_node: str, message_type: str, payload: dict, **kwargs) -> str:
        url = f"{self._config.bridge_local_url}/send"
        envelope = {
            "dest_node": dest_node,
            "payload": {**payload, "type": message_type},
        }
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                resp = await client.post(url, json=envelope)
                resp.raise_for_status()
        except httpx.ConnectError:
            return f"Error: taila2a bridge not reachable at {url}. Is it running?"
        except httpx.HTTPError as e:
            return f"Error sending to {dest_node}: {e}"

        return f"Message sent to {dest_node} (type: {message_type})"
