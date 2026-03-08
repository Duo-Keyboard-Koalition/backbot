package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"darci-go/darci/config"
)

// TaskStore provides JSON file-backed task storage
type TaskStore struct {
	mu          sync.RWMutex
	tasksPath   string
	contextPath string
	counterPath string
	stateDir    string
}

// NewTaskStore creates a new TaskStore with the given config
func NewTaskStore(cfg *config.DarciConfig) (*TaskStore, error) {
	stateDir := cfg.StateDir
	if stateDir == "" {
		stateDir = ".darci-state"
	}

	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create state directory: %w", err)
	}

	store := &TaskStore{
		stateDir:    stateDir,
		tasksPath:   filepath.Join(stateDir, "tasks.json"),
		contextPath: filepath.Join(stateDir, "context.json"),
		counterPath: filepath.Join(stateDir, "counter.txt"),
	}

	if err := store.ensureDefaults(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *TaskStore) ensureDefaults() error {
	// Initialize tasks file
	if _, err := os.Stat(s.tasksPath); os.IsNotExist(err) {
		if err := os.WriteFile(s.tasksPath, []byte("{}"), 0644); err != nil {
			return err
		}
	}

	// Initialize context file
	if _, err := os.Stat(s.contextPath); os.IsNotExist(err) {
		ctx := AgentContext{
			AgentAssignments: make(map[string]*AgentAssignment),
			DarciState: &DarciState{
				ActiveMonitors: []string{},
			},
		}
		data, err := json.MarshalIndent(ctx, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(s.contextPath, data, 0644); err != nil {
			return err
		}
	}

	// Initialize counter file
	if _, err := os.Stat(s.counterPath); os.IsNotExist(err) {
		if err := os.WriteFile(s.counterPath, []byte("0"), 0644); err != nil {
			return err
		}
	}

	return nil
}

func (s *TaskStore) nextID() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.counterPath)
	if err != nil {
		return "T001"
	}

	n, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		n = 0
	}
	n++

	if err := os.WriteFile(s.counterPath, []byte(strconv.Itoa(n)), 0644); err != nil {
		return "T001"
	}

	return fmt.Sprintf("T%03d", n)
}

// readTasks reads all tasks from the JSON file
func (s *TaskStore) readTasks() (map[string]*Task, error) {
	data, err := os.ReadFile(s.tasksPath)
	if err != nil {
		return nil, err
	}

	var tasks map[string]*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = make(map[string]*Task)
	}

	return tasks, nil
}

func (s *TaskStore) writeTasks(tasks map[string]*Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.tasksPath, data, 0644)
}

// Create stores a new task
func (s *TaskStore) Create(task *Task) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.readTasks()
	if err != nil {
		return nil, err
	}

	if task.CreatedAt.IsZero() {
		task.CreatedAt = time.Now().UTC()
	}
	if task.UpdatedAt.IsZero() {
		task.UpdatedAt = time.Now().UTC()
	}

	tasks[task.ID] = task
	if err := s.writeTasks(tasks); err != nil {
		return nil, err
	}

	return task, nil
}

// CreateNew creates a new task with the given parameters
func (s *TaskStore) CreateNew(title, description, priority string, labels, dependencies []string) (*Task, error) {
	if priority == "" {
		priority = "P2"
	}

	task := &Task{
		ID:           s.nextID(),
		Title:        title,
		Description:  description,
		Priority:     priority,
		Labels:       labels,
		Dependencies: dependencies,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Status:       "pending",
	}

	return s.Create(task)
}

// Update updates fields on an existing task
func (s *TaskStore) Update(taskID string, fields map[string]interface{}) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tasks, err := s.readTasks()
	if err != nil {
		return nil, err
	}

	task, exists := tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	// Handle nested updates for darci and sentinel_snapshot
	for key, val := range fields {
		switch key {
		case "darci":
			if v, ok := val.(DarciRoles); ok {
				task.Darci = v
			}
		case "sentinel_snapshot":
			if v, ok := val.(SentinelSnapshot); ok {
				task.SentinelSnapshot = v
			}
		case "status":
			if v, ok := val.(string); ok {
				task.Status = v
			}
		}
	}

	task.UpdatedAt = time.Now().UTC()
	tasks[taskID] = task

	if err := s.writeTasks(tasks); err != nil {
		return nil, err
	}

	return task, nil
}

// Get retrieves a task by ID
func (s *TaskStore) Get(taskID string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.readTasks()
	if err != nil {
		return nil, err
	}

	task, exists := tasks[taskID]
	if !exists {
		return nil, nil
	}

	return task, nil
}

// Query returns tasks matching the given criteria
func (s *TaskStore) Query(status, priority, label string) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, err := s.readTasks()
	if err != nil {
		return nil, err
	}

	var results []*Task
	for _, task := range tasks {
		if status != "" && task.Status != status {
			continue
		}
		if priority != "" && task.Priority != priority {
			continue
		}
		if label != "" {
			found := false
			for _, l := range task.Labels {
				if l == label {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		results = append(results, task)
	}

	return results, nil
}

// All returns all tasks
func (s *TaskStore) All() ([]*Task, error) {
	return s.Query("", "", "")
}

// GetContext retrieves the agent context
func (s *TaskStore) GetContext() (*AgentContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.contextPath)
	if err != nil {
		return nil, err
	}

	var ctx AgentContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

// UpdateContext updates the agent context
func (s *TaskStore) UpdateContext(updates map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, err := s.GetContext()
	if err != nil {
		return err
	}

	// Apply updates (simplified - would need reflection for full implementation)
	_ = updates

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.contextPath, data, 0644)
}

// SetAgentAssignment sets or updates an agent's assignment
func (s *TaskStore) SetAgentAssignment(nodeName, taskID, role string, riskScore float64, status string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ctx, err := s.GetContext()
	if err != nil {
		return err
	}

	if ctx.AgentAssignments == nil {
		ctx.AgentAssignments = make(map[string]*AgentAssignment)
	}

	if status == "" {
		status = "pending"
	}

	ctx.AgentAssignments[nodeName] = &AgentAssignment{
		TaskID:    taskID,
		DARCIROle: role,
		RiskScore: riskScore,
		Status:    status,
	}

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.contextPath, data, 0644)
}
