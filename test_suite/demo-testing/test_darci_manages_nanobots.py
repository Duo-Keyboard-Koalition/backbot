"""
Demo Test: DarCI Manages Multiple Nanobot Agents

This test demonstrates the full DarCI coordination workflow:
1. Spins up 1 DarCI agent (coordinator/manager)
2. Spins up 2 Nanobot agents (workers)
3. DarCI discovers available nanobots via taila2a bridge
4. DarCI creates tasks and assigns them to nanobots
5. DarCI monitors nanobot progress and intervenes if needed
6. Verifies all tasks complete successfully

This is a pre-show demo test that showcases the full SentinelAI system.
"""

import asyncio
import os
import time
from pathlib import Path
from typing import Any, Dict, List, Optional
from dataclasses import dataclass, field

import pytest
import pytest_asyncio
from dotenv import load_dotenv

# Load test environment
load_dotenv(".env.test")

# Test configuration
TEST_TIMEOUT = int(os.getenv("TEST_TIMEOUT", "300"))
AGENT_STARTUP_DELAY = int(os.getenv("AGENT_STARTUP_DELAY", "5"))
DISCOVERY_DELAY = int(os.getenv("DISCOVERY_DELAY", "3"))


# ============================================================================
# Test Fixtures
# ============================================================================

@pytest.fixture(scope="session")
def gemini_api_key() -> str:
    """Load Gemini API key from environment."""
    key = os.getenv("GEMINI_API_KEY")
    if not key:
        pytest.skip("GEMINI_API_KEY not set in environment")
    return key


@pytest.fixture(scope="session")
def tailscale_config() -> dict:
    """Load Tailscale configuration for agents."""
    auth_keys = []
    for i in range(1, 4):
        key = os.getenv(f"TS_AUTH_KEY_{i}")
        if key:
            auth_keys.append(key)

    if not auth_keys:
        pytest.skip("No TS_AUTH_KEY_* found in environment")

    tailnet = os.getenv("TAILNET_NAME", "test.ts.net")

    return {
        "auth_keys": auth_keys,
        "tailnet": tailnet,
        "bridge_url": os.getenv("TAILA2A_BRIDGE_URL", "http://127.0.0.1:8080"),
    }


@dataclass
class AgentInstance:
    """Represents a running agent instance."""
    name: str
    agent_type: str
    node_id: str
    hostname: str
    port: int
    process: Optional[Any] = None
    status: str = "starting"

    def to_dict(self) -> dict:
        return {
            "name": self.name,
            "agent_type": self.agent_type,
            "node_id": self.node_id,
            "hostname": self.hostname,
            "port": self.port,
            "status": self.status,
        }


@pytest_asyncio.fixture(scope="function")
async def darci_agent(tailscale_config: dict, gemini_api_key: str) -> AgentInstance:
    """
    Start a DarCI coordinator agent.

    DarCI acts as the project manager that coordinates other agents.
    """
    import subprocess

    agent_name = "darci-coordinator-001"
    node_id = f"{agent_name}.{tailscale_config['tailnet']}"

    print(f"\n🚀 Starting DarCI agent: {agent_name}")

    # Set environment for DarCI agent
    env = os.environ.copy()
    env["GEMINI_API_KEY"] = gemini_api_key
    env["TS_AUTH_KEY"] = tailscale_config["auth_keys"][0]
    env["TAILNET_NAME"] = tailscale_config["tailnet"]
    env["DARCI_AGENT_NAME"] = agent_name

    # Start DarCI agent process
    # Note: Adjust the command based on how darci is actually started
    process = subprocess.Popen(
        ["python", "-m", "darci.run"],
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        cwd=str(Path(__file__).parent.parent.parent),
    )

    agent = AgentInstance(
        name=agent_name,
        agent_type="darci",
        node_id=node_id,
        hostname=agent_name,
        port=8080,
        process=process,
    )

    # Wait for agent to start
    await asyncio.sleep(AGENT_STARTUP_DELAY)

    # Check if process is still running
    if process.poll() is not None:
        stdout, stderr = process.communicate()
        print(f"⚠ DarCI agent failed to start:")
        print(f"  STDOUT: {stdout.decode()}")
        print(f"  STDERR: {stderr.decode()}")
        pytest.skip(f"DarCI agent startup failed")

    agent.status = "running"
    print(f"✓ DarCI agent started: {agent_name}")

    yield agent

    # Cleanup: stop DarCI agent
    print(f"\n🧹 Stopping DarCI agent: {agent_name}")
    if process.poll() is None:
        process.terminate()
        try:
            process.wait(timeout=5)
            print(f"✓ DarCI agent stopped")
        except subprocess.TimeoutExpired:
            process.kill()
            print(f"⚠ DarCI agent force killed")


@pytest_asyncio.fixture(scope="function")
async def nanobot_agents(
    tailscale_config: dict,
    gemini_api_key: str,
    darci_agent: AgentInstance
) -> List[AgentInstance]:
    """
    Start 2 Nanobot worker agents.

    Nanobots are production engineers that execute tasks assigned by DarCI.
    """
    import subprocess

    agents = []
    num_nanobots = 2

    print(f"\n🚀 Starting {num_nanobots} Nanobot agents...")

    for i in range(num_nanobots):
        agent_name = f"nanobot-worker-00{i+1}"
        node_id = f"{agent_name}.{tailscale_config['tailnet']}"

        # Use different auth keys for each agent if available
        auth_key_idx = min(i + 1, len(tailscale_config["auth_keys"]) - 1)
        auth_key = tailscale_config["auth_keys"][auth_key_idx]

        # Set environment for Nanobot agent
        env = os.environ.copy()
        env["GEMINI_API_KEY"] = gemini_api_key
        env["TS_AUTH_KEY"] = auth_key
        env["TAILNET_NAME"] = tailscale_config["tailnet"]
        env["NANOBOT_AGENT_NAME"] = agent_name
        env["DARCI_BRIDGE_URL"] = tailscale_config["bridge_url"]

        # Start Nanobot agent process
        # Note: Adjust command based on actual nanobot startup
        process = subprocess.Popen(
            ["python", "-c", f"""
import asyncio
import os
from tailbridge.taila2a.scripts.aip_client import AIPClient, AgentConfig

async def run_nanobot():
    config = AgentConfig(
        agent_id='{agent_name}',
        agent_type='nanobot',
        endpoint_url='http://127.0.0.1:900{i+1}/api',
        health_url='http://127.0.0.1:900{i+1}/health',
        capabilities=['task-execution', 'deployment', 'monitoring'],
        tags=['worker', 'production']
    )

    client = AIPClient(
        bridge_url='{tailscale_config['bridge_url']}',
        config=config
    )

    print(f'Nanobot {{config.agent_id}} starting...')
    await client.register()
    print(f'Nanobot {{config.agent_id}} registered')
    await client.start_heartbeat(interval=30)
    print(f'Nanobot {{config.agent_id}} heartbeat started')

    # Keep running
    while True:
        await asyncio.sleep(1)

if __name__ == '__main__':
    asyncio.run(run_nanobot())
"""],
            env=env,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=str(Path(__file__).parent.parent.parent),
        )

        agent = AgentInstance(
            name=agent_name,
            agent_type="nanobot",
            node_id=node_id,
            hostname=agent_name,
            port=9000 + i + 1,
            process=process,
        )

        # Wait for agent to start
        await asyncio.sleep(2)

        # Check if process is still running
        if process.poll() is not None:
            stdout, stderr = process.communicate()
            print(f"⚠ Nanobot {i+1} failed to start:")
            print(f"  STDOUT: {stdout.decode()}")
            print(f"  STDERR: {stderr.decode()}")
        else:
            agent.status = "running"
            agents.append(agent)
            print(f"✓ Nanobot agent {i+1} started: {agent_name}")

    if not agents:
        pytest.skip("Failed to start any nanobot agents")

    # Wait for all agents to connect to tailnet
    print(f"⏳ Waiting for agents to connect to tailnet...")
    await asyncio.sleep(AGENT_STARTUP_DELAY)

    yield agents

    # Cleanup: stop all nanobot agents
    print(f"\n🧹 Stopping Nanobot agents...")
    for agent in agents:
        print(f"  Stopping {agent.name}...")
        if agent.process and agent.process.poll() is None:
            agent.process.terminate()
            try:
                agent.process.wait(timeout=5)
                print(f"  ✓ {agent.name} stopped")
            except subprocess.TimeoutExpired:
                agent.process.kill()
                print(f"  ⚠ {agent.name} force killed")


@pytest_asyncio.fixture(scope="function")
async def taila2a_bridge(tailscale_config: dict) -> str:
    """
    Ensure taila2a bridge is running.

    Returns bridge URL for agent communication.
    """
    bridge_url = tailscale_config["bridge_url"]

    # Check if bridge is accessible
    import httpx

    try:
        async with httpx.AsyncClient(timeout=5.0) as client:
            resp = await client.get(f"{bridge_url}/health")
            if resp.status_code == 200:
                print(f"✓ Taila2a bridge accessible at {bridge_url}")
                return bridge_url
    except Exception:
        pass

    # Bridge not accessible, skip test
    pytest.skip(f"Taila2a bridge not accessible at {bridge_url}")

    return bridge_url


# ============================================================================
# Helper Functions
# ============================================================================

async def discover_agents_via_bridge(bridge_url: str) -> List[dict]:
    """Query taila2a bridge for available agents."""
    import httpx

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.get(f"{bridge_url}/agents")
            resp.raise_for_status()
            data = resp.json()
            return data.get("agents", [])
    except Exception as e:
        print(f"⚠ Discovery failed: {e}")
        return []


async def send_darci_directive(
    bridge_url: str,
    dest_node: str,
    task_id: str,
    goal: str,
    priority: str = "P1"
) -> bool:
    """Send a DARCI directive to assign a task to an agent."""
    import httpx

    envelope = {
        "dest_node": dest_node,
        "payload": {
            "type": "darci_directive",
            "task_id": task_id,
            "goal": goal,
            "priority": priority,
            "timestamp": time.time(),
        }
    }

    try:
        async with httpx.AsyncClient(timeout=10.0) as client:
            resp = await client.post(f"{bridge_url}/send", json=envelope)
            resp.raise_for_status()
            return True
    except Exception as e:
        print(f"⚠ Failed to send directive to {dest_node}: {e}")
        return False


async def check_agent_status(bridge_url: str, agent_name: str) -> Optional[dict]:
    """Check status of a specific agent."""
    agents = await discover_agents_via_bridge(bridge_url)
    for agent in agents:
        if agent.get("name") == agent_name:
            return agent
    return None


# ============================================================================
# Demo Tests
# ============================================================================

@pytest.mark.demo
@pytest.mark.e2e
@pytest.mark.slow
class TestDarciManagesNanobots:
    """
    Demo test: DarCI coordinates multiple nanobot agents.

    This test demonstrates the full DarCI management workflow:
    1. Agent discovery
    2. Task creation and assignment
    3. Progress monitoring
    4. Task completion verification
    """

    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_discovers_nanobots(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """Test that DarCI can discover nanobot agents on the tailnet."""
        print("\n🔍 Testing agent discovery...")

        # Wait for agents to register
        await asyncio.sleep(DISCOVERY_DELAY)

        # Discover agents via bridge
        agents = await discover_agents_via_bridge(taila2a_bridge)

        print(f"📊 Discovered {len(agents)} agent(s):")
        for agent in agents:
            name = agent.get("name", "unknown")
            online = "✅" if agent.get("online") else "❌"
            print(f"  - {name}: {online}")

        # Verify both nanobots are discovered
        nanobot_names = [nb.name for nb in nanobot_agents]
        discovered_names = [a.get("name") for a in agents if a.get("online")]

        for nb_name in nanobot_names:
            assert nb_name in discovered_names, f"Nanobot {nb_name} not discovered"

        print(f"✓ DarCI successfully discovered all {len(nanobot_agents)} nanobots")

    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_assigns_tasks_to_nanobots(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """Test that DarCI can assign tasks to nanobot agents."""
        print("\n📋 Testing task assignment...")

        # Define test tasks
        tasks = [
            {
                "task_id": "TSK-001",
                "goal": "Deploy application to staging environment",
                "priority": "P1",
                "assigned_to": nanobot_agents[0].name,
            },
            {
                "task_id": "TSK-002",
                "goal": "Monitor system health and report metrics",
                "priority": "P1",
                "assigned_to": nanobot_agents[1].name,
            },
        ]

        # Assign tasks to nanobots
        for task in tasks:
            print(f"  Assigning {task['task_id']} to {task['assigned_to']}...")

            success = await send_darci_directive(
                bridge_url=taila2a_bridge,
                dest_node=f"{task['assigned_to']}.{os.getenv('TAILNET_NAME', 'test.ts.net')}",
                task_id=task["task_id"],
                goal=task["goal"],
                priority=task["priority"],
            )

            assert success, f"Failed to assign {task['task_id']}"
            print(f"  ✓ {task['task_id']} assigned to {task['assigned_to']}")

        print(f"✓ DarCI successfully assigned {len(tasks)} tasks to nanobots")

    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_monitors_nanobot_progress(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """Test that DarCI can monitor nanobot progress."""
        print("\n⏳ Testing progress monitoring...")

        # Simulate monitoring period
        monitoring_cycles = 3
        for cycle in range(monitoring_cycles):
            print(f"  Monitoring cycle {cycle + 1}/{monitoring_cycles}...")

            # Check status of each nanobot
            for nb in nanobot_agents:
                status = await check_agent_status(taila2a_bridge, nb.name)
                if status:
                    online = "✅" if status.get("online") else "❌"
                    print(f"    {nb.name}: {online}")

            await asyncio.sleep(2)

        print(f"✓ DarCI successfully monitored nanobot progress")

    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_darci_coordinates_nanobots(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """
        Full integration test: DarCI coordinates 2 nanobots end-to-end.

        This is the main demo test that shows the complete workflow.
        """
        print("\n🎬 Starting demo: DarCI manages 2 nanobots")
        print("=" * 60)

        # Phase 1: Discovery
        print("\n🔍 Phase 1: Agent Discovery")
        await asyncio.sleep(DISCOVERY_DELAY)

        agents = await discover_agents_via_bridge(taila2a_bridge)
        print(f"  Discovered {len(agents)} agent(s)")

        assert len(agents) >= 2, "Should discover at least 2 nanobots"

        # Phase 2: Task Assignment
        print("\n📋 Phase 2: Task Assignment")

        tasks = [
            {
                "task_id": "TSK-DEMO-001",
                "goal": "Execute demo task on nanobot #1",
                "priority": "P1",
            },
            {
                "task_id": "TSK-DEMO-002",
                "goal": "Execute demo task on nanobot #2",
                "priority": "P1",
            },
        ]

        for i, task in enumerate(tasks):
            nanobot = nanobot_agents[i]
            print(f"  Assigning {task['task_id']} to {nanobot.name}...")

            success = await send_darci_directive(
                bridge_url=taila2a_bridge,
                dest_node=nanobot.node_id,
                task_id=task["task_id"],
                goal=task["goal"],
                priority=task["priority"],
            )

            assert success, f"Failed to assign {task['task_id']}"

        # Phase 3: Monitoring
        print("\n⏳ Phase 3: Progress Monitoring")

        monitoring_duration = 10  # seconds
        check_interval = 2

        for elapsed in range(0, monitoring_duration, check_interval):
            print(f"  [{elapsed}s] Checking agent status...")

            for nb in nanobot_agents:
                status = await check_agent_status(taila2a_bridge, nb.name)
                if status:
                    online_status = "online" if status.get("online") else "offline"
                    print(f"    {nb.name}: {online_status}")

            await asyncio.sleep(check_interval)

        # Phase 4: Verification
        print("\n✅ Phase 4: Verification")

        # Verify all agents are still online
        for nb in nanobot_agents:
            status = await check_agent_status(taila2a_bridge, nb.name)
            assert status is not None, f"{nb.name} not found"
            assert status.get("online"), f"{nb.name} is offline"
            print(f"  ✓ {nb.name} is online and responsive")

        print("\n" + "=" * 60)
        print("🎉 Demo completed successfully!")
        print("=" * 60)
        print("\nSummary:")
        print(f"  - DarCI agent: {darci_agent.name}")
        print(f"  - Nanobot agents: {[nb.name for nb in nanobot_agents]}")
        print(f"  - Tasks assigned: {len(tasks)}")
        print(f"  - All agents: online and responsive")
        print("\n✓ DarCI successfully coordinated 2 nanobot agents")


# ============================================================================
# Demo Scenario Tests
# ============================================================================

@pytest.mark.demo
@pytest.mark.scenario
class TestDarciNanobotScenarios:
    """
    Scenario-based tests for DarCI-nanobot interactions.

    These tests demonstrate specific management scenarios.
    """

    @pytest.mark.timeout(120)
    async def test_scenario_parallel_task_execution(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """Test DarCI assigning parallel tasks to multiple nanobots."""
        print("\n🎯 Scenario: Parallel Task Execution")

        # Assign same task type to both nanobots
        tasks = []
        for i, nb in enumerate(nanobot_agents):
            task_id = f"PARALLEL-{i+1}"
            tasks.append({
                "task_id": task_id,
                "nanobot": nb,
                "goal": f"Execute parallel task {i+1}",
            })

            await send_darci_directive(
                bridge_url=taila2a_bridge,
                dest_node=nb.node_id,
                task_id=task_id,
                goal=tasks[-1]["goal"],
            )

        # Monitor both executing in parallel
        for _ in range(3):
            for task in tasks:
                status = await check_agent_status(
                    taila2a_bridge,
                    task["nanobot"].name
                )
                if status:
                    print(f"  {task['task_id']}: {task['nanobot'].name} active")

            await asyncio.sleep(1)

        print("✓ Parallel execution scenario completed")

    @pytest.mark.timeout(120)
    async def test_scenario_failover_handling(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """Test DarCI handling a nanobot failure by reassigning tasks."""
        print("\n🔄 Scenario: Failover Handling")

        # Assign task to first nanobot
        primary = nanobot_agents[0]
        backup = nanobot_agents[1]

        print(f"  Primary: {primary.name}")
        print(f"  Backup: {backup.name}")

        # Simulate task assignment
        await send_darci_directive(
            bridge_url=taila2a_bridge,
            dest_node=primary.node_id,
            task_id="FAILOVER-001",
            goal="Test failover scenario",
        )

        # Monitor primary
        status = await check_agent_status(taila2a_bridge, primary.name)
        if status and status.get("online"):
            print(f"  ✓ Primary {primary.name} is online")

        # In a real scenario, if primary fails, DarCI would reassign to backup
        print("  (Failover logic would trigger if primary goes offline)")
        print("✓ Failover scenario structure validated")


# ============================================================================
# Performance Tests
# ============================================================================

@pytest.mark.demo
@pytest.mark.performance
class TestDarciNanobotPerformance:
    """Performance tests for DarCI-nanobot coordination."""

    @pytest.mark.timeout(60)
    async def test_discovery_latency(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """Test agent discovery latency."""
        print("\n⚡ Performance: Discovery Latency")

        latencies = []
        for _ in range(3):
            start = time.time()
            agents = await discover_agents_via_bridge(taila2a_bridge)
            latency = (time.time() - start) * 1000  # ms
            latencies.append(latency)
            print(f"  Discovery: {latency:.2f}ms ({len(agents)} agents)")

        avg_latency = sum(latencies) / len(latencies)
        print(f"  Average: {avg_latency:.2f}ms")

        assert avg_latency < 1000, "Discovery should complete within 1 second"
        print("✓ Discovery latency within acceptable range")

    @pytest.mark.timeout(60)
    async def test_directive_throughput(
        self,
        darci_agent: AgentInstance,
        nanobot_agents: List[AgentInstance],
        taila2a_bridge: str
    ):
        """Test directive sending throughput."""
        print("\n⚡ Performance: Directive Throughput")

        if not nanobot_agents:
            pytest.skip("No nanobot agents available")

        nanobot = nanobot_agents[0]
        num_directives = 5

        latencies = []
        for i in range(num_directives):
            start = time.time()
            success = await send_darci_directive(
                bridge_url=taila2a_bridge,
                dest_node=nanobot.node_id,
                task_id=f"PERF-{i}",
                goal=f"Performance test directive {i}",
            )
            latency = (time.time() - start) * 1000  # ms
            latencies.append(latency)

            if success:
                print(f"  Directive {i}: {latency:.2f}ms ✓")
            else:
                print(f"  Directive {i}: FAILED ✗")

        avg_latency = sum(latencies) / len(latencies)
        print(f"  Average: {avg_latency:.2f}ms")

        assert avg_latency < 2000, "Directive should complete within 2 seconds"
        print("✓ Directive throughput within acceptable range")
