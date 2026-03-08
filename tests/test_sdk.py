import asyncio
import os
from dotenv import load_dotenv
from backboard import BackboardClient

load_dotenv(override=True)

async def test_sdk():
    api_key = os.getenv("BACKBOARD_API_KEY")
    if not api_key:
        print("[!] No API key found")
        return
        
    client = BackboardClient(api_key=api_key)
    
    try:
        print("[*] Creating assistant...")
        assistant = await client.create_assistant(
            name="Test Assistant",
            system_prompt="You are a helpful assistant."
        )
        print(f"[+] Assistant ID: {assistant.assistant_id}")
        print(f"[*] Assistant object attributes: {dir(assistant)}")
        for attr in dir(assistant):
            if not attr.startswith('__'):
                try:
                    print(f"  - {attr}: {getattr(assistant, attr)}")
                except:
                    pass
        
        print("[*] Creating thread...")
        thread = await client.create_thread(assistant.assistant_id)
        print(f"[+] Thread ID: {thread.thread_id}")
        
        print("[*] Sending message...")
        response = await client.add_message(
            thread_id=thread.thread_id,
            content="Hello!",
            stream=False
        )
        print(f"[*] Response object type: {type(response)}")
        print(f"[*] Response object attributes: {dir(response)}")
        for attr in dir(response):
            if not attr.startswith('__'):
                try:
                    print(f"  - {attr}: {getattr(response, attr)}")
                except:
                    pass
        print("-" * 20)
        print(f"AGENT RESPONSE:\n{response.content}")
        print("-" * 20)
        
    except Exception as e:
        print(f"[!] SDK Error: {type(e).__name__}: {e}")

if __name__ == "__main__":
    asyncio.run(test_sdk())
