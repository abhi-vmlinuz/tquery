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

// BuildTree formats any JSON interface into a beautiful ASCII tree string with multi-pattern match highlighting.
func BuildTree(val any, useColor bool, patterns []string, ignoreCase bool) string {
	var sb strings.Builder
	buildTreeInternal(&sb, "", "root", val, true, useColor, patterns, ignoreCase)
	return sb.String()
}

func buildTreeInternal(sb *strings.Builder, indent string, key string, val any, isLast bool, useColor bool, patterns []string, ignoreCase bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	displayKey := key
	if useColor {
		sb.WriteString(indent + BranchStyle.Render(connector))
		if len(patterns) > 0 {
			displayKey = filter.HighlightMulti(key, patterns, ignoreCase)
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
			buildTreeInternal(sb, nextIndent, k, v[k], i == len(keys)-1, useColor, patterns, ignoreCase)
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
			buildTreeInternal(sb, nextIndent, itemKey, item, i == len(v)-1, useColor, patterns, ignoreCase)
		}

	default:
		formatted := fmt.Sprintf("%v", v)
		if useColor {
			if len(patterns) > 0 {
				formatted = filter.HighlightMulti(formatted, patterns, ignoreCase)
			} else {
				formatted = colorizeCell(formatted)
			}
			sb.WriteString(": " + formatted + "\n")
		} else {
			sb.WriteString(": " + formatted + "\n")
		}
	}
}
