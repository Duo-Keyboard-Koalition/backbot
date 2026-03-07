import json
import os
from typing import Any

import google.generativeai as genai
from dotenv import load_dotenv
from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware

from .agent import AGENT_SYSTEM_PROMPT, run_agent_step
from .intervention import intervene, should_intervene
from .sentinel import score_step
from .state import ExecutionState, FailureType, InterventionType

load_dotenv()

app = FastAPI(title="SentinelAI", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


def make_serializable(obj: Any) -> Any:
    """Recursively convert dataclasses / enums to JSON-serialisable structures."""
    if isinstance(obj, (FailureType, InterventionType)):
        return obj.value
    if hasattr(obj, "__dataclass_fields__"):
        return {k: make_serializable(v) for k, v in obj.__dict__.items()}
    if isinstance(obj, list):
        return [make_serializable(i) for i in obj]
    if isinstance(obj, dict):
        return {k: make_serializable(v) for k, v in obj.items()}
    return obj


@app.get("/health")
async def health() -> dict:
    return {"status": "ok", "service": "SentinelAI"}


@app.websocket("/ws/run")
async def websocket_run(websocket: WebSocket) -> None:
    await websocket.accept()

    try:
        data = await websocket.receive_json()
        goal: str = data.get("goal", "").strip()
        api_key: str = data.get("api_key") or os.getenv("GEMINI_API_KEY", "")
        model_name: str = data.get("model") or os.getenv(
            "GEMINI_FLASH_MODEL", "gemini-2.0-flash"
        )
        max_steps: int = int(data.get("max_steps", 10))

        if not goal:
            await websocket.send_json({"type": "error", "message": "No goal provided."})
            return

        if not api_key:
            await websocket.send_json(
                {"type": "error", "message": "No GEMINI_API_KEY configured."}
            )
            return

        genai.configure(api_key=api_key)
        model = genai.GenerativeModel(model_name)

        state = ExecutionState(goal=goal)

        await websocket.send_json(
            {
                "type": "start",
                "goal": goal,
                "model": model_name,
                "max_steps": max_steps,
            }
        )

        conversation_history: list = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]

        for step_num in range(1, max_steps + 1):
            if state.is_complete:
                break

            await websocket.send_json({"type": "step_start", "step": step_num})

            try:
                step = await run_agent_step(model, conversation_history, step_num)
            except Exception as exc:
                await websocket.send_json(
                    {"type": "error", "message": f"Agent error: {exc}"}
                )
                break

            state.steps.append(step)

            # Extend conversation history with this step
            conversation_history.append(
                {
                    "role": "model",
                    "parts": [
                        f"Thought: {step.thought}\n"
                        f"Action: {step.action}\n"
                        f"Action Input: {json.dumps(step.action_input)}"
                    ],
                }
            )
            conversation_history.append(
                {"role": "user", "parts": [f"Observation: {step.observation}"]}
            )

            # Completion check
            if step.action == "final_answer":
                state.is_complete = True
                state.final_answer = step.observation
                await websocket.send_json(
                    {
                        "type": "complete",
                        "step": make_serializable(step),
                        "state": make_serializable(state),
                    }
                )
                break

            # Sentinel scoring
            risk_score, failure_types = score_step(goal, state.steps)
            step.score = risk_score
            step.failure_types = failure_types

            await websocket.send_json(
                {
                    "type": "step",
                    "step": make_serializable(step),
                    "risk_score": risk_score,
                    "failure_types": [f.value for f in failure_types],
                }
            )

            # Intervention check
            if should_intervene(risk_score, failure_types):
                try:
                    intervention, state, reprompt = intervene(
                        state, risk_score, failure_types, model, step_num
                    )
                    await websocket.send_json(
                        {
                            "type": "intervention",
                            "intervention": make_serializable(intervention),
                        }
                    )

                    if intervention.intervention_type == InterventionType.HALT:
                        break

                    conversation_history.append(
                        {
                            "role": "user",
                            "parts": [f"[SENTINEL INTERVENTION]: {reprompt}"],
                        }
                    )
                except Exception as exc:
                    await websocket.send_json(
                        {"type": "warning", "message": f"Intervention error: {exc}"}
                    )

        if not state.is_complete:
            await websocket.send_json(
                {
                    "type": "timeout",
                    "message": f"Reached max steps ({max_steps}) without completion.",
                    "state": make_serializable(state),
                }
            )

    except WebSocketDisconnect:
        pass
    except Exception as exc:
        try:
            await websocket.send_json({"type": "error", "message": str(exc)})
        except Exception:
            pass
