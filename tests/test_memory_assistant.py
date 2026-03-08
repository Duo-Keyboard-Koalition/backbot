import asyncio
import os
from dotenv import load_dotenv
from agent.agent import Agent

load_dotenv(override=True)

async def test_persistent_memory():
    api_key = os.getenv("BACKBOARD_API_KEY")
    if not api_key:
        print("[!] No API key found")
        return
        
    print("[*] Initializing Agent with memory='Auto'...")
    agent = Agent(
        name="Memory Test Agent",
        instructions="You are a helpful assistant with persistent memory.",
        api_key=api_key,
        memory="Auto"
    )
    
    # Ensure initialized to get assistant_id
    await agent._ensure_initialized()
    assistant_id = agent.assistant_id
    print(f"[+] Assistant ID: {assistant_id}")
    
    # Sequence 1: Share information in Thread 1
    print("\n[*] Thread 1: Sharing information...")
    print(f"[*] Thread ID: {agent.thread_id}")
    response1 = await agent.invoke("My name is Darcy and I am testing persistent memory.")
    print(f"AI Response 1: {response1.content}")
    
    # Wait for memory to be processed (as suggested in the docs)
    print("\n[*] Waiting for memory processing...")
    await asyncio.sleep(5)
    
    # Sequence 2: Ask about the information in a NEW thread
    print("\n[*] Thread 2: Testing memory recall in a new thread...")
    # Create a new agent instance with the same assistant_id but it will create a new thread
    agent2 = Agent(
        name="Memory Test Agent",
        instructions="You are a helpful assistant with persistent memory.",
        api_key=api_key,
        memory="Auto",
        assistant_id=assistant_id
    )
    # This is a bit of a hack to force the same assistant but new thread
    # In a real scenario, you'd likely just create a new thread for the same assistant
    await agent2._ensure_initialized()
    print(f"[*] New Thread ID: {agent2.thread_id}")
    
    response2 = await agent2.invoke("What is my name and what am I doing?")
    print(f"AI Response 2: {response2.content}")
    
    if "Darcy" in response2.content:
        print("\n[SUCCESS] Assistant remembered the name across threads!")
    else:
        print("\n[FAILURE] Assistant did not seem to remember the name.")

if __name__ == "__main__":
    asyncio.run(test_persistent_memory())
