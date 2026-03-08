"""
Pytest configuration and fixtures for SentinelAI integration tests.

ALL TESTS USE REAL APIs - NO MOCKS
"""

import asyncio
import os
import time
from pathlib import Path
from typing import Any, AsyncGenerator, Generator, List

import google.generativeai as genai
import pytest
import pytest_asyncio
from dotenv import load_dotenv

# Load test environment from project root
ROOT_DIR = Path(__file__).parent.parent.parent
load_dotenv(ROOT_DIR / ".env.test")

# Test configuration
TEST_TIMEOUT = int(os.getenv("TEST_TIMEOUT", "300"))
TEST_RETRY_COUNT = int(os.getenv("TEST_RETRY_COUNT", "3"))
TEST_RETRY_DELAY = 2  # seconds between retries


def pytest_configure(config):
    """Register custom markers."""
    config.addinivalue_line(
        "markers",
        "gemini: mark test as requiring Gemini API (real API call)"
    )
    config.addinivalue_line(
        "markers",
        "tailscale: mark test as requiring Tailscale connection (real network)"
    )
    config.addinivalue_line(
        "markers",
        "e2e: mark test as end-to-end integration test"
    )
    config.addinivalue_line(
        "markers",
        "slow: mark test as slow-running (>30 seconds)"
    )
    config.addinivalue_line(
        "markers",
        "api_cost: mark test as consuming API quota"
    )


# ============================================================================
# Gemini API Fixtures
# ============================================================================

@pytest.fixture(scope="session")
def gemini_api_key() -> str:
    """
    Load real Gemini API key from environment.
    
    Skips test if key not available.
    """
    key = os.getenv("GEMINI_API_KEY")
    if not key:
        pytest.skip("GEMINI_API_KEY not set in environment")
    return key


@pytest.fixture(scope="session")
def gemini_flash_model(gemini_api_key: str) -> Any:
    """
    Create real Gemini Flash model instance.
    
    Faster and cheaper than Pro, good for most tests.
    """
    genai.configure(api_key=gemini_api_key)
    model_name = os.getenv("GEMINI_FLASH_MODEL", "gemini-2.0-flash")
    return genai.GenerativeModel(model_name)


@pytest.fixture(scope="session")
def gemini_pro_model(gemini_api_key: str) -> Any:
    """
    Create real Gemini Pro model instance.
    
    More capable but slower/expensive, use for complex reasoning tests.
    """
    genai.configure(api_key=gemini_api_key)
    model_name = os.getenv("GEMINI_PRO_MODEL", "gemini-2.0-pro")
    return genai.GenerativeModel(model_name)


@pytest.fixture(scope="function")
def rate_limited_gemini(gemini_api_key: str) -> Any:
    """
    Create Gemini model with rate limiting to avoid quota errors.
    
    Adds delay between calls to respect rate limits.
    """
    genai.configure(api_key=gemini_api_key)
    model = genai.GenerativeModel("gemini-2.0-flash")
    
    original_generate = model.generate_content
    
    def wrapped_generate(*args, **kwargs):
        time.sleep(TEST_RETRY_DELAY)  # Rate limiting delay
        return original_generate(*args, **kwargs)
    
    model.generate_content = wrapped_generate
    return model


# ============================================================================
# Tailscale Fixtures
# ============================================================================

@pytest.fixture(scope="session")
def tailscale_auth() -> dict:
    """
    Load Tailscale authentication configuration.
    
    Returns dict with auth keys for multiple agents.
    """
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
    }


@pytest_asyncio.fixture(scope="function")
async def tailscale_agents(tailscale_auth: dict) -> AsyncGenerator[List[Any], None]:
    """
    Start real agents on Tailscale network.
    
    Yields list of running agent instances.
    Cleanup automatically stops all agents.
    """
    from tailbridge.taila2a import Agent  # Real agent import
    
    agents = []
    auth_keys = tailscale_auth["auth_keys"]
    
    for i, key in enumerate(auth_keys[:2]):  # Start 2 agents by default
        try:
            agent = Agent(
                name=f"test-agent-{i+1}",
                auth_key=key,
                tailnet=tailscale_auth["tailnet"],
            )
            await agent.start()
            agents.append(agent)
            print(f"✓ Started agent {i+1}: {agent.id}")
        except Exception as e:
            print(f"⚠ Failed to start agent {i+1}: {e}")
            pytest.skip(f"Tailscale agent startup failed: {e}")
    
    # Wait for agents to connect
    print("⏳ Waiting for Tailscale connection...")
    await asyncio.sleep(5)
    
    yield agents
    
    # Cleanup: stop all agents
    print("\n🧹 Cleaning up agents...")
    for agent in agents:
        try:
            await agent.stop()
            print(f"✓ Stopped agent: {agent.id}")
        except Exception as e:
            print(f"⚠ Error stopping agent: {e}")


# ============================================================================
# Backend Fixtures
# ============================================================================

@pytest_asyncio.fixture(scope="function")
async def backend_server(gemini_api_key: str) -> AsyncGenerator[str, None]:
    """
    Start real SentinelAI backend server.
    
    Yields server URL.
    """
    import subprocess
    import socket
    
    # Find available port
    sock = socket.socket()
    sock.bind(("", 0))
    port = sock.getsockname()[1]
    sock.close()
    
    server_url = f"http://localhost:{port}"
    
    # Set environment for server
    env = os.environ.copy()
    env["GEMINI_API_KEY"] = gemini_api_key
    env["PORT"] = str(port)
    
    # Start server
    print(f"🚀 Starting backend server on {server_url}")
    process = subprocess.Popen(
        ["python", "-m", "uvicorn", "backend.main:app", "--port", str(port)],
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    
    # Wait for server to start
    await asyncio.sleep(3)
    
    # Check if server is running
    if process.poll() is not None:
        stdout, stderr = process.communicate()
        pytest.fail(f"Backend server failed to start:\n{stderr.decode()}")
    
    yield server_url
    
    # Cleanup: stop server
    print("\n🧹 Stopping backend server...")
    process.terminate()
    try:
        process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        process.kill()
    print("✓ Backend server stopped")


@pytest.fixture(scope="function")
def execution_state() -> Any:
    """Create fresh ExecutionState for testing."""
    from backend.state import ExecutionState
    return ExecutionState(goal="Test goal")


@pytest.fixture(scope="function")
def sample_steps() -> List[Any]:
    """Create sample steps for testing."""
    from backend.state import Step
    return [
        Step(
            step_number=1,
            thought="I need to search for information",
            action="web_search",
            action_input={"query": "test query"},
            observation="Search results",
        ),
        Step(
            step_number=2,
            thought="Now I'll calculate something",
            action="calculate",
            action_input={"expression": "2 + 2"},
            observation="4",
        ),
    ]


# ============================================================================
# Utility Fixtures
# ============================================================================

@pytest.fixture(scope="function")
def temp_output_dir() -> Generator[Path, None, None]:
    """
    Create temporary output directory for file operations.
    
    Automatically cleaned up after test.
    """
    import tempfile
    import shutil
    
    temp_dir = tempfile.mkdtemp(prefix="sentinel_test_")
    output_dir = Path(temp_dir) / "output"
    output_dir.mkdir(parents=True, exist_ok=True)
    
    # Temporarily override output directory
    original_cwd = os.getcwd()
    os.chdir(temp_dir)
    
    yield output_dir
    
    # Cleanup
    os.chdir(original_cwd)
    shutil.rmtree(temp_dir, ignore_errors=True)


@pytest.fixture(scope="function")
def test_file(temp_output_dir: Path) -> Path:
    """Create a test file with known content."""
    test_file = temp_output_dir / "test.txt"
    test_file.write_text("Test content for integration testing")
    return test_file


# ============================================================================
# Retry Logic
# ============================================================================

def retry_on_failure(max_retries: int = TEST_RETRY_COUNT, delay: float = TEST_RETRY_DELAY):
    """
    Decorator for retrying flaky network tests.
    
    Usage:
        @retry_on_failure(max_retries=3, delay=2)
        async def test_flaky_network_call():
            ...
    """
    def decorator(func):
        async def wrapper(*args, **kwargs):
            last_exception = None
            for attempt in range(max_retries):
                try:
                    return await func(*args, **kwargs)
                except Exception as e:
                    last_exception = e
                    if attempt < max_retries - 1:
                        print(f"⚠ Attempt {attempt + 1} failed: {e}")
                        print(f"⏳ Retrying in {delay}s...")
                        await asyncio.sleep(delay)
            raise last_exception
        return wrapper
    return decorator


# ============================================================================
# Assertion Helpers
# ============================================================================

def assert_valid_risk_score(score: float):
    """Assert risk score is valid (0.0 - 1.0)."""
    assert 0.0 <= score <= 1.0, f"Risk score {score} out of range [0.0, 1.0]"


def assert_step_structure(step: Any):
    """Assert step has required fields."""
    assert hasattr(step, "step_number")
    assert hasattr(step, "thought")
    assert hasattr(step, "action")
    assert hasattr(step, "action_input")
    assert hasattr(step, "observation")
    assert isinstance(step.step_number, int)
    assert step.step_number > 0


def assert_gemini_response_valid(response: Any):
    """Assert Gemini API response is valid."""
    assert response is not None
    assert hasattr(response, "text")
    assert len(response.text) > 0
