package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the interactive TUI visualizer.
func Run(rawJSON []byte, initialQuery string, autoUnwrap bool) error {
	m, err := NewModel(rawJSON, initialQuery, autoUnwrap)
	if err != nil {
		return err
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
