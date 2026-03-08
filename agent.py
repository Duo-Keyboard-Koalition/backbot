"""
Backboard IO - Python Agent with Tool Calling
Integrated with official Backboard SDK
"""

import json
import re
import asyncio
import time
import threading
from typing import Any, Callable, Dict, List, Optional
from dataclasses import dataclass, field
from enum import Enum
import requests
from backboard import BackboardClient

class ToolParameterType(Enum):
    STRING = "string"
    NUMBER = "number"
    BOOLEAN = "boolean"
    OBJECT = "object"
    ARRAY = "array"

@dataclass
class ToolParameter:
    name: str
    type: ToolParameterType
    description: str
    required: bool = True

@dataclass
class Tool:
    name: str
    description: str
    parameters: List[ToolParameter]
    handler: Callable[..., Any]
    
    def to_schema(self) -> Dict:
        properties = {}
        required = []
        for param in self.parameters:
            properties[param.name] = {
                "type": param.type.value,
                "description": param.description
            }
            if param.required:
                required.append(param.name)
        return {
            "name": self.name,
            "description": self.description,
            "parameters": {
                "type": "object",
                "properties": properties,
                "required": required
            }
        }

@dataclass
class ToolCall:
    name: str
    arguments: Dict[str, Any]
    call_id: str = ""

@dataclass
class AgentResponse:
    content: str
    tool_calls: List[ToolCall] = field(default_factory=list)
    is_complete: bool = True

@dataclass
class Invocation:
    id: str
    task: str
    context: Dict[str, Any] = field(default_factory=dict)
    conversation_id: Optional[str] = None
    timestamp: float = field(default_factory=time.time)

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
        base_url: str = "https://api.backboard.ai",
        gateway_url: str = "http://localhost:18789"
    ):
        self.name = name
        self.instructions = instructions
        self.api_key = api_key
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
                system_prompt=self._generate_system_prompt()
            )
            self.assistant_id = assistant.assistant_id
            print(f"[*] Initialized Assistant: {self.assistant_id}")
            
            # Create default thread
            thread = await self.client.create_thread(self.assistant_id)
            self.thread_id = thread.thread_id
            print(f"[*] Initialized Thread: {self.thread_id}")

    def _generate_system_prompt(self) -> str:
        # Simplify system prompt to avoid INVALID_PROMPT errors
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

    async def process_invocation(self, invocation: Invocation) -> AgentResponse:
        await self._ensure_initialized()
        
        if not self.client:
            return AgentResponse(content="Error: Backboard API key not configured.")

        # If a thread_id is provided in context or invocation, we could use it
        thread_id = invocation.context.get("thread_id", self.thread_id)
        
        max_iterations = 5
        for _ in range(max_iterations):
            # Send message to thread
            response = await self.client.add_message(
                thread_id=thread_id,
                content=invocation.task if _ == 0 else tool_results_block,
                stream=False
            )
            
            content = response.content
            tool_calls = self._parse_tool_calls(content)
            
            if tool_calls:
                tool_results = []
                for tc in tool_calls:
                    res = await self._execute_tool(tc)
                    tool_results.append(f"[{tc.name} result: {res}]")
                tool_results_block = "\n".join(tool_results)
                continue
            else:
                return AgentResponse(content=content, tool_calls=tool_calls)
        
        return AgentResponse(content="Iteration limit reached.", is_complete=False)

    async def invoke(self, task: str, context: Optional[Dict] = None) -> AgentResponse:
        invocation = Invocation(id=f"inv_{int(time.time())}", task=task, context=context or {})
        return await self.process_invocation(invocation)

# Example usage
if __name__ == "__main__":
    import os
    from dotenv import load_dotenv
    load_dotenv()
    
    async def chat():
        agent = Agent(
            name="Backboard",
            api_key=os.getenv("BACKBOARD_API_KEY")
        )
        print("Backboard Agent initialized. Type 'quit' to exit.\n")
        while True:
            user_input = input("You: ")
            if user_input.lower() in ('quit', 'exit'): break
            response = await agent.invoke(user_input)
            print(f"Agent: {response.content}\n")

    asyncio.run(chat())
