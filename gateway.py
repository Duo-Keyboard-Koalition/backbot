#!/usr/bin/env python3
"""
Sentinel AI - Python Gateway
Flask-based API Gateway for Backboard AI
Inspired by NanoClaw architecture
"""

from flask import Flask, request, jsonify
from flask_cors import CORS
from dotenv import load_dotenv
import os
import requests
from agent import Agent

load_dotenv()

app = Flask(__name__)
CORS(app)

# Configuration
API_KEY = os.getenv("BACKBOARD_API_KEY", "")
BASE_URL = "https://api.backboard.ai"

# In-memory storage
agents_store = {}
tasks_store = {}
tools_store = {}
conversations_store = {}

# Default agent
default_agent = Agent(
    name="Sentinel",
    instructions="You are a helpful AI assistant with tool calling capabilities.",
    api_key=API_KEY
)


# ============ Health ============

@app.route("/health", methods=["GET"])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "service": "sentinel-ai-gateway",
        "version": "1.0.0"
    })


# ============ Agents ============

@app.route("/agents", methods=["GET"])
def list_agents():
    """List all agents"""
    return jsonify({
        "agents": list(agents_store.values()) or [
            {"id": "agent_001", "name": "Assistant", "status": "active"},
            {"id": "agent_002", "name": "Analyzer", "status": "active"}
        ]
    })


@app.route("/agents", methods=["POST"])
def create_agent():
    """Create a new agent"""
    data = request.json
    
    if not data or not data.get("name") or not data.get("instructions"):
        return jsonify({"error": "Name and instructions are required"}), 400
    
    agent_id = f"agent_{os.urandom(4).hex()}"
    agent = {
        "id": agent_id,
        "name": data["name"],
        "instructions": data["instructions"],
        "description": data.get("description", ""),
        "status": "active"
    }
    agents_store[agent_id] = agent
    
    return jsonify(agent), 201


@app.route("/agents/<agent_id>", methods=["GET"])
def get_agent(agent_id):
    """Get agent by ID"""
    agent = agents_store.get(agent_id)
    if not agent:
        # Return simulated agent for demo
        agent = {
            "id": agent_id,
            "name": "Assistant",
            "instructions": "You are a helpful assistant",
            "status": "active"
        }
    return jsonify(agent)


@app.route("/agents/<agent_id>", methods=["DELETE"])
def delete_agent(agent_id):
    """Delete an agent"""
    if agent_id in agents_store:
        del agents_store[agent_id]
    return jsonify({"success": True, "message": f"Agent {agent_id} deleted"})


# ============ Agent Execution ============

@app.route("/agents/execute", methods=["POST"])
def execute_agent():
    """Execute a task with an agent"""
    data = request.json
    
    if not data or not data.get("task"):
        return jsonify({"error": "Task is required"}), 400
    
    task = data["task"]
    agent_id = data.get("agentId")
    
    # Use default agent for execution
    response = default_agent.chat(task)
    
    return jsonify({
        "id": f"exec_{os.urandom(4).hex()}",
        "agentId": agent_id or "default",
        "task": task,
        "status": "completed",
        "result": response.content,
        "tool_calls": [
            {"name": tc.name, "arguments": tc.arguments}
            for tc in response.tool_calls
        ] if response.tool_calls else []
    })


# ============ Tasks ============

@app.route("/tasks", methods=["GET"])
def list_tasks():
    """List all tasks"""
    return jsonify({
        "tasks": list(tasks_store.values()) or [
            {"id": "task_001", "agent_id": "agent_001", "task": "Analyze data", "status": "completed"},
            {"id": "task_002", "agent_id": "agent_001", "task": "Generate report", "status": "pending"}
        ]
    })


@app.route("/tasks", methods=["POST"])
def create_task():
    """Create a new task"""
    data = request.json
    
    if not data or not data.get("agentId") or not data.get("task"):
        return jsonify({"error": "Agent ID and task are required"}), 400
    
    task_id = f"task_{os.urandom(4).hex()}"
    task = {
        "id": task_id,
        "agent_id": data["agentId"],
        "task": data["task"],
        "priority": data.get("priority", "normal"),
        "status": "pending"
    }
    tasks_store[task_id] = task
    
    return jsonify(task), 201


@app.route("/tasks/<task_id>", methods=["GET"])
def get_task(task_id):
    """Get task by ID"""
    task = tasks_store.get(task_id)
    if not task:
        task = {
            "id": task_id,
            "agent_id": "agent_001",
            "task": "Sample task",
            "status": "completed",
            "result": "Task completed successfully"
        }
    return jsonify(task)


# ============ Tools ============

@app.route("/tools", methods=["GET"])
def list_tools():
    """List all registered tools"""
    return jsonify({
        "tools": list(tools_store.values()) or [
            {"id": "tool_001", "name": "calculator", "description": "Math calculations"},
            {"id": "tool_002", "name": "search", "description": "Web search"}
        ]
    })


@app.route("/tools", methods=["POST"])
def register_tool():
    """Register a new tool"""
    data = request.json
    
    if not data or not data.get("name") or not data.get("schema"):
        return jsonify({"error": "Name and schema are required"}), 400
    
    tool_id = f"tool_{os.urandom(4).hex()}"
    tool = {
        "id": tool_id,
        "name": data["name"],
        "description": data.get("description", ""),
        "schema": data["schema"]
    }
    tools_store[tool_id] = tool
    
    return jsonify(tool), 201


# ============ Conversations ============

@app.route("/conversations", methods=["POST"])
def conversation():
    """Send a message in a conversation"""
    data = request.json
    
    if not data or not data.get("agentId") or not data.get("message"):
        return jsonify({"error": "Agent ID and message are required"}), 400
    
    agent_id = data["agentId"]
    message = data["message"]
    conversation_id = data.get("conversationId") or f"conv_{os.urandom(4).hex()}"
    
    # Process message with agent
    response = default_agent.chat(message)
    
    return jsonify({
        "message": response.content,
        "conversation_id": conversation_id,
        "tool_calls": [
            {"name": tc.name, "arguments": tc.arguments}
            for tc in response.tool_calls
        ] if response.tool_calls else []
    })


# ============ Chat (shortcut) ============

@app.route("/chat", methods=["POST"])
def chat():
    """Quick chat endpoint"""
    data = request.json
    
    if not data or not data.get("message"):
        return jsonify({"error": "Message is required"}), 400
    
    response = default_agent.chat(data["message"])
    
    return jsonify({
        "response": response.content,
        "tool_calls": [
            {"name": tc.name, "arguments": tc.arguments}
            for tc in response.tool_calls
        ] if response.tool_calls else []
    })


# ============ Error Handlers ============

@app.errorhandler(404)
def not_found(error):
    return jsonify({"error": "Endpoint not found"}), 404


@app.errorhandler(500)
def internal_error(error):
    return jsonify({"error": "Internal server error"}), 500


# ============ Main ============

if __name__ == "__main__":
    port = int(os.getenv("PORT", 3000))
    print(f"🚀 Sentinel AI Gateway running on port {port}")
    print(f"📡 Backboard API configured")
    print(f"🔗 Health check: http://localhost:{port}/health")
    app.run(host="0.0.0.0", port=port, debug=True)
