import asyncio
import os
from dotenv import load_dotenv
from agent.agent import Agent

load_dotenv()

async def test_tool_calling():
    api_key = os.getenv("BACKBOARD_API_KEY")
    if not api_key:
        print("[!] No API key found")
        return
        
    print("[*] Initializing agent...")
    agent = Agent(
        name="Sentinel Test",
        instructions="You are a helpful assistant. Use the calculator tool for math.",
        api_key=api_key,
        llm_provider="openai",
        model="gpt-4o-mini"
    )
    
    print("[*] Invoking agent with a math problem...")
    response = await agent.invoke("What is 123 * 456?")
    
    print(f"\n[+] Agent Response: {response.content}")
    if response.tool_calls:
        print("[+] Tool calls made during invocation:")
        for tc in response.tool_calls:
            print(f"    - {tc.name}({tc.arguments})")
    else:
        print("[*] No tool calls recorded in AgentResponse (this is expected as they are handled internally now)")

if __name__ == "__main__":
    asyncio.run(test_tool_calling())
