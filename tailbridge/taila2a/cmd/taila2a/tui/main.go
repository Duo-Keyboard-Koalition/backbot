package tui
 
import (
	"fmt"
	"os"
	tea "github.com/charmbracelet/bubbletea"
)
 
func Run(port int) error {
	m := InitialModel(port)
	p := tea.NewProgram(m, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
	
	return nil
}
