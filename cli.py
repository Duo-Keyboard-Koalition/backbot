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
    """Start the Backclaw WebSocket gateway in the background"""
    ensure_config_dir()
    
    if PID_FILE.exists():
        try:
            pid = int(PID_FILE.read_text())
            if sys.platform == "win32":
                # Check if process exists on Windows
                process_check = subprocess.run(['tasklist', '/FI', f'PID eq {pid}'], capture_output=True, text=True)
                if str(pid) in process_check.stdout:
                    click.echo(f"⚠️ Gateway is already running (PID: {pid})")
                    return
            else:
                os.kill(pid, 0)
                click.echo(f"⚠️ Gateway is already running (PID: {pid})")
                return
        except (ValueError, OSError):
            pass

    gateway_script = Path(__file__).parent / "gateway.py"
    
    # Start process
    if sys.platform == "win32":
        # On Windows, we use creationflags to detach
        process = subprocess.Popen(
            [sys.executable, str(gateway_script)],
            creationflags=subprocess.CREATE_NEW_PROCESS_GROUP | subprocess.DETACHED_PROCESS,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            close_fds=True
        )
    else:
        process = subprocess.Popen(
            [sys.executable, str(gateway_script)],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            preexec_fn=os.setpgrp
        )
    
    PID_FILE.write_text(str(process.pid))
    click.echo(f"🚀 Backclaw Gateway started (PID: {process.pid})")
    click.echo(f"📡 Listening on ws://127.0.0.1:18789")

@gateway.command("stop")
def gateway_stop():
    """Stop the running Backclaw gateway"""
    if not PID_FILE.exists():
        click.echo("❌ No gateway running (PID file not found)")
        return
    
    try:
        pid = int(PID_FILE.read_text())
        if sys.platform == "win32":
            subprocess.run(['taskkill', '/F', '/PID', str(pid)], capture_output=True)
        else:
            os.kill(pid, signal.SIGTERM)
        
        PID_FILE.unlink()
        click.echo(f"🛑 Backclaw Gateway stopped (PID: {pid})")
    except (ValueError, OSError) as e:
        click.echo(f"❌ Failed to stop gateway: {e}")
        if PID_FILE.exists():
            PID_FILE.unlink()

@gateway.command("restart")
@click.pass_context
def gateway_restart(ctx):
    """Restart the Backclaw gateway"""
    ctx.invoke(gateway_stop)
    time.sleep(1)
    ctx.invoke(gateway_start)

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
