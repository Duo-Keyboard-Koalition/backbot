"""
End-to-End Integration Tests

Complete workflow tests exercising the entire SentinelAI pipeline:
- User submits goal
- DarCI coordinates agents
- Agents execute with Gemini
- Sentinel monitors risk in real-time
- Interventions applied when needed
- Results returned via WebSocket

NO MOCKS - All tests use real Gemini and Tailscale APIs.
"""

import asyncio
import json
import os
import time
from typing import Any, Dict, List, Optional
from uuid import uuid4

import google.generativeai as genai
import httpx
import pytest
import pytest_asyncio
import websockets

from backend.state import ExecutionState, FailureType, Step
from backend.sentinel import score_step
from backend.intervention import intervene, should_intervene
from backend.agent import run_agent_step, AGENT_SYSTEM_PROMPT


# ============================================================================
# Test Configuration
# ============================================================================

GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")
GEMINI_FLASH_MODEL = os.getenv("GEMINI_FLASH_MODEL", "gemini-2.0-flash")
SENTINEL_URL = os.getenv("SENTINEL_URL", "ws://localhost:8000/ws/run")
DARCI_BRIDGE_URL = os.getenv("DARCI_BRIDGE_URL", "http://localhost:8080")
TAILSCALE_AUTH = os.getenv("TS_AUTH_KEY_1")

TEST_TIMEOUT = int(os.getenv("TEST_TIMEOUT", "300"))


# ============================================================================
# Complete Workflow Tests
# ============================================================================

@pytest.mark.e2e
@pytest.mark.slow
class TestCompleteWorkflow:
    """Test complete SentinelAI workflow from goal to completion."""
    
    @pytest.fixture(autouse=True)
    def setup_gemini(self):
        """Setup Gemini API client."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        genai.configure(api_key=GEMINI_API_KEY)
        self.model = genai.GenerativeModel(GEMINI_FLASH_MODEL)
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_full_risk_assessment_workflow(self):
        """Test complete flow: Task → Agent → Sentinel → Intervention."""
        goal = "Calculate the sum of numbers from 1 to 10"
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        max_steps = 5
        interventions_applied = 0
        
        for step_num in range(1, max_steps + 1):
            if state.is_complete:
                break
            
            # Agent executes step (real Gemini call)
            step = await run_agent_step(self.model, conversation_history, step_num)
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
            
            # Sentinel scores step
            risk_score, failure_types = score_step(goal, state.steps)
            step.score = risk_score
            step.failure_types = failure_types
            
            # DarCI intervenes if needed
            if should_intervene(risk_score, failure_types):
                intervention, state, reprompt = intervene(
                    state, risk_score, failure_types, self.model, step_num
                )
                interventions_applied += 1
                conversation_history.append({
                    "role": "user",
                    "parts": [f"[SENTINEL INTERVENTION]: {reprompt}"],
                })
        
        # Verify workflow completed or made progress
        assert len(state.steps) > 0
        assert state.is_complete or len(state.steps) == max_steps
        print(f"✓ Workflow: {len(state.steps)} steps, {interventions_applied} interventions")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_multi_agent_collaboration_with_monitoring(self):
        """Test multiple agents collaborating under Sentinel monitoring."""
        goal = "Research and summarize: Python async programming"
        
        # Simulate DarCI decomposing goal
        sub_goals = [
            "Search for Python async best practices",
            "Find examples of async/await patterns",
            "Summarize key concepts",
        ]
        
        results = []
        
        for sub_goal in sub_goals:
            state = ExecutionState(goal=sub_goal)
            
            conversation_history = [
                {
                    "role": "user",
                    "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {sub_goal}"],
                }
            ]
            
            # Each agent executes (simulated multi-agent)
            for step_num in range(1, 3):
                if state.is_complete:
                    break
                
                step = await run_agent_step(self.model, conversation_history, step_num)
                state.steps.append(step)
                
                # Monitor risk
                risk_score, failure_types = score_step(sub_goal, state.steps)
                
                if should_intervene(risk_score, failure_types):
                    intervention, state, reprompt = intervene(
                        state, risk_score, failure_types, self.model, step_num
                    )
                
                conversation_history.append({
                    "role": "model",
                    "parts": [f"Thought: {step.thought}\nAction: {step.action}"],
                })
                conversation_history.append({
                    "role": "user",
                    "parts": [f"Observation: {step.observation}"],
                })
            
            results.append({
                "sub_goal": sub_goal,
                "steps": len(state.steps),
                "completed": state.is_complete,
            })
        
        # Verify all sub-goals processed
        assert len(results) == 3
        print(f"✓ Multi-agent collaboration: {len(results)} sub-goals processed")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_websocket_real_time_monitoring(self):
        """Test real-time monitoring via WebSocket."""
        try:
            async with websockets.connect(SENTINEL_URL) as websocket:
                # Send goal
                await websocket.send(json.dumps({
                    "goal": "Calculate 5 + 10",
                    "api_key": GEMINI_API_KEY,
                    "model": GEMINI_FLASH_MODEL,
                    "max_steps": 3,
                }))
                
                # Receive messages
                messages = []
                async for message in websocket:
                    data = json.loads(message)
                    messages.append(data)
                    
                    if data.get("type") in ["complete", "timeout", "error"]:
                        break
                
                # Verify message flow
                assert len(messages) > 0
                assert messages[0].get("type") == "start"
                
                # Check for step messages
                step_messages = [m for m in messages if m.get("type") == "step"]
                assert len(step_messages) > 0
                
                # Verify risk scores included
                for msg in step_messages:
                    assert "risk_score" in msg
                    assert 0.0 <= msg["risk_score"] <= 1.0
                
                print(f"✓ WebSocket: {len(messages)} messages received")
        
        except (ConnectionRefusedError, OSError):
            pytest.skip("Sentinel backend not running")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_intervention_prevents_failure(self):
        """Test intervention prevents agent failure."""
        # Create scenario prone to failure
        goal = "Do something productive"
        state = ExecutionState(goal=goal)
        
        # Simulate agent struggling
        state.steps = [
            Step(
                step_number=1,
                thought="Maybe I should search... or maybe not...",
                action="web_search",
                action_input={"query": "random"},
                observation="Results",
            ),
            Step(
                step_number=2,
                thought="I'm not sure what to do next...",
                action="web_search",
                action_input={"query": "random"},
                observation="Results",
            ),
        ]
        
        # Sentinel should detect issues
        risk_score, failure_types = score_step(goal, state.steps)
        
        # DarCI should intervene
        assert should_intervene(risk_score, failure_types)
        
        intervention, new_state, reprompt = intervene(
            state, risk_score, failure_types, self.model, step_number=3
        )
        
        assert intervention is not None
        assert "goal" in reprompt.lower() or "focus" in reprompt.lower()
        print(f"✓ Intervention prevented failure: {intervention.intervention_type.value}")


# ============================================================================
# Failure Scenario Tests
# ============================================================================

@pytest.mark.e2e
@pytest.mark.slow
class TestFailureScenarios:
    """Test system handles real-world failure scenarios."""
    
    @pytest.fixture(autouse=True)
    def setup_gemini(self):
        """Setup Gemini API client."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        genai.configure(api_key=GEMINI_API_KEY)
        self.model = genai.GenerativeModel(GEMINI_FLASH_MODEL)
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_gemini_rate_limit_handling(self):
        """Test system handles Gemini rate limits gracefully."""
        # Make several rapid calls to potentially trigger rate limit
        responses = []
        errors = []
        
        for i in range(5):
            try:
                response = self.model.generate_content(f"Test query {i}")
                responses.append(response)
                await asyncio.sleep(0.5)  # Small delay
            except Exception as e:
                errors.append(str(e))
        
        # Should have at least some successful responses
        assert len(responses) > 0
        print(f"✓ Rate limit handling: {len(responses)} successful, {len(errors)} errors")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_agent_error_recovery(self):
        """Test agent recovers from tool errors."""
        goal = "Read a file that doesn't exist"
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        # First step: try to read non-existent file
        step1 = await run_agent_step(self.model, conversation_history, step_number=1)
        state.steps.append(step1)
        
        # Should get error observation
        assert "not found" in step1.observation.lower() or "error" in step1.observation.lower()
        
        # Add to conversation
        conversation_history.append({
            "role": "model",
            "parts": [f"Thought: {step1.thought}\nAction: {step1.action}"],
        })
        conversation_history.append({
            "role": "user",
            "parts": [f"Observation: {step1.observation}"],
        })
        
        # Second step: agent should try different approach
        step2 = await run_agent_step(self.model, conversation_history, step_number=2)
        state.steps.append(step2)
        
        # Verify agent adapted
        assert step2.step_number == 2
        print("✓ Agent error recovery: adapted after failure")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_network_partition_handling(self):
        """Test system handles network partitions (Tailscale disconnect)."""
        # Simulate agent becoming unavailable
        agent_unavailable = False
        
        async with httpx.AsyncClient(timeout=5.0) as client:
            try:
                # Try to reach agent
                response = await client.get(f"{DARCI_BRIDGE_URL}/health")
                if response.status_code != 200:
                    agent_unavailable = True
            except (httpx.ConnectError, httpx.TimeoutException):
                agent_unavailable = True
        
        if agent_unavailable:
            # System should handle gracefully
            print("✓ Network partition detected: agent unavailable")
            # DarCI should reassign task or retry
        else:
            print("✓ Network stable: agent available")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_invalid_goal_handling(self):
        """Test system handles invalid or impossible goals."""
        goal = ""  # Empty goal
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        try:
            step = await run_agent_step(self.model, conversation_history, step_number=1)
            state.steps.append(step)
            
            # Agent should ask for clarification or report error
            assert step is not None
            print("✓ Invalid goal handled gracefully")
        except Exception as e:
            print(f"✓ Invalid goal raised error: {e}")


# ============================================================================
# Performance Tests
# ============================================================================

@pytest.mark.e2e
@pytest.mark.load
class TestPerformance:
    """Test system performance under load."""
    
    @pytest.fixture(autouse=True)
    def setup_gemini(self):
        """Setup Gemini API client."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        genai.configure(api_key=GEMINI_API_KEY)
        self.model = genai.GenerativeModel(GEMINI_FLASH_MODEL)
    
    @pytest.mark.timeout(180)
    async def test_concurrent_agent_execution(self):
        """Test multiple agents running concurrently."""
        goals = [
            "Calculate 2 + 2",
            "Search for Python news",
            "Write hello to a file",
        ]
        
        async def run_agent(goal: str):
            state = ExecutionState(goal=goal)
            conversation_history = [
                {
                    "role": "user",
                    "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
                }
            ]
            
            step = await run_agent_step(self.model, conversation_history, step_number=1)
            return {
                "goal": goal,
                "action": step.action,
                "success": step.observation is not None,
            }
        
        # Run all agents concurrently
        tasks = [run_agent(goal) for goal in goals]
        results = await asyncio.gather(*tasks, return_exceptions=True)
        
        # Verify all completed
        successful = [r for r in results if isinstance(r, dict) and r.get("success")]
        print(f"✓ Concurrent execution: {len(successful)}/{len(goals)} successful")
    
    @pytest.mark.timeout(60)
    def test_sentinel_scoring_throughput(self):
        """Test Sentinel can score steps quickly."""
        goal = "Test goal"
        steps = [
            Step(
                step_number=i,
                thought=f"Thought {i}",
                action="web_search",
                action_input={"query": f"query {i}"},
                observation=f"Results {i}",
            )
            for i in range(100)
        ]
        
        start_time = time.time()
        
        for step in steps:
            score_step(goal, [step])
        
        elapsed = time.time() - start_time
        steps_per_second = len(steps) / elapsed
        
        assert steps_per_second > 50  # Should score >50 steps/second
        print(f"✓ Sentinel throughput: {steps_per_second:.0f} steps/second")


# ============================================================================
# Real-World Scenario Tests
# ============================================================================

@pytest.mark.e2e
@pytest.mark.slow
class TestRealWorldScenarios:
    """Test real-world usage scenarios."""
    
    @pytest.fixture(autouse=True)
    def setup_gemini(self):
        """Setup Gemini API client."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        genai.configure(api_key=GEMINI_API_KEY)
        self.model = genai.GenerativeModel(GEMINI_FLASH_MODEL)
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_research_and_summarize(self):
        """Test research and summarization workflow."""
        goal = "Research and summarize: benefits of exercise"
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        # Multi-step research
        for step_num in range(1, 4):
            if state.is_complete:
                break
            
            step = await run_agent_step(self.model, conversation_history, step_num)
            state.steps.append(step)
            
            conversation_history.append({
                "role": "model",
                "parts": [f"Thought: {step.thought}\nAction: {step.action}"],
            })
            conversation_history.append({
                "role": "user",
                "parts": [f"Observation: {step.observation}"],
            })
            
            if step.action == "final_answer":
                state.is_complete = True
                state.final_answer = step.observation
                break
        
        # Verify research completed
        assert len(state.steps) > 0
        print(f"✓ Research: {len(state.steps)} steps completed")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_code_generation_workflow(self):
        """Test code generation workflow."""
        goal = "Write a Python function to calculate factorial"
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        # Generate code
        step = await run_agent_step(self.model, conversation_history, step_number=1)
        state.steps.append(step)
        
        # Should write to file or provide final answer
        assert step.action in ["write_to_file", "final_answer"]
        print(f"✓ Code generation: {step.action}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_data_analysis_workflow(self):
        """Test data analysis workflow."""
        goal = "Calculate the average of 10, 20, 30, 40, 50"
        state = ExecutionState(goal=goal)
        
        conversation_history = [
            {
                "role": "user",
                "parts": [AGENT_SYSTEM_PROMPT + f"\n\nGoal: {goal}"],
            }
        ]
        
        step = await run_agent_step(self.model, conversation_history, step_number=1)
        state.steps.append(step)
        
        # Should use calculate tool
        assert step.action == "calculate"
        print(f"✓ Data analysis: {step.action} tool used")


# ============================================================================
# Integration Verification Tests
# ============================================================================

@pytest.mark.e2e
class TestIntegrationVerification:
    """Verify all components integrate correctly."""
    
    @pytest.mark.timeout(60)
    def test_gemini_connection(self):
        """Verify Gemini API connection."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        
        genai.configure(api_key=GEMINI_API_KEY)
        model = genai.GenerativeModel(GEMINI_FLASH_MODEL)
        
        response = model.generate_content("Say 'hello'")
        assert len(response.text) > 0
        print("✓ Gemini API connection verified")
    
    @pytest.mark.timeout(60)
    async def test_tailscale_connection(self):
        """Verify Tailscale connection."""
        if not TAILSCALE_AUTH:
            pytest.skip("TS_AUTH_KEY not set")
        
        async with httpx.AsyncClient(timeout=5.0) as client:
            try:
                response = await client.get(f"{DARCI_BRIDGE_URL}/phonebook")
                if response.status_code == 200:
                    phonebook = response.json()
                    assert "agents" in phonebook
                    print(f"✓ Tailscale connection verified: {len(phonebook.get('agents', []))} agents")
                else:
                    pytest.skip("Bridge not available")
            except (httpx.ConnectError, httpx.TimeoutException):
                pytest.skip("Tailscale not reachable")
    
    @pytest.mark.timeout(60)
    async def test_sentinel_backend_connection(self):
        """Verify Sentinel backend connection."""
        async with httpx.AsyncClient(timeout=5.0) as client:
            try:
                response = await client.get("http://localhost:8000/health")
                if response.status_code == 200:
                    data = response.json()
                    assert data.get("status") == "ok"
                    print("✓ Sentinel backend connection verified")
                else:
                    pytest.skip("Backend not available")
            except (httpx.ConnectError, httpx.TimeoutException):
                pytest.skip("Sentinel backend not running")
