from typing import Any, List, Tuple

from .state import ExecutionState, FailureType, Intervention, InterventionType

RISK_THRESHOLD = 0.5


def should_intervene(risk_score: float, failure_types: List[FailureType]) -> bool:
    """Return True when the risk score breaches the threshold or multiple failures detected."""
    return risk_score >= RISK_THRESHOLD or len(failure_types) >= 2


def build_reprompt(
    goal: str, failure_types: List[FailureType], steps: list
) -> str:
    """Construct a corrective reprompt message tailored to the detected failures."""
    parts = [f"Your goal is: {goal}\n\nI noticed issues with your recent steps:"]

    if FailureType.LOOP in failure_types:
        parts.append(
            "- You are repeating the same actions. Try a completely different approach or tool."
        )
    if FailureType.GOAL_DRIFT in failure_types:
        parts.append(
            "- Your recent actions seem unrelated to the goal. Refocus on the original task."
        )
    if FailureType.LOW_CONFIDENCE in failure_types:
        parts.append(
            "- You seem uncertain. Consolidate what you know and proceed decisively."
        )
    if FailureType.INCOHERENT_TOOL in failure_types:
        parts.append(
            "- Your tool usage has been inconsistent or error-prone. "
            "Choose the most appropriate tool and provide valid inputs."
        )

    parts.append(
        f"\nYou have taken {len(steps)} step(s) so far. "
        "Please reassess and continue working toward the goal."
    )
    return "\n".join(parts)


def execute_rollback(state: ExecutionState, rollback_steps: int = 2) -> ExecutionState:
    """Remove the last *rollback_steps* steps from the execution state."""
    if len(state.steps) <= rollback_steps:
        state.steps = []
    else:
        state.steps = state.steps[:-rollback_steps]
    return state


def decompose_goal(goal: str, model: Any) -> List[str]:
    """Ask the model to break the goal into 3-5 concrete sub-tasks."""
    prompt = (
        f"Break the following goal into 3 to 5 specific, actionable sub-tasks.\n"
        f"Return ONLY a numbered list, one sub-task per line.\n\n"
        f"Goal: {goal}"
    )
    try:
        response = model.generate_content(prompt)
        sub_goals: List[str] = []
        for line in response.text.strip().splitlines():
            clean = line.lstrip("0123456789.-) ").strip()
            if clean:
                sub_goals.append(clean)
        return sub_goals if sub_goals else [goal]
    except Exception:
        return [goal]


def intervene(
    state: ExecutionState,
    risk_score: float,
    failure_types: List[FailureType],
    model: Any,
    step_number: int,
) -> Tuple[Intervention, ExecutionState, str]:
    """
    Choose and execute an intervention strategy.
    Returns (intervention, updated_state, reprompt_message).
    """
    reason_parts = [
        f"Risk score: {risk_score:.2f}.",
        f"Failures detected: {[f.value for f in failure_types]}.",
    ]
    reprompt: str

    if FailureType.LOOP in failure_types and len(state.steps) >= 2:
        intervention_type = InterventionType.ROLLBACK
        state = execute_rollback(state)
        reprompt = build_reprompt(state.goal, failure_types, state.steps)
        reason_parts.append("Rolled back the last 2 steps to break the loop.")

    elif len(failure_types) >= 3:
        intervention_type = InterventionType.DECOMPOSE
        sub_goals = decompose_goal(state.goal, model)
        reprompt = (
            f"Let's approach this differently. Focus on this sub-task first: {sub_goals[0]}"
        )
        reason_parts.append(f"Goal decomposed into {len(sub_goals)} sub-task(s).")

    elif risk_score > 0.8:
        intervention_type = InterventionType.HALT
        reprompt = (
            f"Stop and reconsider your entire approach to: {state.goal}"
        )
        reason_parts.append("Very high risk score triggered a halt.")

    else:
        intervention_type = InterventionType.REPROMPT
        reprompt = build_reprompt(state.goal, failure_types, state.steps)
        reason_parts.append("Injecting corrective reprompt.")

    intervention = Intervention(
        step_number=step_number,
        intervention_type=intervention_type,
        failure_types=failure_types,
        reason=" ".join(reason_parts),
        reprompt=reprompt,
    )
    state.interventions.append(intervention)
    return intervention, state, reprompt
