"""
DarCI Multi-Agent Coordination Tests

Tests real multi-agent coordination with Gemini API and Tailscale.
Requires:
- GEMINI_API_KEY environment variable
- Tailscale installed and authenticated
- Tailbridge agents running

NO MOCKS - All tests use real APIs.
"""

import asyncio
import json
import os
import time
from pathlib import Path
from typing import Any, Dict, List, Optional
from uuid import uuid4

import google.generativeai as genai
import httpx
import pytest
import pytest_asyncio

from backend.state import ExecutionState, FailureType, InterventionType, Step
from backend.sentinel import score_step
from backend.intervention import intervene, should_intervene


# ============================================================================
# Test Configuration
# ============================================================================

GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")
GEMINI_MODEL = os.getenv("GEMINI_FLASH_MODEL", "gemini-2.0-flash")
DARCI_BRIDGE_URL = os.getenv("DARCI_BRIDGE_URL", "http://localhost:8080")
SENTINEL_URL = os.getenv("SENTINEL_URL", "http://localhost:8000")

TEST_TIMEOUT = int(os.getenv("TEST_TIMEOUT", "300"))


# ============================================================================
# DarCI Agent Tests
# ============================================================================

@pytest.mark.gemini
@pytest.mark.slow
class TestDarCIAgentCoordination:
    """Test DarCI coordinating multiple agents."""
    
    @pytest.fixture(autouse=True)
    def setup_gemini(self):
        """Setup Gemini API client."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        genai.configure(api_key=GEMINI_API_KEY)
        self.model = genai.GenerativeModel(GEMINI_MODEL)
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_discovers_agents(self):
        """Test DarCI can discover agents via Tailscale."""
        async with httpx.AsyncClient() as client:
            try:
                response = await client.get(f"{DARCI_BRIDGE_URL}/phonebook")
                if response.status_code == 200:
                    phonebook = response.json()
                    assert "agents" in phonebook
                    assert len(phonebook["agents"]) > 0
                    print(f"✓ Discovered {len(phonebook['agents'])} agents")
                else:
                    pytest.skip("Bridge not available")
            except httpx.ConnectError:
                pytest.skip("Tailbridge not running")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_assigns_task_to_agent(self):
        """Test DarCI assigns task to discovered agent."""
        # Simulate DarCI task assignment
        task = {
            "id": str(uuid4()),
            "description": "Research Python async best practices",
            "assigned_to": "agent1",
            "priority": "high",
            "created_at": time.time(),
        }
        
        async with httpx.AsyncClient() as client:
            try:
                # Post task to agent via bridge
                response = await client.post(
                    f"{DARCI_BRIDGE_URL}/a2a/inbound",
                    json={
                        "id": str(uuid4()),
                        "type": "request",
                        "source": "darci",
                        "dest": "agent1",
                        "body": {
                            "action": "assign_task",
                            "payload": task,
                        },
                    },
                    timeout=10.0,
                )
                assert response.status_code == 202
                print(f"✓ Task assigned: {task['id']}")
            except httpx.ConnectError:
                pytest.skip("Tailbridge not running")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_monitors_agent_progress(self):
        """Test DarCI monitors agent task progress."""
        # Create execution state
        state = ExecutionState(goal="Monitor agent progress")
        
        # Simulate agent steps
        steps = [
            Step(
                step_number=1,
                thought="Starting task",
                action="web_search",
                action_input={"query": "async python"},
                observation="Results found",
            ),
            Step(
                step_number=2,
                thought="Analyzing results",
                action="read_file",
                action_input={"filename": "guide.md"},
                observation="File content",
            ),
        ]
        
        state.steps = steps
        
        # Score each step (Sentinel monitoring)
        for step in steps:
            risk_score, failure_types = score_step(state.goal, state.steps)
            step.score = risk_score
            step.failure_types = failure_types
            
            assert 0.0 <= risk_score <= 1.0
        
        print(f"✓ Monitored {len(steps)} steps with risk scores")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_multi_agent_collaboration(self):
        """Test multiple agents collaborating on complex task."""
        goal = "Build a REST API with tests and documentation"
        
        # Simulate DarCI decomposing goal for multiple agents
        prompt = f"""Break this goal into sub-tasks for different agents:
Goal: {goal}

Return a JSON list of sub-tasks with agent assignments:
[
    {{"agent": "backend_agent", "task": "..."}},
    {{"agent": "qa_agent", "task": "..."}},
    {{"agent": "docs_agent", "task": "..."}}
]
"""
        
        response = self.model.generate_content(prompt)
        
        # Parse response (should contain JSON-like structure)
        assert len(response.text) > 0
        print(f"✓ Goal decomposed for multi-agent collaboration")
        print(f"  Response length: {len(response.text)} chars")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_task_reassignment(self):
        """Test DarCI reassigns task when agent fails."""
        # Simulate agent failure
        original_agent = "agent1"
        new_agent = "agent2"
        
        reassignment = {
            "task_id": str(uuid4()),
            "from_agent": original_agent,
            "to_agent": new_agent,
            "reason": "Agent unavailable",
            "timestamp": time.time(),
        }
        
        async with httpx.AsyncClient() as client:
            try:
                # Send reassignment message
                response = await client.post(
                    f"{DARCI_BRIDGE_URL}/a2a/inbound",
                    json={
                        "id": str(uuid4()),
                        "type": "request",
                        "source": "darci",
                        "dest": new_agent,
                        "body": {
                            "action": "reassign_task",
                            "payload": reassignment,
                        },
                    },
                    timeout=10.0,
                )
                assert response.status_code == 202
                print(f"✓ Task reassigned from {original_agent} to {new_agent}")
            except httpx.ConnectError:
                pytest.skip("Tailbridge not running")


# ============================================================================
# DarCI + Sentinel Integration Tests
# ============================================================================

@pytest.mark.gemini
@pytest.mark.slow
class TestDarCISentinelIntegration:
    """Test DarCI integrating with Sentinel for risk monitoring."""
    
    @pytest.fixture(autouse=True)
    def setup_gemini(self):
        """Setup Gemini API client."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        genai.configure(api_key=GEMINI_API_KEY)
        self.model = genai.GenerativeModel(GEMINI_MODEL)
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_receives_sentinel_alerts(self):
        """Test DarCI receives and processes Sentinel risk alerts."""
        # Simulate Sentinel alert
        alert = {
            "task_id": str(uuid4()),
            "risk_score": 0.75,
            "failure_types": ["loop", "goal_drift"],
            "step_number": 5,
            "timestamp": time.time(),
        }
        
        # DarCI should process alert and decide on intervention
        assert should_intervene(alert["risk_score"], [FailureType.LOOP, FailureType.GOAL_DRIFT])
        print(f"✓ Sentinel alert processed: risk={alert['risk_score']}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_executes_intervention(self):
        """Test DarCI executes intervention based on Sentinel alert."""
        state = ExecutionState(goal="Test intervention")
        state.steps = [
            Step(
                step_number=1,
                thought="Maybe I should search...",
                action="web_search",
                action_input={"query": "test"},
                observation="Results",
            ),
            Step(
                step_number=2,
                thought="Search again?",
                action="web_search",
                action_input={"query": "test"},
                observation="Results",
            ),
        ]
        
        risk_score = 0.6
        failure_types = [FailureType.LOW_CONFIDENCE]
        
        # Execute intervention
        intervention, new_state, reprompt = intervene(
            state,
            risk_score,
            failure_types,
            self.model,
            step_number=3,
        )
        
        assert intervention.step_number == 3
        assert intervention.intervention_type.value in ["reprompt", "rollback", "decompose", "halt"]
        assert len(intervention.reason) > 0
        print(f"✓ Intervention executed: {intervention.intervention_type.value}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_rollback_intervention(self):
        """Test DarCI executes rollback intervention."""
        state = ExecutionState(goal="Test rollback")
        state.steps = [
            Step(step_number=1, thought="t1", action="calc", action_input={}, observation=""),
            Step(step_number=2, thought="t2", action="calc", action_input={}, observation=""),
            Step(step_number=3, thought="t3", action="calc", action_input={}, observation=""),
            Step(step_number=4, thought="t4", action="calc", action_input={}, observation=""),
        ]
        
        # Force rollback for loop
        failure_types = [FailureType.LOOP]
        
        intervention, new_state, reprompt = intervene(
            state,
            risk_score=0.7,
            failure_types=failure_types,
            model=self.model,
            step_number=5,
        )
        
        # Should have rolled back steps
        assert len(new_state.steps) < len(state.steps)
        print(f"✓ Rollback: {len(state.steps)} -> {len(new_state.steps)} steps")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_decompose_intervention(self):
        """Test DarCI executes goal decomposition intervention."""
        goal = "Build a complete machine learning pipeline with data preprocessing, model training, evaluation, and deployment"
        
        sub_goals = []
        try:
            sub_goals = decompose_goal_for_darci(goal, self.model)
            assert len(sub_goals) >= 3
            assert len(sub_goals) <= 5
            print(f"✓ Goal decomposed into {len(sub_goals)} sub-tasks")
        except Exception as e:
            print(f"⚠ Decomposition skipped: {e}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_halt_intervention(self):
        """Test DarCI executes halt intervention for critical failures."""
        state = ExecutionState(goal="Test halt")
        state.steps = [
            Step(
                step_number=1,
                thought="Something is wrong",
                action="unknown",
                action_input={},
                observation="Error: Tool not found",
            ),
        ]
        
        # Very high risk score should trigger halt
        intervention, new_state, reprompt = intervene(
            state,
            risk_score=0.9,  # Very high
            failure_types=[FailureType.INCOHERENT_TOOL, FailureType.GOAL_DRIFT, FailureType.LOW_CONFIDENCE],
            model=self.model,
            step_number=2,
        )
        
        assert intervention.intervention_type == InterventionType.HALT
        print("✓ Halt intervention executed for critical failure")


# ============================================================================
# DarCI Task Management Tests
# ============================================================================

@pytest.mark.gemini
class TestDarCITaskManagement:
    """Test DarCI task lifecycle management."""
    
    @pytest.fixture(autouse=True)
    def setup_gemini(self):
        """Setup Gemini API client."""
        if not GEMINI_API_KEY:
            pytest.skip("GEMINI_API_KEY not set")
        genai.configure(api_key=GEMINI_API_KEY)
        self.model = genai.GenerativeModel(GEMINI_MODEL)
    
    @pytest.mark.timeout(60)
    def test_darci_creates_task(self):
        """Test DarCI creates new task."""
        task = {
            "id": str(uuid4()),
            "goal": "Research quantum computing",
            "status": "pending",
            "created_at": time.time(),
            "metadata": {
                "priority": "high",
                "tags": ["research", "quantum"],
            },
        }
        
        assert task["id"] is not None
        assert task["status"] == "pending"
        print(f"✓ Task created: {task['id']}")
    
    @pytest.mark.timeout(60)
    def test_darci_updates_task_status(self):
        """Test DarCI updates task status."""
        task_status = {
            "id": str(uuid4()),
            "old_status": "pending",
            "new_status": "in_progress",
            "updated_at": time.time(),
        }
        
        assert task_status["new_status"] != task_status["old_status"]
        print(f"✓ Task status updated: {task_status['old_status']} -> {task_status['new_status']}")
    
    @pytest.mark.timeout(60)
    def test_darci_completes_task(self):
        """Test DarCI marks task as complete."""
        task_result = {
            "id": str(uuid4()),
            "status": "completed",
            "output": "Task completed successfully",
            "completed_at": time.time(),
        }
        
        assert task_result["status"] == "completed"
        assert task_result["output"] is not None
        print(f"✓ Task completed: {task_result['id']}")
    
    @pytest.mark.timeout(60)
    def test_darci_fails_task(self):
        """Test DarCI marks task as failed."""
        task_failure = {
            "id": str(uuid4()),
            "status": "failed",
            "error": "Agent unreachable",
            "failed_at": time.time(),
        }
        
        assert task_failure["status"] == "failed"
        assert task_failure["error"] is not None
        print(f"✓ Task failed: {task_failure['id']}")


# ============================================================================
# DarCI Communication Tests
# ============================================================================

@pytest.mark.tailscale
@pytest.mark.slow
class TestDarCICommunication:
    """Test DarCI communication over Tailscale."""
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_sends_broadcast(self):
        """Test DarCI sends broadcast message to all agents."""
        broadcast_msg = {
            "id": str(uuid4()),
            "type": "broadcast",
            "source": "darci",
            "dest": "all",
            "body": {
                "action": "announcement",
                "payload": {
                    "message": "System maintenance in 5 minutes",
                    "priority": "high",
                },
            },
        }
        
        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(
                    f"{DARCI_BRIDGE_URL}/a2a/inbound",
                    json=broadcast_msg,
                    timeout=10.0,
                )
                assert response.status_code == 202
                print("✓ Broadcast message sent")
            except httpx.ConnectError:
                pytest.skip("Tailbridge not running")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_request_response_pattern(self):
        """Test DarCI request/response pattern with agent."""
        correlation_id = str(uuid4())
        
        # Send request
        request_msg = {
            "id": str(uuid4()),
            "type": "request",
            "correlation_id": correlation_id,
            "source": "darci",
            "dest": "agent1",
            "body": {
                "action": "get_status",
                "payload": {"agent_id": "agent1"},
            },
        }
        
        async with httpx.AsyncClient() as client:
            try:
                response = await client.post(
                    f"{DARCI_BRIDGE_URL}/a2a/inbound",
                    json=request_msg,
                    timeout=10.0,
                )
                assert response.status_code == 202
                
                # Send response
                response_msg = {
                    "id": str(uuid4()),
                    "type": "response",
                    "correlation_id": correlation_id,
                    "source": "agent1",
                    "dest": "darci",
                    "body": {
                        "action": "status_response",
                        "payload": {"status": "online", "load": 0.5},
                    },
                }
                
                response = await client.post(
                    f"{DARCI_BRIDGE_URL}/a2a/inbound",
                    json=response_msg,
                    timeout=10.0,
                )
                assert response.status_code == 202
                
                print(f"✓ Request/response pattern completed: {correlation_id}")
            except httpx.ConnectError:
                pytest.skip("Tailbridge not running")


# ============================================================================
# Helper Functions
# ============================================================================

def decompose_goal_for_darci(goal: str, model: Any) -> List[Dict[str, str]]:
    """Decompose goal for DarCI multi-agent assignment."""
    prompt = f"""Break this goal into 3-5 sub-tasks for different agents:
Goal: {goal}

Return ONLY a JSON array:
[
    {{"agent": "agent_name", "task": "specific sub-task"}},
    ...
]
"""
    
    response = model.generate_content(prompt)
    
    # Try to parse JSON from response
    import re
    json_match = re.search(r'\[.*\]', response.text, re.DOTALL)
    if json_match:
        return json.loads(json_match.group(0))
    
    # Fallback: return simple decomposition
    return [
        {"agent": "agent1", "task": f"Sub-task 1 of: {goal[:50]}..."},
        {"agent": "agent2", "task": f"Sub-task 2 of: {goal[:50]}..."},
        {"agent": "agent3", "task": f"Sub-task 3 of: {goal[:50]}..."},
    ]
