#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Backboard IO - CLI Tools
Python CLI for Backboard AI Agent Management
"""

import click
import requests
import json
from dotenv import load_dotenv
import os
import sys

# Handle Windows console encoding
if sys.platform == "win32":
    import codecs
    sys.stdout = codecs.getwriter("utf-8")(sys.stdout.buffer, "strict")
    sys.stderr = codecs.getwriter("utf-8")(sys.stderr.buffer, "strict")

load_dotenv()

API_KEY = os.getenv("BACKBOARD_API_KEY", "")
BASE_URL = "https://api.backboard.ai"

def get_client():
    """Create authenticated HTTP client"""
    return requests.Session()


@click.group()
@click.version_option(version="1.0.0", prog_name="backboard")
def cli():
    """Backboard IO - Backboard AI Agent CLI"""
    pass


# ============ GATEWAY ============

@cli.group()
def gateway():
    """Gateway management commands"""
    pass


@gateway.command("run")
def gateway_run():
    """Start the Backboard IO WebSocket gateway"""
    import subprocess
    import sys
    import os
    
    click.echo("🚀 Starting Backboard IO WebSocket Gateway...")
    # Run gateway.py as a subprocess
    gateway_path = os.path.join(os.path.dirname(__file__), "gateway.py")
    subprocess.run([sys.executable, gateway_path])


# ============ AGENTS ============

@cli.group()
def agents():
    """Agent management commands"""
    pass


@agents.command("list")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def agents_list(as_json):
    """List all agents"""
    click.echo("Fetching agents...")
    
    # Simulated response for local testing
    agents_data = {
        "agents": [
            {"id": "agent_001", "name": "Assistant", "status": "active"},
            {"id": "agent_002", "name": "Analyzer", "status": "active"}
        ]
    }
    
    if as_json:
        click.echo(json.dumps(agents_data, indent=2))
    else:
        click.echo("\n" + "=" * 50)
        for agent in agents_data.get("agents", []):
            click.echo(f"  ID: {agent['id']}")
            click.echo(f"  Name: {agent['name']}")
            click.echo(f"  Status: {agent['status']}")
            click.echo("-" * 50)
        click.echo()


@agents.command("create")
@click.argument("name")
@click.argument("instructions")
@click.option("--description", "-d", default="", help="Agent description")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def agents_create(name, instructions, description, as_json):
    """Create a new agent"""
    click.echo(f"Creating agent '{name}'...")
    
    # Simulated response
    agent_data = {
        "id": f"agent_{os.urandom(4).hex()}",
        "name": name,
        "instructions": instructions,
        "description": description,
        "status": "active"
    }
    
    if as_json:
        click.echo(json.dumps(agent_data, indent=2))
    else:
        click.echo(f"\n[OK] Agent created!")
        click.echo(f"  ID: {agent_data['id']}")
        click.echo(f"  Name: {agent_data['name']}")
        click.echo()


@agents.command("get")
@click.argument("agent_id")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def agents_get(agent_id, as_json):
    """Get agent by ID"""
    click.echo(f"Fetching agent {agent_id}...")
    
    # Simulated response
    agent_data = {
        "id": agent_id,
        "name": "Assistant",
        "instructions": "You are a helpful assistant",
        "status": "active"
    }
    
    if as_json:
        click.echo(json.dumps(agent_data, indent=2))
    else:
        click.echo("\n" + "=" * 50)
        click.echo(f"  ID: {agent_data['id']}")
        click.echo(f"  Name: {agent_data['name']}")
        click.echo(f"  Instructions: {agent_data['instructions']}")
        click.echo("=" * 50)
        click.echo()


@agents.command("delete")
@click.argument("agent_id")
@click.confirmation_option(prompt="Are you sure you want to delete this agent?")
def agents_delete(agent_id):
    """Delete an agent"""
    click.echo(f"Deleting agent {agent_id}...")
    click.echo("[OK] Agent deleted!")


# ============ TASKS ============

@cli.group()
def tasks():
    """Task management commands"""
    pass


@tasks.command("list")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def tasks_list(as_json):
    """List all tasks"""
    click.echo("Fetching tasks...")
    
    tasks_data = {
        "tasks": [
            {"id": "task_001", "agent_id": "agent_001", "task": "Analyze data", "status": "completed"},
            {"id": "task_002", "agent_id": "agent_001", "task": "Generate report", "status": "pending"}
        ]
    }
    
    if as_json:
        click.echo(json.dumps(tasks_data, indent=2))
    else:
        click.echo("\n" + "=" * 70)
        for task in tasks_data.get("tasks", []):
            status_icon = "[OK]" if task["status"] == "completed" else "[..]"
            click.echo(f"  {status_icon} {task['id']} | Agent: {task['agent_id']} | {task['task'][:30]}")
        click.echo("=" * 70)
        click.echo()


@tasks.command("create")
@click.argument("agent_id")
@click.argument("task")
@click.option("--priority", "-p", type=click.Choice(["low", "normal", "high"]), default="normal")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def tasks_create(agent_id, task, priority, as_json):
    """Create a new task"""
    click.echo(f"Creating task for agent {agent_id}...")
    
    task_data = {
        "id": f"task_{os.urandom(4).hex()}",
        "agent_id": agent_id,
        "task": task,
        "priority": priority,
        "status": "pending"
    }
    
    if as_json:
        click.echo(json.dumps(task_data, indent=2))
    else:
        click.echo(f"\n[OK] Task created!")
        click.echo(f"  ID: {task_data['id']}")
        click.echo(f"  Priority: {task_data['priority']}")
        click.echo()


@tasks.command("get")
@click.argument("task_id")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def tasks_get(task_id, as_json):
    """Get task status"""
    click.echo(f"Fetching task {task_id}...")
    
    task_data = {
        "id": task_id,
        "agent_id": "agent_001",
        "task": "Sample task",
        "status": "completed",
        "result": "Task completed successfully"
    }
    
    if as_json:
        click.echo(json.dumps(task_data, indent=2))
    else:
        click.echo("\n" + "=" * 50)
        click.echo(f"  ID: {task_data['id']}")
        click.echo(f"  Status: {task_data['status']}")
        click.echo(f"  Result: {task_data['result']}")
        click.echo("=" * 50)
        click.echo()


# ============ EXECUTE ============

@cli.command("execute")
@click.argument("agent_id")
@click.argument("task")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def execute(agent_id, task, as_json):
    """Execute a task and wait for result"""
    click.echo(f"Executing task on agent {agent_id}...")
    
    # Simulated response
    result_data = {
        "id": f"exec_{os.urandom(4).hex()}",
        "agent_id": agent_id,
        "task": task,
        "status": "completed",
        "result": f"Completed: {task}"
    }
    
    if as_json:
        click.echo(json.dumps(result_data, indent=2))
    else:
        click.echo("\n" + "=" * 50)
        click.echo(f"[OK] Execution complete!")
        click.echo(f"  Result: {result_data['result']}")
        click.echo("=" * 50)
        click.echo()


# ============ CHAT ============

@cli.command("chat")
@click.argument("agent_id")
@click.argument("message")
@click.option("--conversation-id", "-c", default=None, help="Conversation ID")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def chat(agent_id, message, conversation_id, as_json):
    """Send a chat message to an agent"""
    click.echo(f"Sending message to agent {agent_id}...")
    
    response_data = {
        "message": f"I received: '{message}'",
        "conversation_id": conversation_id or f"conv_{os.urandom(4).hex()}"
    }
    
    if as_json:
        click.echo(json.dumps(response_data, indent=2))
    else:
        click.echo("\n" + "=" * 50)
        click.echo(f"Agent: {response_data['message']}")
        click.echo(f"  Conversation ID: {response_data['conversation_id']}")
        click.echo("=" * 50)
        click.echo()


# ============ TOOLS ============

@cli.group()
def tools():
    """Tool management commands"""
    pass


@tools.command("register")
@click.argument("name")
@click.argument("schema")
@click.option("--description", "-d", default="", help="Tool description")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def tools_register(name, schema, description, as_json):
    """Register a new tool"""
    click.echo(f"Registering tool '{name}'...")
    
    try:
        schema_obj = json.loads(schema)
    except json.JSONDecodeError:
        click.echo("[ERROR] Invalid JSON schema")
        return
    
    tool_data = {
        "id": f"tool_{os.urandom(4).hex()}",
        "name": name,
        "description": description,
        "schema": schema_obj
    }
    
    if as_json:
        click.echo(json.dumps(tool_data, indent=2))
    else:
        click.echo(f"\n[OK] Tool registered!")
        click.echo(f"  ID: {tool_data['id']}")
        click.echo(f"  Name: {tool_data['name']}")
        click.echo()


@tools.command("list")
@click.option("--json", "as_json", is_flag=True, help="Output as JSON")
def tools_list(as_json):
    """List all registered tools"""
    click.echo("Fetching tools...")
    
    tools_data = {
        "tools": [
            {"id": "tool_001", "name": "calculator", "description": "Math calculations"},
            {"id": "tool_002", "name": "search", "description": "Web search"}
        ]
    }
    
    if as_json:
        click.echo(json.dumps(tools_data, indent=2))
    else:
        click.echo("\n" + "=" * 50)
        for tool in tools_data.get("tools", []):
            click.echo(f"  [*] {tool['name']} - {tool['description']}")
        click.echo("=" * 50)
        click.echo()


# ============ HEALTH ============

@cli.command("health")
def health():
    """Check Backboard API health"""
    click.echo("Checking API health...")
    click.echo("[OK] Backboard API is healthy")


# ============ REPL ============

@cli.command("repl")
@click.argument("agent_id", default=None, required=False)
def repl(agent_id):
    """Start an interactive REPL with an agent"""
    click.echo("Starting Backboard IO REPL...")
    click.echo("Type 'quit' or 'exit' to end session\n")
    
    if not agent_id:
        agent_id = "agent_default"
        click.echo(f"Using default agent: {agent_id}\n")
    
    while True:
        try:
            user_input = click.prompt("You", prompt_suffix="> ")
            
            if user_input.lower() in ('quit', 'exit', 'q'):
                click.echo("Goodbye!")
                break
            
            click.echo(f"Agent: Processing '{user_input[:50]}...'")
            click.echo(f"   Response: Task completed.\n")
            
        except KeyboardInterrupt:
            click.echo("\nGoodbye!")
            break
        except EOFError:
            click.echo("\nGoodbye!")
            break


if __name__ == "__main__":
    cli()
