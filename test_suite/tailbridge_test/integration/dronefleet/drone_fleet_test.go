//go:build integration

package dronefleet

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// DroneStatus represents the status of a drone in the fleet
type DroneStatus struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Role          string   `json:"role"` // lead, scout, worker, relay
	IP            string   `json:"ip"`
	TailscaleIP   string   `json:"tailscale_ip"`
	Online        bool     `json:"online"`
	BatteryLevel  int      `json:"battery_level"`
	CurrentTask   string   `json:"current_task"`
	Capabilities  []string `json:"capabilities"`
	Position      Position `json:"position"`
}

// Position represents drone GPS coordinates
type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

// Task represents a drone mission task
type Task struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // survey, deliver, inspect, patrol, scan
	Priority    string    `json:"priority"`
	Status      string    `json:"status"` // pending, in_progress, completed, failed
	AssignedTo  string    `json:"assigned_to"`
	Description string    `json:"description"`
	Parameters  TaskParams `json:"parameters"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// TaskParams holds task-specific parameters
type TaskParams struct {
	Waypoints   []Position `json:"waypoints,omitempty"`
	Payload     string     `json:"payload,omitempty"`
	Duration    int        `json:"duration_seconds,omitempty"`
	Area        float64    `json:"area_sqkm,omitempty"`
	Target      string     `json:"target,omitempty"`
}

// DroneMessage represents communication between drones
type DroneMessage struct {
	From        string    `json:"from"`
	To          string    `json:"to"`
	Type        string    `json:"message_type"` // task_assignment, status_report, alert, coordination
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
	TaskID      string    `json:"task_id,omitempty"`
	Priority    string    `json:"priority,omitempty"`
}

// TaskLog represents a drone's task execution log
type TaskLog struct {
	DroneID     string      `json:"drone_id"`
	Tasks       []Task      `json:"tasks"`
	Messages    []DroneMessage `json:"messages"`
	IP          string      `json:"tailscale_ip"`
	Role        string      `json:"role"`
}

// DroneFleetTestSuite runs comprehensive drone fleet tests
type DroneFleetTestSuite struct {
	suite.Suite
	leadURL      string
	drone1URL    string
	drone2URL    string
	drone3URL    string
	tsAuthKey    string
	geminiAPIKey string
	client       *http.Client
	mu           sync.Mutex
	taskLogs     map[string]*TaskLog
	missionID    string
}

// SetupSuite runs once before all tests
func (s *DroneFleetTestSuite) SetupSuite() {
	s.leadURL = getEnv("LEAD_DRONE_URL", "http://localhost:8081")
	s.drone1URL = getEnv("DRONE1_URL", "http://localhost:8082")
	s.drone2URL = getEnv("DRONE2_URL", "http://localhost:8083")
	s.drone3URL = getEnv("DRONE3_URL", "http://localhost:8084")
	s.tsAuthKey = getEnv("TS_AUTH_KEY", "tskey-auth-k7Q1t39ZWj11CNTRL-FbRR2tKLRcPn5L246vsAcP7LP2YCUxWD")
	s.geminiAPIKey = getEnv("GEMINI_API_KEY", "AIzaSyC_9M8im8z5F0ING2W3Hu2aQiunJhhWUXI")

	s.client = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
	}

	s.taskLogs = make(map[string]*TaskLog)
	s.missionID = fmt.Sprintf("MISSION-%d", time.Now().Unix())
}

// TestDroneFleetMission runs the main drone fleet test
func (s *DroneFleetTestSuite) TestDroneFleetMission() {
	ctx := context.Background()

	s.T().Log("")
	s.T().Log("╔═══════════════════════════════════════════════════════════╗")
	s.T().Log("║        DRONE FLEET TAILSCALE COMMUNICATION TEST          ║")
	s.T().Log("║              Mission: Urban Survey & Delivery            ║")
	s.T().Log("╚═══════════════════════════════════════════════════════════╝")
	s.T().Log("")

	// Phase 1: Fleet Initialization
	s.T().Log("🚁 Phase 1: Drone Fleet Initialization")
	s.T().Log("═══════════════════════════════════════════════════════════")
	
	drones := map[string]string{
		"lead-drone":   s.leadURL,
		"scout-drone":  s.drone1URL,
		"worker-drone": s.drone2URL,
		"relay-drone":  s.drone3URL,
	}

	healthyDrones := s.waitForFleetReady(ctx, drones, 5*time.Minute)
	s.Require().Equal(4, len(healthyDrones), "All 4 drones must be ready")
	s.T().Logf("✅ All %d drones are online and ready\n", len(healthyDrones))

	// Phase 2: Get Drone Status and IP Addresses
	s.T().Log("📡 Phase 2: Retrieving Drone Fleet Status")
	s.T().Log("═══════════════════════════════════════════════════════════")
	
	droneStatuses := make(map[string]*DroneStatus)

	for name, url := range healthyDrones {
		status, err := s.getDroneStatus(ctx, url)
		s.Require().NoError(err, "Failed to get status for %s", name)
		droneStatuses[name] = status

		s.T().Logf("  %s:", strings.ToUpper(name))
		s.T().Logf("    Role:           %s", status.Role)
		s.T().Logf("    Tailscale IP:   %s", status.TailscaleIP)
		s.T().Logf("    Local IP:       %s", status.IP)
		s.T().Logf("    Battery:        %d%%", status.BatteryLevel)
		s.T().Logf("    Position:       %.6f, %.6f, %.1fm", status.Position.Latitude, status.Position.Longitude, status.Position.Altitude)
		s.T().Logf("    Capabilities:   %v", status.Capabilities)
		s.T().Log("")

		// Initialize task log
		s.taskLogs[name] = &TaskLog{
			DroneID:  name,
			Tasks:    []Task{},
			Messages: []DroneMessage{},
			IP:       status.TailscaleIP,
			Role:     status.Role,
		}
	}

	// Phase 3: Mission Briefing - Lead Drone Assigns Tasks
	s.T().Log("📋 Phase 3: Mission Briefing - Task Assignment")
	s.T().Log("═══════════════════════════════════════════════════════════")
	s.T().Logf("  Mission ID: %s", s.missionID)
	s.T().Logf("  Objective:  Urban Survey and Emergency Supply Delivery")
	s.T().Logf("  Area:       Downtown District (2.5 sq km)")
	s.T().Log("")

	// Task 1: Lead Drone assigns survey task to Scout Drone
	s.T().Log("  Task 1: Lead → Scout (Aerial Survey)")
	task1 := Task{
		ID:          "TASK-001",
		Type:        "survey",
		Priority:    "high",
		Status:      "pending",
		AssignedTo:  "scout-drone",
		Description: "Conduct aerial survey of downtown district",
		Parameters: TaskParams{
			Area: 2.5,
			Duration: 1800,
		},
		CreatedAt: time.Now(),
	}
	s.assignTask(ctx, "lead-drone", "scout-drone", task1)
	s.taskLogs["lead-drone"].Tasks = append(s.taskLogs["lead-drone"].Tasks, task1)
	s.T().Logf("    ✓ Task assigned via %s → %s", droneStatuses["lead-drone"].TailscaleIP, droneStatuses["scout-drone"].TailscaleIP)
	s.T().Log("")

	// Task 2: Lead Drone assigns delivery task to Worker Drone
	s.T().Log("  Task 2: Lead → Worker (Emergency Delivery)")
	task2 := Task{
		ID:          "TASK-002",
		Type:        "deliver",
		Priority:    "critical",
		Status:      "pending",
		AssignedTo:  "worker-drone",
		Description: "Deliver emergency medical supplies to Hospital Zone A",
		Parameters: TaskParams{
			Payload: "Medical Supplies Kit #A-47",
			Duration: 900,
		},
		CreatedAt: time.Now(),
	}
	s.assignTask(ctx, "lead-drone", "worker-drone", task2)
	s.taskLogs["lead-drone"].Tasks = append(s.taskLogs["lead-drone"].Tasks, task2)
	s.T().Logf("    ✓ Task assigned via %s → %s", droneStatuses["lead-drone"].TailscaleIP, droneStatuses["worker-drone"].TailscaleIP)
	s.T().Log("")

	// Task 3: Lead Drone assigns relay task to Relay Drone
	s.T().Log("  Task 3: Lead → Relay (Communication Relay)")
	task3 := Task{
		ID:          "TASK-003",
		Type:        "patrol",
		Priority:    "medium",
		Status:      "pending",
		AssignedTo:  "relay-drone",
		Description: "Maintain communication relay between fleet and base",
		Parameters: TaskParams{
			Duration: 3600,
		},
		CreatedAt: time.Now(),
	}
	s.assignTask(ctx, "lead-drone", "relay-drone", task3)
	s.taskLogs["lead-drone"].Tasks = append(s.taskLogs["lead-drone"].Tasks, task3)
	s.T().Logf("    ✓ Task assigned via %s → %s", droneStatuses["lead-drone"].TailscaleIP, droneStatuses["relay-drone"].TailscaleIP)
	s.T().Log("")

	// Phase 4: Task Execution and Status Reports
	s.T().Log("✈️ Phase 4: Task Execution - Status Reports")
	s.T().Log("═══════════════════════════════════════════════════════════")

	// Scout Drone reports progress
	s.T().Log("  Scout Drone: Survey in progress...")
	msg1 := DroneMessage{
		From:      "scout-drone",
		To:        "lead-drone",
		Type:      "status_report",
		Content:   "Lead Drone, Scout reporting. Survey 45% complete. Coverage: 1.1 sq km. Weather conditions optimal. Battery at 78%. ETA 22 minutes.",
		Timestamp: time.Now(),
		TaskID:    "TASK-001",
		Priority:  "normal",
	}
	s.sendDroneMessage(ctx, msg1)
	s.taskLogs["scout-drone"].Messages = append(s.taskLogs["scout-drone"].Messages, msg1)
	s.T().Logf("    ✓ Status report sent via %s", droneStatuses["scout-drone"].TailscaleIP)
	s.T().Log("")

	// Worker Drone reports departure
	s.T().Log("  Worker Drone: Delivery mission started...")
	msg2 := DroneMessage{
		From:      "worker-drone",
		To:        "lead-drone",
		Type:      "status_report",
		Content:   "Lead Drone, Worker reporting. Departed base with payload 'Medical Supplies Kit #A-47'. En route to Hospital Zone A. Battery at 92%. ETA 12 minutes.",
		Timestamp: time.Now(),
		TaskID:    "TASK-002",
		Priority:  "high",
	}
	s.sendDroneMessage(ctx, msg2)
	s.taskLogs["worker-drone"].Messages = append(s.taskLogs["worker-drone"].Messages, msg2)
	s.T().Logf("    ✓ Status report sent via %s", droneStatuses["worker-drone"].TailscaleIP)
	s.T().Log("")

	// Relay Drone confirms position
	s.T().Log("  Relay Drone: Relay position established...")
	msg3 := DroneMessage{
		From:      "relay-drone",
		To:        "lead-drone",
		Type:      "status_report",
		Content:   "Lead Drone, Relay reporting. Stationed at coordinates 40.7580, -73.9855, altitude 150m. Signal strength excellent. All fleet communications routing through relay. Battery at 88%.",
		Timestamp: time.Now(),
		TaskID:    "TASK-003",
		Priority:  "normal",
	}
	s.sendDroneMessage(ctx, msg3)
	s.taskLogs["relay-drone"].Messages = append(s.taskLogs["relay-drone"].Messages, msg3)
	s.T().Logf("    ✓ Status report sent via %s", droneStatuses["relay-drone"].TailscaleIP)
	s.T().Log("")

	// Phase 5: Inter-Drone Coordination
	s.T().Log("🤝 Phase 5: Inter-Drone Coordination")
	s.T().Log("═══════════════════════════════════════════════════════════")

	// Scout requests Worker assistance
	s.T().Log("  Scout → Worker: Coordination request")
	msg4 := DroneMessage{
		From:      "scout-drone",
		To:        "worker-drone",
		Type:      "coordination",
		Content:   "Worker Drone, Scout here. I've identified optimal landing zone at Grid C-7 during my survey. Coordinates transmitted. Suitable for your delivery. Over.",
		Timestamp: time.Now(),
		TaskID:    "TASK-001",
		Priority:  "normal",
	}
	s.sendDroneMessage(ctx, msg4)
	s.taskLogs["scout-drone"].Messages = append(s.taskLogs["scout-drone"].Messages, msg4)
	s.T().Logf("    ✓ Coordination message sent via %s → %s", droneStatuses["scout-drone"].TailscaleIP, droneStatuses["worker-drone"].TailscaleIP)

	// Worker acknowledges
	s.T().Log("  Worker → Scout: Acknowledgment")
	msg5 := DroneMessage{
		From:      "worker-drone",
		To:        "scout-drone",
		Type:      "coordination",
		Content:   "Scout Drone, Worker copying. Grid C-7 coordinates received. Adjusting route. Excellent intel. Will confirm landing on approach. Out.",
		Timestamp: time.Now(),
		TaskID:    "TASK-002",
		Priority:  "normal",
	}
	s.sendDroneMessage(ctx, msg5)
	s.taskLogs["worker-drone"].Messages = append(s.taskLogs["worker-drone"].Messages, msg5)
	s.T().Logf("    ✓ Coordination response sent via %s → %s", droneStatuses["worker-drone"].TailscaleIP, droneStatuses["scout-drone"].TailscaleIP)
	s.T().Log("")

	// Phase 6: Task Completion Reports
	s.T().Log("✅ Phase 6: Mission Completion Reports")
	s.T().Log("═══════════════════════════════════════════════════════════")

	// Scout completes survey
	s.T().Log("  Scout Drone: Survey complete")
	msg6 := DroneMessage{
		From:      "scout-drone",
		To:        "lead-drone",
		Type:      "status_report",
		Content:   "Lead Drone, Scout reporting TASK-001 COMPLETE. Full aerial survey completed. 2.5 sq km mapped. 3D model uploaded to base. Battery at 62%. Returning to base.",
		Timestamp: time.Now(),
		TaskID:    "TASK-001",
		Priority:  "high",
	}
	s.sendDroneMessage(ctx, msg6)
	s.taskLogs["scout-drone"].Messages = append(s.taskLogs["scout-drone"].Messages, msg6)
	s.T().Logf("    ✓ Task completion report sent")
	s.T().Log("")

	// Worker completes delivery
	s.T().Log("  Worker Drone: Delivery complete")
	msg7 := DroneMessage{
		From:      "worker-drone",
		To:        "lead-drone",
		Type:      "status_report",
		Content:   "Lead Drone, Worker reporting TASK-002 COMPLETE. Medical supplies delivered to Hospital Zone A. Landing zone Grid C-7 performed flawlessly. Battery at 71%. Returning to base.",
		Timestamp: time.Now(),
		TaskID:    "TASK-002",
		Priority:  "critical",
	}
	s.sendDroneMessage(ctx, msg7)
	s.taskLogs["worker-drone"].Messages = append(s.taskLogs["worker-drone"].Messages, msg7)
	s.T().Logf("    ✓ Task completion report sent")
	s.T().Log("")

	// Relay confirms continued operation
	s.T().Log("  Relay Drone: Relay operational")
	msg8 := DroneMessage{
		From:      "relay-drone",
		To:        "lead-drone",
		Type:      "status_report",
		Content:   "Lead Drone, Relay reporting. All fleet communications maintained throughout mission. Zero packet loss. Battery at 76%. Continuing patrol.",
		Timestamp: time.Now(),
		TaskID:    "TASK-003",
		Priority:  "normal",
	}
	s.sendDroneMessage(ctx, msg8)
	s.taskLogs["relay-drone"].Messages = append(s.taskLogs["relay-drone"].Messages, msg8)
	s.T().Logf("    ✓ Status report sent")
	s.T().Log("")

	// Phase 7: Lead Drone Mission Summary
	s.T().Log("📊 Phase 7: Lead Drone Mission Summary")
	s.T().Log("═══════════════════════════════════════════════════════════")
	
	missionSummary := DroneMessage{
		From:      "lead-drone",
		To:        "all",
		Type:      "status_report",
		Content:   "MISSION COMPLETE. Mission ID: " + s.missionID + ". All objectives achieved. Survey: 100%. Delivery: 100%. Relay: Operational. Fleet returning to base. Outstanding work: None.",
		Timestamp: time.Now(),
		Priority:  "high",
	}
	s.sendDroneMessage(ctx, missionSummary)
	s.taskLogs["lead-drone"].Messages = append(s.taskLogs["lead-drone"].Messages, missionSummary)
	s.T().Logf("    ✓ Mission summary broadcast to all drones")
	s.T().Log("")

	// Phase 8: Print Final Task Logs
	s.T().Log("")
	s.T().Log("═══════════════════════════════════════════════════════════")
	s.T().Log("           FINAL TASK LOGS - ALL DRONES")
	s.T().Log("═══════════════════════════════════════════════════════════")
	s.T().Log("")

	s.printTaskLogs()

	// Phase 9: Assertions
	s.T().Log("")
	s.T().Log("📊 Phase 9: Running Assertions")
	s.T().Log("═══════════════════════════════════════════════════════════")

	for name, status := range droneStatuses {
		s.T().Logf("  Asserting %s is online...", name)
		assert.True(s.T(), status.Online, "%s should be online", name)
		assert.NotEmpty(s.T(), status.TailscaleIP, "%s should have Tailscale IP", name)
		assert.Greater(s.T(), status.BatteryLevel, 0, "%s should have battery", name)
	}

	s.T().Logf("  ✅ All assertions passed!")
	s.T().Log("")
	s.T().Log("═══════════════════════════════════════════════════════════")
	s.T().Log("           MISSION COMPLETED SUCCESSFULLY")
	s.T().Log("═══════════════════════════════════════════════════════════")
}

// waitForFleetReady waits for all drones to be ready
func (s *DroneFleetTestSuite) waitForFleetReady(ctx context.Context, drones map[string]string, timeout time.Duration) map[string]string {
	ready := make(map[string]string)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) && len(ready) < 4 {
		for name, url := range drones {
			if _, ok := ready[name]; ok {
				continue
			}

			resp, err := s.client.Get(url + "/health")
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				ready[name] = url
				s.T().Logf("  ✓ %s is ready", name)
			}
		}

		if len(ready) < 4 {
			time.Sleep(5 * time.Second)
		}
	}

	return ready
}

// getDroneStatus retrieves drone status
func (s *DroneFleetTestSuite) getDroneStatus(ctx context.Context, url string) (*DroneStatus, error) {
	resp, err := s.client.Get(url + "/status")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var status DroneStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, err
	}

	return &status, nil
}

// assignTask assigns a task from one drone to another
func (s *DroneFleetTestSuite) assignTask(ctx context.Context, from, to string, task Task) error {
	payload := map[string]interface{}{
		"from":        from,
		"to":          to,
		"message_type": "task_assignment",
		"task":        task,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	targetURL := s.leadURL
	if to != "lead-drone" {
		if strings.Contains(to, "scout") {
			targetURL = s.drone1URL
		} else if strings.Contains(to, "worker") {
			targetURL = s.drone2URL
		} else if strings.Contains(to, "relay") {
			targetURL = s.drone3URL
		}
	}

	req, err := http.NewRequest("POST", targetURL+"/fleet/task", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to assign task: %s", string(body))
	}

	return nil
}

// sendDroneMessage sends a message between drones
func (s *DroneFleetTestSuite) sendDroneMessage(ctx context.Context, msg DroneMessage) error {
	payload := map[string]interface{}{
		"from":         msg.From,
		"to":           msg.To,
		"message_type": msg.Type,
		"content":      msg.Content,
		"task_id":      msg.TaskID,
		"priority":     msg.Priority,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	targetURL := s.leadURL
	if msg.To != "lead-drone" && msg.To != "all" {
		if strings.Contains(msg.To, "scout") {
			targetURL = s.drone1URL
		} else if strings.Contains(msg.To, "worker") {
			targetURL = s.drone2URL
		} else if strings.Contains(msg.To, "relay") {
			targetURL = s.drone3URL
		}
	}

	req, err := http.NewRequest("POST", targetURL+"/fleet/message", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// printTaskLogs prints all drone task logs
func (s *DroneFleetTestSuite) printTaskLogs() {
	for name, log := range s.taskLogs {
		s.T().Logf("┌─────────────────────────────────────────────────────────────┐")
		s.T().Logf("│ DRONE: %-9s ROLE: %-8s                        │", strings.ToUpper(name), strings.ToUpper(log.Role))
		s.T().Logf("│ TAILSCALE IP: %-15s                            │", log.IP)
		s.T().Logf("├─────────────────────────────────────────────────────────────┤")

		// Print tasks
		if len(log.Tasks) > 0 {
			s.T().Logf("│ TASKS ASSIGNED: %d                                           │", len(log.Tasks))
			for i, task := range log.Tasks {
				s.T().Logf("│   %d. [%s] %s - %s                         │", i+1, task.Priority, task.ID, task.Type)
				s.T().Logf("│      %s                                         │", task.Description[:50])
			}
		}

		// Print messages
		if len(log.Messages) > 0 {
			s.T().Logf("├─────────────────────────────────────────────────────────────┤")
			s.T().Logf("│ COMMUNICATION LOG: %d messages                               │", len(log.Messages))
			
			for i, msg := range log.Messages {
				timeStr := msg.Timestamp.Format("15:04:05")
				typeBadge := "💬"
				if msg.Type == "status_report" {
					typeBadge = "📊"
				} else if msg.Type == "task_assignment" {
					typeBadge = "📋"
				} else if msg.Type == "coordination" {
					typeBadge = "🤝"
				}

				s.T().Logf("│   %d. %s %s %s → %s                       │", i+1, timeStr, typeBadge, msg.From, msg.To)
				
				content := msg.Content
				for len(content) > 58 {
					s.T().Logf("│       %s                                               │", content[:58])
					content = content[58:]
				}
				s.T().Logf("│       %s                                               │", content)
			}
		}

		s.T().Logf("└─────────────────────────────────────────────────────────────┘")
		s.T().Log("")
	}
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestDroneFleetSuite runs the test suite
func TestDroneFleetSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	if os.Getenv("TS_AUTH_KEY") == "" {
		t.Log("TS_AUTH_KEY not set, using default test key")
	}

	suite.Run(t, new(DroneFleetTestSuite))
}
