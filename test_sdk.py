import asyncio
import os
from dotenv import load_dotenv
from backboard import BackboardClient

load_dotenv()

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
        
        print("[*] Creating thread...")
        thread = await client.create_thread(assistant.assistant_id)
        print(f"[+] Thread ID: {thread.thread_id}")
        
        print("[*] Sending message...")
        response = await client.add_message(
            thread_id=thread.thread_id,
            content="Hello!",
            stream=False
        )
        print(f"[+] Response content: {response.content}")
        
    except Exception as e:
        print(f"[!] SDK Error: {type(e).__name__}: {e}")

if __name__ == "__main__":
    asyncio.run(test_sdk())
