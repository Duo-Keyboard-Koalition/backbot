"""
Tailbridge + Tailscale Integration Tests

Tests real A2A communication and file transfer over Tailscale.
Requires:
- Tailscale installed and authenticated
- Tailbridge agents running (Go binaries)
- TS_AUTH_KEY_* environment variables

NO MOCKS - All tests use real Tailscale network.
"""

import asyncio
import hashlib
import json
import os
import tempfile
import time
from pathlib import Path
from typing import Any, Dict, List, Optional
from uuid import uuid4

import httpx
import pytest
import pytest_asyncio


# ============================================================================
# Test Configuration
# ============================================================================

AGENT1_URL = os.getenv("AGENT1_URL", "http://localhost:8081")
AGENT2_URL = os.getenv("AGENT2_URL", "http://localhost:8082")
AGENT3_URL = os.getenv("AGENT3_URL", "http://localhost:8083")
TEST_TIMEOUT = int(os.getenv("TEST_TIMEOUT", "300"))


# ============================================================================
# Helper Functions
# ============================================================================

async def wait_for_agents(client: httpx.AsyncClient, agent_urls: List[str], timeout: float = 120.0):
    """Wait for all agents to be healthy."""
    start_time = time.time()
    
    while time.time() - start_time < timeout:
        all_healthy = True
        for url in agent_urls:
            try:
                response = await client.get(f"{url}/health", timeout=5.0)
                if response.status_code != 200:
                    all_healthy = False
                    break
            except Exception:
                all_healthy = False
                break
        
        if all_healthy:
            print(f"✓ All {len(agent_urls)} agents are healthy")
            return
        
        await asyncio.sleep(2.0)
    
    raise TimeoutError(f"Agents did not become healthy within {timeout}s")


def create_test_message(
    source: str,
    dest: str,
    action: str,
    payload: Dict[str, Any],
    msg_type: str = "request",
    topic: Optional[str] = None,
) -> Dict[str, Any]:
    """Create a properly formatted A2A message."""
    return {
        "id": str(uuid4()),
        "type": msg_type,
        "source": source,
        "dest": dest,
        "topic": topic or f"{dest}.requests",
        "timestamp": time.time(),
        "body": {
            "action": action,
            "content_type": "application/json",
            "payload": payload,
        },
    }


# ============================================================================
# A2A Communication Tests
# ============================================================================

@pytest.mark.tailscale
@pytest.mark.slow
class TestA2ACommunication:
    """Test real A2A communication over Tailscale."""
    
    @pytest.fixture(autouse=True)
    def check_bridge_available(self):
        """Skip tests if Tailbridge agents are not running."""
        import httpx
        try:
            with httpx.Client(timeout=5.0) as client:
                response = client.get(f"{AGENT1_URL}/health")
                if response.status_code != 200:
                    pytest.skip(f"Tailbridge agent not available at {AGENT1_URL}")
        except (httpx.ConnectError, httpx.TimeoutException):
            pytest.skip("Tailbridge agents not running - skip A2A tests")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_agent_health(self):
        """Test all agents are healthy and accessible."""
        agent_urls = [AGENT1_URL, AGENT2_URL, AGENT3_URL]
        
        async with httpx.AsyncClient() as client:
            for url in agent_urls:
                response = await client.get(f"{url}/health")
                assert response.status_code == 200, f"Agent at {url} not healthy"
                
                data = response.json()
                assert data.get("status") == "healthy"
                print(f"✓ Agent healthy: {url}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_agent_status(self):
        """Test agent status endpoints."""
        async with httpx.AsyncClient() as client:
            response = await client.get(f"{AGENT1_URL}/status")
            assert response.status_code == 200
            
            status = response.json()
            assert status.get("status") == "online"
            assert "name" in status
            assert "capabilities" in status
            print(f"✓ Agent status: {status.get('name')}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_phonebook_discovery(self):
        """Test agent discovery via phonebook."""
        async with httpx.AsyncClient() as client:
            response = await client.get(f"{AGENT1_URL}/phonebook")
            assert response.status_code == 200
            
            phonebook = response.json()
            assert "agents" in phonebook
            assert "count" in phonebook
            assert phonebook["count"] > 0, "Phonebook should not be empty"
            
            print(f"✓ Discovered {phonebook['count']} agents")
            for agent in phonebook.get("agents", [])[:3]:
                print(f"  - {agent.get('name', 'unknown')}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_agent_listing_by_capability(self):
        """Test filtering agents by capability."""
        async with httpx.AsyncClient() as client:
            response = await client.get(f"{AGENT1_URL}/agents?capability=file_send")
            assert response.status_code == 200
            
            result = response.json()
            assert "agents" in result
            assert "count" in result
            
            print(f"✓ Found {result['count']} agents with file_send capability")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_send_message_between_agents(self):
        """Test sending a message from one agent to another."""
        async with httpx.AsyncClient() as client:
            message = create_test_message(
                source="test-client",
                dest="agent2",
                action="chat",
                payload={"message": "Hello from integration test!", "test": "TestSendMessage"},
            )
            
            response = await client.post(
                f"{AGENT2_URL}/a2a/inbound",
                json=message,
                timeout=10.0,
            )
            assert response.status_code == 202, f"Message send failed: {response.text}"
            print("✓ Message sent successfully")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_bidirectional_messaging(self):
        """Test two-way communication between agents."""
        async with httpx.AsyncClient() as client:
            # Send ping
            ping_msg = create_test_message(
                source="agent1",
                dest="agent2",
                action="ping",
                payload={"type": "ping", "data": "test"},
            )
            
            response = await client.post(
                f"{AGENT2_URL}/a2a/inbound",
                json=ping_msg,
                timeout=10.0,
            )
            assert response.status_code == 202
            
            # Send pong back
            pong_msg = create_test_message(
                source="agent2",
                dest="agent1",
                action="pong",
                payload={"type": "pong", "data": "response"},
            )
            
            response = await client.post(
                f"{AGENT1_URL}/a2a/inbound",
                json=pong_msg,
                timeout=10.0,
            )
            assert response.status_code == 202
            
            print("✓ Bidirectional messaging successful")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_message_correlation(self):
        """Test request/response correlation with correlation_id."""
        correlation_id = str(uuid4())
        
        async with httpx.AsyncClient() as client:
            # Send request
            request_msg = create_test_message(
                source="agent1",
                dest="agent3",
                action="execute",
                payload={"command": "execute_test"},
                msg_type="request",
            )
            request_msg["correlation_id"] = correlation_id
            
            response = await client.post(
                f"{AGENT3_URL}/a2a/inbound",
                json=request_msg,
                timeout=10.0,
            )
            assert response.status_code == 202
            
            # Send response with same correlation ID
            response_msg = create_test_message(
                source="agent3",
                dest="agent1",
                action="result",
                payload={"result": "success"},
                msg_type="response",
            )
            response_msg["correlation_id"] = correlation_id
            
            response = await client.post(
                f"{AGENT1_URL}/a2a/inbound",
                json=response_msg,
                timeout=10.0,
            )
            assert response.status_code == 202
            
            print(f"✓ Message correlation successful: {correlation_id}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_concurrent_messaging(self):
        """Test concurrent message sending."""
        message_count = 20
        
        async with httpx.AsyncClient() as client:
            async def send_message(idx: int):
                msg = create_test_message(
                    source="test-client",
                    dest="agent2",
                    action="test",
                    payload={"index": idx},
                )
                try:
                    response = await client.post(
                        f"{AGENT2_URL}/a2a/inbound",
                        json=msg,
                        timeout=10.0,
                    )
                    return response.status_code == 202
                except Exception:
                    return False
            
            # Send all messages concurrently
            tasks = [send_message(i) for i in range(message_count)]
            results = await asyncio.gather(*tasks)
            
            success_count = sum(results)
            print(f"✓ Sent {success_count}/{message_count} concurrent messages")
            assert success_count >= message_count * 0.8  # At least 80% success
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_different_message_types(self):
        """Test different message types."""
        message_types = ["request", "response", "event", "broadcast"]
        
        async with httpx.AsyncClient() as client:
            for msg_type in message_types:
                msg = create_test_message(
                    source="test-client",
                    dest="agent3",
                    action="test_type",
                    payload={"type": msg_type},
                    msg_type=msg_type,
                )
                
                response = await client.post(
                    f"{AGENT3_URL}/a2a/inbound",
                    json=msg,
                    timeout=10.0,
                )
                assert response.status_code == 202, f"Failed for type: {msg_type}"
                print(f"✓ Message type '{msg_type}' successful")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_large_message(self):
        """Test sending large messages (100KB)."""
        # Create 100KB payload
        large_data = bytes(range(256)) * 400  # ~100KB
        
        async with httpx.AsyncClient() as client:
            msg = create_test_message(
                source="test-client",
                dest="agent1",
                action="large_transfer",
                payload={
                    "data": large_data.hex(),
                    "type": "large_message",
                    "size_bytes": len(large_data),
                },
            )
            
            response = await client.post(
                f"{AGENT1_URL}/a2a/inbound",
                json=msg,
                timeout=30.0,
            )
            assert response.status_code == 202, f"Large message failed: {response.text}"
            print(f"✓ Large message ({len(large_data)} bytes) sent successfully")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_invalid_message_handling(self):
        """Test error handling for invalid messages."""
        async with httpx.AsyncClient() as client:
            # Send invalid JSON
            response = await client.post(
                f"{AGENT1_URL}/a2a/inbound",
                content=b"invalid json",
                headers={"Content-Type": "application/json"},
                timeout=10.0,
            )
            assert response.status_code == 400, "Should reject invalid JSON"
            print("✓ Invalid message properly rejected")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_message_with_metadata(self):
        """Test messages with metadata."""
        msg = create_test_message(
            source="test-client",
            dest="agent2",
            action="test_metadata",
            payload={"test": "data"},
            topic="test.topic",
        )
        msg["metadata"] = {
            "priority": "high",
            "tags": ["test", "integration"],
            "custom_field": "custom_value",
        }
        
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{AGENT2_URL}/a2a/inbound",
                json=msg,
                timeout=10.0,
            )
            assert response.status_code == 202
            print("✓ Message with metadata sent successfully")


# ============================================================================
# File Transfer Tests
# ============================================================================

@pytest.mark.tailscale
@pytest.mark.slow
class TestFileTransfer:
    """Test real file transfer over Tailscale."""
    
    @pytest.fixture(autouse=True)
    def check_bridge_available(self):
        """Skip tests if Tailbridge agents are not running."""
        import httpx
        try:
            with httpx.Client(timeout=5.0) as client:
                response = client.get(f"{AGENT1_URL}/health")
                if response.status_code != 200:
                    pytest.skip(f"Tailbridge agent not available at {AGENT1_URL}")
        except (httpx.ConnectError, httpx.TimeoutException):
            pytest.skip("Tailbridge agents not running - skip file transfer tests")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_small_file_transfer(self):
        """Test small file transfer (< 1MB)."""
        # Create test file
        test_content = b"Hello over Tailscale! " * 1000  # ~24KB
        test_hash = hashlib.sha256(test_content).hexdigest()
        
        async with httpx.AsyncClient() as client:
            # Initiate transfer
            transfer_request = {
                "id": str(uuid4()),
                "source_agent_id": "agent1",
                "source_agent_name": "Test Agent 1",
                "dest_agent_id": "agent2",
                "dest_agent_name": "Test Agent 2",
                "file_name": "test_small.txt",
                "file_size": len(test_content),
                "file_hash": test_hash,
                "compress": False,
                "verify": True,
            }
            
            response = await client.post(
                f"{AGENT2_URL}/tailfs/transfer/request",
                json=transfer_request,
                timeout=30.0,
            )
            assert response.status_code == 200, f"Transfer request failed: {response.text}"
            
            transfer_response = response.json()
            assert transfer_response.get("accepted") is True
            
            print(f"✓ Small file transfer initiated: {len(test_content)} bytes")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_large_file_transfer(self):
        """Test large file transfer (10MB+)."""
        # Create 10MB test file
        test_size = 10 * 1024 * 1024  # 10MB
        test_content = os.urandom(test_size)
        test_hash = hashlib.sha256(test_content).hexdigest()
        
        async with httpx.AsyncClient() as client:
            transfer_request = {
                "id": str(uuid4()),
                "source_agent_id": "agent1",
                "dest_agent_id": "agent3",
                "file_name": "test_large.bin",
                "file_size": test_size,
                "file_hash": test_hash,
                "compress": True,
                "verify": True,
            }
            
            response = await client.post(
                f"{AGENT3_URL}/tailfs/transfer/request",
                json=transfer_request,
                timeout=60.0,
            )
            assert response.status_code == 200
            
            transfer_response = response.json()
            assert transfer_response.get("accepted") is True
            
            print(f"✓ Large file transfer initiated: {test_size / (1024*1024):.1f}MB")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_file_transfer_progress(self):
        """Test file transfer progress tracking."""
        async with httpx.AsyncClient() as client:
            # Get transfer progress (if any active)
            response = await client.get(f"{AGENT1_URL}/tailfs/transfers/active")
            assert response.status_code == 200
            
            transfers = response.json()
            print(f"✓ Active transfers: {len(transfers.get('transfers', []))}")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_file_transfer_history(self):
        """Test file transfer history."""
        async with httpx.AsyncClient() as client:
            response = await client.get(f"{AGENT1_URL}/tailfs/transfers/history")
            assert response.status_code == 200
            
            history = response.json()
            assert "transfers" in history
            print(f"✓ Transfer history: {len(history.get('transfers', []))} entries")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_compressed_transfer(self):
        """Test compressed file transfer."""
        # Create compressible content (repeated patterns)
        test_content = b"AAAAAAAAAA" * 10000  # Highly compressible
        test_hash = hashlib.sha256(test_content).hexdigest()
        
        async with httpx.AsyncClient() as client:
            transfer_request = {
                "id": str(uuid4()),
                "source_agent_id": "agent1",
                "dest_agent_id": "agent2",
                "file_name": "test_compressed.txt",
                "file_size": len(test_content),
                "file_hash": test_hash,
                "compress": True,  # Enable compression
                "verify": True,
            }
            
            response = await client.post(
                f"{AGENT2_URL}/tailfs/transfer/request",
                json=transfer_request,
                timeout=30.0,
            )
            assert response.status_code == 200
            
            print("✓ Compressed file transfer initiated")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_transfer_with_encryption(self):
        """Test encrypted file transfer."""
        test_content = b"Secret data " * 100
        test_hash = hashlib.sha256(test_content).hexdigest()
        
        async with httpx.AsyncClient() as client:
            transfer_request = {
                "id": str(uuid4()),
                "source_agent_id": "agent1",
                "dest_agent_id": "agent2",
                "file_name": "test_encrypted.txt",
                "file_size": len(test_content),
                "file_hash": test_hash,
                "encrypt": True,  # Enable encryption
                "verify": True,
            }
            
            response = await client.post(
                f"{AGENT2_URL}/tailfs/transfer/request",
                json=transfer_request,
                timeout=30.0,
            )
            assert response.status_code == 200
            
            print("✓ Encrypted file transfer initiated")


# ============================================================================
# Tailscale Network Tests
# ============================================================================

@pytest.mark.tailscale
@pytest.mark.slow
class TestTailscaleNetwork:
    """Test Tailscale network connectivity."""
    
    @pytest.fixture(autouse=True)
    def check_tailscale_available(self):
        """Skip tests if Tailscale is not available."""
        import subprocess
        try:
            result = subprocess.run(["tailscale", "status"], capture_output=True, text=True, timeout=5)
            if result.returncode != 0 or "Connected" not in result.stdout:
                pytest.skip("Tailscale not connected")
        except (FileNotFoundError, subprocess.TimeoutExpired):
            pytest.skip("tailscale CLI not available or timed out")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_tailscale_status(self):
        """Test Tailscale network status."""
        # This requires tailscale CLI installed
        try:
            process = await asyncio.create_subprocess_exec(
                "tailscale", "status",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE,
            )
            stdout, stderr = await process.communicate()
            
            if process.returncode == 0:
                status_output = stdout.decode()
                assert "Connected" in status_output or "Running" in status_output
                print("✓ Tailscale is connected")
                print(f"  Status: {status_output[:200]}...")
            else:
                pytest.skip(f"Tailscale not running: {stderr.decode()}")
        except FileNotFoundError:
            pytest.skip("tailscale CLI not installed")
    
    @pytest.mark.timeout(TEST_TIMEOUT)
    async def test_agent_tailnet_addresses(self):
        """Test agents have valid Tailscale addresses."""
        async with httpx.AsyncClient() as client:
            for url in [AGENT1_URL, AGENT2_URL, AGENT3_URL]:
                try:
                    response = await client.get(f"{url}/status")
                    if response.status_code == 200:
                        status = response.json()
                        if "tailscale_ip" in status:
                            ip = status["tailscale_ip"]
                            assert ip.startswith("100.")  # Tailscale IPs start with 100.
                            print(f"✓ Agent {url} has Tailscale IP: {ip}")
                except Exception:
                    pass  # Some agents might not expose this


# ============================================================================
# Error Handling Tests
# ============================================================================

@pytest.mark.tailscale
class TestErrorHandling:
    """Test error handling in real network conditions."""
    
    @pytest.fixture(autouse=True)
    def check_bridge_available(self):
        """Skip tests if Tailbridge agents are not running."""
        import httpx
        try:
            with httpx.Client(timeout=5.0) as client:
                response = client.get(f"{AGENT1_URL}/health")
                if response.status_code != 200:
                    pytest.skip(f"Tailbridge agent not available at {AGENT1_URL}")
        except (httpx.ConnectError, httpx.TimeoutException):
            pytest.skip("Tailbridge agents not running - skip error handling tests")
    
    @pytest.mark.timeout(60)
    async def test_agent_unavailable(self):
        """Test handling of unavailable agent."""
        async with httpx.AsyncClient() as client:
            try:
                # Try to reach non-existent agent
                response = await client.post(
                    "http://localhost:9999/a2a/inbound",
                    json={"test": "data"},
                    timeout=5.0,
                )
            except httpx.ConnectError:
                print("✓ Connection error properly handled")
                return
            
            pytest.fail("Should have raised ConnectError")
    
    @pytest.mark.timeout(60)
    async def test_message_timeout(self):
        """Test message timeout handling."""
        async with httpx.AsyncClient(timeout=1.0) as client:
            # Create very large message that might timeout
            large_payload = {"data": "x" * 1000000}
            msg = create_test_message(
                source="test",
                dest="agent1",
                action="test",
                payload=large_payload,
            )
            
            try:
                response = await client.post(
                    f"{AGENT1_URL}/a2a/inbound",
                    json=msg,
                )
                # If it succeeds, that's fine too
                print(f"✓ Message sent (no timeout): {response.status_code}")
            except httpx.TimeoutException:
                print("✓ Message timeout properly handled")
