#!/usr/bin/env python3
"""
Sentinel AI - Terminal UI (TUI)
Python TUI for Backboard AI Agent Management using Textual
"""

from textual.app import App, ComposeResult
from textual.screen import Screen
from textual.containers import Container, Horizontal, Vertical, ScrollableContainer
from textual.widgets import Header, Footer, Static, Button, Input, Label, DataTable, TabbedContent, TabPane
from textual.binding import Binding
from textual.message import Message
from textual import work
import json


class AgentRow(Static):
    """Widget for displaying an agent in the list"""
    
    def __init__(self, agent_id: str, name: str, status: str, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.agent_id = agent_id
        self.name = name
        self.status = status
    
    def compose(self) -> ComposeResult:
        status_color = "green" if self.status == "active" else "yellow"
        yield Static(f"[bold]{self.name}[/bold] (ID: {self.agent_id})")
        yield Static(f"Status: [{status_color}]{self.status}[/]")


class TaskRow(Static):
    """Widget for displaying a task in the list"""
    
    def __init__(self, task_id: str, agent_id: str, task: str, status: str, *args, **kwargs):
        super().__init__(*args, **kwargs)
        self.task_id = task_id
        self.agent_id = agent_id
        self.task = task
        self.status = status
    
    def compose(self) -> ComposeResult:
        status_icon = "✅" if self.status == "completed" else "⏳"
        task_preview = self.task[:50] + "..." if len(self.task) > 50 else self.task
        yield Static(f"{status_icon} [bold]{self.task_id}[/] | Agent: {self.agent_id} | {task_preview}")


class AgentsScreen(Screen):
    """Screen for managing agents"""
    
    BINDINGS = [
        Binding("r", "refresh", "Refresh"),
        Binding("a", "add_agent", "Add Agent"),
        Binding("d", "delete_agent", "Delete"),
        Binding("e", "execute", "Execute"),
        Binding("s", "select", "Select"),
    ]
    
    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("🤖 Agents", id="screen-title")
        
        with Container(id="agents-container"):
            with ScrollableContainer(id="agents-list"):
                yield Static("Loading agents...", id="agents-placeholder")
        
        with Horizontal(id="agents-actions"):
            yield Button("🔄 Refresh", id="refresh-btn", variant="primary")
            yield Button("➕ Add Agent", id="add-btn", variant="success")
            yield Button("▶️ Execute", id="execute-btn", variant="warning")
        
        yield Footer()
    
    def on_mount(self) -> None:
        self.refresh_agents()
    
    def refresh_agents(self) -> None:
        """Refresh the agents list"""
        # Simulated agents data
        self.agents = [
            {"id": "agent_001", "name": "Assistant", "status": "active"},
            {"id": "agent_002", "name": "Analyzer", "status": "active"},
            {"id": "agent_003", "name": "Coder", "status": "idle"},
        ]
        self._update_agents_display()
    
    def _update_agents_display(self) -> None:
        """Update the agents list display"""
        agents_list = self.query_one("#agents-list", ScrollableContainer)
        agents_list.remove_children()
        
        if not self.agents:
            agents_list.mount(Static("No agents found. Press 'a' to add one.", id="no-agents"))
            return
        
        for agent in self.agents:
            row = AgentRow(agent["id"], agent["name"], agent["status"])
            row.add_class("agent-row")
            agents_list.mount(row)
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "refresh-btn":
            self.refresh_agents()
            self.notify("Agents refreshed")
        elif event.button.id == "add-btn":
            self.app.push_screen("add_agent_screen")
        elif event.button.id == "execute-btn":
            self.app.push_screen("execute_screen")
    
    def action_refresh(self) -> None:
        self.refresh_agents()
        self.notify("Agents refreshed")
    
    def action_add_agent(self) -> None:
        self.app.push_screen("add_agent_screen")
    
    def action_execute(self) -> None:
        self.app.push_screen("execute_screen")
    
    def action_delete(self) -> None:
        self.notify("Delete functionality - select an agent first")
    
    def action_select(self) -> None:
        if self.agents:
            self.app.selected_agent = self.agents[0]
            self.notify(f"Selected agent: {self.agents[0]['name']}")


class TasksScreen(Screen):
    """Screen for managing tasks"""
    
    BINDINGS = [
        Binding("r", "refresh", "Refresh"),
        Binding("c", "create_task", "Create"),
    ]
    
    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("📋 Tasks", id="screen-title")
        
        with Container(id="tasks-container"):
            with ScrollableContainer(id="tasks-list"):
                yield Static("Loading tasks...", id="tasks-placeholder")
        
        with Horizontal(id="tasks-actions"):
            yield Button("🔄 Refresh", id="refresh-btn", variant="primary")
            yield Button("➕ Create Task", id="create-btn", variant="success")
        
        yield Footer()
    
    def on_mount(self) -> None:
        self.refresh_tasks()
    
    def refresh_tasks(self) -> None:
        """Refresh the tasks list"""
        self.tasks = [
            {"id": "task_001", "agent_id": "agent_001", "task": "Analyze data and generate report", "status": "completed"},
            {"id": "task_002", "agent_id": "agent_001", "task": "Process incoming requests", "status": "pending"},
            {"id": "task_003", "agent_id": "agent_002", "task": "Monitor system health", "status": "running"},
        ]
        self._update_tasks_display()
    
    def _update_tasks_display(self) -> None:
        """Update the tasks list display"""
        tasks_list = self.query_one("#tasks-list", ScrollableContainer)
        tasks_list.remove_children()
        
        if not self.tasks:
            tasks_list.mount(Static("No tasks found. Press 'c' to create one.", id="no-tasks"))
            return
        
        for task in self.tasks:
            row = TaskRow(task["id"], task["agent_id"], task["task"], task["status"])
            row.add_class("task-row")
            tasks_list.mount(row)
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "refresh-btn":
            self.refresh_tasks()
            self.notify("Tasks refreshed")
        elif event.button.id == "create-btn":
            self.app.push_screen("create_task_screen")
    
    def action_refresh(self) -> None:
        self.refresh_tasks()
        self.notify("Tasks refreshed")
    
    def action_create_task(self) -> None:
        self.app.push_screen("create_task_screen")


class ExecuteScreen(Screen):
    """Screen for executing tasks"""
    
    BINDINGS = [
        Binding("escape", "go_back", "Back"),
    ]
    
    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("⚡ Execute Task", id="screen-title")
        
        with Container(id="execute-container"):
            yield Static("Selected Agent: ", id="selected-agent")
            yield Input(placeholder="Enter task description...", id="task-input")
            yield Button("🚀 Execute", id="execute-btn", variant="warning")
            yield Static("", id="result-display")
        
        yield Footer()
    
    def on_mount(self) -> None:
        if hasattr(self.app, 'selected_agent') and self.app.selected_agent:
            self.query_one("#selected-agent", Static).update(
                f"Selected Agent: [bold]{self.app.selected_agent['name']}[/]"
            )
        else:
            self.query_one("#selected-agent", Static).update(
                "Selected Agent: [red]None - Go to Agents tab to select one[/]"
            )
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "execute-btn":
            task_input = self.query_one("#task-input", Input)
            result_display = self.query_one("#result-display", Static)
            
            if task_input.value:
                result_display.update(f"[yellow]Executing: {task_input.value}...[/]")
                # Simulate execution
                self._execute_task(task_input.value)
            else:
                self.notify("Please enter a task", severity="error")
    
    def _execute_task(self, task: str) -> None:
        """Execute the task"""
        result_display = self.query_one("#result-display", Static)
        result_display.update(f"[green]✅ Result: Task completed successfully![/]\n\nOutput: {task}")
        self.notify("Task executed!")
    
    def action_go_back(self) -> None:
        self.app.pop_screen()


class CreateTaskScreen(Screen):
    """Screen for creating a new task"""
    
    BINDINGS = [
        Binding("escape", "go_back", "Back"),
    ]
    
    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("📝 Create Task", id="screen-title")
        
        with Container(id="create-task-container"):
            yield Static("Agent ID:", id="agent-label")
            yield Input(placeholder="agent_001", id="agent-input")
            yield Static("Task:", id="task-label")
            yield Input(placeholder="Enter task description...", id="task-input")
            yield Static("Priority:", id="priority-label")
            with Horizontal(id="priority-buttons"):
                yield Button("Low", id="priority-low", variant="default")
                yield Button("Normal", id="priority-normal", variant="primary")
                yield Button("High", id="priority-high", variant="error")
            yield Static("", id="priority-selected", classes="priority-display")
            yield Button("✅ Create", id="create-btn", variant="success")
        
        yield Footer()
    
    def on_mount(self) -> None:
        self.priority = "normal"
        self.query_one("#priority-selected", Static).update("Priority: [yellow]Normal[/]")
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "priority-low":
            self.priority = "low"
            self.query_one("#priority-selected", Static).update("Priority: [green]Low[/]")
        elif event.button.id == "priority-normal":
            self.priority = "normal"
            self.query_one("#priority-selected", Static).update("Priority: [yellow]Normal[/]")
        elif event.button.id == "priority-high":
            self.priority = "high"
            self.query_one("#priority-selected", Static).update("Priority: [red]High[/]")
        elif event.button.id == "create-btn":
            self._create_task()
    
    def _create_task(self) -> None:
        """Create the task"""
        agent_input = self.query_one("#agent-input", Input)
        task_input = self.query_one("#task-input", Input)
        
        if agent_input.value and task_input.value:
            self.notify(f"Task created for agent {agent_input.value}!")
            self.app.pop_screen()
        else:
            self.notify("Please fill in all fields", severity="error")
    
    def action_go_back(self) -> None:
        self.app.pop_screen()


class AddAgentScreen(Screen):
    """Screen for adding a new agent"""
    
    BINDINGS = [
        Binding("escape", "go_back", "Back"),
    ]
    
    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("➕ Add Agent", id="screen-title")
        
        with Container(id="add-agent-container"):
            yield Static("Name:", id="name-label")
            yield Input(placeholder="Agent name...", id="name-input")
            yield Static("Instructions:", id="instructions-label")
            yield Input(placeholder="Agent instructions...", id="instructions-input")
            yield Static("Description (optional):", id="desc-label")
            yield Input(placeholder="Description...", id="desc-input")
            yield Button("✅ Create Agent", id="create-btn", variant="success")
        
        yield Footer()
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "create-btn":
            self._create_agent()
    
    def _create_agent(self) -> None:
        """Create the agent"""
        name_input = self.query_one("#name-input", Input)
        instructions_input = self.query_one("#instructions-input", Input)
        
        if name_input.value and instructions_input.value:
            self.notify(f"Agent '{name_input.value}' created!")
            self.app.pop_screen()
        else:
            self.notify("Please fill in name and instructions", severity="error")
    
    def action_go_back(self) -> None:
        self.app.pop_screen()


class ChatScreen(Screen):
    """Screen for chatting with an agent"""
    
    BINDINGS = [
        Binding("escape", "go_back", "Back"),
    ]
    
    def compose(self) -> ComposeResult:
        yield Header()
        yield Static("💬 Chat", id="screen-title")
        
        with Container(id="chat-container"):
            with ScrollableContainer(id="chat-history"):
                yield Static("[bold]Agent:[/] Hello! How can I help you today?")
            yield Input(placeholder="Type your message...", id="chat-input")
            yield Button("📤 Send", id="send-btn", variant="primary")
        
        yield Footer()
    
    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "send-btn":
            self._send_message()
    
    def on_input_submitted(self, event: Input.Submitted) -> None:
        if event.input.id == "chat-input":
            self._send_message()
    
    def _send_message(self) -> None:
        """Send a chat message"""
        chat_input = self.query_one("#chat-input", Input)
        chat_history = self.query_one("#chat-history", ScrollableContainer)
        
        if chat_input.value:
            # Add user message
            chat_history.mount(Static(f"[bold]You:[/] {chat_input.value}"))
            
            # Simulate agent response
            chat_history.mount(Static(f"[bold]Agent:[/] I received: '{chat_input.value}'"))
            
            chat_input.value = ""
            chat_history.scroll_end()
    
    def action_go_back(self) -> None:
        self.app.pop_screen()


class SentinelApp(App):
    """Sentinel AI TUI Application"""
    
    CSS = """
    Screen {
        background: $surface;
    }
    
    #screen-title {
        height: 3;
        content-align: center middle;
        background: $primary;
        color: $text;
        text-style: bold;
    }
    
    Container {
        height: 1fr;
    }
    
    #agents-container, #tasks-container, #execute-container, 
    #create-task-container, #add-agent-container, #chat-container {
        margin: 1 2;
        border: solid $primary;
        padding: 1;
    }
    
    #agents-list, #tasks-list, #chat-history {
        height: 1fr;
        border: solid $secondary;
        padding: 1;
    }
    
    .agent-row, .task-row {
        padding: 1;
        margin: 1 0;
        background: $surface;
        border: solid $secondary;
    }
    
    .agent-row:hover, .task-row:hover {
        background: $primary;
    }
    
    #agents-actions, #tasks-actions {
        height: 3;
        align: center middle;
    }
    
    #agents-actions Button, #tasks-actions Button {
        margin: 0 1;
    }
    
    #selected-agent, #result-display {
        padding: 1;
        margin: 1 0;
    }
    
    #task-input, #agent-input, #name-input, #instructions-input, 
    #desc-input, #chat-input {
        margin: 1 0;
        width: 100%;
    }
    
    #priority-buttons {
        height: 3;
        align: center middle;
    }
    
    #priority-buttons Button {
        margin: 0 1;
    }
    
    .priority-display {
        padding: 1;
        margin: 1 0;
    }
    
    #execute-btn, #create-btn {
        margin: 1 0;
    }
    """
    
    BINDINGS = [
        Binding("1", "show_agents", "Agents"),
        Binding("2", "show_tasks", "Tasks"),
        Binding("3", "show_execute", "Execute"),
        Binding("4", "show_chat", "Chat"),
        Binding("q", "quit", "Quit"),
    ]
    
    def __init__(self):
        super().__init__()
        self.selected_agent = None
    
    def on_mount(self) -> None:
        self.push_screen("agents_screen")
    
    def action_show_agents(self) -> None:
        self.push_screen("agents_screen")
    
    def action_show_tasks(self) -> None:
        self.push_screen("tasks_screen")
    
    def action_show_execute(self) -> None:
        self.push_screen("execute_screen")
    
    def action_show_chat(self) -> None:
        self.push_screen("chat_screen")
    
    def action_quit(self) -> None:
        self.exit()


def main():
    app = SentinelApp()
    
    # Register screens
    app.register_screen("agents_screen", AgentsScreen)
    app.register_screen("tasks_screen", TasksScreen)
    app.register_screen("execute_screen", ExecuteScreen)
    app.register_screen("chat_screen", ChatScreen)
    app.register_screen("create_task_screen", CreateTaskScreen)
    app.register_screen("add_agent_screen", AddAgentScreen)
    
    app.run()


if __name__ == "__main__":
    main()
