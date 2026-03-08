import asyncio
import json
import websockets

async def test_gateway():
    uri = "ws://127.0.0.1:18789"
    try:
        async with websockets.connect(uri) as websocket:
            print("[*] Connected to gateway")
            
            # 1. Register
            print("[*] Registering...")
            await websocket.send(json.dumps({
                "type": "register",
                "client_id": "VerificationScript"
            }))
            reg_resp = await websocket.recv()
            print(f"[<] {reg_resp}")
            
            # 2. Send task
            print("[*] Sending task: Calculate 10 + 20")
            await websocket.send(json.dumps({
                "type": "message",
                "text": "Calculate 10 + 20"
            }))
            
            # Status
            status = await websocket.recv()
            print(f"[<] {status}")
            
            # Response
            resp = await websocket.recv()
            print(f"[<] {resp}")
            
            # 3. Send shell command
            print("[*] Sending shell command: !echo WebSocket Test")
            await websocket.send(json.dumps({
                "type": "message",
                "text": "!echo WebSocket Test"
            }))
            shell_resp = await websocket.recv()
            print(f"[<] {shell_resp}")
            
    except Exception as e:
        print(f"[!] Error: {e}")

if __name__ == "__main__":
    asyncio.run(test_gateway())
