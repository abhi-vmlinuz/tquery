package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1)

	PillActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#22C55E")).
			Padding(0, 1)

	PillInactive = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 1)

	PromptStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F43F5E"))

	ErrStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#64748B"))

	InspectHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F59E0B")).
			Background(lipgloss.Color("#1E1B4B")).
			Padding(0, 1)
)

func (m Model) View() string {
	var sb strings.Builder

	// Header Bar
	title := TitleStyle.Render("tquery")
	var modePills []string
	modes := []string{"Table", "Tree", "JSON"}
	for i, name := range modes {
		if ViewMode(i) == m.ViewMode {
			modePills = append(modePills, PillActive.Render(name))
		} else {
			modePills = append(modePills, PillInactive.Render(name))
		}
	}
	header := title + "  " + strings.Join(modePills, " ")
	sb.WriteString(header + "\n\n")

	// Search Input Bar
	prompt := PromptStyle.Render("jq > ")
	sb.WriteString(prompt + m.TextInput.View() + "\n")

	// Error Line
	if m.QueryErr != nil {
		sb.WriteString(ErrStyle.Render(fmt.Sprintf("  ⚠ %v", m.QueryErr)) + "\n\n")
	} else {
		sb.WriteString("\n")
	}

	// Main Viewport / Table / Overlay
	if m.ShowInspect {
		inspectBar := InspectHeader.Render(" INSPECT ROW DETAIL (Press ESC to return) ")
		sb.WriteString(inspectBar + "\n" + m.Viewport.View() + "\n")
	} else {
		switch m.ViewMode {
		case ViewTable:
			if m.DataStruct != nil && len(m.Table.Rows()) > 0 {
				sb.WriteString(m.Table.View() + "\n")
			} else {
				sb.WriteString(HelpStyle.Render("No rows to display in table view.") + "\n")
			}
		case ViewTree, ViewJSON:
			sb.WriteString(m.Viewport.View() + "\n")
		}
	}

	// Footer / Help Line
	footer := HelpStyle.Render("Tab: switch view  •  Enter: inspect row  •  Ctrl+C / q: quit")
	sb.WriteString("\n" + footer)

	return sb.String()
}
