package render

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/olekukonko/tablewriter"
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
	HeaderStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")) // Pink/Magenta
	NumberStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))            // Yellow
	BoolStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("120"))            // Light Green
	NullStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Italic(true) // Grey
	StringStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))            // White
	BorderColor  = lipgloss.Color("240")
)

type RenderOptions struct {
	Format     Format
	ShowHeader bool
	UseColor   bool
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
	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)

	table.SetAutoWrapText(true)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("┼")
	table.SetColumnSeparator("│")
	table.SetRowSeparator("─")
	table.SetHeaderLine(true)
	table.SetBorder(true)

	if opts.ShowHeader && len(ds.Headers) > 0 {
		headers := make([]string, len(ds.Headers))
		for i, h := range ds.Headers {
			if opts.UseColor {
				headers[i] = HeaderStyle.Render(h)
			} else {
				headers[i] = h
			}
		}
		table.SetHeader(headers)
	}

	for _, row := range ds.Rows {
		styledRow := make([]string, len(row))
		for i, cell := range row {
			if opts.UseColor {
				styledRow[i] = colorizeCell(cell)
			} else {
				styledRow[i] = cell
			}
		}
		table.Append(styledRow)
	}

	table.Render()

	tableStr := buf.String()
	if opts.UseColor {
		tableStr = lipgloss.NewStyle().BorderForeground(BorderColor).Render(tableStr)
	}

	_, err := io.WriteString(w, tableStr+"\n")
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
		return nil
	}

	var sb strings.Builder
	// Header line
	sb.WriteString("| " + strings.Join(ds.Headers, " | ") + " |\n")
	// Separator line
	seps := make([]string, len(ds.Headers))
	for i := range seps {
		seps[i] = "---"
	}
	sb.WriteString("| " + strings.Join(seps, " | ") + " |\n")

	// Data rows
	for _, row := range ds.Rows {
		// Escape pipeline characters in markdown table cells
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

	for _, row := range ds.Rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}

func renderJSON(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	b, err := json.MarshalIndent(ds.Unwrapped, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(b))
	return err
}

func renderTree(w io.Writer, ds *parser.DataStructure, opts RenderOptions) error {
	treeStr := BuildTree(ds.Unwrapped, opts.UseColor)
	_, err := io.WriteString(w, treeStr+"\n")
	return err
}
