import asyncio
import os
from dotenv import load_dotenv
from agent.agent import Agent

load_dotenv(override=True)

async def test_new_tools():
    api_key = os.getenv("BACKBOARD_API_KEY")
    if not api_key:
        print("[!] No API key found")
        return
        
    print("[*] Initializing agent...")
    agent = Agent(
        name="Sentinel New Tools Test",
        instructions="You are a helpful assistant with filesystem, CLI, and document tools.",
        api_key=api_key,
        llm_provider="openai",
        model="gpt-4o-mini"
    )
    
    print("\n" + "="*50)
    print("TESTING TOOL REGISTRATION")
    print("="*50)
    registered_tools = list(agent.tools.keys())
    print(f"[+] Registered tools: {registered_tools}")
    
    expected_tools = ["list_files", "read_file", "write_file", "delete_file", "make_directory", "execute_command", "upload_document", "get_document_status"]
    for tool in expected_tools:
        if tool in registered_tools:
            print(f"  [PASS] {tool} is registered")
        else:
            print(f"  [FAIL] {tool} is NOT registered")

    print("\n" + "="*50)
    print("TESTING FILESYSTEM TOOLS MANUALLY")
    print("="*50)
    
    # Test directory creation
    print("[*] Testing make_directory...")
    res = await agent.tool_manager.execute("make_directory", {"path": "test_dir"})
    print(f"  Result: {res}")
    
    # Test file writing
    print("[*] Testing write_file...")
    res = await agent.tool_manager.execute("write_file", {"path": "test_dir/hello.txt", "content": "Hello from tools!"})
    print(f"  Result: {res}")
    
    # Test file reading
    print("[*] Testing read_file...")
    res = await agent.tool_manager.execute("read_file", {"path": "test_dir/hello.txt"})
    print(f"  Result: {res}")
    
    # Test file listing
    print("[*] Testing list_files...")
    res = await agent.tool_manager.execute("list_files", {"directory": "test_dir"})
    print(f"  Result: {res}")
    
    # Test CLI command
    print("\n" + "="*50)
    print("TESTING CLI TOOL")
    print("="*50)
    print("[*] Testing execute_command (echo)...")
    res = await agent.tool_manager.execute("execute_command", {"command": "echo 'CLI test successful'"})
    print(f"  Result: {res}")
    
    # Clean up
    print("\n" + "="*50)
    print("CLEANING UP")
    print("="*50)
    print("[*] Testing delete_file...")
    res = await agent.tool_manager.execute("delete_file", {"path": "test_dir/hello.txt"})
    print(f"  Result: {res}")
    
    # Python's Path.rmdir needs the directory to be empty, which it should be now
    import shutil
    if os.path.exists("test_dir"):
        shutil.rmtree("test_dir")
        print("[+] Removed test_dir manually")

if __name__ == "__main__":
    asyncio.run(test_new_tools())
