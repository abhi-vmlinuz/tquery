package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tquery/tquery/pkg/filter"
)

var (
	KeyStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")) // Cyan
	BranchStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))          // Grey
)

// BuildTree formats any JSON interface into a beautiful ASCII tree string with optional match highlighting.
func BuildTree(val any, useColor bool, highlightPattern string, ignoreCase bool) string {
	var sb strings.Builder
	buildTreeInternal(&sb, "", "root", val, true, useColor, highlightPattern, ignoreCase)
	return sb.String()
}

func buildTreeInternal(sb *strings.Builder, indent string, key string, val any, isLast bool, useColor bool, pattern string, ignoreCase bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	displayKey := key
	if useColor {
		sb.WriteString(indent + BranchStyle.Render(connector))
		if pattern != "" {
			displayKey = filter.Highlight(key, pattern, ignoreCase)
		} else {
			displayKey = KeyStyle.Render(key)
		}
		sb.WriteString(displayKey)
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
			buildTreeInternal(sb, nextIndent, k, v[k], i == len(keys)-1, useColor, pattern, ignoreCase)
		}

	case []any:
		if len(v) == 0 {
			sb.WriteString(" [] (empty)\n")
			return
		}
		sb.WriteString(fmt.Sprintf(" [%d items]\n", len(v)))
		nextIndent := indent
		if isLast {
			nextIndent += "    "
		} else {
			nextIndent += "│   "
		}

		for i, item := range v {
			itemKey := fmt.Sprintf("[%d]", i)
			buildTreeInternal(sb, nextIndent, itemKey, item, i == len(v)-1, useColor, pattern, ignoreCase)
		}

	default:
		formatted := fmt.Sprintf("%v", v)
		if useColor {
			if pattern != "" {
				formatted = filter.Highlight(formatted, pattern, ignoreCase)
			} else {
				formatted = colorizeCell(formatted)
			}
			sb.WriteString(": " + formatted + "\n")
		} else {
			sb.WriteString(": " + formatted + "\n")
		}
	}
}
