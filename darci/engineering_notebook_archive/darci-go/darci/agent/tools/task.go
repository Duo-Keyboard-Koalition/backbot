package tools

import (
	"context"
	"fmt"
	"strings"

	"darci-go/darci/state"
)

// TaskCreateTool creates a new tracked task in DarCI
type TaskCreateTool struct {
	store *state.TaskStore
}

func NewTaskCreateTool(store *state.TaskStore) *TaskCreateTool {
	return &TaskCreateTool{store: store}
}

func (t *TaskCreateTool) Name() string {
	return "task_create"
}

func (t *TaskCreateTool) Description() string {
	return "Create a new tracked task in DarCI. Returns the task ID and summary. Use this before assigning or monitoring any work."
}

func (t *TaskCreateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"title": map[string]interface{}{
				"type":        "string",
				"description": "Short task title",
			},
			"description": map[string]interface{}{
				"type":        "string",
				"description": "Detailed description",
			},
			"priority": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"P0", "P1", "P2", "P3"},
				"description": "P0=critical, P1=high, P2=normal, P3=low",
			},
			"labels": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Optional tags (e.g. ['sentinel', 'tailbridge'])",
			},
			"dependencies": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "Task IDs this task depends on",
			},
		},
		"required": []string{"title"},
	}
}

func (t *TaskCreateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	title, _ := args["title"].(string)
	description, _ := args["description"].(string)
	priority, _ := args["priority"].(string)
	if priority == "" {
		priority = "P2"
	}

	var labels []string
	if l, ok := args["labels"].([]interface{}); ok {
		for _, v := range l {
			if s, ok := v.(string); ok {
				labels = append(labels, s)
			}
		}
	}

	var dependencies []string
	if d, ok := args["dependencies"].([]interface{}); ok {
		for _, v := range d {
			if s, ok := v.(string); ok {
				dependencies = append(dependencies, s)
			}
		}
	}

	task, err := t.store.CreateNew(title, description, priority, labels, dependencies)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	return fmt.Sprintf("Task created: %s\nTitle: %s\nPriority: %s | Status: %s\nLabels: %s",
		task.ID, task.Title, task.Priority, task.Status, strings.Join(task.Labels, ", ")), nil
}

// TaskUpdateTool updates a task's status, priority, or description
type TaskUpdateTool struct {
	store *state.TaskStore
}

func NewTaskUpdateTool(store *state.TaskStore) *TaskUpdateTool {
	return &TaskUpdateTool{store: store}
}

func (t *TaskUpdateTool) Name() string {
	return "task_update"
}

func (t *TaskUpdateTool) Description() string {
	return "Update a task's status, priority, or description."
}

func (t *TaskUpdateTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "Task ID (e.g. T001)",
			},
			"status": map[string]interface{}{
				"type": "string",
				"enum": []string{"pending", "in_progress", "at_risk", "blocked", "completed"},
			},
			"priority": map[string]interface{}{
				"type": "string",
				"enum": []string{"P0", "P1", "P2", "P3"},
			},
			"description": map[string]interface{}{
				"type": "string",
			},
		},
		"required": []string{"task_id"},
	}
}

func (t *TaskUpdateTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	taskID, _ := args["task_id"].(string)
	if taskID == "" {
		return "", fmt.Errorf("task_id is required")
	}

	updates := make(map[string]interface{})
	if status, ok := args["status"].(string); ok {
		updates["status"] = status
	}
	if priority, ok := args["priority"].(string); ok {
		updates["priority"] = priority
	}
	if description, ok := args["description"].(string); ok {
		updates["description"] = description
	}

	if len(updates) == 0 {
		return fmt.Sprintf("Error: no fields to update for %s", taskID), nil
	}

	task, err := t.store.Update(taskID, updates)
	if err != nil || task == nil {
		return fmt.Sprintf("Error: task %s not found", taskID), nil
	}

	var changes []string
	for k, v := range updates {
		changes = append(changes, fmt.Sprintf("%s=%v", k, v))
	}

	return fmt.Sprintf("Updated %s: %s", taskID, strings.Join(changes, ", ")), nil
}

// TaskQueryTool queries tasks by status, priority, or label
type TaskQueryTool struct {
	store *state.TaskStore
}

func NewTaskQueryTool(store *state.TaskStore) *TaskQueryTool {
	return &TaskQueryTool{store: store}
}

func (t *TaskQueryTool) Name() string {
	return "task_query"
}

func (t *TaskQueryTool) Description() string {
	return "Query tasks by status, priority, or label. Returns a markdown table."
}

func (t *TaskQueryTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"status": map[string]interface{}{
				"type": "string",
				"enum": []string{"pending", "in_progress", "at_risk", "blocked", "completed"},
			},
			"priority": map[string]interface{}{
				"type": "string",
				"enum": []string{"P0", "P1", "P2", "P3"},
			},
			"label": map[string]interface{}{
				"type":        "string",
				"description": "Filter by label",
			},
		},
	}
}

func (t *TaskQueryTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	status, _ := args["status"].(string)
	priority, _ := args["priority"].(string)
	label, _ := args["label"].(string)

	tasks, err := t.store.Query(status, priority, label)
	if err != nil || len(tasks) == 0 {
		return "No tasks found matching the query.", nil
	}

	var lines []string
	lines = append(lines, "| ID | Title | Priority | Status | Responsible |")
	lines = append(lines, "|---|---|---|---|---|")

	for _, task := range tasks {
		responsible := "unassigned"
		if task.Darci.Responsible != "" {
			responsible = task.Darci.Responsible
		}
		lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s |",
			task.ID, task.Title, task.Priority, task.Status, responsible))
	}

	return strings.Join(lines, "\n"), nil
}

// StatusReportTool generates a full DarCI project status board
type StatusReportTool struct {
	store *state.TaskStore
}

func NewStatusReportTool(store *state.TaskStore) *StatusReportTool {
	return &StatusReportTool{store: store}
}

func (t *StatusReportTool) Name() string {
	return "status_report"
}

func (t *StatusReportTool) Description() string {
	return "Generate a full DarCI project status board grouped by status."
}

func (t *StatusReportTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	}
}

func (t *StatusReportTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	tasks, err := t.store.All()
	if err != nil || len(tasks) == 0 {
		return "No tasks yet. Use task_create to get started.", nil
	}

	groups := map[string][]*state.Task{
		"at_risk":     {},
		"blocked":     {},
		"in_progress": {},
		"pending":     {},
		"completed":   {},
	}

	for _, task := range tasks {
		groups[task.Status] = append(groups[task.Status], task)
	}

	var lines []string
	lines = append(lines, "# DarCI Project Status\n")

	total := len(tasks)
	done := len(groups["completed"])
	lines = append(lines, fmt.Sprintf("**Progress:** %d/%d tasks complete\n", done, total))

	icons := map[string]string{
		"at_risk":     "⚠️",
		"blocked":     "🛑",
		"in_progress": "🔄",
		"pending":     "⏳",
		"completed":   "✅",
	}

	order := []string{"at_risk", "blocked", "in_progress", "pending", "completed"}

	for _, status := range order {
		group := groups[status]
		if len(group) == 0 {
			continue
		}

		icon := icons[status]
		lines = append(lines, fmt.Sprintf("\n## %s %s (%d)\n", icon, strings.Title(strings.ReplaceAll(status, "_", " ")), len(group)))
		lines = append(lines, "| ID | Title | Priority | Responsible | Risk |")
		lines = append(lines, "|---|---|---|---|---|")

		for _, task := range group {
			responsible := "unassigned"
			if task.Darci.Responsible != "" {
				responsible = task.Darci.Responsible
			}
			risk := "-"
			if task.SentinelSnapshot.RiskScore > 0 {
				risk = fmt.Sprintf("%.2f", task.SentinelSnapshot.RiskScore)
			}
			lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s |",
				task.ID, task.Title, task.Priority, status, responsible, risk))
		}
	}

	return strings.Join(lines, "\n"), nil
}

// AssignTaskTool assigns a task to a Responsible agent
type AssignTaskTool struct {
	store    *state.TaskStore
	sendTool *SendDarciMessageTool
}

func NewAssignTaskTool(store *state.TaskStore, sendTool *SendDarciMessageTool) *AssignTaskTool {
	return &AssignTaskTool{
		store:    store,
		sendTool: sendTool,
	}
}

func (t *AssignTaskTool) Name() string {
	return "assign_task"
}

func (t *AssignTaskTool) Description() string {
	return "Assign a task to a Responsible agent on the tailnet. Sets darci.responsible and sends a darci_directive via tailbridge."
}

func (t *AssignTaskTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"task_id": map[string]interface{}{
				"type":        "string",
				"description": "Task ID to assign (e.g. T001)",
			},
			"node_name": map[string]interface{}{
				"type":        "string",
				"description": "Tailnet node name of the Responsible agent",
			},
			"goal_description": map[string]interface{}{
				"type":        "string",
				"description": "What the agent should do",
			},
		},
		"required": []string{"task_id", "node_name", "goal_description"},
	}
}

func (t *AssignTaskTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	taskID, _ := args["task_id"].(string)
	nodeName, _ := args["node_name"].(string)
	goalDesc, _ := args["goal_description"].(string)

	if taskID == "" || nodeName == "" || goalDesc == "" {
		return "", fmt.Errorf("task_id, node_name, and goal_description are required")
	}

	task, err := t.store.Get(taskID)
	if err != nil || task == nil {
		return fmt.Sprintf("Error: task %s not found", taskID), nil
	}

	t.store.Update(taskID, map[string]interface{}{
		"status": "in_progress",
	})
	t.store.SetAgentAssignment(nodeName, taskID, "responsible", 0.0, "in_progress")

	directiveResult, err := t.sendTool.Execute(ctx, map[string]interface{}{
		"dest_node":    nodeName,
		"message_type": "darci_directive",
		"payload": map[string]interface{}{
			"task_id":    taskID,
			"task_title": task.Title,
			"goal":       goalDesc,
			"priority":   task.Priority,
		},
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Task %s assigned to %s.\nStatus: in_progress | Responsible: %s\nDirective sent: %s",
		taskID, nodeName, nodeName, directiveResult), nil
}
