# Backboard IO 🚀

Python Agentic Gateway for Backboard AI - Inspired by NanoClaw

A lightweight Python agent framework with tool calling capabilities, CLI tools, TUI, and API gateway.

## Features

- 🤖 **Python Agent** - NanoClaw-inspired agent with tool calling
- 🔧 **Tool System** - Register and execute custom tools
- 💻 **CLI Tools** - Full-featured command-line interface
- 🖥️ **TUI** - Beautiful terminal user interface with Textual
- 🌐 **API Gateway** - Flask-based REST API

## Installation

```bash
# Install Python dependencies
pip install requests click textual python-dotenv flask flask-cors

# Or use requirements.txt
pip install -r requirements.txt
```

## Quick Start

### 1. Configure Environment

Create a `.env` file:
```
BACKBOARD_API_KEY=espr_qsVeIx7RGSkv0sczKPqWHyPnxm3rOdx76Rv7alXqQFw
PORT=3000
```

### 2. Run the Gateway

```bash
python gateway.py
```

### 3. Use the CLI

```bash
# List agents
python cli.py agents list

# Create an agent
python cli.py agents create "MyAgent" "You are a helpful assistant"

# Execute a task
python cli.py execute agent_001 "Calculate 2+2"

# Chat with an agent
python cli.py chat agent_001 "Hello!"

# Start REPL
python cli.py repl
```

### 4. Launch the TUI

```bash
python tui.py
```

## Agent Usage

```python
from agent import Agent, ToolParameter, ToolParameterType

# Create an agent
agent = Agent(
    name="Sentinel",
    instructions="You are a helpful AI assistant."
)

# Register a custom tool
agent.register_tool(
    name="weather",
    description="Get weather information",
    parameters=[
        ToolParameter("city", ToolParameterType.STRING, "City name")
    ],
    handler=lambda city: f"Weather in {city}: Sunny, 25°C"
)

# Chat with the agent
response = agent.chat("What's the weather in Toronto?")
print(response.content)
```

## CLI Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `agents list` | `al` | List all agents |
| `agents create <name> <instructions>` | `ac` | Create agent |
| `agents get <id>` | `ag` | Get agent details |
| `agents delete <id>` | `ad` | Delete agent |
| `tasks list` | `tl` | List tasks |
| `tasks create <agent> <task>` | `tc` | Create task |
| `tasks get <id>` | `tg` | Get task status |
| `execute <agent> <task>` | `e` | Execute task |
| `chat <agent> <message>` | `c` | Chat with agent |
| `tools list` | `tol` | List tools |
| `tools register <name> <schema>` | `tr` | Register tool |
| `repl` | - | Interactive REPL |
| `health` | `h` | Health check |

## TUI Controls

| Key | Action |
|-----|--------|
| `1` | Agents tab |
| `2` | Tasks tab |
| `3` | Execute tab |
| `4` | Chat tab |
| `q` | Quit |
| `r` | Refresh |
| `a` | Add agent |
| `c` | Create task |
| `e` | Execute |
| `s` | Select agent |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/agents` | List agents |
| POST | `/agents` | Create agent |
| GET | `/agents/<id>` | Get agent |
| DELETE | `/agents/<id>` | Delete agent |
| POST | `/agents/execute` | Execute task |
| GET | `/tasks` | List tasks |
| POST | `/tasks` | Create task |
| GET | `/tasks/<id>` | Get task |
| POST | `/tools` | Register tool |
| POST | `/conversations` | Chat conversation |
| POST | `/chat` | Quick chat |

## Tool Calling Format

The agent uses XML-style tool calling:

```
<tool:tool_name>{"arg1": "value1", "arg2": "value2"}
```

Example:
```
<tool:calculator>{"expression": "2+2"}
<tool:save_memory>{"key": "name", "value": "John"}
```

## License

ISC
