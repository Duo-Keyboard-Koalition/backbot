import asyncio
import json
from .core import AgentResponse, ToolCall

async def process_agent_invocation(agent, invocation) -> AgentResponse:
    """The core reasoning loop for the agent"""
    await agent._ensure_initialized()
    
    if not agent.client:
        return AgentResponse(content="Error: Backboard API key not configured.")

    thread_id = invocation.context.get("thread_id", agent.thread_id)
    
    max_iterations = 5
    run_id = None
    
    # First message in this invocation
    kwargs = {
        "thread_id": thread_id,
        "content": invocation.task,
        "memory": agent.memory,
        "stream": False
    }
    if agent.llm_provider:
        kwargs["llm_provider"] = agent.llm_provider
    if agent.model:
        kwargs["model_name"] = agent.model
        
    response = await agent.client.add_message(**kwargs)
    
    for _ in range(max_iterations):
        if response.status == "REQUIRES_ACTION" and response.tool_calls:
            tool_outputs = []
            agent_tool_calls = []
            
            for tc in response.tool_calls:
                # SDK provides tool_call object with function name and parsed_arguments
                name = tc.function.name
                args = tc.function.parsed_arguments
                
                # Register tool call for the response
                agent_tool_calls.append(ToolCall(name=name, arguments=args, call_id=tc.id))
                
                # Execute the tool
                print(f"[*] Executing tool: {name}({args})")
                result = await agent.tool_manager.execute(name, args)
                
                tool_outputs.append({
                    "tool_call_id": tc.id,
                    "output": json.dumps(result)
                })

            # Submit tool outputs back to continue the conversation
            response = await agent.client.submit_tool_outputs(
                thread_id=thread_id,
                run_id=response.run_id,
                tool_outputs=tool_outputs,
                stream=False
            )
            continue
        else:
            # We have a final response or reached an unexpected state
            return AgentResponse(
                content=response.content or "", 
                tool_calls=[] # Tool calls are handled within the loop now
            )
    
    return AgentResponse(content="Iteration limit reached.", is_complete=False)
