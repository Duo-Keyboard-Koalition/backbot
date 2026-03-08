import json
import re
from typing import Any, Dict, List, Optional, Tuple

from .state import Step
from .tools import TOOL_REGISTRY, execute_tool

AGENT_SYSTEM_PROMPT = """You are a helpful AI agent that solves tasks step by step.

For each step, respond in EXACTLY this format:

Thought: <your reasoning about what to do next>
Action: <tool_name>
Action Input: <valid JSON object with the tool parameters>

When you have gathered enough information to fully answer the goal, respond with:

Thought: <your final reasoning>
Final Answer: <your complete answer to the goal>

Available tools:
- web_search: Search the web for information. Input: {"query": "your search query"}
- write_to_file: Write content to a file. Input: {"filename": "name.txt", "content": "content here"}
- read_file: Read content from a file. Input: {"filename": "name.txt"}
- calculate: Evaluate a math expression. Input: {"expression": "2 ** 10"}

Rules:
- Always include a Thought before every Action or Final Answer.
- Action Input must be valid JSON.
- Use Final Answer only when you are confident you have fully addressed the goal.
- Do not repeat the same action with the same input more than once.
"""


def parse_agent_response(
    response_text: str,
) -> Tuple[Optional[str], Optional[str], Optional[Dict[str, Any]], Optional[str]]:
    """Parse agent response into (thought, action, action_input, final_answer)."""
    thought: Optional[str] = None
    action: Optional[str] = None
    action_input: Optional[Dict[str, Any]] = None
    final_answer: Optional[str] = None

    thought_match = re.search(
        r"Thought:\s*(.+?)(?=\n(?:Action|Final Answer)|\Z)", response_text, re.DOTALL
    )
    if thought_match:
        thought = thought_match.group(1).strip()

    final_match = re.search(r"Final Answer:\s*(.+)", response_text, re.DOTALL)
    if final_match:
        final_answer = final_match.group(1).strip()
        return thought, None, None, final_answer

    action_match = re.search(r"Action:\s*(\w+)", response_text)
    if action_match:
        action = action_match.group(1).strip()

    input_match = re.search(r"Action Input:\s*(\{.*\})", response_text, re.DOTALL)
    if input_match:
        try:
            action_input = json.loads(input_match.group(1))
        except json.JSONDecodeError:
            action_input = {"raw": input_match.group(1)}

    return thought, action, action_input, final_answer


async def run_agent_step(
    model: Any, conversation_history: List[Dict[str, Any]], step_number: int
) -> Step:
    """Run one agent step, execute the chosen tool, and return a Step."""
    response = model.generate_content(conversation_history)
    response_text = response.text

    thought, action, action_input, final_answer = parse_agent_response(response_text)

    if final_answer is not None:
        return Step(
            step_number=step_number,
            thought=thought or "",
            action="final_answer",
            action_input={},
            observation=final_answer,
        )

    if not action or action not in TOOL_REGISTRY:
        observation = (
            f"Tool '{action}' not recognised. "
            f"Available tools: {list(TOOL_REGISTRY.keys())}"
        )
        return Step(
            step_number=step_number,
            thought=thought or "",
            action=action or "unknown",
            action_input=action_input or {},
            observation=observation,
        )

    observation = execute_tool(action, action_input or {})

    return Step(
        step_number=step_number,
        thought=thought or "",
        action=action,
        action_input=action_input or {},
        observation=observation,
    )
