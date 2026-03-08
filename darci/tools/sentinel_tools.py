"""Sentinel WebSocket risk monitoring tools for DarCI."""
import asyncio
import json
from typing import Any

from scorpion.agent.tools.base import Tool
from scorpion.bus.events import InboundMessage

from darci.config import DarciConfig
from darci.state.store import TaskStore


class SentinelMonitorRegistry:
    """Tracks running asyncio.Tasks for sentinel monitoring (one per agent)."""

    def __init__(self):
        self._tasks: dict[str, asyncio.Task] = {}

    def is_monitoring(self, node_name: str) -> bool:
        t = self._tasks.get(node_name)
        return t is not None and not t.done()

    def start(self, node_name: str, coro) -> asyncio.Task:
        task = asyncio.create_task(coro, name=f"sentinel-{node_name}")
        self._tasks[node_name] = task
        return task

    def stop(self, node_name: str):
        t = self._tasks.pop(node_name, None)
        if t and not t.done():
            t.cancel()

    def active(self) -> list[str]:
        return [n for n, t in self._tasks.items() if not t.done()]


async def _monitor_loop(
    node_name: str,
    sentinel_url: str,
    goal: str,
    task_id: str,
    api_key: str,
    store: TaskStore,
    bus,  # MessageBus
):
    """Persistent WebSocket monitor for a single agent's Sentinel stream."""
    try:
        import websockets
    except ImportError:
        # Publish error back to agent loop
        await bus.publish_inbound(InboundMessage(
            channel="system",
            sender_id="sentinel-monitor",
            chat_id="darci:direct",
            content=f"[ERROR] websockets package not installed. Run: pip install websockets",
        ))
        return

    try:
        async with websockets.connect(sentinel_url) as ws:
            await ws.send(json.dumps({
                "goal": goal,
                "api_key": api_key or "",
                "max_steps": 50,
            }))

            async for raw in ws:
                try:
                    event = json.loads(raw)
                except json.JSONDecodeError:
                    continue

                evt_type = event.get("type")

                if evt_type == "step":
                    risk = float(event.get("risk_score", 0.0))
                    failure_types = event.get("failure_types", [])
                    step_num = event.get("step", {}).get("step_number", 0)

                    store.update(task_id, sentinel_snapshot={
                        "risk_score": risk,
                        "failure_types": failure_types,
                        "step_count": step_num,
                    })
                    store.set_agent_assignment(node_name, task_id, "responsible", risk_score=risk)

                    if risk >= 0.5:
                        store.update(task_id, status="at_risk")
                        await bus.publish_inbound(InboundMessage(
                            channel="system",
                            sender_id="sentinel-monitor",
                            chat_id="darci:direct",
                            content=(
                                f"[RISK ALERT] Agent '{node_name}' on task {task_id}: "
                                f"risk_score={risk:.2f}, failures={failure_types}. "
                                f"As Driver, send a darci_directive to refocus the agent on its goal."
                            ),
                        ))

                elif evt_type == "intervention":
                    intervention = event.get("intervention", {})
                    itype = intervention.get("intervention_type", "")
                    store.update(task_id, sentinel_snapshot={"intervention_type": itype})

                    if itype == "HALT":
                        store.update(task_id, status="blocked")
                        await bus.publish_inbound(InboundMessage(
                            channel="system",
                            sender_id="sentinel-monitor",
                            chat_id="darci:direct",
                            content=(
                                f"[HALT] Agent '{node_name}' task {task_id} has been halted by Sentinel "
                                f"(Approver veto). Task is now blocked. Notify the user and create a "
                                f"notebook entry documenting this intervention."
                            ),
                        ))
                    else:
                        await bus.publish_inbound(InboundMessage(
                            channel="system",
                            sender_id="sentinel-monitor",
                            chat_id="darci:direct",
                            content=(
                                f"[INTERVENTION] Agent '{node_name}' task {task_id}: "
                                f"Sentinel issued {itype}. Monitor continues."
                            ),
                        ))

                elif evt_type in ("complete", "timeout"):
                    store.update(task_id, status="completed")
                    await bus.publish_inbound(InboundMessage(
                        channel="system",
                        sender_id="sentinel-monitor",
                        chat_id="darci:direct",
                        content=(
                            f"[COMPLETE] Agent '{node_name}' task {task_id} has finished "
                            f"({'completed' if evt_type == 'complete' else 'timed out'}). "
                            f"Create a notebook entry to document the session."
                        ),
                    ))
                    break

    except asyncio.CancelledError:
        pass  # Normal shutdown
    except Exception as e:
        await bus.publish_inbound(InboundMessage(
            channel="system",
            sender_id="sentinel-monitor",
            chat_id="darci:direct",
            content=f"[ERROR] Sentinel monitor for '{node_name}' failed: {e}",
        ))


class MonitorAgentTool(Tool):
    def __init__(self, config: DarciConfig, store: TaskStore, bus, registry: SentinelMonitorRegistry):
        self._config = config
        self._store = store
        self._bus = bus
        self._registry = registry

    @property
    def name(self) -> str:
        return "monitor_agent"

    @property
    def description(self) -> str:
        return (
            "Start monitoring a Responsible agent's Sentinel risk stream. "
            "Runs in the background — alerts you automatically when risk_score >= 0.5 (Approver signal) "
            "or when a HALT intervention fires. Requires a task_id to track state."
        )

    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "node_name": {"type": "string", "description": "Tailnet name of the agent to monitor"},
                "sentinel_url": {
                    "type": "string",
                    "description": "WebSocket URL of the agent's Sentinel backend (e.g. ws://192.168.x.x:8000/ws/run)",
                },
                "goal": {"type": "string", "description": "The agent's current goal to send to Sentinel"},
                "task_id": {"type": "string", "description": "DarCI task ID this agent is working on"},
                "api_key": {"type": "string", "description": "Optional Gemini API key for the Sentinel backend"},
            },
            "required": ["node_name", "sentinel_url", "goal", "task_id"],
        }

    async def execute(self, node_name: str, sentinel_url: str, goal: str,
                      task_id: str, api_key: str = "", **kwargs) -> str:
        if self._registry.is_monitoring(node_name):
            return f"Already monitoring '{node_name}'. Use monitor_agent with a different node, or wait."

        task = self._store.get(task_id)
        if not task:
            return f"Error: task {task_id} not found. Create it first with task_create."

        self._registry.start(
            node_name,
            _monitor_loop(node_name, sentinel_url, goal, task_id, api_key, self._store, self._bus),
        )

        active = self._registry.active()
        return (
            f"Monitoring started for '{node_name}' on task {task_id}.\n"
            f"Sentinel URL: {sentinel_url}\n"
            f"Active monitors: {active}"
        )
