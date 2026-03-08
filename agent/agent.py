import os
import json
import re
import asyncio
import time
from typing import Any, Dict, List, Optional
from backboard import BackboardClient
from .core import Tool, ToolParameter, ToolParameterType, ToolCall, AgentResponse, Invocation
from .loop import process_agent_invocation

class Agent:
    """
    Python Agent integrated with Backboard SDK.
    Uses Assistant and Thread pattern for persistent conversations.
    """
    
    def __init__(
        self,
        name: str,
        instructions: str = "You are a helpful AI assistant.",
        api_key: Optional[str] = None,
        model: str = "gemini-2.0-flash",
        base_url: str = "https://api.backboard.ai",
        gateway_url: str = "http://localhost:18789"
    ):
        self.name = name
        self.instructions = instructions
        self.api_key = api_key
        self.model = model
        self.base_url = base_url
        self.gateway_url = gateway_url
        self.tools: Dict[str, Tool] = {}
        self._tool_call_counter = 0
        
        # SDK components (initialized lazily)
        self.client: Optional[BackboardClient] = None
        self.assistant_id: Optional[str] = None
        self.thread_id: Optional[str] = None
        
        self._register_default_tools()
    
    def _register_default_tools(self):
        def calculator(expression: str) -> str:
            try:
                allowed_chars = set("0123456789+-*/.() ")
                if not all(c in allowed_chars for c in expression):
                    return "Error: Invalid characters"
                result = eval(expression, {"__builtins__": {}}, {})
                return str(result)
            except Exception as e:
                return f"Error: {str(e)}"
        
        self.register_tool(
            name="calculator",
            description="Evaluate math expressions",
            parameters=[ToolParameter("expression", ToolParameterType.STRING, "Math expression")],
            handler=calculator
        )

    def register_tool(self, name, description, parameters, handler):
        tool = Tool(name=name, description=description, parameters=parameters, handler=handler)
        self.tools[name] = tool
        return tool

    async def _ensure_initialized(self):
        """Initialize SDK client, assistant, and thread if not already done"""
        if not self.client and self.api_key:
            self.client = BackboardClient(api_key=self.api_key)
            
            # Create assistant
            assistant = await self.client.create_assistant(
                name=self.name,
                model=self.model,
                system_prompt=self._generate_system_prompt()
            )
            self.assistant_id = assistant.assistant_id
            
            # Create default thread
            thread = await self.client.create_thread(self.assistant_id)
            self.thread_id = thread.thread_id

    def _generate_system_prompt(self) -> str:
        return f"{self.instructions}"

    def _parse_tool_calls(self, content: str) -> List[ToolCall]:
        tool_calls = []
        pattern = r'<tool:(\w+)>(\{[^}]+\})'
        matches = re.findall(pattern, content)
        for name, args_json in matches:
            if name in self.tools:
                try:
                    args = json.loads(args_json)
                    self._tool_call_counter += 1
                    tool_calls.append(ToolCall(name=name, arguments=args, call_id=f"call_{self._tool_call_counter}"))
                except json.JSONDecodeError: pass
        return tool_calls

    async def _execute_tool(self, tool_call: ToolCall) -> str:
        if tool_call.name not in self.tools:
            return f"Error: Unknown tool '{tool_call.name}'"
        tool = self.tools[tool_call.name]
        try:
            if asyncio.iscoroutinefunction(tool.handler):
                result = await tool.handler(**tool_call.arguments)
            else:
                result = tool.handler(**tool_call.arguments)
            return str(result)
        except Exception as e:
            return f"Error executing {tool_call.name}: {str(e)}"

    async def invoke(self, task: str, context: Optional[Dict] = None) -> AgentResponse:
        invocation = Invocation(id=f"inv_{int(time.time())}", task=task, context=context or {})
        return await process_agent_invocation(self, invocation)
