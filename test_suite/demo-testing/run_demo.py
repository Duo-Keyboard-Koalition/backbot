#!/usr/bin/env python3
"""
Run DarCI Manages Nanobots Demo

This script runs the demo test without requiring pytest.
It demonstrates DarCI coordinating 2 nanobot agents.
"""

import asyncio
from dataclasses import dataclass
from typing import List, Dict, Any


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


async def run_demo():
    """Run the full DarCI manages nanobots demo."""
    print("=" * 70)
    print("🎬 Demo: DarCI Manages 2 Nanobot Agents")
    print("=" * 70)

    # Initialize DarCI coordinator
    darci = SimulatedDarCICoordinator(name="darci-coordinator-001")

    # Create and register 2 nanobot agents
    print("\n🚀 Starting nanobot agents...")
    nanobots = []
    for i in range(2):
        agent = SimulatedAgent(
            name=f"nanobot-worker-00{i+1}",
            agent_type="nanobot",
            status="online"
        )
        darci.register_agent(agent)
        nanobots.append(agent)

    # Phase 1: Discovery
    print("\n🔍 Phase 1: Agent Discovery")
    discovered = await darci.discover_agents()
    assert len(discovered) == 2, "Should discover 2 nanobots"
    print(f"  ✓ Confirmed: {[a.name for a in discovered]}")

    # Phase 2: Task Creation
    print("\n📋 Phase 2: Task Creation")
    tasks = [
        ("TSK-DEMO-001", "Deploy application to staging environment"),
        ("TSK-DEMO-002", "Monitor system health and report metrics"),
    ]

    for task_id, goal in tasks:
        await darci.create_task(task_id, goal, "P1")

    assert len(darci.tasks) == 2, "Should create 2 tasks"

    # Phase 3: Task Assignment
    print("\n📤 Phase 3: Task Assignment")
    for i, (task_id, goal) in enumerate(tasks):
        agent_name = f"nanobot-worker-00{i+1}"
        success = await darci.assign_task(task_id, agent_name)
        assert success, f"Failed to assign {task_id}"

    # Phase 4: Task Execution
    print("\n⚡ Phase 4: Task Execution")
    for i, (task_id, goal) in enumerate(tasks):
        result = await nanobots[i].execute_task(task_id)
        print(f"  {nanobots[i].name}: {result}")

    # Phase 5: Monitoring
    print("\n⏳ Phase 5: Progress Monitoring")
    status = await darci.monitor_agents(cycles=2)

    # Phase 6: Verification
    print("\n✅ Phase 6: Verification")
    all_completed = True
    for task_id, task in darci.tasks.items():
        agent_name = task["assigned_to"]
        agent = darci.agents[agent_name]
        task_status = [t["status"] for t in agent.tasks if t["task_id"] == task_id][0]
        status_icon = "✓" if task_status == "completed" else "✗"
        print(f"  {status_icon} {task_id}: {task_status}")
        if task_status != "completed":
            all_completed = False

    # Summary
    print("\n" + "=" * 70)
    if all_completed:
        print("🎉 Demo completed successfully!")
    else:
        print("⚠ Demo completed with issues")
    print("=" * 70)
    print(f"\nSummary:")
    print(f"  - DarCI agent: {darci.name}")
    print(f"  - Nanobot agents: {[a.name for a in nanobots]}")
    print(f"  - Tasks created: {len(darci.tasks)}")
    print(f"  - Tasks completed: {sum(1 for t in darci.tasks.values() if t['status'] == 'completed')}/{len(darci.tasks)}")

    if all_completed:
        print("\n✅ SUCCESS: DarCI successfully coordinated 2 nanobot agents!")
    else:
        print("\n❌ FAILURE: Some tasks did not complete")

    return all_completed


if __name__ == "__main__":
    result = asyncio.run(run_demo())
    exit(0 if result else 1)
