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

	NavPill = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#F59E0B")).
			Padding(0, 1)

	FilterPill = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#10B981")).
			Padding(0, 1)

	StatusSuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#10B981"))

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

	var inputBadge string
	if m.InputMode == ModeFilter {
		inputBadge = FilterPill.Render("FILTER")
	} else {
		inputBadge = NavPill.Render("NAV")
	}

	var modePills []string
	modes := []string{"Table", "Tree", "JSON"}
	for i, name := range modes {
		if ViewMode(i) == m.ViewMode {
			modePills = append(modePills, PillActive.Render(name))
		} else {
			modePills = append(modePills, PillInactive.Render(name))
		}
	}
	header := title + "  " + inputBadge + "  " + strings.Join(modePills, " ")
	sb.WriteString(header + "\n\n")

	// Search Input Bar
	prompt := PromptStyle.Render("jq > ")
	sb.WriteString(prompt + m.TextInput.View())
	if m.StatusMsg != "" {
		sb.WriteString("  " + StatusSuccessStyle.Render(m.StatusMsg))
	}
	sb.WriteString("\n")

	// Error Line
	if m.QueryErr != nil {
		sb.WriteString(ErrStyle.Render(fmt.Sprintf("  ⚠ %v", m.QueryErr)) + "\n\n")
	} else {
		sb.WriteString("\n")
	}

	// Main Viewport / Table / Overlay
	if m.ShowInspect {
		inspectBar := InspectHeader.Render(" INSPECT ROW DETAIL (Press ESC or Enter to return) ")
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
	var footer string
	if m.ShowInspect {
		footer = HelpStyle.Render("Esc / Enter: return to table  •  j/k: scroll  •  Ctrl+C: quit")
	} else if m.InputMode == ModeFilter {
		footer = HelpStyle.Render("Esc / Enter: navigate rows  •  Tab: switch view  •  Ctrl+Y: copy query  •  Ctrl+C: quit")
	} else {
		footer = HelpStyle.Render("/ or i: filter  •  j/k: move rows  •  Enter: inspect  •  Tab: view  •  Ctrl+Y: copy query  •  q: quit")
	}
	sb.WriteString("\n" + footer)

	return sb.String()
}
