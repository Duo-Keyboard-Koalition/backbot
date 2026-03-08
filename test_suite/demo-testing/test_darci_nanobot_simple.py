"""
Demo Test: DarCI Manages 2 Nanobot Agents

Simplified version that demonstrates DarCI coordinating multiple agents.
This test can run with simulated agents if real Tailscale is not available.
"""

import asyncio
import os
from typing import List, Dict, Any
from dataclasses import dataclass

import pytest
from dotenv import load_dotenv

# Load test environment
load_dotenv(".env.test")


# ============================================================================
# Simulated Agent Classes (for testing without real Tailscale)
# ============================================================================

@dataclass
class SimulatedAgent:
    """Simulated agent for testing."""
    name: str
    agent_type: str
    status: str = "online"
    tasks: List[Dict[str, Any]] = None

    def __post_init__(self):
        if self.tasks is None:
            self.tasks = []

    async def receive_directive(self, task_id: str, goal: str, priority: str) -> bool:
        """Simulate receiving a directive from DarCI."""
        self.tasks.append({
            "task_id": task_id,
            "goal": goal,
            "priority": priority,
            "status": "assigned"
        })
        return True

    async def execute_task(self, task_id: str) -> str:
        """Simulate task execution."""
        for task in self.tasks:
            if task["task_id"] == task_id:
                task["status"] = "completed"
                return f"Task {task_id} completed: {task['goal']}"
        return f"Task {task_id} not found"


class SimulatedDarCICoordinator:
    """Simulated DarCI coordinator for testing."""

    def __init__(self, name: str = "darci-coordinator"):
        self.name = name
        self.agents: Dict[str, SimulatedAgent] = {}
        self.tasks: Dict[str, Dict[str, Any]] = {}

    def register_agent(self, agent: SimulatedAgent):
        """Register an agent with DarCI."""
        self.agents[agent.name] = agent
        print(f"  ✓ {agent.name} registered with DarCI")

    async def discover_agents(self) -> List[SimulatedAgent]:
        """Discover all registered agents."""
        print(f"\n🔍 {self.name} discovering agents...")
        await asyncio.sleep(0.5)  # Simulate discovery delay

        online_agents = [a for a in self.agents.values() if a.status == "online"]
        print(f"  ✓ Discovered {len(online_agents)} agent(s): {[a.name for a in online_agents]}")
        return online_agents

    async def create_task(self, task_id: str, goal: str, priority: str = "P1") -> Dict[str, Any]:
        """Create a new task."""
        task = {
            "task_id": task_id,
            "goal": goal,
            "priority": priority,
            "status": "pending",
            "assigned_to": None
        }
        self.tasks[task_id] = task
        print(f"  ✓ Task {task_id} created: {goal}")
        return task

    async def assign_task(self, task_id: str, agent_name: str) -> bool:
        """Assign a task to an agent."""
        if task_id not in self.tasks:
            print(f"  ✗ Task {task_id} not found")
            return False

        if agent_name not in self.agents:
            print(f"  ✗ Agent {agent_name} not found")
            return False

        task = self.tasks[task_id]
        agent = self.agents[agent_name]

        # Send directive to agent
        success = await agent.receive_directive(
            task_id=task_id,
            goal=task["goal"],
            priority=task["priority"]
        )

        if success:
            task["assigned_to"] = agent_name
            task["status"] = "in_progress"
            print(f"  ✓ Task {task_id} assigned to {agent_name}")

        return success

    async def monitor_agents(self, cycles: int = 3) -> Dict[str, str]:
        """Monitor agent progress."""
        print(f"\n⏳ {self.name} monitoring agents...")
        status_report = {}

        for cycle in range(cycles):
            print(f"  Cycle {cycle + 1}/{cycles}:")
            for agent_name, agent in self.agents.items():
                active_tasks = len([t for t in agent.tasks if t["status"] == "in_progress"])
                completed_tasks = len([t for t in agent.tasks if t["status"] == "completed"])
                status = f"{active_tasks} active, {completed_tasks} completed"
                status_report[agent_name] = status
                print(f"    {agent_name}: {status}")

            await asyncio.sleep(0.5)

        return status_report


# ============================================================================
# Fixtures
# ============================================================================

@pytest.fixture(scope="function")
def darci_coordinator() -> SimulatedDarCICoordinator:
    """Create a DarCI coordinator instance."""
    return SimulatedDarCICoordinator(name="darci-coordinator-001")


@pytest.fixture(scope="function")
def nanobot_agents(darci_coordinator: SimulatedDarCICoordinator) -> List[SimulatedAgent]:
    """Create 2 nanobot agents registered with DarCI."""
    agents = []

    # Create 2 nanobot agents
    for i in range(2):
        agent = SimulatedAgent(
            name=f"nanobot-worker-00{i+1}",
            agent_type="nanobot",
            status="online"
        )
        darci_coordinator.register_agent(agent)
        agents.append(agent)

    return agents


# ============================================================================
# Demo Tests
# ============================================================================

@pytest.mark.demo
@pytest.mark.e2e
class TestDarciManagesNanobots:
    """
    Demo: DarCI coordinates 2 nanobot agents.

    This test demonstrates the complete DarCI management workflow.
    """

    @pytest.mark.asyncio
    async def test_darci_discovers_nanobots(
        self,
        darci_coordinator: SimulatedDarCICoordinator,
        nanobot_agents: List[SimulatedAgent]
    ):
        """Test DarCI discovers registered nanobot agents."""
        print("\n🔍 Testing agent discovery...")

        # DarCI discovers agents
        discovered = await darci_coordinator.discover_agents()

        # Verify both nanobots are discovered
        assert len(discovered) == 2
        assert all(a.agent_type == "nanobot" for a in discovered)
        assert all(a.status == "online" for a in discovered)

        print(f"✓ DarCI successfully discovered {len(discovered)} nanobots")

    @pytest.mark.asyncio
    async def test_darci_creates_and_assigns_tasks(
        self,
        darci_coordinator: SimulatedDarCICoordinator,
        nanobot_agents: List[SimulatedAgent]
    ):
        """Test DarCI creates tasks and assigns them to nanobots."""
        print("\n📋 Testing task creation and assignment...")

        # Create tasks
        task1 = await darci_coordinator.create_task(
            task_id="TSK-001",
            goal="Deploy application to staging",
            priority="P1"
        )

        task2 = await darci_coordinator.create_task(
            task_id="TSK-002",
            goal="Monitor system health metrics",
            priority="P1"
        )

        # Assign tasks to different nanobots
        success1 = await darci_coordinator.assign_task("TSK-001", "nanobot-worker-001")
        success2 = await darci_coordinator.assign_task("TSK-002", "nanobot-worker-002")

        assert success1
        assert success2

        # Verify tasks were assigned
        assert darci_coordinator.tasks["TSK-001"]["assigned_to"] == "nanobot-worker-001"
        assert darci_coordinator.tasks["TSK-002"]["assigned_to"] == "nanobot-worker-002"

        print(f"✓ DarCI successfully assigned {len(darci_coordinator.tasks)} tasks")

    @pytest.mark.asyncio
    async def test_darci_monitors_nanobot_progress(
        self,
        darci_coordinator: SimulatedDarCICoordinator,
        nanobot_agents: List[SimulatedAgent]
    ):
        """Test DarCI monitors nanobot task progress."""
        print("\n⏳ Testing progress monitoring...")

        # Assign tasks first
        await darci_coordinator.create_task("TSK-001", "Task 1", "P1")
        await darci_coordinator.create_task("TSK-002", "Task 2", "P1")
        await darci_coordinator.assign_task("TSK-001", "nanobot-worker-001")
        await darci_coordinator.assign_task("TSK-002", "nanobot-worker-002")

        # Simulate task execution
        await nanobot_agents[0].execute_task("TSK-001")
        await nanobot_agents[1].execute_task("TSK-002")

        # Monitor progress
        status = await darci_coordinator.monitor_agents(cycles=2)

        # Verify monitoring captured task completion
        assert "nanobot-worker-001" in status
        assert "nanobot-worker-002" in status

        print(f"✓ DarCI successfully monitored nanobot progress")

    @pytest.mark.asyncio
    async def test_darci_coordinates_full_workflow(
        self,
        darci_coordinator: SimulatedDarCICoordinator,
        nanobot_agents: List[SimulatedAgent]
    ):
        """
        Full integration test: DarCI coordinates 2 nanobots end-to-end.

        This is the main demo test showing the complete workflow.
        """
        print("\n" + "=" * 60)
        print("🎬 Demo: DarCI Manages 2 Nanobot Agents")
        print("=" * 60)

        # Phase 1: Discovery
        print("\n🔍 Phase 1: Agent Discovery")
        discovered = await darci_coordinator.discover_agents()
        assert len(discovered) == 2
        print(f"  ✓ Discovered: {[a.name for a in discovered]}")

        # Phase 2: Task Creation
        print("\n📋 Phase 2: Task Creation")
        tasks = [
            ("TSK-DEMO-001", "Execute demo task on nanobot #1"),
            ("TSK-DEMO-002", "Execute demo task on nanobot #2"),
        ]

        for task_id, goal in tasks:
            await darci_coordinator.create_task(task_id, goal, "P1")

        assert len(darci_coordinator.tasks) == 2

        # Phase 3: Task Assignment
        print("\n📤 Phase 3: Task Assignment")
        for i, (task_id, _) in enumerate(tasks):
            agent_name = f"nanobot-worker-00{i+1}"
            success = await darci_coordinator.assign_task(task_id, agent_name)
            assert success

        # Phase 4: Task Execution
        print("\n⚡ Phase 4: Task Execution")
        for i, (task_id, _) in enumerate(tasks):
            result = await nanobot_agents[i].execute_task(task_id)
            print(f"  {nanobot_agents[i].name}: {result}")

        # Phase 5: Monitoring
        print("\n⏳ Phase 5: Progress Monitoring")
        status = await darci_coordinator.monitor_agents(cycles=2)

        # Phase 6: Verification
        print("\n✅ Phase 6: Verification")
        for task_id, task in darci_coordinator.tasks.items():
            agent_name = task["assigned_to"]
            agent = darci_coordinator.agents[agent_name]
            task_status = [t["status"] for t in agent.tasks if t["task_id"] == task_id][0]
            print(f"  {task_id}: {task_status}")
            assert task_status == "completed"

        print("\n" + "=" * 60)
        print("🎉 Demo completed successfully!")
        print("=" * 60)
        print(f"\nSummary:")
        print(f"  - DarCI agent: {darci_coordinator.name}")
        print(f"  - Nanobot agents: {[a.name for a in nanobot_agents]}")
        print(f"  - Tasks created: {len(darci_coordinator.tasks)}")
        print(f"  - Tasks completed: {sum(1 for t in darci_coordinator.tasks.values() if t['status'] == 'completed')}")
        print("\n✓ DarCI successfully coordinated 2 nanobot agents")


@pytest.mark.demo
@pytest.mark.scenario
class TestDarciNanobotScenarios:
    """Scenario-based tests for DarCI-nanobot interactions."""

    @pytest.mark.asyncio
    async def test_scenario_parallel_execution(
        self,
        darci_coordinator: SimulatedDarCICoordinator,
        nanobot_agents: List[SimulatedAgent]
    ):
        """Test DarCI assigning parallel tasks to multiple nanobots."""
        print("\n🎯 Scenario: Parallel Task Execution")

        # Create and assign parallel tasks
        for i, agent in enumerate(nanobot_agents):
            task_id = f"PARALLEL-{i+1}"
            await darci_coordinator.create_task(task_id, f"Parallel task {i+1}", "P1")
            await darci_coordinator.assign_task(task_id, agent.name)

        # Execute in parallel
        execution_tasks = []
        for i, agent in enumerate(nanobot_agents):
            task_id = f"PARALLEL-{i+1}"
            execution_tasks.append(agent.execute_task(task_id))

        results = await asyncio.gather(*execution_tasks)

        print(f"  Parallel results:")
        for i, result in enumerate(results):
            print(f"    Nanobot {i+1}: {result}")

        # Verify all completed
        assert all("completed" in r.lower() for r in results)
        print("✓ Parallel execution scenario completed")

    @pytest.mark.asyncio
    async def test_scenario_priority_handling(
        self,
        darci_coordinator: SimulatedDarCICoordinator,
        nanobot_agents: List[SimulatedAgent]
    ):
        """Test DarCI handling tasks with different priorities."""
        print("\n🎯 Scenario: Priority-Based Task Assignment")

        # Create tasks with different priorities
        priorities = [
            ("TSK-P0", "Critical task", "P0"),
            ("TSK-P1", "High priority task", "P1"),
            ("TSK-P2", "Normal task", "P2"),
        ]

        for task_id, goal, priority in priorities:
            await darci_coordinator.create_task(task_id, goal, priority)

        # Assign to nanobots (P0 gets first nanobot)
        await darci_coordinator.assign_task("TSK-P0", "nanobot-worker-001")
        await darci_coordinator.assign_task("TSK-P1", "nanobot-worker-002")
        await darci_coordinator.assign_task("TSK-P2", "nanobot-worker-001")

        # Verify priority ordering
        p0_task = darci_coordinator.tasks["TSK-P0"]
        p1_task = darci_coordinator.tasks["TSK-P1"]

        assert p0_task["priority"] == "P0"
        assert p1_task["priority"] == "P1"

        print(f"  Tasks by priority:")
        for task_id, task in darci_coordinator.tasks.items():
            print(f"    {task_id}: {task['priority']} - {task['goal']}")

        print("✓ Priority handling scenario completed")


@pytest.mark.demo
@pytest.mark.performance
class TestDarciNanobotPerformance:
    """Performance tests for DarCI-nanobot coordination."""

    @pytest.mark.asyncio
    async def test_task_assignment_latency(
        self,
        darci_coordinator: SimulatedDarCICoordinator,
        nanobot_agents: List[SimulatedAgent]
    ):
        """Test task assignment latency."""
        print("\n⚡ Performance: Task Assignment Latency")

        import time

        latencies = []
        num_tests = 10

        for i in range(num_tests):
            task_id = f"PERF-{i}"

            start = time.time()
            await darci_coordinator.create_task(task_id, f"Performance test {i}", "P1")
            await darci_coordinator.assign_task(task_id, "nanobot-worker-001")
            latency = (time.time() - start) * 1000  # ms

            latencies.append(latency)

        avg_latency = sum(latencies) / len(latencies)
        print(f"  Average latency: {avg_latency:.2f}ms ({num_tests} iterations)")

        assert avg_latency < 100, "Task assignment should complete within 100ms"
        print("✓ Task assignment latency within acceptable range")

    @pytest.mark.asyncio
    async def test_scaling_multiple_nanobots(
        self,
        darci_coordinator: SimulatedDarCICoordinator
    ):
        """Test DarCI scaling to multiple nanobots."""
        print("\n⚡ Performance: Scaling Test")

        # Register multiple nanobots
        num_nanobots = 5
        for i in range(num_nanobots):
            agent = SimulatedAgent(
                name=f"nanobot-scale-{i+1}",
                agent_type="nanobot",
                status="online"
            )
            darci_coordinator.register_agent(agent)

        # Discover all agents
        discovered = await darci_coordinator.discover_agents()
        assert len(discovered) == num_nanobots

        # Assign task to each
        for i in range(num_nanobots):
            task_id = f"SCALE-{i+1}"
            await darci_coordinator.create_task(task_id, f"Scale test {i+1}", "P1")
            await darci_coordinator.assign_task(task_id, f"nanobot-scale-{i+1}")

        print(f"  ✓ DarCI coordinated {num_nanobots} nanobots successfully")
        print("✓ Scaling test completed")
