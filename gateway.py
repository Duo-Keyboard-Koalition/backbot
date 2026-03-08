#!/usr/bin/env python3
"""
Backclaw - WebSocket Gateway
Inspired by OpenClaw/NanoClaw architecture
"""

import asyncio
import json
import os
import subprocess
import time
import websockets
from pathlib import Path
from dotenv import load_dotenv
from agent import Agent, Invocation, AgentResponse
from config import load_config

OPENCLAW_DIR = Path.home() / ".backclaw"
PID_FILE = OPENCLAW_DIR / "gateway.pid"

load_dotenv()

# Configuration
CONFIG = load_config()
HOST = CONFIG["gateway"]["host"]
PORT = CONFIG["gateway"]["port"]
MODE = CONFIG["gateway"]["mode"]
MODEL = CONFIG["gateway"].get("model", "gemini-2.0-flash")
API_KEY = os.getenv("BACKBOARD_API_KEY", "")

# Shared state
clients = set()
sessions = {}

# Default agent
default_agent = Agent(
    name="Backclaw",
    instructions="You are a helpful AI assistant with tool calling capabilities.",
    api_key=API_KEY,
    model=MODEL,
    gateway_url=f"http://{HOST}:{PORT}"
)

async def handle_shell_command(command):
    """Execute local shell command and return output"""
    try:
        # Remove ! prefix
        cmd = command[1:].strip()
        print(f"[*] Executing local command: {cmd}")
        
        # Run command
        process = await asyncio.create_subprocess_shell(
            cmd,
            stdout=asyncio.subprocess.PIPE,
            stderr=asyncio.subprocess.PIPE
        )
        stdout, stderr = await process.communicate()
        
        output = stdout.decode().strip() or stderr.decode().strip() or "Command executed (no output)"
        return output
    except Exception as e:
        return f"Error executing shell command: {str(e)}"

async def broadcast(message):
    """Send message to all connected clients"""
    if clients:
        await asyncio.gather(*(client.send(json.dumps(message)) for client in clients))

async def handler(websocket):
    """WebSocket connection handler"""
    clients.add(websocket)
    print(f"[+] Client connected. Total: {len(clients)}")
    
    try:
        async for message in websocket:
            data = json.loads(message)
            msg_type = data.get("type", "message")
            
            if msg_type == "message":
                text = data.get("text", "")
                
                # Check for shell command
                if text.startswith("!"):
                    output = await handle_shell_command(text)
                    await websocket.send(json.dumps({
                        "type": "response",
                        "content": output,
                        "source": "shell"
                    }))
                else:
                    # Process with agent (simulated async for now)
                    # In a real system, this would trigger an invocation
                    print(f"[*] Task: {text}")
                    await websocket.send(json.dumps({
                        "type": "status",
                        "content": "Processing task..."
                    }))
                    
                    # Direct agent invocation for now
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
        clients.remove(websocket)
        print(f"[-] Client disconnected. Total: {len(clients)}")

async def main():
    print(f"🚀 Backclaw WebSocket Gateway")
    print(f"📡 Mode: {MODE}")
    print(f"🔗 Listening on ws://{HOST}:{PORT}")
    
    async with websockets.serve(handler, HOST, PORT):
        await asyncio.Future()  # run forever

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n[*] Shutting down...")
