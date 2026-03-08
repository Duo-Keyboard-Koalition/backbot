#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Backclaw - CLI Tools
Python CLI for Backclaw Agent Management
"""

import click
import requests
import json
import os
import sys
import subprocess
import signal
import time
from pathlib import Path
from dotenv import load_dotenv, set_key

# Handle Windows console encoding
if sys.platform == "win32":
    import codecs
    sys.stdout = codecs.getwriter("utf-8")(sys.stdout.buffer, "strict")
    sys.stderr = codecs.getwriter("utf-8")(sys.stderr.buffer, "strict")

load_dotenv()

API_KEY = os.getenv("BACKBOARD_API_KEY", "")
BASE_URL = "https://api.backboard.ai"
OPENCLAW_DIR = Path.home() / ".backclaw"
PID_FILE = OPENCLAW_DIR / "gateway.pid"

def ensure_config_dir():
    OPENCLAW_DIR.mkdir(parents=True, exist_ok=True)

@click.group()
@click.version_option(version="0.1.0", prog_name="backclaw")
def cli():
    """Backclaw - AI Agent Gateway & CLI"""
    pass

# ============ ONBOARD ============

@cli.command("onboard")
def onboard():
    """Setup Backclaw with your API key, port, and model"""
    click.echo("🎨 Welcome to Backclaw Onboarding!")
    api_key = click.prompt("Please enter your Backboard API Key", type=str)
    port = click.prompt("Please enter the Gateway Port", type=int, default=18789)
    model = click.prompt("Please enter the Model ID", type=str, default="gemini-2.0-flash")
    
    # Save API key to .env
    env_path = Path(".env")
    if not env_path.exists():
        env_path.touch()
    
    set_key(str(env_path), "BACKBOARD_API_KEY", api_key)
    
    # Save config to .backclaw/config.json
    from config import save_config, DEFAULT_CONFIG
    config = DEFAULT_CONFIG.copy()
    config["gateway"]["port"] = port
    config["gateway"]["model"] = model
    save_config(config)
    
    click.echo(f"✅ API Key saved to {env_path}")
    click.echo(f"✅ Config saved to {OPENCLAW_DIR}/config.json")
    click.echo(f"🚀 Backclaw is ready on port {port} using {model}!")

# ============ GATEWAY ============

@cli.group()
def gateway():
    """Gateway management commands"""
    pass

@gateway.command("start")
def gateway_start():
    """Start the Backclaw WebSocket gateway (foreground, Ctrl+C to stop)"""
    import asyncio
    from gateway import run_server
    click.echo("Starting Backclaw gateway... (Press Ctrl+C to stop)")
    try:
        asyncio.run(run_server())
    except KeyboardInterrupt:
        click.echo("\n[*] Gateway stopped.")


# ============ TUI ============

@cli.command("tui")
def tui_cmd():
    """Launch the Backclaw TUI"""
    from tui import main as tui_main
    click.echo("🎨 Launching Backclaw TUI...")
    tui_main()

# ============ AGENTS ============

@cli.group("agentp")
def agentp():
    """Agent management commands"""
    pass

@agentp.command("list")
def agents_list():
    """List all agents"""
    click.echo("Fetching agents...")
    # Simulated for now
    click.echo("\n  - Backclaw (local)")

# ============ REPL ============

@cli.command("repl")
def repl():
    """Start an interactive REPL with an agent"""
    click.echo("Starting Backclaw REPL...")
    # Launcher or embedded repl
    import asyncio
    from agent.agent import Agent
    async def run_chat():
        agent = Agent(name="Backclaw", api_key=os.getenv("BACKBOARD_API_KEY"))
        while True:
            text = input("You> ")
            if text.lower() in ('q', 'quit', 'exit'): break
            resp = await agent.invoke(text)
            print(f"Agent> {resp.content}")
    asyncio.run(run_chat())

if __name__ == "__main__":
    cli()
