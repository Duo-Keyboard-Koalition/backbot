import math
import os
from typing import Any, Dict

TOOL_REGISTRY = {
    "web_search": {
        "description": "Search the web for information.",
        "parameters": {"query": "string"},
    },
    "write_to_file": {
        "description": "Write content to a file in the output directory.",
        "parameters": {"filename": "string", "content": "string"},
    },
    "read_file": {
        "description": "Read content from a file in the output directory.",
        "parameters": {"filename": "string"},
    },
    "calculate": {
        "description": "Evaluate a mathematical expression.",
        "parameters": {"expression": "string"},
    },
}


def web_search(query: str) -> str:
    return f"Search results for '{query}': No relevant results found. Try a different or more specific search query."


def write_to_file(filename: str, content: str) -> str:
    output_dir = "output"
    os.makedirs(output_dir, exist_ok=True)
    filepath = os.path.join(output_dir, os.path.basename(filename))
    with open(filepath, "w") as f:
        f.write(content)
    return f"Successfully wrote {len(content)} characters to {filepath}"


def read_file(filename: str) -> str:
    output_dir = "output"
    filepath = os.path.join(output_dir, os.path.basename(filename))
    if not os.path.exists(filepath):
        return f"File not found: {filepath}"
    with open(filepath, "r") as f:
        return f.read()


def calculate(expression: str) -> str:
    try:
        allowed = {k: v for k, v in math.__dict__.items() if not k.startswith("__")}
        result = eval(expression, {"__builtins__": {}}, allowed)  # noqa: S307
        return str(result)
    except Exception as e:
        return f"Calculation error: {e}"


def execute_tool(tool_name: str, tool_input: Dict[str, Any]) -> str:
    if tool_name == "web_search":
        return web_search(**tool_input)
    elif tool_name == "write_to_file":
        return write_to_file(**tool_input)
    elif tool_name == "read_file":
        return read_file(**tool_input)
    elif tool_name == "calculate":
        return calculate(**tool_input)
    else:
        return f"Unknown tool: '{tool_name}'. Available tools: {list(TOOL_REGISTRY.keys())}"
