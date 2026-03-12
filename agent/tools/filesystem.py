import os
from pathlib import Path
from typing import List, Union
from ..core import Tool, ToolParameter, ToolParameterType
from config import load_config

def get_root() -> Path:
    """Get the project root directory from config."""
    config = load_config()
    root_str = config.get("workspace", {}).get("root", ".")
    if root_str == "~":
        return Path.home().resolve()
    return Path(root_str).expanduser().resolve()

def secure_path(path: str) -> Path:
    """Resolve and secure a path to ensure it stays within the project root."""
    root = get_root()
    resolved = (root / path).resolve()
    if not resolved.is_relative_to(root):
        raise PermissionError(f"Access denied: {path} is outside the project root.")
    return resolved

def list_files(directory: str = ".") -> str:
    """List files and directories in a given path."""
    try:
        target = secure_path(directory)
        if not target.exists():
            return f"Error: Directory '{directory}' does not exist."
        if not target.is_dir():
            return f"Error: '{directory}' is not a directory."
            
        items = []
        for item in target.iterdir():
            prefix = "[DIR] " if item.is_dir() else "[FILE]"
            items.append(f"{prefix} {item.name}")
        
        return "\n".join(sorted(items)) if items else "Directory is empty."
    except Exception as e:
        return f"Error: {str(e)}"

def read_file(path: str) -> str:
    """Read the content of a file."""
    try:
        target = secure_path(path)
        if not target.exists():
            return f"Error: File '{path}' does not exist."
        if not target.is_file():
            return f"Error: '{path}' is not a file."
            
        with open(target, "r", encoding="utf-8") as f:
            return f.read()
    except Exception as e:
        return f"Error: {str(e)}"

def write_file(path: str, content: str) -> str:
    """Write content to a file."""
    try:
        target = secure_path(path)
        # Ensure parent directory exists
        target.parent.mkdir(parents=True, exist_ok=True)
        
        with open(target, "w", encoding="utf-8") as f:
            f.write(content)
        return f"Successfully wrote to '{path}'."
    except Exception as e:
        return f"Error: {str(e)}"

def delete_file(path: str) -> str:
    """Delete a file."""
    try:
        target = secure_path(path)
        if not target.exists():
            return f"Error: File '{path}' does not exist."
        if not target.is_file():
            return f"Error: '{path}' is not a file."
            
        target.unlink()
        return f"Successfully deleted '{path}'."
    except Exception as e:
        return f"Error: {str(e)}"

def make_directory(path: str) -> str:
    """Create a directory."""
    try:
        target = secure_path(path)
        target.mkdir(parents=True, exist_ok=True)
        return f"Successfully created directory '{path}'."
    except Exception as e:
        return f"Error: {str(e)}"

# Define the tools
list_files_tool = Tool(
    name="list_files",
    description="List files and directories in a given path (relative to project root)",
    parameters=[
        ToolParameter("directory", ToolParameterType.STRING, "The directory to list (default: '.')")
    ],
    handler=list_files
)

read_file_tool = Tool(
    name="read_file",
    description="Read the content of a file (relative to project root)",
    parameters=[
        ToolParameter("path", ToolParameterType.STRING, "The path to the file to read")
    ],
    handler=read_file
)

write_file_tool = Tool(
    name="write_file",
    description="Write content to a file (relative to project root)",
    parameters=[
        ToolParameter("path", ToolParameterType.STRING, "The path where the file should be written"),
        ToolParameter("content", ToolParameterType.STRING, "The content to write to the file")
    ],
    handler=write_file
)

delete_file_tool = Tool(
    name="delete_file",
    description="Delete a file (relative to project root)",
    parameters=[
        ToolParameter("path", ToolParameterType.STRING, "The path to the file to delete")
    ],
    handler=delete_file
)

make_directory_tool = Tool(
    name="make_directory",
    description="Create a new directory (relative to project root)",
    parameters=[
        ToolParameter("path", ToolParameterType.STRING, "The path of the directory to create")
    ],
    handler=make_directory
)
