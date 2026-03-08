"""
Sentinel AI - Python Agent with Tool Calling
Inspired by NanoClaw architecture
"""

import json
import re
from typing import Any, Callable, Dict, List, Optional
from dataclasses import dataclass, field
from enum import Enum


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
        """Convert tool to JSON schema format"""
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
class AgentMessage:
    role: str  # "system", "user", "assistant"
    content: str


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


class Agent:
    """
    Python Agent with tool calling capabilities.
    Inspired by NanoClaw's lightweight agent architecture.
    """
    
    def __init__(
        self,
        name: str,
        instructions: str = "You are a helpful AI assistant.",
        api_key: Optional[str] = None,
        base_url: str = "https://api.backboard.ai"
    ):
        self.name = name
        self.instructions = instructions
        self.api_key = api_key
        self.base_url = base_url
        self.tools: Dict[str, Tool] = {}
        self.message_history: List[AgentMessage] = []
        self._tool_call_counter = 0
        
        # Add default tools
        self._register_default_tools()
    
    def _register_default_tools(self):
        """Register built-in tools"""
        
        def calculator(expression: str) -> str:
            """Evaluate a mathematical expression safely"""
            try:
                # Safe evaluation of math expressions
                allowed_chars = set("0123456789+-*/.() ")
                if not all(c in allowed_chars for c in expression):
                    return "Error: Invalid characters in expression"
                result = eval(expression, {"__builtins__": {}}, {})
                return str(result)
            except Exception as e:
                return f"Error: {str(e)}"
        
        self.register_tool(
            name="calculator",
            description="Evaluate mathematical expressions",
            parameters=[
                ToolParameter(
                    name="expression",
                    type=ToolParameterType.STRING,
                    description="The mathematical expression to evaluate"
                )
            ],
            handler=calculator
        )
        
        def save_memory(key: str, value: str) -> str:
            """Save a key-value pair to memory"""
            if not hasattr(self, '_memory'):
                self._memory = {}
            self._memory[key] = value
            return f"Saved: {key} = {value}"
        
        self.register_tool(
            name="save_memory",
            description="Save information to persistent memory",
            parameters=[
                ToolParameter(
                    name="key",
                    type=ToolParameterType.STRING,
                    description="The key to store the value under"
                ),
                ToolParameter(
                    name="value",
                    type=ToolParameterType.STRING,
                    description="The value to store"
                )
            ],
            handler=save_memory
        )
        
        def get_memory(key: str) -> str:
            """Retrieve a value from memory"""
            if not hasattr(self, '_memory'):
                self._memory = {}
            return self._memory.get(key, f"Key '{key}' not found in memory")
        
        self.register_tool(
            name="get_memory",
            description="Retrieve information from persistent memory",
            parameters=[
                ToolParameter(
                    name="key",
                    type=ToolParameterType.STRING,
                    description="The key to retrieve"
                )
            ],
            handler=get_memory
        )
    
    def register_tool(
        self,
        name: str,
        description: str,
        parameters: List[ToolParameter],
        handler: Callable[..., Any]
    ):
        """Register a custom tool"""
        tool = Tool(
            name=name,
            description=description,
            parameters=parameters,
            handler=handler
        )
        self.tools[name] = tool
        return tool
    
    def _parse_tool_calls(self, content: str) -> List[ToolCall]:
        """
        Parse tool calls from agent response.
        Looks for patterns like: <tool:name>{"args": "value"}
        """
        tool_calls = []
        pattern = r'<tool:(\w+)>(\{[^}]+\})'
        matches = re.findall(pattern, content)
        
        for name, args_json in matches:
            if name in self.tools:
                try:
                    args = json.loads(args_json)
                    self._tool_call_counter += 1
                    tool_calls.append(ToolCall(
                        name=name,
                        arguments=args,
                        call_id=f"call_{self._tool_call_counter}"
                    ))
                except json.JSONDecodeError:
                    pass
        
        return tool_calls
    
    def _execute_tool(self, tool_call: ToolCall) -> str:
        """Execute a tool call and return the result"""
        if tool_call.name not in self.tools:
            return f"Error: Unknown tool '{tool_call.name}'"
        
        tool = self.tools[tool_call.name]
        try:
            result = tool.handler(**tool_call.arguments)
            return str(result)
        except Exception as e:
            return f"Error executing {tool_call.name}: {str(e)}"
    
    def _generate_system_prompt(self) -> str:
        """Generate the system prompt with tool definitions"""
        prompt = f"""You are {self.name}. {self.instructions}

You have access to the following tools. Use them by outputting:
<tool:tool_name>{{"arg1": "value1", "arg2": "value2"}}

Available tools:
"""
        for tool in self.tools.values():
            params = ", ".join(f"{p.name}: {p.type.value}" for p in tool.parameters)
            prompt += f"- {tool.name}({params}): {tool.description}\n"
        
        prompt += """
When you need to use a tool, output ONLY the tool call.
When done, output your final response normally.
"""
        return prompt
    
    def chat(self, message: str, max_iterations: int = 5) -> AgentResponse:
        """
        Process a chat message with tool calling support.
        
        Args:
            message: User message
            max_iterations: Maximum tool call iterations
            
        Returns:
            AgentResponse with content and any tool calls
        """
        self.message_history.append(AgentMessage(role="user", content=message))
        
        for iteration in range(max_iterations):
            # Build conversation context
            messages = [
                {"role": "system", "content": self._generate_system_prompt()}
            ]
            for msg in self.message_history:
                messages.append({"role": msg.role, "content": msg.content})
            
            # For local testing without API, simulate agent response
            # In production, this would call the Backboard API
            response_content = self._simulate_response(messages)
            
            # Parse tool calls
            tool_calls = self._parse_tool_calls(response_content)
            
            if tool_calls:
                # Execute tools
                for tool_call in tool_calls:
                    result = self._execute_tool(tool_call)
                    self.message_history.append(
                        AgentMessage(role="tool", content=f"[{tool_call.name} result: {result}]")
                    )
                continue
            else:
                # No tool calls, return final response
                self.message_history.append(AgentMessage(role="assistant", content=response_content))
                return AgentResponse(content=response_content, tool_calls=tool_calls)
        
        return AgentResponse(
            content="I reached the maximum number of iterations. Let me summarize what I found...",
            is_complete=False
        )
    
    def _simulate_response(self, messages: List[Dict]) -> str:
        """
        Simulate agent response for local testing.
        In production, replace this with actual Backboard API call.
        """
        last_message = messages[-1]["content"] if messages else ""
        
        # Simple rule-based responses for demonstration
        if "calculate" in last_message.lower() or any(c in last_message for c in "+-*/"):
            # Extract math expression
            match = re.search(r'(\d+[\s\+\-\*\/]*\d+)', last_message)
            if match:
                expr = match.group(1).replace(' ', '')
                return f'<tool:calculator>{{"expression": "{expr}"}}'
        
        if "remember" in last_message.lower() or "save" in last_message.lower():
            match = re.search(r'(\w+)\s+(?:is|was|=)\s+(.+)', last_message)
            if match:
                key, value = match.groups()
                return f'<tool:save_memory>{{"key": "{key}", "value": "{value.strip()}"}}'
        
        if "what is" in last_message.lower() and "memory" in last_message.lower():
            match = re.search(r'what is (\w+)', last_message.lower())
            if match:
                key = match.group(1)
                return f'<tool:get_memory>{{"key": "{key}"}}'
        
        # Default response
        return f"I received your message: '{last_message[:50]}...' How can I help you further?"
    
    def clear_history(self):
        """Clear conversation history"""
        self.message_history = []
    
    def get_tools_schema(self) -> List[Dict]:
        """Get all tools in schema format"""
        return [tool.to_schema() for tool in self.tools.values()]


# Example usage
if __name__ == "__main__":
    agent = Agent(
        name="Sentinel",
        instructions="You are a helpful assistant with tool calling capabilities."
    )
    
    print("Agent initialized. Type 'quit' to exit.\n")
    
    while True:
        user_input = input("You: ")
        if user_input.lower() in ('quit', 'exit'):
            break
        
        response = agent.chat(user_input)
        print(f"Agent: {response.content}\n")
