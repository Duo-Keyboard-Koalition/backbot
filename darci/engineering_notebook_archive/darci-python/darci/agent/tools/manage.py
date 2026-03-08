"""Tool for managing background subagents."""

from typing import Any, TYPE_CHECKING

from darci.agent.tools.base import Tool

if TYPE_CHECKING:
    from darci.agent.subagent import SubagentManager


class ManageAgentsTool(Tool):
    """Tool to list or stop background subagents."""
    
    def __init__(self, manager: "SubagentManager"):
        self._manager = manager
    
    @property
    def name(self) -> str:
        return "manage_agents"
    
    @property
    def description(self) -> str:
        return (
            "Manage background subagents. Use this to list all active subagents "
            "or to stop a specific subagent. Listing provides IDs and labels. "
            "Stopping requires the task ID."
        )
    
    @property
    def parameters(self) -> dict[str, Any]:
        return {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string",
                    "enum": ["list", "stop"],
                    "description": "The action to perform: 'list' all agents or 'stop' a specific one.",
                },
                "task_id": {
                    "type": "string",
                    "description": "The ID of the subagent to stop (required for action='stop').",
                },
            },
            "required": ["action"],
        }
    
    async def execute(self, action: str, task_id: str | None = None, **kwargs: Any) -> str:
        """Execute the management action."""
        if action == "list":
            return self._manager.list_agents()
        elif action == "stop":
            if not task_id:
                return "Error: task_id is required to stop a subagent."
            success = await self._manager.stop_agent(task_id)
            if success:
                return f"Subagent [{task_id}] has been stopped."
            else:
                return f"Error: Could not find or stop subagent [{task_id}]."
        else:
            return f"Error: Unknown action '{action}'."
