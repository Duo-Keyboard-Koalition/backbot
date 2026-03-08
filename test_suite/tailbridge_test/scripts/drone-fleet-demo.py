#!/usr/bin/env python3
"""
Drone Fleet Tailscale Communication Test - Demo Mode

This script demonstrates the expected output of the drone fleet test
with simulated Tailscale IP addresses, task assignments, and communication logs.

Run this to see what the drone fleet test output will look like!
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

class DroneStatus:
    def __init__(self, id: str, name: str, role: str, tailscale_ip: str, 
                 local_ip: str, battery: int, position: Dict, capabilities: List[str]):
        self.id = id
        self.name = name
        self.role = role
        self.tailscale_ip = tailscale_ip
        self.local_ip = local_ip
        self.battery = battery
        self.position = position
        self.capabilities = capabilities
        self.online = True

class Task:
    def __init__(self, id: str, task_type: str, priority: str, assigned_to: str, 
                 description: str, params: Dict):
        self.id = id
        self.type = task_type
        self.priority = priority
        self.assigned_to = assigned_to
        self.description = description
        self.params = params
        self.status = "pending"
        self.created_at = datetime.now()

class DroneMessage:
    def __init__(self, from_drone: str, to_drone: str, msg_type: str, 
                 content: str, task_id: str = ""):
        self.from_drone = from_drone
        self.to_drone = to_drone
        self.type = msg_type
        self.content = content
        self.task_id = task_id
        self.timestamp = datetime.now()

    def type_emoji(self) -> str:
        if self.type == "task_assignment":
            return "📋"
        elif self.type == "status_report":
            return "📊"
        elif self.type == "coordination":
            return "🤝"
        elif self.type == "alert":
            return "⚠️"
        return "💬"

class TaskLog:
    def __init__(self, drone_id: str, ip: str, role: str):
        self.drone_id = drone_id
        self.ip = ip
        self.role = role
        self.tasks: List[Task] = []
        self.messages: List[DroneMessage] = []

def print_header(text: str, color: str = Colors.CYAN):
    """Print a formatted header"""
    print(f"\n{color}{'=' * 67}{Colors.RESET}")
    print(f"{color}{text.center(67)}{Colors.RESET}")
    print(f"{color}{'=' * 67}{Colors.RESET}")

def print_subheader(text: str, color: str = Colors.BLUE):
    """Print a subheader"""
    print(f"\n{color}{text}{Colors.RESET}")

def print_drone_status(drones: Dict[str, DroneStatus]):
    """Print drone fleet status"""
    print_header("DRONE FLEET STATUS", Colors.BLUE)
    
    for name, drone in drones.items():
        role_badge = "🚁"
        if drone.role == "lead":
            role_badge = "👑"
        elif drone.role == "scout":
            role_badge = "🔍"
        elif drone.role == "worker":
            role_badge = "📦"
        elif drone.role == "relay":
            role_badge = "📡"
        
        print(f"\n{Colors.CYAN}{role_badge} {name.upper()}:{Colors.RESET}")
        print(f"  {Colors.WHITE}Drone ID:{Colors.RESET}     {drone.id}")
        print(f"  {Colors.WHITE}Role:{Colors.RESET}          {drone.role.upper()}")
        print(f"  {Colors.WHITE}Local IP:{Colors.RESET}      {drone.local_ip}")
        print(f"  {Colors.GREEN}Tailscale IP:{Colors.RESET}  {drone.tailscale_ip}")
        print(f"  {Colors.WHITE}Battery:{Colors.RESET}       {drone.battery}%")
        print(f"  {Colors.WHITE}Position:{Colors.RESET}      {drone.position['lat']:.6f}, {drone.position['lon']:.6f}, {drone.position['alt']:.0f}m")
        print(f"  {Colors.WHITE}Capabilities:{Colors.RESET}  {', '.join(drone.capabilities)}")

def print_task_logs(task_logs: Dict[str, TaskLog]):
    """Print formatted task logs"""
    print_header("FINAL TASK LOGS - ALL DRONES", Colors.GREEN)
    print()
    
    for name, log in task_logs.items():
        role_badge = "🚁"
        if log.role == "lead":
            role_badge = "👑"
        elif log.role == "scout":
            role_badge = "🔍"
        elif log.role == "worker":
            role_badge = "📦"
        elif log.role == "relay":
            role_badge = "📡"
        
        print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
        print(f"{Colors.CYAN}|{Colors.RESET} {role_badge} {Colors.BOLD}DRONE:{Colors.RESET} {name.upper().ljust(54)} {Colors.CYAN}|{Colors.RESET}")
        print(f"{Colors.CYAN}|{Colors.RESET}   {Colors.WHITE}ROLE:{Colors.RESET} {log.role.upper().ljust(56)} {Colors.CYAN}|{Colors.RESET}")
        print(f"{Colors.CYAN}|{Colors.RESET}   {Colors.WHITE}TAILSCALE IP:{Colors.RESET} {log.ip.ljust(50)} {Colors.CYAN}|{Colors.RESET}")
        print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
        
        # Print tasks
        if len(log.tasks) > 0:
            print(f"{Colors.CYAN}|{Colors.RESET} {Colors.BOLD}TASKS ASSIGNED:{Colors.RESET} {len(log.tasks)}                                        {Colors.CYAN}|{Colors.RESET}")
            print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
            
            for i, task in enumerate(log.tasks):
                priority_color = Colors.RED if task.priority == "critical" else Colors.YELLOW if task.priority == "high" else Colors.WHITE
                print(f"{Colors.CYAN}|{Colors.RESET}   {i+1}. [{task.priority.upper()}] {task.id} - {task.type.upper()}")
                print(f"{Colors.CYAN}|{Colors.RESET}      {task.description[:60]}")
        
        # Print messages
        if len(log.messages) > 0:
            print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
            print(f"{Colors.CYAN}|{Colors.RESET} {Colors.BOLD}COMMUNICATION LOG:{Colors.RESET} {len(log.messages)} messages                           {Colors.CYAN}|{Colors.RESET}")
            print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
            
            for i, msg in enumerate(log.messages):
                time_str = msg.timestamp.strftime("%H:%M:%S")
                type_badge = msg.type_emoji()
                
                print(f"{Colors.CYAN}|{Colors.RESET} {i+1}. {time_str} {type_badge} {msg.from_drone:<15} -> {msg.to_drone:<15}")
                
                # Word wrap content
                content = msg.content
                while len(content) > 58:
                    print(f"{Colors.CYAN}|{Colors.RESET}       {content[:58]}")
                    content = content[58:]
                print(f"{Colors.CYAN}|{Colors.RESET}       {content}")
                print(f"{Colors.CYAN}|{Colors.RESET} {' ' * 65}")
        
        print(f"{Colors.CYAN}+{'-' * 65}+{Colors.RESET}")
        print()

def run_demo():
    """Run the drone fleet mission demo"""
    print(f"\n{Colors.CYAN}+{'=' * 65}+{Colors.RESET}")
    print(f"{Colors.CYAN}|{Colors.RESET}  {Colors.BOLD}DRONE FLEET TAILSCALE COMMUNICATION TEST - DEMO{Colors.RESET}  {Colors.CYAN}|{Colors.RESET}")
    print(f"{Colors.CYAN}|{Colors.RESET}  {Colors.BOLD}Mission: Urban Survey & Emergency Delivery{Colors.RESET}       {Colors.CYAN}|{Colors.RESET}")
    print(f"{Colors.CYAN}+{'=' * 65}+{Colors.RESET}")
    
    # Simulated Tailscale configuration
    print_subheader("🚁 Phase 1: Drone Fleet Initialization...", Colors.YELLOW)
    print("  Connecting drones to Tailscale network...")
    print(f"  Using auth key: tskey-auth-k7Q1t39ZWj11CNTRL-...")
    print(f"  {Colors.GREEN}✓{Colors.RESET} Fleet connected to Tailscale")
    
    # Create drone statuses with simulated Tailscale IPs
    drones = {
        "lead-drone": DroneStatus(
            id="DRONE-LEAD-001",
            name="lead-drone",
            role="lead",
            tailscale_ip="100.76.142.10",
            local_ip="172.29.0.2",
            battery=100,
            position={"lat": 40.7580, "lon": -73.9855, "alt": 200},
            capabilities=["orchestrate", "coordinate", "task_assign", "monitor"]
        ),
        "scout-drone": DroneStatus(
            id="DRONE-SCOUT-002",
            name="scout-drone",
            role="scout",
            tailscale_ip="100.89.156.23",
            local_ip="172.29.0.3",
            battery=95,
            position={"lat": 40.7590, "lon": -73.9845, "alt": 150},
            capabilities=["survey", "mapping", "reconnaissance"]
        ),
        "worker-drone": DroneStatus(
            id="DRONE-WORKER-003",
            name="worker-drone",
            role="worker",
            tailscale_ip="100.104.67.89",
            local_ip="172.29.0.4",
            battery=98,
            position={"lat": 40.7570, "lon": -73.9865, "alt": 100},
            capabilities=["delivery", "transport", "payload_carry"]
        ),
        "relay-drone": DroneStatus(
            id="DRONE-RELAY-004",
            name="relay-drone",
            role="relay",
            tailscale_ip="100.118.201.45",
            local_ip="172.29.0.5",
            battery=92,
            position={"lat": 40.7585, "lon": -73.9850, "alt": 180},
            capabilities=["relay", "communication", "signal_boost"]
        )
    }
    
    # Print drone status
    print_drone_status(drones)
    
    # Initialize task logs
    task_logs = {
        "lead-drone": TaskLog("lead-drone", drones["lead-drone"].tailscale_ip, "lead"),
        "scout-drone": TaskLog("scout-drone", drones["scout-drone"].tailscale_ip, "scout"),
        "worker-drone": TaskLog("worker-drone", drones["worker-drone"].tailscale_ip, "worker"),
        "relay-drone": TaskLog("relay-drone", drones["relay-drone"].tailscale_ip, "relay")
    }
    
    mission_id = f"MISSION-{int(datetime.now().timestamp())}"
    
    # Phase 2: Mission Briefing
    print_subheader("📋 Phase 2: Mission Briefing - Task Assignment...", Colors.YELLOW)
    print(f"  Mission ID: {mission_id}")
    print(f"  Objective:  Urban Survey and Emergency Supply Delivery")
    print(f"  Area:       Downtown District (2.5 sq km)")
    print()
    
    # Task 1: Lead assigns survey to Scout
    print("  Task 1: Lead Drone → Scout Drone (Aerial Survey)")
    task1 = Task("TASK-001", "survey", "high", "scout-drone",
                 "Conduct aerial survey of downtown district",
                 {"area": 2.5, "duration": 1800})
    task_logs["lead-drone"].tasks.append(task1)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Task assigned via {drones['lead-drone'].tailscale_ip} → {drones['scout-drone'].tailscale_ip}")
    
    # Task 2: Lead assigns delivery to Worker
    print("  Task 2: Lead Drone → Worker Drone (Emergency Delivery)")
    task2 = Task("TASK-002", "deliver", "critical", "worker-drone",
                 "Deliver emergency medical supplies to Hospital Zone A",
                 {"payload": "Medical Supplies Kit #A-47", "duration": 900})
    task_logs["lead-drone"].tasks.append(task2)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Task assigned via {drones['lead-drone'].tailscale_ip} → {drones['worker-drone'].tailscale_ip}")
    
    # Task 3: Lead assigns relay to Relay Drone
    print("  Task 3: Lead Drone → Relay Drone (Communication Relay)")
    task3 = Task("TASK-003", "patrol", "medium", "relay-drone",
                 "Maintain communication relay between fleet and base",
                 {"duration": 3600})
    task_logs["lead-drone"].tasks.append(task3)
    print(f"    {Colors.GREEN}✓{Colors.RESET} Task assigned via {drones['lead-drone'].tailscale_ip} → {drones['relay-drone'].tailscale_ip}")
    
    # Phase 3: Status Reports
    print_subheader("✈️ Phase 3: Task Execution - Status Reports...", Colors.YELLOW)
    
    # Scout progress report
    msg1 = DroneMessage("scout-drone", "lead-drone", "status_report",
                        "Lead Drone, Scout reporting. Survey 45% complete. Coverage: 1.1 sq km. Weather conditions optimal. Battery at 78%. ETA 22 minutes.",
                        "TASK-001")
    task_logs["scout-drone"].messages.append(msg1)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Scout Drone status report")
    
    # Worker departure report
    msg2 = DroneMessage("worker-drone", "lead-drone", "status_report",
                        "Lead Drone, Worker reporting. Departed base with payload 'Medical Supplies Kit #A-47'. En route to Hospital Zone A. Battery at 92%. ETA 12 minutes.",
                        "TASK-002")
    task_logs["worker-drone"].messages.append(msg2)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Worker Drone status report")
    
    # Relay position report
    msg3 = DroneMessage("relay-drone", "lead-drone", "status_report",
                        "Lead Drone, Relay reporting. Stationed at coordinates 40.7580, -73.9855, altitude 150m. Signal strength excellent. All fleet communications routing through relay. Battery at 88%.",
                        "TASK-003")
    task_logs["relay-drone"].messages.append(msg3)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Relay Drone status report")
    
    # Phase 4: Inter-Drone Coordination
    print_subheader("🤝 Phase 4: Inter-Drone Coordination...", Colors.YELLOW)
    
    # Scout coordinates with Worker
    msg4 = DroneMessage("scout-drone", "worker-drone", "coordination",
                        "Worker Drone, Scout here. I've identified optimal landing zone at Grid C-7 during my survey. Coordinates transmitted. Suitable for your delivery. Over.",
                        "TASK-001")
    task_logs["scout-drone"].messages.append(msg4)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Scout → Worker coordination")
    
    # Worker acknowledges
    msg5 = DroneMessage("worker-drone", "scout-drone", "coordination",
                        "Scout Drone, Worker copying. Grid C-7 coordinates received. Adjusting route. Excellent intel. Will confirm landing on approach. Out.",
                        "TASK-002")
    task_logs["worker-drone"].messages.append(msg5)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Worker → Scout acknowledgment")
    
    # Phase 5: Mission Completion
    print_subheader("✅ Phase 5: Mission Completion Reports...", Colors.YELLOW)
    
    # Scout completes
    msg6 = DroneMessage("scout-drone", "lead-drone", "status_report",
                        "Lead Drone, Scout reporting TASK-001 COMPLETE. Full aerial survey completed. 2.5 sq km mapped. 3D model uploaded to base. Battery at 62%. Returning to base.",
                        "TASK-001")
    task_logs["scout-drone"].messages.append(msg6)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Scout Drone mission complete")
    
    # Worker completes
    msg7 = DroneMessage("worker-drone", "lead-drone", "status_report",
                        "Lead Drone, Worker reporting TASK-002 COMPLETE. Medical supplies delivered to Hospital Zone A. Landing zone Grid C-7 performed flawlessly. Battery at 71%. Returning to base.",
                        "TASK-002")
    task_logs["worker-drone"].messages.append(msg7)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Worker Drone mission complete")
    
    # Relay continues
    msg8 = DroneMessage("relay-drone", "lead-drone", "status_report",
                        "Lead Drone, Relay reporting. All fleet communications maintained throughout mission. Zero packet loss. Battery at 76%. Continuing patrol.",
                        "TASK-003")
    task_logs["relay-drone"].messages.append(msg8)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Relay Drone operational")
    
    # Lead Drone mission summary
    msg9 = DroneMessage("lead-drone", "all", "status_report",
                        f"MISSION COMPLETE. Mission ID: {mission_id}. All objectives achieved. Survey: 100%. Delivery: 100%. Relay: Operational. Fleet returning to base. Outstanding work: None.",
                        "")
    task_logs["lead-drone"].messages.append(msg9)
    print(f"  {Colors.GREEN}✓{Colors.RESET} Lead Drone mission summary broadcast")
    
    # Print final task logs
    print_task_logs(task_logs)
    
    # Summary
    print_header("MISSION SUMMARY", Colors.GREEN)
    print(f"\n{Colors.WHITE}Drones Deployed:{Colors.RESET} 4")
    print(f"{Colors.WHITE}Tasks Assigned:{Colors.RESET} 3")
    print(f"{Colors.WHITE}Messages Exchanged:{Colors.RESET} 9")
    print(f"{Colors.WHITE}Tailscale Network:{Colors.RESET} Connected")
    print(f"{Colors.WHITE}Mission Result:{Colors.RESET} {Colors.GREEN}SUCCESS ✓{Colors.RESET}")
    
    print(f"\n{Colors.GREEN}+{'=' * 65}+{Colors.RESET}")
    print(f"{Colors.GREEN}|{Colors.RESET}  {Colors.BOLD}MISSION COMPLETED SUCCESSFULLY{Colors.RESET}                     {Colors.GREEN}|{Colors.RESET}")
    print(f"{Colors.GREEN}+{'=' * 65}+{Colors.RESET}\n")
    
    # Save to JSON
    output = {
        "timestamp": datetime.now().isoformat(),
        "mission_id": mission_id,
        "drones": {
            name: {
                "id": drone.id,
                "role": drone.role,
                "tailscale_ip": drone.tailscale_ip,
                "local_ip": drone.local_ip,
                "battery": drone.battery,
                "position": drone.position,
                "capabilities": drone.capabilities
            }
            for name, drone in drones.items()
        },
        "task_logs": {
            name: {
                "role": log.role,
                "ip": log.ip,
                "tasks": [
                    {
                        "id": task.id,
                        "type": task.type,
                        "priority": task.priority,
                        "description": task.description
                    }
                    for task in log.tasks
                ],
                "messages": [
                    {
                        "from": msg.from_drone,
                        "to": msg.to_drone,
                        "type": msg.type,
                        "content": msg.content,
                        "task_id": msg.task_id,
                        "timestamp": msg.timestamp.isoformat()
                    }
                    for msg in log.messages
                ]
            }
            for name, log in task_logs.items()
        }
    }
    
    output_file = "test_logs/drone-fleet-demo-output.json"
    os.makedirs("test_logs", exist_ok=True)
    with open(output_file, 'w') as f:
        json.dump(output, f, indent=2)
    
    print(f"{Colors.YELLOW}!{Colors.RESET} Mission output saved to: {Colors.CYAN}{output_file}{Colors.RESET}\n")

if __name__ == "__main__":
    run_demo()
