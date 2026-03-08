# Tmux Skill

Terminal multiplexer integration for the darci agent.

## Description

This skill enables the darci agent to interact with tmux sessions for terminal management and long-running processes.

## Tools

- `tmux_new_session`: Create a new tmux session
- `tmux_list_sessions`: List all tmux sessions
- `tmux_send_keys`: Send commands to a tmux session
- `tmux_capture_pane`: Capture output from a tmux session
- `tmux_kill_session`: Kill a tmux session

## Usage

Ask the agent to:
- "Start a new tmux session for the build process"
- "Show me the output from the dev session"
- "Run the server in a tmux session"
