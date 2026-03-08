package tui

import (
	"fmt"
	"strings"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)

	tabStyle = lipgloss.NewStyle().
		Padding(0, 1)

	activeTabStyle = tabStyle.Copy().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Underline(true)

	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	sepStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))
)

// headerLines is the exact number of lines the header occupies:
//   1: title
//   2: stats
//   3: separator
//   4: tabs
//   5: separator
const headerLines = 5

// footerLines is the exact number of lines the footer occupies:
//   1: blank
//   2: help
const footerLines = 2

func (m model) View() string {
	if m.width == 0 {
		return "Initializing Monitor..."
	}

	doc := strings.Builder{}
	sep := sepStyle.Render(strings.Repeat("─", m.width))

	// Line 1: Title
	doc.WriteString(titleStyle.Render("🤖 SentinelAI Fleet Monitor") + "\n")

	// Line 2: Stats
	onlineCount := 0
	for _, a := range m.agents {
		if a.Online {
			onlineCount++
		}
	}
	doc.WriteString(fmt.Sprintf("Fleet Status: 🟢 %d/%d Online | Last updated: %s\n",
		onlineCount, len(m.agents), m.lastUpdate.Format("15:04:05")))

	// Line 3: Separator
	doc.WriteString(sep + "\n")

	// Line 4: Tabs (single-line, no borders)
	tabs := []string{"[1] Agents", "[2] Messages"}
	renderedTabs := make([]string, len(tabs))
	for i, t := range tabs {
		if i == int(m.activeTab) {
			renderedTabs[i] = activeTabStyle.Render(t)
		} else {
			renderedTabs[i] = tabStyle.Render(t)
		}
	}
	doc.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...) + "\n")

	// Line 5: Separator
	doc.WriteString(sep + "\n")

	// Lines 6 to (m.height - footerLines): Content
	switch m.activeTab {
	case AgentsTab:
		doc.WriteString(m.agentTable.View())
	case MessagesTab:
		doc.WriteString(m.logsViewport.View())
	}

	// Line m.height - 1: blank separator before footer
	doc.WriteString("\n")

	// Line m.height: Help
	helpText := "[q] Quit  [1-2] Switch Tabs"
	if m.activeTab == AgentsTab {
		helpText += "  [↑/↓] Navigate"
	}
	if m.err != nil {
		helpText += fmt.Sprintf(" | ERROR: %v", m.err)
	}
	doc.WriteString(helpStyle.Render(helpText))

	return doc.String()
}
