import os
import json
import re
import asyncio
import time
from pathlib import Path
from typing import Any, Dict, List, Optional
from backboard import BackboardClient
from .core import Tool, ToolCall, AgentResponse, Invocation
from .loop import process_agent_invocation
from .tools import ToolManager
from config import load_config, get_config_dir

class Agent:
    """
    Python Agent integrated with Backboard SDK.
    Uses Assistant and Thread pattern for persistent conversations.
    """
    
    def __init__(
        self,
        name: Optional[str] = None,
        instructions: Optional[str] = None,
        api_key: Optional[str] = None,
        model: Optional[str] = None,
        llm_provider: Optional[str] = None,
        base_url: Optional[str] = None,
        gateway_url: Optional[str] = None,
        memory: Optional[str] = "Auto",
        assistant_id: Optional[str] = None
    ):
        # Load defaults from config
        from config import DEFAULT_CONFIG
        config = load_config()
        
        self.name = name or config.get("agent", {}).get("name", DEFAULT_CONFIG["agent"]["name"])
        self.instructions = instructions or config.get("agent", {}).get("instructions", DEFAULT_CONFIG["agent"]["instructions"])
        self.api_key = api_key
        
        self.model = model or config.get("gateway", {}).get("model", DEFAULT_CONFIG["gateway"]["model"])
        self.llm_provider = llm_provider or config.get("gateway", {}).get("llm_provider", DEFAULT_CONFIG["gateway"]["llm_provider"])
        
        self.base_url = base_url or config.get("api", {}).get("base_url", DEFAULT_CONFIG["api"]["base_url"])
        self.gateway_url = gateway_url or config.get("gateway", {}).get("url", DEFAULT_CONFIG["gateway"]["url"])
        self.memory = memory
        self.assistant_id = assistant_id
        
        # Load identity from SOUL.md if available
        self._load_soul()
        
        # Import ToolManager here to avoid circular dependency
        from .tools import ToolManager
        self.tool_manager = ToolManager(self)
        
        # SDK components (initialized lazily)
        self.client: Optional[BackboardClient] = None
        self.thread_id: Optional[str] = None
        self._tool_call_counter = 0

    def _load_soul(self):
        """Load name and instructions from SOUL.md in .backclaw/workplace/"""
        soul_path = get_config_dir() / "workplace" / "SOUL.md"
        if soul_path.exists():
            try:
                content = soul_path.read_text()
                name_match = re.search(r'^NAME:\s*(.*)$', content, re.MULTILINE)
                role_match = re.search(r'^ROLE:\s*(.*)$', content, re.MULTILINE)
                if name_match:
                    self.name = name_match.group(1).strip()
                if role_match:
                    self.instructions = role_match.group(1).strip()
            except Exception as e:
                print(f"[*] Warning: Could not load SOUL.md: {e}")

    @property
    def tools(self) -> Dict[str, Tool]:
        return self.tool_manager.tools

    async def _ensure_initialized(self):
        """Initialize SDK client, assistant, and thread if not already done"""
        if not self.client and self.api_key:
            self.client = BackboardClient(api_key=self.api_key)
            
            # Create assistant with tools if not already provided
            if not self.assistant_id:
                assistant = await self.client.create_assistant(
                    name=self.name,
                    system_prompt=self._generate_system_prompt(),
                    tools=self.tool_manager.get_schemas()
                )
                self.assistant_id = assistant.assistant_id
            
            # Create default thread
            if not self.thread_id:
                thread = await self.client.create_thread(self.assistant_id)
                self.thread_id = thread.thread_id

    def _generate_system_prompt(self) -> str:
        return f"{self.instructions}"

    def _parse_tool_calls(self, content: str) -> List[ToolCall]:
        tool_calls = []
        pattern = r'<tool:(\w+)>(\{[^}]+\})'
        matches = re.findall(pattern, content)
        for name, args_json in matches:
            if name in self.tools:
                try:
                    args = json.loads(args_json)
                    self._tool_call_counter += 1
                    tool_calls.append(ToolCall(name=name, arguments=args, call_id=f"call_{self._tool_call_counter}"))
                except json.JSONDecodeError: pass
        return tool_calls

    async def _execute_tool(self, tool_call: ToolCall) -> str:
        return await self.tool_manager.execute(tool_call.name, tool_call.arguments)

    async def invoke(self, task: str, context: Optional[Dict] = None) -> AgentResponse:
        invocation = Invocation(id=f"inv_{int(time.time())}", task=task, context=context or {})
        return await process_agent_invocation(self, invocation)
