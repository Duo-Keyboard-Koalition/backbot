from typing import List, Tuple

from .state import FailureType, Step

WEIGHTS = {
    FailureType.LOOP: 0.35,
    FailureType.GOAL_DRIFT: 0.30,
    FailureType.LOW_CONFIDENCE: 0.20,
    FailureType.INCOHERENT_TOOL: 0.15,
}

_HEDGE_WORDS = {
    "maybe", "perhaps", "not sure", "uncertain", "unclear",
    "might", "could", "possibly", "i think", "i believe", "probably",
}


def _judge(value: float, threshold: float) -> bool:
    """Return True when value exceeds threshold (failure detected)."""
    return value > threshold


def check_loop(steps: List[Step], window: int = 3) -> Tuple[bool, float]:
    """Detect repeated action+input combos within the last *window* steps."""
    if len(steps) < 2:
        return False, 0.0

    recent = steps[-window:]
    fingerprints = [f"{s.action}:{s.action_input}" for s in recent]
    duplicates = len(fingerprints) - len(set(fingerprints))
    score = duplicates / max(len(fingerprints) - 1, 1)
    return _judge(score, 0.4), score


def check_goal_drift(goal: str, steps: List[Step]) -> Tuple[bool, float]:
    """Detect if recent steps are semantically drifting from the original goal."""
    if not steps:
        return False, 0.0

    goal_words = set(goal.lower().split())
    recent = steps[-3:]
    drift_scores: List[float] = []

    for step in recent:
        step_text = f"{step.thought} {step.action} {step.action_input}".lower()
        step_words = set(step_text.split())
        if not step_words:
            drift_scores.append(1.0)
            continue
        overlap = len(goal_words & step_words) / max(len(goal_words), 1)
        drift_scores.append(1.0 - min(overlap * 3.0, 1.0))

    avg_drift = sum(drift_scores) / len(drift_scores)
    return _judge(avg_drift, 0.7), avg_drift


def check_confidence(steps: List[Step]) -> Tuple[bool, float]:
    """Detect low confidence via hedging language in the agent's thoughts."""
    if not steps:
        return False, 0.0

    recent = steps[-2:]
    hedge_scores: List[float] = []

    for step in recent:
        thought_lower = step.thought.lower()
        count = sum(1 for w in _HEDGE_WORDS if w in thought_lower)
        hedge_scores.append(min(count / 3.0, 1.0))

    avg_hedge = sum(hedge_scores) / len(hedge_scores)
    return _judge(avg_hedge, 0.5), avg_hedge


def check_tool_coherence(steps: List[Step], goal: str) -> Tuple[bool, float]:
    """Detect incoherent tool usage (empty inputs, errors, unknown tools)."""
    if not steps:
        return False, 0.0

    recent = steps[-3:]
    incoherence_scores: List[float] = []

    for step in recent:
        if step.action in ("final_answer", "unknown"):
            incoherence_scores.append(0.0)
            continue
        if not step.action_input:
            incoherence_scores.append(0.8)
            continue
        obs_lower = step.observation.lower()
        if "error" in obs_lower or "not found" in obs_lower or "not recognised" in obs_lower:
            incoherence_scores.append(0.6)
        else:
            incoherence_scores.append(0.0)

    avg_incoherence = sum(incoherence_scores) / len(incoherence_scores)
    return _judge(avg_incoherence, 0.5), avg_incoherence


def score_step(goal: str, steps: List[Step]) -> Tuple[float, List[FailureType]]:
    """
    Compute a weighted risk score for the current execution state.
    Returns (risk_score, failure_types). Higher score = higher risk.

    Weights:  loop 0.35 | drift 0.30 | confidence 0.20 | coherence 0.15
    """
    if not steps:
        return 0.0, []

    loop_fail, loop_score = check_loop(steps)
    drift_fail, drift_score = check_goal_drift(goal, steps)
    conf_fail, conf_score = check_confidence(steps)
    coh_fail, coh_score = check_tool_coherence(steps, goal)

    risk_score = (
        WEIGHTS[FailureType.LOOP] * loop_score
        + WEIGHTS[FailureType.GOAL_DRIFT] * drift_score
        + WEIGHTS[FailureType.LOW_CONFIDENCE] * conf_score
        + WEIGHTS[FailureType.INCOHERENT_TOOL] * coh_score
    )

    failure_types: List[FailureType] = []
    if loop_fail:
        failure_types.append(FailureType.LOOP)
    if drift_fail:
        failure_types.append(FailureType.GOAL_DRIFT)
    if conf_fail:
        failure_types.append(FailureType.LOW_CONFIDENCE)
    if coh_fail:
        failure_types.append(FailureType.INCOHERENT_TOOL)

    return risk_score, failure_types
