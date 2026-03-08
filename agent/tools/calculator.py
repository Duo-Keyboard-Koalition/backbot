import asyncio
from typing import Any, Dict, List
from ..core import Tool, ToolParameter, ToolParameterType

def calculator(expression: str) -> str:
    """Evaluate math expressions safely"""
    try:
        allowed_chars = set("0123456789+-*/.() ")
        if not all(c in allowed_chars for c in expression):
            return "Error: Invalid characters"
        # Using a restricted eval for basic math
        result = eval(expression, {"__builtins__": {}}, {})
        return str(result)
    except Exception as e:
        return f"Error: {str(e)}"

# Define the tool metadata
calculator_tool = Tool(
    name="calculator",
    description="Evaluate math expressions",
    parameters=[
        ToolParameter("expression", ToolParameterType.STRING, "The math expression to evaluate (e.g., '2 + 2')")
    ],
    handler=calculator
)
