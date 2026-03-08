"""
Pytest configuration for demo testing suite.

Demo tests are pre-show demonstrations that showcase the full SentinelAI system.
"""

import os
import pytest


def pytest_configure(config):
    """Register custom markers for demo tests."""
    config.addinivalue_line(
        "markers",
        "demo: mark test as a demo/pre-show test"
    )
    config.addinivalue_line(
        "markers",
        "scenario: mark test as a scenario-based demonstration"
    )
    config.addinivalue_line(
        "markers",
        "performance: mark test as a performance benchmark"
    )


@pytest.fixture(scope="session")
def test_timeout() -> int:
    """Default timeout for demo tests."""
    return int(os.getenv("DEMO_TEST_TIMEOUT", "300"))


@pytest.fixture(scope="function")
def demo_goal() -> str:
    """Sample goal for demo tasks."""
    return "Demonstrate DarCI coordinating multiple nanobot agents"


@pytest.fixture(scope="function")
def nanobot_count() -> int:
    """Number of nanobot agents for demo."""
    return int(os.getenv("NANOBOT_COUNT", "2"))
