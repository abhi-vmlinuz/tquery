package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	KeyStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")) // Cyan
	BranchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))           // Grey
)

// BuildTree formats any JSON interface into a beautiful ASCII tree string.
func BuildTree(val any, useColor bool) string {
	var sb strings.Builder
	buildTreeInternal(&sb, "", "root", val, true, useColor)
	return sb.String()
}

func buildTreeInternal(sb *strings.Builder, indent string, key string, val any, isLast bool, useColor bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	if useColor {
		sb.WriteString(indent + BranchStyle.Render(connector))
		sb.WriteString(KeyStyle.Render(key))
	} else {
		sb.WriteString(indent + connector + key)
	}

	switch v := val.(type) {
	case map[string]any:
		sb.WriteString("\n")
		nextIndent := indent
		if isLast {
			nextIndent += "    "
		} else {
			nextIndent += "│   "
		}

		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for i, k := range keys {
			buildTreeInternal(sb, nextIndent, k, v[k], i == len(keys)-1, useColor)
		}

	case []any:
		sb.WriteString(fmt.Sprintf(" [%d items]\n", len(v)))
		nextIndent := indent
		if isLast {
			nextIndent += "    "
		} else {
			nextIndent += "│   "
		}

		for i, item := range v {
			itemKey := fmt.Sprintf("[%d]", i)
			buildTreeInternal(sb, nextIndent, itemKey, item, i == len(v)-1, useColor)
		}

	default:
		formatted := fmt.Sprintf("%v", v)
		if useColor {
			sb.WriteString(": " + colorizeCell(formatted) + "\n")
		} else {
			sb.WriteString(": " + formatted + "\n")
		}
	}
}
