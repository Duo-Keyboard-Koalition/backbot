import importlib
import inspect
import os
from typing import Dict, List, Any
from ..core import Tool

class ToolManager:
    def __init__(self, agent=None):
        self.agent = agent
        self.tools: Dict[str, Tool] = {}
        self._load_tools()
        self._register_document_tools()

    def _register_document_tools(self):
        """Register tools that require the agent instance"""
        if self.agent:
            try:
                from .document import get_document_tools
                for t_args in get_document_tools(self.agent):
                    # get_document_tools returns list of dicts with name, description, parameters, handler
                    tool = Tool(**t_args)
                    self.register_tool(tool)
            except (ImportError, AttributeError):
                pass

    def _load_tools(self):
        """Automatically load tools from the current directory"""
        tools_dir = os.path.dirname(__file__)
        for filename in os.listdir(tools_dir):
            if filename.endswith(".py") and filename != "__init__.py" and filename != "manager.py":
                module_name = f".{filename[:-3]}"
                try:
                    module = importlib.import_module(module_name, package="agent.tools")
                    # Look for objects ending with _tool or named Tool
                    for name, obj in inspect.getmembers(module):
                        if isinstance(obj, Tool):
                            self.register_tool(obj)
                except Exception as e:
                    print(f"Error loading tool module {module_name}: {e}")

    def register_tool(self, tool: Tool):
        self.tools[tool.name] = tool

    def get_tool(self, name: str) -> Tool:
        return self.tools.get(name)

    def get_schemas(self) -> List[Dict[str, Any]]:
        """Return tool schemas in Backboard SDK format"""
        return [{"type": "function", "function": tool.to_schema()} for tool in self.tools.values()]

    async def execute(self, name: str, arguments: Dict[str, Any]) -> str:
        tool = self.get_tool(name)
        if not tool:
            return f"Error: Tool '{name}' not found."
        
        try:
            import asyncio
            if asyncio.iscoroutinefunction(tool.handler):
                result = await tool.handler(**arguments)
            else:
                result = tool.handler(**arguments)
            return str(result)
        except Exception as e:
            return f"Error executing {name}: {str(e)}"
