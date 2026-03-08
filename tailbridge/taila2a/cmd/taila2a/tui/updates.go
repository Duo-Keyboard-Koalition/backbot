package tui
 
import (
	"fmt"
	"time"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)
 
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		}),
		fetchAgents(m.apiClient),
		fetchNotifications(m.apiClient),
	)
}
 
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
 
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "1":
			m.activeTab = AgentsTab
		case "2":
			m.activeTab = MessagesTab
		}
	
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Adjust table height
		m.agentTable.SetHeight(m.height - 10)
		
		// Adjust viewport
		m.logsViewport.Width = m.width - 2
		m.logsViewport.Height = m.height - 10
 
	case tickMsg:
		m.lastUpdate = time.Time(msg)
		cmds = append(cmds, 
			fetchAgents(m.apiClient),
			fetchNotifications(m.apiClient),
			tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
				return tickMsg(t)
			}),
		)
 
	case agentsLoadedMsg:
		m.agents = msg
		rows := make([]table.Row, 0, len(msg))
		for _, a := range msg {
			status := "❌ Offline"
			if a.Online {
				status = "🟢 Online"
			}
			
			services := ""
			for i, s := range a.Services {
				if i > 0 {
					services += ", "
				}
				services += fmt.Sprintf("%s:%d", s.Service, s.Port)
			}
			
			rows = append(rows, table.Row{
				a.Name,
				a.IP,
				status,
				services,
			})
		}
		m.agentTable.SetRows(rows)
		
	case notificationsLoadedMsg:
		m.notifications = msg
		content := ""
		for _, n := range msg {
			content += fmt.Sprintf("[%s] %s: %s\n", n.Timestamp, n.Level, n.Message)
		}
		m.logsViewport.SetContent(content)
		m.logsViewport.GotoBottom()
 
	case errMsg:
		m.err = msg.err
	}
 
	// Handle component updates
	if m.activeTab == AgentsTab {
		m.agentTable, cmd = m.agentTable.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.activeTab == MessagesTab {
		m.logsViewport, cmd = m.logsViewport.Update(msg)
		cmds = append(cmds, cmd)
	}
 
	return m, tea.Batch(cmds...)
}
 
// Commands
func fetchAgents(client *Taila2aClient) tea.Cmd {
	return func() tea.Msg {
		agents, err := client.FetchAgents()
		if err != nil {
			return errMsg{err}
		}
		return agentsLoadedMsg(agents)
	}
}
 
func fetchNotifications(client *Taila2aClient) tea.Cmd {
	return func() tea.Msg {
		notes, err := client.FetchNotifications()
		if err != nil {
			return errMsg{err}
		}
		return notificationsLoadedMsg(notes)
	}
}
