package tui
 
import (
	"time"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)
 
type ActiveTab int
 
const (
	AgentsTab ActiveTab = iota
	MessagesTab
)
 
// Agent structure mimicking what /agents returns
type Agent struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Online   bool   `json:"online"`
	Services []struct {
		Service string `json:"service"`
		Port    int    `json:"port"`
	} `json:"gateways"`
}
 
type Notification struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Data      map[string]interface{} `json:"data"`
}
 
// The Bubbletea model
type model struct {
	activeTab     ActiveTab
	
	// Data
	agents        []Agent
	notifications []Notification
	
	// UI Components
	agentTable    table.Model
	logsViewport  viewport.Model
	
	// State
	err           error
	lastUpdate    time.Time
	apiPort       int
	apiClient     *Taila2aClient
	
	// Layout
	width         int
	height        int
}
 
// TUI messages
type tickMsg time.Time
type agentsLoadedMsg []Agent
type notificationsLoadedMsg []Notification
type errMsg struct{ err error }
 
func InitialModel(port int) model {
	// Setup agents table
	cols := []table.Column{
		{Title: "Node", Width: 20},
		{Title: "IP", Width: 15},
		{Title: "Status", Width: 10},
		{Title: "Services", Width: 30},
	}
	
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(make([]table.Row, 0)),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	
	// Styling the table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)
 
	vp := viewport.New(0, 0)
	vp.SetContent("Waiting for messages...")
 
	m := model{
		activeTab:    AgentsTab,
		agentTable:   t,
		logsViewport: vp,
		apiPort:      port,
		apiClient:    NewClient(port),
		lastUpdate:   time.Now(),
	}
 
	return m
}
