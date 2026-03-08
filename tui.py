#!/usr/bin/env python3
"""
Backclaw - Simplified CLI Chat
Direct console interface for interacting with the Backclaw Gateway.
"""

import asyncio
import json
import websockets
import sys
from config import load_config

async def chat_loop():
    config = load_config()
    host = config["gateway"]["host"]
    port = config["gateway"]["port"]
    uri = f"ws://{host}:{port}"
    
    print("\n" + "="*50)
    print(f"🎨 Backclaw CLI Chat Interface")
    print(f"🔗 Connecting to gateway at {uri}...")
    print("="*50 + "\n")
    
    try:
        async with websockets.connect(uri) as websocket:
            # Register with the gateway
            await websocket.send(json.dumps({
                "type": "register",
                "client_id": "BackclawCLI"
            }))
            
            reg_resp_raw = await websocket.recv()
            reg_resp = json.loads(reg_resp_raw)
            print(f"✅ {reg_resp.get('content', 'Registered')}")
            print("Type '?' for help, 'quit' or 'exit' to leave.")
            print("-" * 50)

            while True:
                # Use run_in_executor to handle non-blocking input
                user_input = await asyncio.get_event_loop().run_in_executor(
                    None, lambda: input("You> ").strip()
                )
                
                if not user_input:
                    continue
                    
                if user_input.lower() in ('quit', 'exit', 'q'):
                    print("\n[*] Closing connection...")
                    break
                
                if user_input == "?":
                    print("\n📖 Backclaw CLI Help:")
                    print("  [bold]!command[/]  - Execute a local shell command (e.g., !dir)")
                    print("  [bold]? [/]         - Show this help message")
                    print("  [bold]quit[/]      - Exit the chat interface\n")
                    continue
                
                # Send message to gateway
                await websocket.send(json.dumps({
                    "type": "message",
                    "text": user_input
                }))
                
                # Listen for responses
                while True:
                    resp_raw = await websocket.recv()
                    resp = json.loads(resp_raw)
                    
                    if resp.get("type") == "status":
                        print(f"[*] {resp.get('content')}")
                    elif resp.get("type") == "response":
                        source = resp.get("source", "Agent").capitalize()
                        content = resp.get("content", "")
                        print(f"\n[{source}]: {content}\n")
                        break
                    elif resp.get("type") == "ack":
                        continue
                    else:
                        print(f"[Gateway]: {resp}")
                        break
                        
    except ConnectionRefusedError:
        print(f"❌ Error: Gateway not found at {uri}. Is it running? (backclaw gateway start)")
    except Exception as e:
        print(f"❌ Error: {e}")

def main():
    try:
        asyncio.run(chat_loop())
    except KeyboardInterrupt:
        print("\n\n[*] Shutting down CLI chat...")
        sys.exit(0)

if __name__ == "__main__":
    main()
