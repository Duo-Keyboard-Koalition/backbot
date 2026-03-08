from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable, Dict, List, Optional
import time

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
