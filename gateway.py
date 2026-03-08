#!/usr/bin/env python3
"""
Backclaw - WebSocket Gateway
Inspired by OpenClaw/NanoClaw architecture
"""

import asyncio
import json
import os
import sys
import signal
import websockets
from pathlib import Path
from dotenv import load_dotenv
from agent import Agent, Invocation, AgentResponse
from config import load_config


load_dotenv()

# Configuration
CONFIG = load_config()
HOST = CONFIG["gateway"]["host"]
PORT = CONFIG["gateway"]["port"]
MODE = CONFIG["gateway"]["mode"]
MODEL = CONFIG["gateway"].get("model", "")
PROVIDER = CONFIG["gateway"].get("llm_provider", "")
API_KEY = os.getenv("BACKBOARD_API_KEY", "")

# Shared state
clients = set()
sessions = {}


def make_agent():
    """Create the default agent from config."""
    kwargs = dict(
        name="Backclaw",
        instructions="You are a helpful AI assistant with tool calling capabilities.",
        api_key=API_KEY,
    )
    if MODEL:
        kwargs["model"] = MODEL
    if PROVIDER:
        kwargs["llm_provider"] = PROVIDER
    return Agent(**kwargs)


async def handle_shell_command(command):
    """Execute local shell command and return output"""
    try:
        cmd = command[1:].strip()
        print(f"[*] Executing local command: {cmd}")
        process = await asyncio.create_subprocess_shell(
            cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()
        return stdout.decode().strip() or stderr.decode().strip() or "Command executed (no output)"
    except Exception as e:
        return f"Error executing shell command: {str(e)}"


async def broadcast(message):
    """Send message to all connected clients"""
    if clients:
        await asyncio.gather(*(client.send(json.dumps(message)) for client in clients))


async def handler(websocket, default_agent):
    """WebSocket connection handler"""
    clients.add(websocket)
    print(f"[+] Client connected. Total: {len(clients)}")

    try:
        async for message in websocket:
            data = json.loads(message)
            msg_type = data.get("type", "message")

            if msg_type == "message":
                text = data.get("text", "")

                if text.startswith("!"):
                    output = await handle_shell_command(text)
                    await websocket.send(json.dumps({
                        "type": "response",
                        "content": output,
                        "source": "shell"
                    }))
                else:
                    print(f"[*] Task: {text}")
                    await websocket.send(json.dumps({
                        "type": "status",
                        "content": "Processing task..."
                    }))

                    response = await default_agent.invoke(text)

                    await websocket.send(json.dumps({
                        "type": "response",
                        "content": response.content,
                        "tool_calls": [
                            {"name": tc.name, "arguments": tc.arguments}
                            for tc in response.tool_calls
                        ]
                    }))

            elif msg_type == "register":
                client_id = data.get("client_id", "unknown")
                print(f"[*] Registered client: {client_id}")
                await websocket.send(json.dumps({
                    "type": "ack",
                    "content": f"Registered as {client_id}"
                }))

    except websockets.exceptions.ConnectionClosedError:
        pass
    finally:
        clients.discard(websocket)
        print(f"[-] Client disconnected. Total: {len(clients)}")



async def run_server(host=HOST, port=PORT):
    default_agent = make_agent()
    print(f"[*] Backclaw WebSocket Gateway listening on ws://{host}:{port}")

    stop_event = asyncio.Event()

    # Define a clean shutdown function
    def shutdown():
        print("\n[*] Shutting down...")
        stop_event.set()

    # Register signals for stop/restart commands
    if sys.platform != "win32":
        loop = asyncio.get_running_loop()
        for sig in (signal.SIGTERM, signal.SIGINT):
            loop.add_signal_handler(sig, shutdown)
    # Note: On Windows, add_signal_handler is not supported for ProactorEventLoop.
    # Foreground SIGINT (Ctrl+C) is handled by the try/except in __main__.

    async def _handler(ws):
        await handler(ws, default_agent)

    async with websockets.serve(_handler, host, port):
        await stop_event.wait()  # Wait until stop_event is set

if __name__ == "__main__":
    try:
        asyncio.run(run_server())
    except KeyboardInterrupt:
        print("\n[*] Shutting down...")
