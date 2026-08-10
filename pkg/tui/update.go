package tui

import (
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tquery/tquery/pkg/engine"
	"github.com/tquery/tquery/pkg/parser"
	"github.com/tquery/tquery/pkg/render"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc":
			if m.ShowInspect {
				m.ShowInspect = false
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
			if m.ViewMode == ViewTable && len(m.Table.Rows()) > 0 {
				m.openInspectDrawer()
				return m, nil
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

	// Update text input if not inspect drawer
	if !m.ShowInspect {
		oldVal := m.TextInput.Value()
		var cmd tea.Cmd
		m.TextInput, cmd = m.TextInput.Update(msg)
		cmds = append(cmds, cmd)

		if m.TextInput.Value() != oldVal {
			m.Query = m.TextInput.Value()
			m.evaluateQuery()
		}
	}

	// Handle Table / Viewport updates
	if m.ViewMode == ViewTable && !m.ShowInspect {
		var cmd tea.Cmd
		m.Table, cmd = m.Table.Update(msg)
		cmds = append(cmds, cmd)
	} else if m.ViewMode != ViewTable || m.ShowInspect {
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
		m.Viewport.SetContent(render.BuildTree(m.DataStruct.Unwrapped, true))
	case ViewJSON:
		b, _ := json.MarshalIndent(m.DataStruct.Unwrapped, "", "  ")
		m.Viewport.SetContent(string(b))
	}
}

func (m *Model) openInspectDrawer() {
	selectedRow := m.Table.SelectedRow()
	if len(selectedRow) == 0 {
		return
	}

	obj := make(map[string]string)
	for i, h := range m.DataStruct.Headers {
		if i < len(selectedRow) {
			obj[h] = selectedRow[i]
		}
	}

	b, _ := json.MarshalIndent(obj, "", "  ")
	m.InspectContent = string(b)
	m.ShowInspect = true
	m.Viewport.SetContent(m.InspectContent)
}
