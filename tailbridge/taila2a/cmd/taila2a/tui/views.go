package tui
 
import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
)
 
var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		MarginBottom(1)
	
	tabStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)
		
	activeTabStyle = tabStyle.Copy().
		Foreground(lipgloss.Color("205")).
		BorderForeground(lipgloss.Color("205"))
		
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1)
)
 
func (m model) View() string {
	if m.width == 0 {
		return "Initializing Monitor..."
	}
 
	doc := strings.Builder{}
 
	// Title
	doc.WriteString(titleStyle.Render("🤖 SentinelAI Fleet Monitor") + "\n")
	
	// Stats
	onlineCount := 0
	for _, a := range m.agents {
		if a.Online {
			onlineCount++
		}
	}
	stats := fmt.Sprintf("Fleet Status: 🟢 %d/%d Online | Last updated: %s\n", 
		onlineCount, len(m.agents), m.lastUpdate.Format("15:04:05"))
	doc.WriteString(stats + "\n")
 
	// Tabs
	tabs := []string{"[1] Agents", "[2] Messages"}
	for i, t := range tabs {
		if i == int(m.activeTab) {
			doc.WriteString(activeTabStyle.Render(t))
		} else {
			doc.WriteString(tabStyle.Render(t))
		}
		doc.WriteString("  ")
	}
	doc.WriteString("\n\n")
 
	// Content
	switch m.activeTab {
	case AgentsTab:
		doc.WriteString(m.agentTable.View())
	case MessagesTab:
		doc.WriteString(m.logsViewport.View())
	}
 
	// Help
	helpText := "\n[q] Quit  [1-2] Switch Tabs  "
	if m.activeTab == AgentsTab {
		helpText += " [↑/↓] Navigate "
	}
	if m.err != nil {
		helpText += fmt.Sprintf(" | ERROR: %v", m.err)
	}
	doc.WriteString(helpStyle.Render(helpText))
 
	return doc.String()
}
