#!/usr/bin/env python3
"""
Backclaw - CLI Chat Interface
Starts the gateway in-process and connects the TUI to it.
"""

import asyncio
import json
import sys
import websockets
from config import load_config
from gateway import run_server


async def chat_loop():
    config = load_config()
    host = config["gateway"]["host"]
    port = config["gateway"]["port"]
    uri = f"ws://{host}:{port}"

    print("\n" + "=" * 50)
    print("🎨 Backclaw CLI Chat Interface")
    print(f"🔗 Connecting to gateway at {uri}...")
    print("=" * 50 + "\n")

    # Retry loop — gateway needs a moment to bind the port
    max_retries = 8
    for attempt in range(max_retries):
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
                    user_input = await asyncio.get_event_loop().run_in_executor(
                        None, lambda: input("You> ").strip()
                    )

                    if not user_input:
                        continue

                    if user_input.lower() in ('quit', 'exit', 'q'):
                        print("\n[*] Closing connection...")
                        return

                    if user_input == "?":
                        print("\n📖 Backclaw CLI Help:")
                        print("  !command   - Execute a local shell command (e.g., !dir)")
                        print("  ?          - Show this help message")
                        print("  quit       - Exit the chat interface\n")
                        continue

                    await websocket.send(json.dumps({
                        "type": "message",
                        "text": user_input
                    }))

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
            return  # clean exit

        except (ConnectionRefusedError, OSError):
            if attempt < max_retries - 1:
                await asyncio.sleep(0.5)
            else:
                print(f"❌ Could not connect to gateway at {uri}.")
        except Exception as e:
            print(f"❌ Error: {e}")
            return


async def run_tui():
    """Start the gateway server as a background task, then run the TUI."""
    # Start the gateway in the background within this process
    server_task = asyncio.create_task(run_server())

    try:
        await chat_loop()
    finally:
        server_task.cancel()
        try:
            await server_task
        except asyncio.CancelledError:
            pass


def main():
    try:
        asyncio.run(run_tui())
    except KeyboardInterrupt:
        print("\n\n[*] Shutting down...")
        sys.exit(0)


if __name__ == "__main__":
    main()
