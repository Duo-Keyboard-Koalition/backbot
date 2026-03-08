"""
AIP (Agent Identification Protocol) Client for DarCI Python

This module handles agent registration and heartbeat management
with the local taila2a bridge.

Usage:
    from darci.channels.aip import AIPClient

    client = AIPClient(
        bridge_url="http://127.0.0.1:8080",
        agent_id="darci-python-001",
        agent_type="darci-python",
        endpoint_url="http://127.0.0.1:9090/api"
    )

    # Register agent
    await client.register()

    # Start heartbeat (runs in background)
    await client.start_heartbeat(interval=30)

    # Stop heartbeat
    await client.stop()
"""

import asyncio
import logging
import socket
import time
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional

import aiohttp

logger = logging.getLogger(__name__)


class AgentStatus(str, Enum):
    """Agent registration status."""
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    OFFLINE = "offline"


@dataclass
class AgentConfig:
    """Agent configuration for AIP registration."""
    agent_id: str
    agent_type: str
    agent_version: str = "1.0.0"
    capabilities: List[str] = field(default_factory=lambda: [
        "task-execution",
        "notebook",
        "file-ops"
    ])
    endpoint_url: str = ""
    health_url: str = ""
    hostname: str = ""
    tags: List[str] = field(default_factory=lambda: ["development"])

    def __post_init__(self):
        if not self.hostname:
            self.hostname = socket.gethostname()
        if not self.endpoint_url:
            self.endpoint_url = "http://127.0.0.1:9090/api"
        if not self.health_url:
            self.health_url = "http://127.0.0.1:9090/health"


@dataclass
class RegistrationResponse:
    """Response from agent registration."""
    status: AgentStatus
    message: str
    registration_id: Optional[str] = None
    next_steps: Optional[str] = None


@dataclass
class HeartbeatResponse:
    """Response from heartbeat submission."""
    status: str
    agent_id: str
    approved: bool


class AIPClient:
    """
    AIP (Agent Identification Protocol) client for DarCI agents.

    Handles:
    - Agent registration with local bridge
    - Periodic heartbeat submission
    - Status monitoring
    """

    def __init__(
        self,
        bridge_url: str = "http://127.0.0.1:8080",
        config: Optional[AgentConfig] = None,
        **kwargs
    ):
        """
        Initialize AIP client.

        Args:
            bridge_url: URL of the local taila2a bridge
            config: Agent configuration (optional, will use defaults)
            **kwargs: Additional configuration parameters
        """
        self.bridge_url = bridge_url.rstrip('/')
        self.config = config or AgentConfig(**kwargs)
        self._session: Optional[aiohttp.ClientSession] = None
        self._heartbeat_task: Optional[asyncio.Task] = None
        self._heartbeat_interval: int = 30
        self._last_heartbeat: Optional[datetime] = None
        self._status: AgentStatus = AgentStatus.PENDING
        self._approved: bool = False

    async def _get_session(self) -> aiohttp.ClientSession:
        """Get or create HTTP session."""
        if self._session is None or self._session.closed:
            self._session = aiohttp.ClientSession(
                timeout=aiohttp.ClientTimeout(total=10)
            )
        return self._session

    async def register(self) -> RegistrationResponse:
        """
        Register agent with the bridge.

        Returns:
            RegistrationResponse with status and next steps

        Raises:
            aiohttp.ClientError: If registration request fails
        """
        session = await self._get_session()

        payload = {
            "agent_id": self.config.agent_id,
            "agent_type": self.config.agent_type,
            "agent_version": self.config.agent_version,
            "capabilities": self.config.capabilities,
            "endpoints": {
                "primary": self.config.endpoint_url,
                "health": self.config.health_url
            },
            "metadata": {
                "hostname": self.config.hostname,
                "os": "linux",  # Could be dynamic
                "tags": self.config.tags
            }
        }

        logger.info(f"Registering agent {self.config.agent_id} with bridge {self.bridge_url}")

        try:
            async with session.post(
                f"{self.bridge_url}/aip/register",
                json=payload,
                headers={"Content-Type": "application/json"}
            ) as resp:
                data = await resp.json()

                status_str = data.get("status", "unknown")
                try:
                    status = AgentStatus(status_str)
                except ValueError:
                    status = AgentStatus.PENDING

                self._status = status

                response = RegistrationResponse(
                    status=status,
                    message=data.get("message", ""),
                    registration_id=data.get("registration_id"),
                    next_steps=data.get("next_steps")
                )

                logger.info(f"Registration status: {status.value}")
                if status == AgentStatus.PENDING:
                    logger.info("Waiting for admin approval")
                elif status == AgentStatus.APPROVED:
                    self._approved = True
                    logger.info("Agent approved!")

                return response

        except aiohttp.ClientError as e:
            logger.error(f"Registration failed: {e}")
            raise

    async def send_heartbeat(self) -> Optional[HeartbeatResponse]:
        """
        Send heartbeat to bridge.

        Returns:
            HeartbeatResponse if successful, None if not approved yet

        Raises:
            aiohttp.ClientError: If heartbeat request fails
        """
        session = await self._get_session()

        payload = {
            "agent_id": self.config.agent_id,
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "status": "healthy",
            "metrics": {
                "cpu_usage": 0.0,  # Could add real metrics
                "memory_mb": 0,
                "active_tasks": 0
            }
        }

        try:
            async with session.post(
                f"{self.bridge_url}/aip/heartbeat",
                json=payload,
                headers={"Content-Type": "application/json"}
            ) as resp:
                data = await resp.json()

                self._approved = data.get("approved", False)
                self._last_heartbeat = datetime.utcnow()

                response = HeartbeatResponse(
                    status=data.get("status", "ok"),
                    agent_id=data.get("agent_id", self.config.agent_id),
                    approved=self._approved
                )

                if self._approved:
                    logger.debug(f"Heartbeat sent - agent approved")
                else:
                    logger.debug(f"Heartbeat sent - awaiting approval")

                return response

        except aiohttp.ClientError as e:
            logger.error(f"Heartbeat failed: {e}")
            raise

    async def _heartbeat_loop(self):
        """Background heartbeat loop."""
        logger.info(f"Starting heartbeat loop (interval: {self._heartbeat_interval}s)")

        while True:
            try:
                await asyncio.sleep(self._heartbeat_interval)
                await self.send_heartbeat()
            except asyncio.CancelledError:
                logger.info("Heartbeat loop cancelled")
                break
            except Exception as e:
                logger.error(f"Heartbeat error: {e}")
                # Continue loop, retry on next interval

    async def start_heartbeat(self, interval: int = 30):
        """
        Start background heartbeat task.

        Args:
            interval: Heartbeat interval in seconds (default: 30)
        """
        self._heartbeat_interval = interval

        if self._heartbeat_task is None or self._heartbeat_task.done():
            self._heartbeat_task = asyncio.create_task(self._heartbeat_loop())
            logger.info("Heartbeat task started")

    async def stop(self):
        """Stop heartbeat and close session."""
        if self._heartbeat_task:
            self._heartbeat_task.cancel()
            try:
                await self._heartbeat_task
            except asyncio.CancelledError:
                pass
            self._heartbeat_task = None

        if self._session and not self._session.closed:
            await self._session.close()

        logger.info("AIP client stopped")

    async def check_status(self) -> AgentStatus:
        """
        Check current registration status.

        Returns:
            Current agent status
        """
        session = await self._get_session()

        try:
            async with session.get(f"{self.bridge_url}/aip/agents") as resp:
                if resp.status == 200:
                    agents = await resp.json()
                    for agent in agents:
                        if agent.get("agent_id") == self.config.agent_id:
                            status_str = agent.get("status", "unknown")
                            try:
                                self._status = AgentStatus(status_str)
                            except ValueError:
                                pass
                            break
        except Exception as e:
            logger.error(f"Status check failed: {e}")

        return self._status

    @property
    def is_approved(self) -> bool:
        """Check if agent is approved."""
        return self._approved

    @property
    def last_heartbeat(self) -> Optional[datetime]:
        """Get last heartbeat timestamp."""
        return self._last_heartbeat


# Convenience function for quick registration
async def register_agent(
    bridge_url: str = "http://127.0.0.1:8080",
    agent_id: Optional[str] = None,
    agent_type: str = "darci-python",
    **kwargs
) -> AIPClient:
    """
    Quick registration helper.

    Usage:
        client = await register_agent(
            agent_id="my-agent",
            endpoint_url="http://127.0.0.1:9090/api"
        )
        await client.start_heartbeat()
    """
    import socket

    if not agent_id:
        agent_id = f"darci-python-{socket.gethostname()}"

    client = AIPClient(
        bridge_url=bridge_url,
        agent_id=agent_id,
        agent_type=agent_type,
        **kwargs
    )

    await client.register()
    await client.start_heartbeat()

    return client
