from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional


class FailureType(Enum):
    LOOP = "loop"
    GOAL_DRIFT = "goal_drift"
    LOW_CONFIDENCE = "low_confidence"
    INCOHERENT_TOOL = "incoherent_tool"


class InterventionType(Enum):
    REPROMPT = "reprompt"
    ROLLBACK = "rollback"
    DECOMPOSE = "decompose"
    HALT = "halt"


@dataclass
class Step:
    step_number: int
    thought: str
    action: str
    action_input: Dict[str, Any]
    observation: str
    score: float = 0.0
    failure_types: List[FailureType] = field(default_factory=list)


@dataclass
class Intervention:
    step_number: int
    intervention_type: InterventionType
    failure_types: List[FailureType]
    reason: str
    reprompt: Optional[str] = None


@dataclass
class ExecutionState:
    goal: str
    steps: List[Step] = field(default_factory=list)
    interventions: List[Intervention] = field(default_factory=list)
    is_complete: bool = False
    final_answer: Optional[str] = None
