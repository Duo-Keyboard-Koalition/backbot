#!/usr/bin/env python3
"""
Multi-Agent Tailscale Communication Test - Demo Mode

This script demonstrates the expected output of the multi-agent test
with simulated Tailscale IP addresses and genuine-looking chat logs.

Run this to see what the test output will look like!
"""

import json
import sys
import os
from datetime import datetime
from typing import Dict, List

# Set UTF-8 encoding for Windows
if sys.platform == 'win32':
    sys.stdout.reconfigure(encoding='utf-8')

# ANSI color codes
class Colors:
    CYAN = '\033[0;36m'
    BLUE = '\033[0;34m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    RED = '\033[0;31m'
    MAGENTA = '\033[0;35m'
    WHITE = '\033[1;37m'
    BOLD = '\033[1m'
    RESET = '\033[0m'

class AgentStatus:
    def __init__(self, name: str, hostname: str, tailscale_ip: str, local_ip: str, capabilities: List[str]):
        self.name = name
        self.hostname = hostname
        self.tailscale_ip = tailscale_ip
        self.local_ip = local_ip
        self.capabilities = capabilities
        self.online = True

class ChatMessage:
    def __init__(self, from_agent: str, to_agent: str, content: str, msg_type: str = "chat"):
        self.from_agent = from_agent
        self.to_agent = to_agent
        self.content = content
        self.message_type = msg_type
        self.timestamp = datetime.now()

    def type_emoji(self) -> str:
        if self.message_type == "directive":
            return "📋"
        elif self.message_type == "status":
            return "📊"
        return "💬"

class ChatLog:
    def __init__(self, agent_name: str, ip: str):
        self.agent_name = agent_name
        self.ip = ip
        self.messages: List[ChatMessage] = []

def print_header(text: str, color: str = Colors.CYAN):
    """Print a formatted header"""
    print(f"\n{color}{'=' * 67}{Colors.RESET}")
    print(f"{color}{text.center(67)}{Colors.RESET}")
    print(f"{color}{'=' * 67}{Colors.RESET}")

def print_subheader(text: str, color: str = Colors.BLUE):
    """Print a subheader"""
    print(f"\n{color}{text}{Colors.RESET}")

def print_agent_status(agents: Dict[str, AgentStatus]):
    """Print agent IP addresses and status"""
    print_header("AGENT IP ADDRESSES", Colors.BLUE)
    
    for name, agent in agents.items():
        print(f"\n{Colors.CYAN}{name.upper()}:{Colors.RESET}")
        print(f"  {Colors.WHITE}Hostname:{Colors.RESET}      {agent.hostname}")
        print(f"  {Colors.WHITE}Local IP:{Colors.RESET}      {agent.local_ip}")
        print(f"  {Colors.GREEN}Tailscale IP:{Colors.RESET}  {agent.tailscale_ip}")
        print(f"  {Colors.WHITE}Status:{Colors.RESET}        {Colors.GREEN}Online [OK]{Colors.RESET}")
        print(f"  {Colors.WHITE}Capabilities:{Colors.RESET}  {', '.join(agent.capabilities)}")

def print_chat_logs(chat_logs: Dict[str, ChatLog]):
    """Print formatted chat logs"""
    print_header("FINAL CHAT LOGS - ALL AGENTS", Colors.GREEN)
    print()
    
    for name, log in chat_logs.items():
        print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
        print(f"{Colors.CYAN}|{Colors.RESET} {Colors.BOLD}AGENT:{Colors.RESET} {name.upper().ljust(58)} {Colors.CYAN}|{Colors.RESET}")
        print(f"{Colors.CYAN}|{Colors.RESET} {Colors.WHITE}TAILSCALE IP:{Colors.RESET} {log.ip.ljust(52)} {Colors.CYAN}|{Colors.RESET}")
        print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
        
        for i, msg in enumerate(log.messages):
            time_str = msg.timestamp.strftime("%H:%M:%S")
            from_badge = f"[{msg.from_agent}]"
            to_badge = f"-> [{msg.to_agent}]"
            type_badge = msg.type_emoji()
            
            print(f"{Colors.CYAN}|{Colors.RESET} {i+1}. {time_str} {type_badge} {from_badge:<12} {to_badge:<15} {Colors.CYAN}|{Colors.RESET}")
            
            # Word wrap content
            content = msg.content
            while len(content) > 55:
                print(f"{Colors.CYAN}|{Colors.RESET}       {content[:55]:<55} {Colors.CYAN}|{Colors.RESET}")
                content = content[55:]
            print(f"{Colors.CYAN}|{Colors.RESET}       {content:<55} {Colors.CYAN}|{Colors.RESET}")
            print(f"{Colors.CYAN}|{Colors.RESET} {' ' * 65} {Colors.CYAN}|{Colors.RESET}")
        
        print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
        print()

def run_demo():
    """Run the multi-agent communication demo"""
    print(f"\n{Colors.CYAN}+{'=' * 65}+{Colors.RESET}")
    print(f"{Colors.CYAN}|{Colors.RESET}  {Colors.BOLD}MULTI-AGENT TAILSCALE COMMUNICATION TEST - DEMO{Colors.RESET}  {Colors.CYAN}|{Colors.RESET}")
    print(f"{Colors.CYAN}+{'=' * 65}+{Colors.RESET}")
    
    # Simulated Tailscale configuration
    # In real test, these IPs would be assigned by Tailscale
    print_subheader("📡 Step 1: Simulating Tailscale network setup...", Colors.YELLOW)
    print("  Connecting agents to Tailscale network...")
    print(f"  Using auth key: tskey-auth-k7Q1t39ZWj11CNTRL-...")
    print(f"  {Colors.GREEN}✓{Colors.RESET} Network connected")
    
    # Create agent statuses with simulated Tailscale IPs
    # Tailscale IPs are typically in 100.x.y.z range
    agents = {
        "agent1": AgentStatus(
            name="agent1",
            hostname="agent1",
            tailscale_ip="100.76.142.10",
            local_ip="172.28.0.2",
            capabilities=["file_send", "file_receive", "chat", "command"]
        ),
        "agent2": AgentStatus(
            name="agent2",
            hostname="agent2",
            tailscale_ip="100.89.156.23",
            local_ip="172.28.0.3",
            capabilities=["file_receive", "chat", "stream"]
        ),
        "agent3": AgentStatus(
            name="agent3",
            hostname="agent3",
            tailscale_ip="100.104.67.89",
            local_ip="172.28.0.4",
            capabilities=["file_send", "file_receive", "chat", "command", "stream"]
        )
    }
    
    # Print agent IP addresses
    print_agent_status(agents)
    
    # Simulate chat communications
    print_subheader("💬 Step 2: Running A2A communication tests...", Colors.YELLOW)
    
    chat_logs = {
        "agent1": ChatLog("agent1", agents["agent1"].tailscale_ip),
        "agent2": ChatLog("agent2", agents["agent2"].tailscale_ip),
        "agent3": ChatLog("agent3", agents["agent3"].tailscale_ip)
    }
    
    # Test 1: Agent1 -> Agent2 (Initial contact)
    print("  Test 1: Agent1 initiates contact with Agent2...")
    msg1 = ChatMessage("agent1", "agent2", 
        "Hello Agent 2! This is Agent 1 initiating contact over Tailscale. Can you confirm receipt?",
        "chat")
    chat_logs["agent1"].messages.append(msg1)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Message sent via {agents['agent1'].tailscale_ip} → {agents['agent2'].tailscale_ip}")
    
    # Test 2: Agent2 -> Agent1 (Response)
    print("  Test 2: Agent2 responds...")
    msg2 = ChatMessage("agent2", "agent1",
        "Agent 1, this is Agent 2. Message received loud and clear! Tailscale connection stable at 100.89.156.23. I confirm communication is established.",
        "chat")
    chat_logs["agent2"].messages.append(msg2)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Response received")
    
    # Test 3: Agent1 -> Agent3 (Directive)
    print("  Test 3: Agent1 sends directive to Agent3...")
    msg3 = ChatMessage("agent1", "agent3",
        "Agent 3, please execute diagnostic scan and report status. This is a priority directive.",
        "directive")
    chat_logs["agent1"].messages.append(msg3)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Directive sent")
    
    # Test 4: Agent3 -> Agent1 (Status Report)
    print("  Test 4: Agent3 reports status...")
    msg4 = ChatMessage("agent3", "agent1",
        "Agent 1, diagnostic complete. All systems nominal. Network connectivity: 100%. Tailscale interface active at 100.104.67.89. Ready for task assignment.",
        "status")
    chat_logs["agent3"].messages.append(msg4)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Status report received")
    
    # Test 5: Agent2 -> Agent3 (Collaboration)
    print("  Test 5: Agent2 requests collaboration...")
    msg5 = ChatMessage("agent2", "agent3",
        "Agent 3, Agent 2 here. Let's coordinate on the next task. I'll handle data collection, you handle analysis. Agreed?",
        "chat")
    chat_logs["agent2"].messages.append(msg5)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Collaboration request sent")
    
    # Test 6: Agent3 -> Agent2 (Agreement)
    print("  Test 6: Agent3 confirms collaboration...")
    msg6 = ChatMessage("agent3", "agent2",
        "Agent 2, agreement confirmed. I'll prepare the analysis pipeline. Send data to my endpoint at 100.104.67.89 when ready. Standing by.",
        "chat")
    chat_logs["agent3"].messages.append(msg6)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Agreement confirmed")
    
    # Test 7: Agent1 Broadcast
    print("  Test 7: Agent1 broadcasts to all agents...")
    msg7 = ChatMessage("agent1", "all",
        "ATTENTION ALL AGENTS: Multi-agent communication test successful. Tailscale network verified. All agents operational. Test sequence complete.",
        "directive")
    chat_logs["agent1"].messages.append(msg7)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Broadcast sent")
    
    # Test 8: Agent2 Acknowledgment
    print("  Test 8: Agent2 acknowledges...")
    msg8 = ChatMessage("agent2", "all",
        "Agent 2 acknowledging. Communication test successful. Network stable at 100.89.156.23. Ready for production tasks.",
        "status")
    chat_logs["agent2"].messages.append(msg8)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Acknowledgment received")
    
    # Test 9: Agent3 Acknowledgment
    print("  Test 9: Agent3 acknowledges...")
    msg9 = ChatMessage("agent3", "all",
        "Agent 3 acknowledging. All systems green. Tailscale connection stable at 100.104.67.89. Awaiting further instructions.",
        "status")
    chat_logs["agent3"].messages.append(msg9)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Acknowledgment received")
    
    # Print final chat logs
    print_chat_logs(chat_logs)
    
    # Summary
    print_header("TEST SUMMARY", Colors.GREEN)
    print(f"\n{Colors.WHITE}Agents Tested:{Colors.RESET} 3")
    print(f"{Colors.WHITE}Messages Sent:{Colors.RESET} 9")
    print(f"{Colors.WHITE}Tailscale Network:{Colors.RESET} Connected")
    print(f"{Colors.WHITE}Test Result:{Colors.RESET} {Colors.GREEN}PASSED [OK]{Colors.RESET}")
    
    print(f"\n{Colors.GREEN}+{'=' * 65}+{Colors.RESET}")
    print(f"{Colors.GREEN}|{Colors.RESET}  {Colors.BOLD}TEST COMPLETED SUCCESSFULLY{Colors.RESET}                      {Colors.GREEN}|{Colors.RESET}")
    print(f"{Colors.GREEN}+{'=' * 65}+{Colors.RESET}\n")
    
    # Save to file
    output = {
        "timestamp": datetime.now().isoformat(),
        "agents": {
            name: {
                "hostname": agent.hostname,
                "tailscale_ip": agent.tailscale_ip,
                "local_ip": agent.local_ip,
                "capabilities": agent.capabilities
            }
            for name, agent in agents.items()
        },
        "chat_logs": {
            name: [
                {
                    "from": msg.from_agent,
                    "to": msg.to_agent,
                    "content": msg.content,
                    "type": msg.message_type,
                    "timestamp": msg.timestamp.isoformat()
                }
                for msg in log.messages
            ]
            for name, log in chat_logs.items()
        }
    }
    
    output_file = "test_logs/multiagent-demo-output.json"
    import os
    os.makedirs("test_logs", exist_ok=True)
    with open(output_file, 'w') as f:
        json.dump(output, f, indent=2)
    
    print(f"{Colors.YELLOW}!{Colors.RESET} Test output saved to: {Colors.CYAN}{output_file}{Colors.RESET}\n")

if __name__ == "__main__":
    run_demo()
