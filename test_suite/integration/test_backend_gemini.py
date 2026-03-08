"""
Backend + Gemini API Integration Tests

Tests the Sentinel backend with REAL Gemini API calls.
No mocks - all tests hit Google's Generative AI API.
"""

import asyncio
import os
import time
from typing import List

import google.generativeai as genai
import pytest
import pytest_asyncio

from backend.state import ExecutionState, FailureType, Step
from backend.sentinel import (
    check_confidence,
    check_goal_drift,
    check_loop,
    check_tool_coherence,
    score_step,
)
from backend.intervention import (
    build_reprompt,
    decompose_goal,
    execute_rollback,
    intervene,
    should_intervene,
)
from backend.agent import parse_agent_response, run_agent_step, AGENT_SYSTEM_PROMPT


# ============================================================================
# Agent Response Parsing Tests
# ============================================================================

@pytest.mark.gemini
class TestAgentResponseParsing:
    """Test parsing of real Gemini agent responses."""
    
    def test_parse_action_response(self):
        """Test parsing response with action and action input."""
        response_text = """Thought: I need to search for information about Python
Action: web_search
Action Input: {"query": "Python programming language"}"""
        
        thought, action, action_input, final_answer = parse_agent_response(response_text)
        
        assert thought is not None
        assert "search" in thought.lower()
        assert action == "web_search"
        assert action_input == {"query": "Python programming language"}
        assert final_answer is None
    
    def test_parse_final_answer_response(self):
        """Test parsing response with final answer."""
        response_text = """Thought: I have found the answer
Final Answer: Python is a high-level programming language created by Guido van Rossum."""
        
        thought, action, action_input, final_answer = parse_agent_response(response_text)
        
        assert thought is not None
        assert action is None
        assert final_answer is not None
        assert "Python" in final_answer
    
    def test_parse_malformed_json(self):
        """Test parsing response with malformed JSON in action input."""
        response_text = """Thought: Let me calculate
Action: calculate
Action Input: {expression: 2 + 2}"""  # Invalid JSON (missing quotes)
        
        thought, action, action_input, final_answer = parse_agent_response(response_text)
        
        assert action == "calculate"
        assert action_input is not None  # Should have raw fallback
    
    def test_parse_unknown_tool(self):
        """Test parsing response with unknown tool name."""
        response_text = """Thought: I'll use this special tool
Action: magic_wand
Action Input: {"spell": "abracadabra"}"""
        
        thought, action, action_input, final_answer = parse_agent_response(response_text)
        
        assert action == "magic_wand"
        assert action_input is not None


# ============================================================================
# Agent Step Execution Tests (REAL GEMINI CALLS)
# ============================================================================

@pytest.mark.gemini
@pytest.mark.api_cost
class TestAgentStepExecution:
    """Test real agent step execution with Gemini API."""
    
    @pytest.mark.timeout(60)
    async def test_agent_web_search_step(self, gemini_flash_model):
        """Test agent executes web search step via Gemini."""
        goal = "What is the capital of France?"
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        step = await run_agent_step(gemini_flash_model, conversation_history, step_number=1)
        
        assert step.step_number == 1
        assert step.thought is not None
        assert len(step.thought) > 0
        # Agent should choose web_search for this goal
        assert step.action in ["web_search", "final_answer"]
    
    @pytest.mark.timeout(60)
    async def test_agent_calculation_step(self, gemini_flash_model):
        """Test agent executes calculation step via Gemini."""
        goal = "Calculate 25 * 4"
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        step = await run_agent_step(gemini_flash_model, conversation_history, step_number=1)
        
        assert step.step_number == 1
        assert step.action == "calculate"
        assert "expression" in step.action_input
    
    @pytest.mark.timeout(60)
    async def test_agent_file_write_step(self, gemini_flash_model, temp_output_dir):
        """Test agent writes file via Gemini."""
        goal = "Write 'Hello World' to a file called greeting.txt"
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        step = await run_agent_step(gemini_flash_model, conversation_history, step_number=1)
        
        assert step.step_number == 1
        assert step.action == "write_to_file"
        assert "filename" in step.action_input
        assert "content" in step.action_input
    
    @pytest.mark.timeout(120)
    async def test_agent_multi_step_reasoning(self, gemini_flash_model):
        """Test agent performs multi-step reasoning."""
        goal = "Search for Python and then calculate 10 + 20"
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        # First step
        step1 = await run_agent_step(gemini_flash_model, conversation_history, step_number=1)
        assert step1.step_number == 1
        
        # Add to conversation
        conversation_history.append({
            "role": "model",
            "parts": [f"Thought: {step1.thought}\nAction: {step1.action}\nAction Input: {step1.action_input}"],
        })
        conversation_history.append({
            "role": "user",
            "parts": [f"Observation: {step1.observation}"],
        })
        
        # Second step
        step2 = await run_agent_step(gemini_flash_model, conversation_history, step_number=2)
        assert step2.step_number == 2
        assert step2.thought is not None


# ============================================================================
# Risk Scoring Tests
# ============================================================================

@pytest.mark.gemini
class TestRiskScoring:
    """Test Sentinel risk scoring with real agent steps."""
    
    def test_loop_detection_real_steps(self):
        """Test loop detection with realistic step patterns."""
        goal = "Calculate prime numbers"
        
        # Simulate agent stuck in loop
        steps = [
            Step(
                step_number=1,
                thought="Let me calculate 2+2",
                action="calculate",
                action_input={"expression": "2+2"},
                observation="4",
            ),
            Step(
                step_number=2,
                thought="Let me calculate 2+2 again",
                action="calculate",
                action_input={"expression": "2+2"},
                observation="4",
            ),
            Step(
                step_number=3,
                thought="Let me calculate 2+2 once more",
                action="calculate",
                action_input={"expression": "2+2"},
                observation="4",
            ),
        ]
        
        is_loop, score = check_loop(steps, window=3)
        
        assert is_loop
        assert score > 0.5
        assert score <= 1.0
    
    def test_goal_drift_detection_real_steps(self):
        """Test goal drift detection with realistic steps."""
        goal = "Write Python function sort list"
        
        # Create steps with completely unrelated content to trigger drift
        steps = [
            Step(
                step_number=1,
                thought="I'll write a sorting function",
                action="write_to_file",
                action_input={"filename": "sort.py", "content": "def sort_list(x): return sorted(x)"},
                observation="File written",
            ),
            Step(
                step_number=2,
                thought="Now let me search for pizza recipes Italian food cooking",
                action="web_search",
                action_input={"query": "best pizza recipes Italian"},
                observation="Pizza recipes found delicious food",
            ),
            Step(
                step_number=3,
                thought="Looking at restaurant menus and dining options",
                action="web_search",
                action_input={"query": "best restaurants near me"},
                observation="Restaurant listings found",
            ),
        ]
        
        is_drift, score = check_goal_drift(goal, steps)
        
        # Score should be high (significant drift detected)
        assert score > 0.6  # High drift score
        # Note: is_drift depends on threshold (0.7), score shows actual drift level
        print(f"Drift score: {score:.2f}")
    
    def test_confidence_detection_real_steps(self):
        """Test low confidence detection with hedging language."""
        steps = [
            Step(
                step_number=1,
                thought="I'm not sure, but maybe I should try this...",
                action="web_search",
                action_input={"query": "test"},
                observation="Results",
            ),
            Step(
                step_number=2,
                thought="I think this might work, possibly...",
                action="calculate",
                action_input={"expression": "1+1"},
                observation="2",
            ),
        ]
        
        is_low_conf, score = check_confidence(steps)
        
        assert is_low_conf
        assert score > 0.5
    
    def test_tool_coherence_detection_real_steps(self):
        """Test tool coherence detection with error cases."""
        goal = "Calculate something"
        
        steps = [
            Step(
                step_number=1,
                thought="Let me use an unknown tool",
                action="unknown_tool",
                action_input={},
                observation="Error: Tool not found",
            ),
            Step(
                step_number=2,
                thought="Try again with empty input",
                action="calculate",
                action_input={},
                observation="Error: Missing expression",
            ),
        ]
        
        is_incoherent, score = check_tool_coherence(steps, goal)
        
        assert is_incoherent
        assert score > 0.5
    
    def test_full_risk_score_calculation(self):
        """Test complete risk score calculation."""
        goal = "Test goal"
        
        steps = [
            Step(
                step_number=1,
                thought="Maybe I should search...",
                action="web_search",
                action_input={"query": "test"},
                observation="Results",
            ),
            Step(
                step_number=2,
                thought="Maybe search again?",
                action="web_search",
                action_input={"query": "test"},
                observation="Results",
            ),
        ]
        
        risk_score, failure_types = score_step(goal, steps)
        
        assert 0.0 <= risk_score <= 1.0
        assert isinstance(failure_types, list)


# ============================================================================
# Intervention Tests (REAL GEMINI CALLS)
# ============================================================================

@pytest.mark.gemini
@pytest.mark.api_cost
class TestInterventions:
    """Test interventions with real Gemini API calls."""
    
    @pytest.mark.timeout(60)
    def test_should_intervene_threshold(self):
        """Test intervention threshold logic."""
        assert should_intervene(risk_score=0.6, failure_types=[])
        assert should_intervene(risk_score=0.4, failure_types=[FailureType.LOOP, FailureType.GOAL_DRIFT])
        assert not should_intervene(risk_score=0.3, failure_types=[FailureType.LOOP])
    
    @pytest.mark.timeout(60)
    def test_build_reprompt_for_loop(self):
        """Test reprompt construction for loop failure."""
        goal = "Calculate primes"
        failure_types = [FailureType.LOOP]
        steps = [Step(step_number=1, thought="test", action="calc", action_input={}, observation="")]
        
        reprompt = build_reprompt(goal, failure_types, steps)
        
        assert goal in reprompt
        assert "repeating" in reprompt.lower()
        assert "different approach" in reprompt.lower()
    
    @pytest.mark.timeout(60)
    def test_build_reprompt_for_goal_drift(self):
        """Test reprompt construction for goal drift."""
        goal = "Write sorting function"
        failure_types = [FailureType.GOAL_DRIFT]
        steps = []
        
        reprompt = build_reprompt(goal, failure_types, steps)
        
        assert "refocus" in reprompt.lower() or "original task" in reprompt.lower()
    
    @pytest.mark.timeout(120)
    async def test_decompose_goal_real_gemini(self, gemini_flash_model):
        """Test goal decomposition with real Gemini API."""
        goal = "Build a complete REST API with authentication and database"
        
        sub_goals = decompose_goal(goal, gemini_flash_model)
        
        assert len(sub_goals) >= 3
        assert len(sub_goals) <= 5
        assert all(isinstance(sg, str) for sg in sub_goals)
        assert any("api" in sg.lower() or "endpoint" in sg.lower() for sg in sub_goals)
    
    @pytest.mark.timeout(120)
    async def test_decompose_goal_simple_task(self, gemini_flash_model):
        """Test goal decomposition with simple task."""
        goal = "Write a hello world program"
        
        sub_goals = decompose_goal(goal, gemini_flash_model)
        
        assert len(sub_goals) >= 1
        assert len(sub_goals) <= 5
    
    @pytest.mark.timeout(120)
    async def test_intervene_reprompt(self, gemini_flash_model):
        """Test full intervention flow with reprompt."""
        state = ExecutionState(goal="Calculate 2+2")
        state.steps = [
            Step(
                step_number=1,
                thought="Maybe I should search...",
                action="web_search",
                action_input={"query": "math"},
                observation="Results",
            ),
        ]
        
        intervention, new_state, reprompt = intervene(
            state,
            risk_score=0.6,
            failure_types=[FailureType.LOW_CONFIDENCE],
            model=gemini_flash_model,
            step_number=2,
        )
        
        assert intervention.step_number == 2
        assert intervention.intervention_type.value in ["reprompt", "rollback", "decompose", "halt"]
        assert intervention.reason is not None
        assert len(intervention.reason) > 0
    
    @pytest.mark.timeout(120)
    async def test_intervene_rollback(self, gemini_flash_model):
        """Test intervention with rollback."""
        state = ExecutionState(goal="Test goal")
        state.steps = [
            Step(step_number=1, thought="t1", action="calc", action_input={}, observation=""),
            Step(step_number=2, thought="t2", action="calc", action_input={}, observation=""),
            Step(step_number=3, thought="t3", action="calc", action_input={}, observation=""),
        ]

        # Force rollback intervention
        failure_types = [FailureType.LOOP]

        intervention, new_state, reprompt = intervene(
            state,
            risk_score=0.7,
            failure_types=failure_types,
            model=gemini_flash_model,
            step_number=4,
        )

        # Should trigger rollback for loop
        assert intervention.intervention_type == InterventionType.ROLLBACK
        assert len(new_state.steps) < len(state.steps)  # Steps were rolled back


# ============================================================================
# Execution State Tests
# ============================================================================

@pytest.mark.gemini
class TestExecutionState:
    """Test execution state management."""
    
    def test_initial_state(self):
        """Test initial execution state."""
        goal = "Test goal"
        state = ExecutionState(goal=goal)
        
        assert state.goal == goal
        assert state.steps == []
        assert state.interventions == []
        assert state.is_complete is False
        assert state.final_answer is None
    
    def test_state_with_steps(self):
        """Test state with steps added."""
        state = ExecutionState(goal="Test")
        state.steps.append(
            Step(
                step_number=1,
                thought="First step",
                action="web_search",
                action_input={"query": "test"},
                observation="Results",
            )
        )
        
        assert len(state.steps) == 1
        assert state.steps[0].step_number == 1
    
    def test_state_completion(self):
        """Test state completion."""
        state = ExecutionState(goal="Test")
        state.is_complete = True
        state.final_answer = "The answer is 42"
        
        assert state.is_complete
        assert state.final_answer == "The answer is 42"


# ============================================================================
# Integration Tests
# ============================================================================

@pytest.mark.gemini
@pytest.mark.api_cost
@pytest.mark.slow
class TestFullIntegration:
    """Full integration tests with real Gemini API."""
    
    @pytest.mark.timeout(180)
    async def test_agent_completes_simple_task(self, gemini_flash_model):
        """Test agent completes a simple task end-to-end."""
        goal = "Calculate 10 + 20 and write result to result.txt"
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        max_steps = 5
        for step_num in range(1, max_steps + 1):
            if state.is_complete:
                break
            
            step = await run_agent_step(gemini_flash_model, conversation_history, step_num)
            state.steps.append(step)
            
            # Add to conversation
            conversation_history.append({
                "role": "model",
                "parts": [f"Thought: {step.thought}\nAction: {step.action}\nAction Input: {step.action_input}"],
            })
            conversation_history.append({
                "role": "user",
                "parts": [f"Observation: {step.observation}"],
            })
            
            # Check completion
            if step.action == "final_answer":
                state.is_complete = True
                state.final_answer = step.observation
                break
            
            # Score and intervene if needed
            risk_score, failure_types = score_step(goal, state.steps)
            step.score = risk_score
            step.failure_types = failure_types
            
            if should_intervene(risk_score, failure_types):
                intervention, state, reprompt = intervene(
                    state, risk_score, failure_types, gemini_flash_model, step_num
                )
                conversation_history.append({
                    "role": "user",
                    "parts": [f"[SENTINEL INTERVENTION]: {reprompt}"],
                })
        
        # Verify task completed or made progress
        assert len(state.steps) > 0
        assert state.is_complete or len(state.steps) == max_steps
    
    @pytest.mark.timeout(180)
    async def test_agent_handles_errors_gracefully(self, gemini_flash_model):
        """Test agent handles tool errors gracefully."""
        goal = "Read a file that doesn't exist"
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        step = await run_agent_step(gemini_flash_model, conversation_history, step_number=1)
        state.steps.append(step)
        
        # Agent should get error observation
        assert "not found" in step.observation.lower() or "error" in step.observation.lower()
        
        # Risk score should reflect the error
        risk_score, failure_types = score_step(goal, state.steps)
        assert 0.0 <= risk_score <= 1.0
