# SentinelAI Integration Testing Brainstorm

**Testing Strategy:** End-to-end integration tests using **real APIs only** (no mocks)
- **Gemini API**: Real Google Generative AI calls for agent reasoning
- **Tailscale API**: Real tailnet connectivity for agent-to-agent communication

---

## 🎯 Test Philosophy

### Why No Mocks?
1. **Catch Real Integration Issues**: Network latency, rate limits, API quirks
2. **Validate Authentication**: Real API key validation and permissions
3. **Test Failure Modes**: Actual API errors, timeouts, quota limits
4. **Confidence in Production**: If it passes here, it works in the wild

### Trade-offs
| Pros | Cons |
|------|------|
| ✅ Catches real integration bugs | ❌ Slower test execution |
| ✅ Validates auth & permissions | ❌ Requires API keys (cost) |
| ✅ Tests actual failure modes | ❌ Non-deterministic results |
| ✅ Production confidence | ❌ Rate limit constraints |

---

## 🏗️ Test Architecture

```
┌─────────────────────────────────────────────────────────────┐
│              Integration Test Runner (pytest)                │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │   Backend Tests  │         │   DarCI Tests    │         │
│  │   (FastAPI)      │         │   (Agent Orch.)  │         │
│  │                  │         │                  │         │
│  │  • Sentinel API  │         │  • Task Mgmt     │         │
│  │  • Risk Scoring  │         │  • Interventions │         │
│  │  • WebSocket     │         │  • Multi-agent   │         │
│  └──────────────────┘         └──────────────────┘         │
│                                                              │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │  Tailbridge      │         │   Full System    │         │
│  │  (Tailscale)     │         │   Tests          │         │
│  │                  │         │                  │         │
│  │  • A2A Comm      │         │  • E2E Scenarios │         │
│  │  • File Transfer │         │  • Load Testing  │         │
│  └──────────────────┘         └──────────────────┘         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
              ┌───────────────────────────────┐
              │      External APIs            │
              │  ┌─────────┐  ┌────────────┐  │
              │  │ Gemini  │  │ Tailscale  │  │
              │  │  API    │  │   API      │  │
              │  └─────────┘  └────────────┘  │
              └───────────────────────────────┘
```

---

## 📋 Test Categories

### 1. Backend + Gemini Integration Tests

#### 1.1 Agent Reasoning Tests
**Goal**: Validate Gemini-powered agent step execution

```python
# test_backend_gemini.py

async def test_agent_web_search_step():
    """Test agent can search web and get results via Gemini"""
    goal = "What is the capital of France?"
    state = ExecutionState(goal=goal)
    
    # Real Gemini call
    model = genai.GenerativeModel("gemini-2.0-flash")
    step = await run_agent_step(model, conversation_history, step_number=1)
    
    assert step.action == "web_search"
    assert "query" in step.action_input
    assert step.observation  # Should have search results
    
async def test_agent_loop_detection():
    """Test Sentinel detects loops in real agent execution"""
    goal = "Count from 1 to 5"
    
    # Simulate agent getting stuck in loop (real Gemini responses)
    steps = [
        Step(action="calculate", action_input={"expr": "1+1"}, observation="2"),
        Step(action="calculate", action_input={"expr": "1+1"}, observation="2"),  # Loop!
        Step(action="calculate", action_input={"expr": "1+1"}, observation="2"),  # Loop!
    ]
    
    risk_score, failures = score_step(goal, steps)
    assert FailureType.LOOP in failures
    assert risk_score > 0.5
    
async def test_agent_goal_drift_detection():
    """Test Sentinel detects when agent drifts from goal"""
    goal = "Calculate the square root of 144"
    
    # Agent starts on-topic, then drifts
    steps = [
        Step(thought="I need to calculate", action="calculate", 
             action_input={"expr": "144 ** 0.5"}, observation="12.0"),
        Step(thought="Now I'll search for something else", 
             action="web_search", 
             action_input={"query": "best pizza recipes"},  # Drift!
             observation="..."),
    ]
    
    risk_score, failures = score_step(goal, steps)
    assert FailureType.GOAL_DRIFT in failures
```

#### 1.2 Intervention Tests
**Goal**: Validate real Gemini-powered interventions

```python
async def test_reprompt_intervention():
    """Test Sentinel generates corrective reprompt via Gemini"""
    state = ExecutionState(goal="Write a hello world program")
    state.steps = [
        Step(action="write_to_file", action_input={"filename": "test.py"}, 
             observation="Error: file not found"),
    ]
    
    model = genai.GenerativeModel("gemini-2.0-flash")
    intervention, new_state, reprompt = intervene(
        state, risk_score=0.6, 
        failure_types=[FailureType.INCOHERENT_TOOL],
        model=model,
        step_number=2
    )
    
    assert intervention.intervention_type == InterventionType.REPROMPT
    assert "goal" in reprompt.lower()
    
async def test_decompose_goal_intervention():
    """Test Gemini decomposes complex goals"""
    goal = "Build a complete web scraper for news articles"
    model = genai.GenerativeModel("gemini-2.0-flash")
    
    sub_goals = decompose_goal(goal, model)
    
    assert len(sub_goals) >= 3
    assert len(sub_goals) <= 5
    assert any("scrape" in sg.lower() for sg in sub_goals)
```

#### 1.3 WebSocket API Tests
**Goal**: Test real WebSocket communication with frontend

```python
async def test_websocket_full_run():
    """Test complete agent run via WebSocket"""
    async with websocket_connect("ws://localhost:8000/ws/run") as ws:
        await ws.send_json({
            "goal": "Search for today's weather in Toronto",
            "api_key": os.getenv("GEMINI_API_KEY"),
            "model": "gemini-2.0-flash",
            "max_steps": 5
        })
        
        # Expect start message
        msg = await ws.receive_json()
        assert msg["type"] == "start"
        
        # Expect step messages
        step_msgs = []
        async for message in ws:
            if message["type"] == "step":
                step_msgs.append(message)
            elif message["type"] in ["complete", "timeout", "error"]:
                break
        
        assert len(step_msgs) >= 1
        assert step_msgs[0]["risk_score"] >= 0
        
async def test_websocket_intervention_broadcast():
    """Test interventions are broadcast via WebSocket"""
    async with websocket_connect("ws://localhost:8000/ws/run") as ws:
        await ws.send_json({
            "goal": "Do the same thing repeatedly",  # Will trigger loop detection
            "api_key": os.getenv("GEMINI_API_KEY"),
            "max_steps": 10
        })
        
        intervention_received = False
        async for message in ws:
            if message["type"] == "intervention":
                intervention_received = True
                assert message["intervention"]["step_number"] > 0
                break
        
        assert intervention_received
```

---

### 2. DarCI + Multi-Agent Tests

#### 2.1 Task Coordination Tests
**Goal**: Test DarCI coordinating multiple agents via Tailscale

```python
# test_darci_coordination.py

async def test_darci_assigns_task_to_agent():
    """Test DarCI assigns task and tracks via Tailscale"""
    darci = DarCI(api_key=os.getenv("GEMINI_API_KEY"))
    
    # Discover agents via Tailscale
    agents = await darci.discover_agents(tailnet="your-tailnet.ts.net")
    assert len(agents) >= 1
    
    # Assign task
    task = Task(
        description="Research quantum computing basics",
        assigned_to=agents[0].id,
        priority=Priority.HIGH
    )
    
    result = await darci.execute_task(task)
    assert result.status == TaskStatus.COMPLETED
    assert len(result.output) > 0
    
async def test_darci_multi_agent_collaboration():
    """Test multiple agents collaborating on complex task"""
    darci = DarCI(api_key=os.getenv("GEMINI_API_KEY"))
    
    goal = "Build a simple Flask API with documentation"
    
    # DarCI decomposes and assigns to multiple agents
    plan = await darci.create_plan(goal)
    
    # Agent 1: Write code
    # Agent 2: Write tests
    # Agent 3: Write docs
    results = await darci.execute_plan(plan)
    
    assert all(r.status == TaskStatus.COMPLETED for r in results)
```

#### 2.2 Intervention Escalation Tests
**Goal**: Test DarCI escalating when Sentinel flags issues

```python
async def test_darci_escalates_sentinel_alert():
    """Test DarCI responds to Sentinel risk alerts"""
    darci = DarCI()
    sentinel_client = SentinelClient("http://localhost:8000")
    
    # Start agent task
    task_id = await darci.start_task("Calculate prime numbers up to 100")
    
    # Monitor risk scores
    async for risk_update in sentinel_client.monitor_task(task_id):
        if risk_update.risk_score > 0.8:
            # DarCI should intervene
            intervention = await darci.intervene(
                task_id, 
                intervention_type=InterventionType.ROLLBACK
            )
            assert intervention.success
            break
```

---

### 3. Tailbridge + Tailscale Tests

#### 3.1 Agent-to-Agent Communication Tests
**Goal**: Test real Tailscale connectivity between agents

```python
# test_tailbridge_tailscale.py
# Requires: TAILSCALE_AUTH_KEY env var, Tailscale installed

@pytest.mark.tailscale
async def test_a2a_message_delivery():
    """Test real A2A message over Tailscale tailnet"""
    # Setup: 2 agents on same tailnet
    agent1 = TailbridgeAgent(
        name="agent1",
        auth_key=os.getenv("TS_AUTH_KEY_1"),
        tailnet="test.ts.net"
    )
    await agent1.start()
    
    agent2 = TailbridgeAgent(
        name="agent2",
        auth_key=os.getenv("TS_AUTH_KEY_2"),
        tailnet="test.ts.net"
    )
    await agent2.start()
    
    # Wait for Tailscale connection
    await asyncio.sleep(5)  # Real connection time
    
    # Send message
    msg = Message(
        from_agent="agent1",
        to_agent="agent2",
        payload={"task": "process_data", "data": [1, 2, 3]}
    )
    
    result = await agent1.send_message(msg)
    assert result.delivered
    assert result.ack_received
    
    # Cleanup
    await agent1.stop()
    await agent2.stop()
    
async def test_a2a_phonebook_discovery():
    """Test agents discover each other via Tailscale"""
    agent = TailbridgeAgent(auth_key=os.getenv("TS_AUTH_KEY_1"))
    await agent.start()
    
    # Discover agents on tailnet
    phonebook = await agent.discover_peers()
    
    assert len(phonebook.agents) >= 1
    assert any(a.capabilities for a in phonebook.agents)
    
    await agent.stop()
```

#### 3.2 File Transfer Tests
**Goal**: Test TailFS file transfers over Tailscale

```python
@pytest.mark.tailscale
async def test_tailfs_file_transfer():
    """Test real file transfer over Tailscale"""
    sender = TailbridgeAgent(auth_key=os.getenv("TS_AUTH_KEY_1"))
    receiver = TailbridgeAgent(auth_key=os.getenv("TS_AUTH_KEY_2"))
    
    await sender.start()
    await receiver.start()
    
    # Create test file
    test_file = Path("/tmp/test_transfer.txt")
    test_file.write_text("Hello over Tailscale!" * 1000)
    
    # Transfer
    transfer = await sender.send_file(
        file_path=test_file,
        to_agent=receiver.id,
        use_compression=True
    )
    
    assert transfer.success
    assert transfer.bytes_transferred == test_file.stat().st_size
    
    # Verify integrity
    received_file = Path("/tmp/test_received.txt")
    assert received_file.read_text() == test_file.read_text()
    
    await sender.stop()
    await receiver.stop()
    
async def test_tailfs_large_file_transfer():
    """Test large file transfer (100MB+) over Tailscale"""
    sender = TailbridgeAgent(auth_key=os.getenv("TS_AUTH_KEY_1"))
    await sender.start()
    
    # Create 100MB test file
    test_file = Path("/tmp/large_test.bin")
    with open(test_file, "wb") as f:
        f.write(os.urandom(100 * 1024 * 1024))
    
    # Track progress
    progress_updates = []
    async def on_progress(bytes_sent, total_bytes):
        progress_updates.append(bytes_sent / total_bytes)
    
    transfer = await sender.send_file(
        file_path=test_file,
        to_agent="agent2",
        on_progress=on_progress
    )
    
    assert transfer.success
    assert len(progress_updates) > 1  # Multiple progress updates
    assert progress_updates[-1] == 1.0
    
    await sender.stop()
```

---

### 4. Full System End-to-End Tests

#### 4.1 Complete Workflow Tests
**Goal**: Test entire SentinelAI pipeline with real APIs

```python
# test_e2e_workflow.py

@pytest.mark.e2e
async def test_full_risk_assessment_workflow():
    """Test complete flow: Task → Agent → Sentinel → Intervention"""
    
    # 1. User submits task via DarCI
    darci = DarCI(api_key=os.getenv("GEMINI_API_KEY"))
    goal = "Research and summarize recent AI breakthroughs"
    
    # 2. DarCI assigns to agent via Tailscale
    task_id = await darci.create_task(goal)
    
    # 3. Agent executes steps (real Gemini calls)
    async for step in darci.monitor_task(task_id):
        # 4. Sentinel scores each step (real-time)
        risk_score, failures = score_step(goal, [step])
        
        # 5. Intervention if needed
        if should_intervene(risk_score, failures):
            intervention = intervene(...)
            await darci.apply_intervention(task_id, intervention)
    
    # 6. Task completes
    result = await darci.get_task_result(task_id)
    assert result.status == "completed"
    assert len(result.output) > 0
    
@pytest.mark.e2e
async def test_multi_agent_collaboration_with_monitoring():
    """Test multiple agents collaborating under Sentinel monitoring"""
    
    goal = "Build a REST API with tests and deployment config"
    
    # DarCI coordinates 3 agents
    agents = [
        await create_agent(role="backend_dev"),
        await create_agent(role="qa_engineer"),
        await create_agent(role="devops")
    ]
    
    # Each agent's steps are monitored
    async with SentinelMonitor() as sentinel:
        tasks = [
            sentinel.monitor_agent(agent, goal)
            for agent in agents
        ]
        
        results = await asyncio.gather(*tasks)
        
        # All agents completed without critical failures
        assert all(r.risk_score < 0.8 for r in results)
```

#### 4.2 Failure Scenario Tests
**Goal**: Test system handles real API failures gracefully

```python
@pytest.mark.e2e
async def test_gemini_rate_limit_handling():
    """Test system handles Gemini rate limits"""
    
    # Exhaust rate limit intentionally
    model = genai.GenerativeModel("gemini-2.0-flash")
    
    for i in range(100):  # Exceed quota
        try:
            model.generate_content(f"Test {i}")
        except ResourceExhausted:
            # System should retry with backoff
            break
    
    # Verify retry logic works
    result = await retry_with_backoff(
        model.generate_content,
        "Final test",
        max_retries=3
    )
    assert result is not None
    
@pytest.mark.e2e
async def test_tailscale_network_partition():
    """Test system handles Tailscale disconnection"""
    
    agent1 = TailbridgeAgent(auth_key=os.getenv("TS_AUTH_KEY_1"))
    agent2 = TailbridgeAgent(auth_key=os.getenv("TS_AUTH_KEY_2"))
    
    await agent1.start()
    await agent2.start()
    
    # Simulate network partition (stop agent2)
    await agent2.stop()
    
    # Agent1 should handle gracefully
    with pytest.raises(AgentUnavailableError):
        await agent1.send_message(to_agent="agent2", msg="Hello")
    
    # DarCI should reassign task
    darci = DarCI()
    reassigned = await darci.reassign_task(
        task_id="task123",
        from_agent="agent2",
        to_agent="agent3"
    )
    assert reassigned.success
    
    await agent1.stop()
```

---

### 5. Load & Performance Tests

#### 5.1 Concurrent Agent Tests
**Goal**: Test system under load with multiple concurrent agents

```python
# test_load_performance.py

@pytest.mark.load
async def test_concurrent_agent_execution():
    """Test 10 agents running simultaneously"""
    
    agents = [
        DarCI(api_key=os.getenv(f"GEMINI_API_KEY_{i}"))
        for i in range(10)
    ]
    
    goals = [f"Research topic {i}" for i in range(10)]
    
    # Run all concurrently
    tasks = [agent.execute_task(goal) for agent, goal in zip(agents, goals)]
    
    start_time = time.time()
    results = await asyncio.gather(*tasks, return_exceptions=True)
    elapsed = time.time() - start_time
    
    # All completed within acceptable time
    assert elapsed < 60  # 1 minute max
    assert all(not isinstance(r, Exception) for r in results)
    
@pytest.mark.load
async def test_sentinel_scoring_throughput():
    """Test Sentinel can score 100 steps/second"""
    
    goal = "Test goal"
    steps = [
        Step(action="web_search", action_input={"query": f"query {i}"})
        for i in range(100)
    ]
    
    start_time = time.time()
    
    for step in steps:
        score_step(goal, [step])
    
    elapsed = time.time() - start_time
    
    # Should score 100 steps in < 1 second (local computation)
    assert elapsed < 1.0
```

---

## 🔧 Test Infrastructure

### Environment Setup

```bash
# .env.test
GEMINI_API_KEY=AIzaSy...  # Real API key
GEMINI_PRO_MODEL=gemini-2.0-pro
GEMINI_FLASH_MODEL=gemini-2.0-flash

TS_AUTH_KEY_1=tskey-auth-xxx1
TS_AUTH_KEY_2=tskey-auth-xxx2
TS_AUTH_KEY_3=tskey-auth-xxx3
TAILNET_NAME=test.ts.net

# Test-specific settings
TEST_TIMEOUT=300  # 5 minutes for real API calls
TEST_RETRY_COUNT=3
```

### Pytest Configuration

```python
# conftest.py

import pytest
import google.generativeai as genai
from tailbridge import TailbridgeAgent

@pytest.fixture(scope="session")
def gemini_api_key():
    """Load real Gemini API key"""
    key = os.getenv("GEMINI_API_KEY")
    if not key:
        pytest.skip("GEMINI_API_KEY not set")
    return key

@pytest.fixture(scope="session")
def gemini_model(gemini_api_key):
    """Create real Gemini model"""
    genai.configure(api_key=gemini_api_key)
    return genai.GenerativeModel("gemini-2.0-flash")

@pytest.fixture(scope="function")
async def tailscale_agents():
    """Start real agents on Tailscale"""
    auth_keys = [
        os.getenv("TS_AUTH_KEY_1"),
        os.getenv("TS_AUTH_KEY_2"),
    ]
    
    agents = []
    for key in auth_keys:
        if not key:
            pytest.skip("Tailscale auth keys not set")
        agent = TailbridgeAgent(auth_key=key)
        await agent.start()
        agents.append(agent)
    
    yield agents
    
    # Cleanup
    for agent in agents:
        await agent.stop()

@pytest.mark.load
def pytest_configure(config):
    """Register custom markers"""
    config.addinivalue_line(
        "markers",
        "load: mark test as load/performance test"
    )
    config.addinivalue_line(
        "markers",
        "e2e: mark test as end-to-end integration test"
    )
    config.addinivalue_line(
        "markers",
        "tailscale: mark test as requiring Tailscale connection"
    )
```

### Running Tests

```bash
# Run all integration tests (requires API keys)
pytest test_integration/ -v

# Run only Gemini tests
pytest test_integration/ -k "gemini" -v

# Run only Tailscale tests
pytest test_integration/ -m tailscale -v

# Run E2E tests
pytest test_integration/ -m e2e -v

# Run load tests
pytest test_integration/ -m load -v

# Run with coverage
pytest test_integration/ --cov=backend --cov=darci --cov-report=html

# Run specific test file
pytest test_integration/test_backend_gemini.py::test_agent_web_search_step -v

# Run with retries (for flaky network tests)
pytest test_integration/ --reruns 3 --reruns-delay 5
```

---

## 📊 Test Coverage Matrix

| Component | Gemini API | Tailscale API | WebSocket | Risk Scoring | Interventions |
|-----------|-----------|---------------|-----------|--------------|---------------|
| **Backend** | ✅ | ❌ | ✅ | ✅ | ✅ |
| **DarCI** | ✅ | ✅ | ❌ | ✅ | ✅ |
| **Tailbridge** | ❌ | ✅ | ❌ | ❌ | ❌ |
| **Full E2E** | ✅ | ✅ | ✅ | ✅ | ✅ |

---

## ⚠️ Risk Mitigation

### API Cost Management
```python
# conftest.py

@pytest.fixture(scope="function")
def rate_limited_gemini(gemini_api_key):
    """Rate-limited Gemini client to control costs"""
    genai.configure(api_key=gemini_api_key)
    
    # Add delay between calls
    import time
    original_generate = genai.GenerativeModel.generate_content
    
    def wrapped_generate(self, *args, **kwargs):
        time.sleep(2)  # 2 second delay
        return original_generate(self, *args, **kwargs)
    
    genai.GenerativeModel.generate_content = wrapped_generate
    return genai.GenerativeModel("gemini-2.0-flash")
```

### Test Timeout Protection
```python
# Prevent tests hanging on network issues
@pytest.mark.tailscale
@pytest.mark.timeout(300)  # 5 minute max
async def test_a2a_message_delivery():
    ...
```

### Cleanup on Failure
```python
@pytest.fixture(scope="function")
async def cleanup_agents():
    """Ensure agents are stopped even if test fails"""
    agents = []
    
    yield agents
    
    # Always cleanup
    for agent in agents:
        try:
            await agent.stop()
        except Exception:
            pass  # Ignore cleanup errors
```

---

## 📈 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| **Test Pass Rate** | > 90% | CI/CD pipeline |
| **E2E Test Duration** | < 10 min | pytest timing |
| **API Cost per Run** | < $5/month | Google Cloud billing |
| **Flaky Test Rate** | < 5% | Retry analysis |
| **Coverage** | > 80% | pytest-cov |

---

## 🚀 Next Steps

1. **Setup Test Environment**
   - [ ] Create dedicated test API keys
   - [ ] Configure test tailnet
   - [ ] Setup pytest fixtures

2. **Implement Core Tests**
   - [ ] Backend + Gemini tests (priority: HIGH)
   - [ ] Tailbridge + Tailscale tests (priority: HIGH)
   - [ ] DarCI coordination tests (priority: MEDIUM)
   - [ ] E2E workflow tests (priority: MEDIUM)
   - [ ] Load tests (priority: LOW)

3. **CI/CD Integration**
   - [ ] GitHub Actions workflow
   - [ ] API key secrets management
   - [ ] Test result reporting

4. **Documentation**
   - [ ] Test setup guide
   - [ ] Troubleshooting guide
   - [ ] Cost management tips

---

*Last updated: March 8, 2026*
*Version: 1.0.0 - Brainstorming Draft*
