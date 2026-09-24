package tui

import (
	"encoding/json"
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/abhi-vmlinuz/tquery/pkg/engine"
	"github.com/abhi-vmlinuz/tquery/pkg/parser"
	"github.com/abhi-vmlinuz/tquery/pkg/render"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+y":
			copyText := m.Query
			if copyText == "" {
				copyText = "."
			}
			if err := clipboard.WriteAll(copyText); err == nil {
				m.StatusMsg = fmt.Sprintf("✓ Copied query %q to clipboard!", copyText)
			} else {
				m.StatusMsg = fmt.Sprintf("⚠ Clipboard error: %v", err)
			}
			return m, nil

		case "esc":
			if m.ShowInspect {
				m.ShowInspect = false
				return m, nil
			}
			if m.InputMode == ModeFilter {
				m.InputMode = ModeNav
				m.TextInput.Blur()
				return m, nil
			}

		case "tab":
			if !m.ShowInspect {
				m.ViewMode = (m.ViewMode + 1) % 3
				m.refreshView()
				return m, nil
			}

		case "enter":
			if m.ShowInspect {
				m.ShowInspect = false
				return m, nil
			}
			if m.InputMode == ModeNav && m.ViewMode == ViewTable && len(m.Table.Rows()) > 0 {
				m.openInspectDrawer()
				return m, nil
			}
			if m.InputMode == ModeFilter {
				// Pressing enter in filter switches to navigation mode
				m.InputMode = ModeNav
				m.TextInput.Blur()
				return m, nil
			}
		}

		// When in ModeNav (Vim-style row & viewport movement)
		if m.InputMode == ModeNav && !m.ShowInspect {
			switch msg.String() {
			case "/", "i":
				m.InputMode = ModeFilter
				m.TextInput.Focus()
				return m, nil
			case "j", "down":
				if m.ViewMode == ViewTable {
					m.Table.MoveDown(1)
				} else {
					m.Viewport.LineDown(1)
				}
				return m, nil
			case "k", "up":
				if m.ViewMode == ViewTable {
					m.Table.MoveUp(1)
				} else {
					m.Viewport.LineUp(1)
				}
				return m, nil
			case "g":
				if m.ViewMode == ViewTable {
					m.Table.GotoTop()
				} else {
					m.Viewport.GotoTop()
				}
				return m, nil
			case "G":
				if m.ViewMode == ViewTable {
					m.Table.GotoBottom()
				} else {
					m.Viewport.GotoBottom()
				}
				return m, nil
			case "q":
				return m, tea.Quit
			}
		}

		// When in ModeFilter, allow Up/Down / Ctrl+N/Ctrl+P to navigate rows
		if m.InputMode == ModeFilter && !m.ShowInspect {
			switch msg.String() {
			case "up", "ctrl+p":
				if m.ViewMode == ViewTable {
					m.Table.MoveUp(1)
					return m, nil
				}
			case "down", "ctrl+n":
				if m.ViewMode == ViewTable {
					m.Table.MoveDown(1)
					return m, nil
				}
			}
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		m.TextInput.Width = msg.Width - 10
		m.Viewport.Width = msg.Width - 4
		m.Viewport.Height = msg.Height - 7
		m.refreshView()
	}

	// Update text input if in ModeFilter and not inspecting
	if m.InputMode == ModeFilter && !m.ShowInspect {
		oldVal := m.TextInput.Value()
		var cmd tea.Cmd
		m.TextInput, cmd = m.TextInput.Update(msg)
		cmds = append(cmds, cmd)

		if m.TextInput.Value() != oldVal {
			m.Query = m.TextInput.Value()
			m.StatusMsg = ""
			m.evaluateQuery()
		}
	}

	// Handle Viewport scrolling during Inspect
	if m.ShowInspect {
		var cmd tea.Cmd
		m.Viewport, cmd = m.Viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) evaluateQuery() {
	if m.Query == "" || m.Query == "." {
		m.CurrentData = m.RawData
		m.QueryErr = nil
		m.updateDataStructure()
		return
	}

	res, err := engine.Evaluate(m.Query, m.RawData)
	if err != nil {
		m.QueryErr = err
		return
	}

	m.QueryErr = nil
	m.CurrentData = res
	m.updateDataStructure()
}

func (m *Model) updateDataStructure() {
	target := m.CurrentData
	if m.AutoUnwrap && m.Query == "" {
		target = parser.UnwrapRoot(m.CurrentData)
	}

	b, _ := json.Marshal(target)
	ds, err := parser.Parse(b, false)
	if err != nil {
		// Fallback
		ds = &parser.DataStructure{
			Raw:       target,
			Unwrapped: target,
			Type:      parser.KindValue,
			Headers:   []string{"Value"},
			Rows:      [][]string{{fmt.Sprintf("%v", target)}},
		}
	}
	m.DataStruct = ds
	m.refreshView()
}

func (m *Model) refreshView() {
	// Rebuild Table
	if m.DataStruct != nil && len(m.DataStruct.Headers) > 0 {
		columns := make([]table.Column, len(m.DataStruct.Headers))
		colWidth := (m.Width - 6) / len(m.DataStruct.Headers)
		if colWidth < 12 {
			colWidth = 12
		}

		for i, h := range m.DataStruct.Headers {
			columns[i] = table.Column{Title: h, Width: colWidth}
		}

		rows := make([]table.Row, len(m.DataStruct.Rows))
		for i, r := range m.DataStruct.Rows {
			row := make(table.Row, len(r))
			copy(row, r)
			rows[i] = row
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithFocused(true),
			table.WithHeight(m.Height-8),
		)
		s := table.DefaultStyles()
		s.Header = s.Header.BorderStyle(table.DefaultStyles().Header.GetBorderStyle()).Bold(true)
		s.Selected = s.Selected.Foreground(table.DefaultStyles().Selected.GetForeground()).Bold(true)
		t.SetStyles(s)
		m.Table = t
	}

	// Rebuild Viewport content for Tree / JSON
	switch m.ViewMode {
	case ViewTree:
		m.Viewport.SetContent(render.BuildTree(m.DataStruct.Unwrapped, true, nil, false))
	case ViewJSON:
		b, _ := json.MarshalIndent(m.DataStruct.Unwrapped, "", "  ")
		m.Viewport.SetContent(string(b))
	}
}

func (m *Model) openInspectDrawer() {
	cursor := m.Table.Cursor()
	if cursor < 0 {
		return
	}

	var inspectData any

	// Extract original deep nested record from Unwrapped slice if available
	if slice, ok := m.DataStruct.Unwrapped.([]any); ok && cursor < len(slice) {
		inspectData = slice[cursor]
	} else if len(m.DataStruct.Rows) > cursor {
		selectedRow := m.DataStruct.Rows[cursor]
		obj := make(map[string]any)
		for i, h := range m.DataStruct.Headers {
			if i < len(selectedRow) {
				obj[h] = selectedRow[i]
			}
		}
		inspectData = obj
	}

	if inspectData == nil {
		return
	}

	b, err := json.MarshalIndent(inspectData, "", "  ")
	if err != nil {
		m.InspectContent = fmt.Sprintf("%v", inspectData)
	} else {
		m.InspectContent = string(b)
	}

	m.ShowInspect = true
	m.Viewport.SetContent(m.InspectContent)
}
