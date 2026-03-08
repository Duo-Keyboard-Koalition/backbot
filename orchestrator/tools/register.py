"""Wire all DarCI tools into the DarCI AdkAgentLoop."""
from darci.agent.loop import AdkAgentLoop

from darci.config import DarciConfig
from darci.state.store import TaskStore
from darci.tools.task_tools import (
    AssignTaskTool, StatusReportTool, TaskCreateTool, TaskQueryTool, TaskUpdateTool,
)
from darci.tools.taila2a_tools import DiscoverAgentsTool, SendDarciMessageTool
from darci.tools.sentinel_tools import MonitorAgentTool, SentinelMonitorRegistry
from darci.tools.notebook_tools import NotebookAppendTool, NotebookCreateTool


def register_darci_tools(
    loop: AdkAgentLoop,
    config: DarciConfig,
    store: TaskStore,
) -> SentinelMonitorRegistry:
    """Register all DarCI tools on the agent loop. Returns the sentinel registry."""
    registry = SentinelMonitorRegistry()
    send_tool = SendDarciMessageTool(config=config)

    tools = [
        TaskCreateTool(store=store),
        TaskUpdateTool(store=store),
        TaskQueryTool(store=store),
        StatusReportTool(store=store),
        AssignTaskTool(store=store, send_tool=send_tool),
        DiscoverAgentsTool(config=config, store=store),
        send_tool,
        MonitorAgentTool(config=config, store=store, bus=loop.bus, registry=registry),
        NotebookCreateTool(config=config),
        NotebookAppendTool(config=config),
    ]

    for tool in tools:
        loop.tools.register(tool)

    return registry
