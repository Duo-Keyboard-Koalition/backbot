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
from config import load_config, get_config_dir, DEFAULT_CONFIG
from dotenv import set_key

config = load_config()
API_KEY = os.getenv("BACKBOARD_API_KEY", "")
BASE_URL = config.get("api", {}).get("base_url", DEFAULT_CONFIG["api"]["base_url"])
OPENCLAW_DIR = get_config_dir()
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
    from config import DEFAULT_CONFIG
    click.echo("🎨 Welcome to Backclaw Onboarding!")
    api_key = click.prompt("Please enter your Backboard API Key", type=str)
    port = click.prompt("Please enter the Gateway Port", type=int, default=DEFAULT_CONFIG["gateway"]["port"])
    model = click.prompt("Please enter the Model ID", type=str, default=DEFAULT_CONFIG["gateway"]["model"])
    
    # Save API key to .env
    env_path = Path(".env")
    if not env_path.exists():
        env_path.touch()
    
    set_key(str(env_path), "BACKBOARD_API_KEY", api_key)
    
    # Save config to .backclaw/config.json
    from config import save_config, DEFAULT_CONFIG
    config = load_config()
    if "gateway" not in config:
        config["gateway"] = DEFAULT_CONFIG["gateway"].copy()
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
@click.option('--daemon', is_flag=True, help="Run in the background")
def gateway_start(daemon):
    """Start the Backclaw WebSocket gateway"""
    ensure_config_dir()
    
    if PID_FILE.exists():
        click.echo("⚠️ Gateway is already running (PID file exists).")
        return

    if daemon:
        # Launch as a detached background process
        gateway_script = Path(__file__).parent / "gateway.py"
        log_file = open(OPENCLAW_DIR / "gateway.log", "a")
        
        env = os.environ.copy()
        env["PYTHONUTF8"] = "1"
        
        p = subprocess.Popen(
            [sys.executable, str(gateway_script)],
            stdout=log_file,
            stderr=log_file,
            cwd=str(Path(__file__).parent),
            env=env,
            start_new_session=True # Detach from terminal
        )
        PID_FILE.write_text(str(p.pid))
        click.echo(f"🚀 Gateway started in background (PID: {p.pid})")
    else:
        # Normal foreground execution
        from gateway import run_server
        import asyncio
        asyncio.run(run_server())

@gateway.command("stop")
def gateway_stop():
    """Stop the background gateway process"""
    if not PID_FILE.exists():
        click.echo("❌ Gateway is not running.")
        return

    pid = int(PID_FILE.read_text())
    try:
        if sys.platform == "win32":
            # On Windows, taskkill is more reliable for stopping background processes
            subprocess.run(["taskkill", "/F", "/T", "/PID", str(pid)], capture_output=True)
        else:
            # Send SIGTERM for graceful shutdown on Unix
            os.kill(pid, signal.SIGTERM)
        
        if PID_FILE.exists():
            PID_FILE.unlink()
        click.echo(f"🛑 Gateway (PID: {pid}) stopped.")
    except (ProcessLookupError, OSError):
        click.echo("⚠️ Process not found, cleaning up stale PID file.")
        if PID_FILE.exists():
            PID_FILE.unlink()

@gateway.command("restart")
@click.pass_context
def gateway_restart(ctx):
    """Restart the gateway"""
    ctx.invoke(gateway_stop)
    time.sleep(1)
    ctx.invoke(gateway_start, daemon=True)


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
        agent = Agent(api_key=os.getenv("BACKBOARD_API_KEY"))
        while True:
            text = input("You> ")
            if text.lower() in ('q', 'quit', 'exit'): break
            resp = await agent.invoke(text)
            print(f"Agent> {resp.content}")
    asyncio.run(run_chat())

if __name__ == "__main__":
    cli()
