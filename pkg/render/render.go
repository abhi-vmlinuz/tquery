package render

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/tquery/tquery/pkg/filter"
	"github.com/tquery/tquery/pkg/parser"
)

type Format string

const (
	FormatTable    Format = "table"
	FormatMarkdown Format = "markdown"
	FormatCSV      Format = "csv"
	FormatTSV      Format = "tsv"
	FormatTree     Format = "tree"
	FormatJSON     Format = "json"
)

// Styles using Lipgloss
var (
	HeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EC4899"))   // Bright Pink
	NumberStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24"))              // Amber/Yellow
	BoolStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ADE80"))              // Emerald Green
	NullStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Italic(true) // Slate Grey
	StringStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F8FAFC"))              // Crisp White
	BorderColor = lipgloss.Color("#475569")                                              // Subtle Border
)

type RenderOptions struct {
	Format              Format
	ShowHeader          bool
	UseColor            bool
	Limit               int
	HighlightPatterns   []string
	HighlightIgnoreCase bool
}

func DefaultOptions() RenderOptions {
	return RenderOptions{
		Format:     FormatTable,
		ShowHeader: true,
		UseColor:   true,
	}
}

// Render renders a DataStructure according to options into writer.
func Render(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	switch opts.Format {
	case FormatTable:
		return renderTable(w, ds, opts)
	case FormatMarkdown:
		return renderMarkdown(w, ds, opts)
	case FormatCSV, FormatTSV:
		return renderCSV(w, ds, opts)
	case FormatTree:
		return renderTree(w, ds, opts)
	case FormatJSON:
		return renderJSON(w, ds, opts)
	default:
		return renderTable(w, ds, opts)
	}
}

func renderTable(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	if len(ds.Headers) == 0 && len(ds.Rows) == 0 {
		msg := "[] (empty)"
		if opts.UseColor {
			msg = NullStyle.Render(msg)
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(BorderColor)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				return HeaderStyle.Padding(0, 1)
			}
			return lipgloss.NewStyle().Padding(0, 1)
		})

	if opts.ShowHeader && len(ds.Headers) > 0 {
		headers := make([]string, len(ds.Headers))
		for i, h := range ds.Headers {
			if opts.UseColor {
				headers[i] = HeaderStyle.Render(h)
			} else {
				headers[i] = h
			}
		}
		t.Headers(headers...)
	}

	rows := ds.Rows
	if opts.Limit > 0 && len(rows) > opts.Limit {
		rows = rows[:opts.Limit]
	}

	for _, row := range rows {
		styledRow := make([]string, len(row))
		for i, cell := range row {
			if opts.UseColor {
				if len(opts.HighlightPatterns) > 0 {
					styledRow[i] = filter.HighlightMulti(cell, opts.HighlightPatterns, opts.HighlightIgnoreCase)
				} else {
					styledRow[i] = colorizeCell(cell)
				}
			} else {
				styledRow[i] = cell
			}
		}
		t.Row(styledRow...)
	}

	_, err := fmt.Fprintln(w, t.String())
	return err
}

func colorizeCell(val string) string {
	if val == "null" {
		return NullStyle.Render("null")
	}
	if val == "true" || val == "false" {
		return BoolStyle.Render(val)
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil && val != "" {
		return NumberStyle.Render(val)
	}
	return StringStyle.Render(val)
}

func renderMarkdown(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	if len(ds.Headers) == 0 {
		_, err := fmt.Fprintln(w, "_(empty)_")
		return err
	}

	var sb strings.Builder
	sb.WriteString("| " + strings.Join(ds.Headers, " | ") + " |\n")
	seps := make([]string, len(ds.Headers))
	for i := range seps {
		seps[i] = "---"
	}
	sb.WriteString("| " + strings.Join(seps, " | ") + " |\n")

	rows := ds.Rows
	if opts.Limit > 0 && len(rows) > opts.Limit {
		rows = rows[:opts.Limit]
	}

	for _, row := range rows {
		escapedRow := make([]string, len(row))
		for i, cell := range row {
			escapedRow[i] = strings.ReplaceAll(cell, "|", "\\|")
		}
		sb.WriteString("| " + strings.Join(escapedRow, " | ") + " |\n")
	}

	_, err := io.WriteString(w, sb.String())
	return err
}

func renderCSV(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	writer := csv.NewWriter(w)
	if opts.Format == FormatTSV {
		writer.Comma = '\t'
	}

	if opts.ShowHeader && len(ds.Headers) > 0 {
		if err := writer.Write(ds.Headers); err != nil {
			return err
		}
	}

	rows := ds.Rows
	if opts.Limit > 0 && len(rows) > opts.Limit {
		rows = rows[:opts.Limit]
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func renderJSON(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	var target = ds.Unwrapped
	if opts.Limit > 0 {
		if slice, ok := target.([]any); ok && len(slice) > opts.Limit {
			target = slice[:opts.Limit]
		}
	}
	b, err := json.MarshalIndent(target, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

func renderTree(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	treeStr := BuildTree(ds.Unwrapped, opts.UseColor, opts.HighlightPatterns, opts.HighlightIgnoreCase)
	if opts.Limit > 0 {
		lines := strings.Split(treeStr, "\n")
		var cleanLines []string
		for _, line := range lines {
			if line != "" {
				cleanLines = append(cleanLines, line)
			}
		}
		if len(cleanLines) > opts.Limit {
			cleanLines = cleanLines[:opts.Limit]
			treeStr = strings.Join(cleanLines, "\n") + "\n..."
		} else {
			treeStr = strings.Join(cleanLines, "\n")
		}
	}
	_, err := io.WriteString(w, treeStr+"\n")
	return err
}
