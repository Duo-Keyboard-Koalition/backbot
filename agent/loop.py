import asyncio
import re
import json
from .core import AgentResponse, ToolCall

async def process_agent_invocation(agent, invocation) -> AgentResponse:
    """The core reasoning loop for the agent"""
    await agent._ensure_initialized()
    
    if not agent.client:
        return AgentResponse(content="Error: Backboard API key not configured.")

    # If a thread_id is provided in context or invocation, we could use it
    thread_id = invocation.context.get("thread_id", agent.thread_id)
    
    max_iterations = 5
    tool_results_block = ""
    
    for _ in range(max_iterations):
        # Send message to thread
        response = await agent.client.add_message(
            thread_id=thread_id,
            content=invocation.task if _ == 0 else tool_results_block,
            stream=False
        )
        
        content = response.content
        tool_calls = agent._parse_tool_calls(content)
        
        if tool_calls:
            tool_results = []
            for tc in tool_calls:
                res = await agent._execute_tool(tc)
                tool_results.append(f"[{tc.name} result: {res}]")
            tool_results_block = "\n".join(tool_results)
            continue
        else:
            return AgentResponse(content=content, tool_calls=tool_calls)
    
    return AgentResponse(content="Iteration limit reached.", is_complete=False)
