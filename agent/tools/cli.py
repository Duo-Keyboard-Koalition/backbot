import subprocess
import os
import shlex
from ..core import Tool, ToolParameter, ToolParameterType

def execute_command(command: str) -> str:
    """Execute a CLI command and return its output."""
    try:
        # Security: Simple check to prevent some obvious dangerous commands
        # In a real environment, this should be much more robust or containerized
        dangerous_keywords = ["rm -rf", "mkfs", "dd if=", "> /dev/"]
        if any(keyword in command for keyword in dangerous_keywords):
            return "Error: Command contains potentially dangerous operations."

        # Execute the command
        # Use shell=True for convenience, but be cautious with user input
        result = subprocess.run(
            command,
            shell=True,
            capture_output=True,
            text=True,
            timeout=30  # Timeout to prevent hanging
        )
        
        output = result.stdout
        if result.stderr:
            output += "\n--- STDERR ---\n" + result.stderr
            
        if result.returncode != 0 and not output:
             output = f"Command failed with return code {result.returncode}"
             
        return output or "Command executed successfully (no output)."
        
    except subprocess.TimeoutExpired:
        return "Error: Command timed out after 30 seconds."
    except Exception as e:
        return f"Error executing command: {str(e)}"

execute_command_tool = Tool(
    name="execute_command",
    description="Execute a shell command (CLI) on the local system.",
    parameters=[
        ToolParameter("command", ToolParameterType.STRING, "The command to execute (e.g., 'dir', 'npm list')")
    ],
    handler=execute_command
)
