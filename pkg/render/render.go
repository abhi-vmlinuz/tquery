package render

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/mattn/go-isatty"
	"github.com/mattn/go-runewidth"
	"github.com/tquery/tquery/pkg/filter"
	"github.com/tquery/tquery/pkg/parser"
	"golang.org/x/term"
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

	numCols := len(ds.Headers)
	if numCols == 0 && len(ds.Rows) > 0 {
		numCols = len(ds.Rows[0])
	}

	rows := ds.Rows
	if opts.Limit > 0 && len(rows) > opts.Limit {
		rows = rows[:opts.Limit]
	}

	// 1. Calculate natural content width for each column
	naturalWidths := make([]int, numCols)
	for j := 0; j < numCols; j++ {
		if j < len(ds.Headers) {
			naturalWidths[j] = lipgloss.Width(ds.Headers[j])
		}
		for _, row := range rows {
			if j < len(row) {
				w := lipgloss.Width(row[j])
				if w > naturalWidths[j] {
					naturalWidths[j] = w
				}
			}
		}
	}

	// 2. Terminal-aware adaptive width allocation
	termWidth := getTerminalWidth()
	colWidths := allocateColumnWidths(naturalWidths, ds.Headers, termWidth)

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
			headerText := h
			if i < len(colWidths) && colWidths[i] > 0 && lipgloss.Width(h) > colWidths[i] {
				headerText = truncateString(h, colWidths[i])
			}
			if opts.UseColor {
				headers[i] = HeaderStyle.Render(headerText)
			} else {
				headers[i] = headerText
			}
		}
		t.Headers(headers...)
	}

	for _, row := range rows {
		styledRow := make([]string, len(row))
		for i, cell := range row {
			maxW := 0
			if i < len(colWidths) {
				maxW = colWidths[i]
			}

			truncated := cell
			if maxW > 0 && lipgloss.Width(cell) > maxW {
				truncated = truncateCellWithMatch(cell, maxW, opts.HighlightPatterns, opts.HighlightIgnoreCase)
			}

			if opts.UseColor {
				if len(opts.HighlightPatterns) > 0 {
					styledRow[i] = filter.HighlightMulti(truncated, opts.HighlightPatterns, opts.HighlightIgnoreCase)
				} else {
					styledRow[i] = colorizeCell(truncated)
				}
			} else {
				styledRow[i] = truncated
			}
		}
		t.Row(styledRow...)
	}

	_, err := fmt.Fprintln(w, t.String())
	return err
}

func getTerminalWidth() int {
	if isatty.IsTerminal(os.Stdout.Fd()) {
		w, _, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && w > 20 {
			return w
		}
	}
	return 0
}

func allocateColumnWidths(naturalWidths []int, headers []string, termWidth int) []int {
	numCols := len(naturalWidths)
	if numCols == 0 {
		return nil
	}

	// If unconstrained (piped or non-TTY), use natural widths
	if termWidth <= 0 {
		return naturalWidths
	}

	// Table overhead: 1 left border, 1 right border, (numCols - 1) separators, 2 spaces padding per col
	overhead := (numCols * 3) + 1
	totalNatural := overhead
	for _, w := range naturalWidths {
		totalNatural += w
	}

	// Fits comfortably without any truncation
	if totalNatural <= termWidth {
		return naturalWidths
	}

	availableBudget := termWidth - overhead
	if availableBudget <= 0 {
		return naturalWidths
	}

	minWidths := make([]int, numCols)
	totalMin := 0
	for j := 0; j < numCols; j++ {
		hLen := 4
		if j < len(headers) {
			hLen = lipgloss.Width(headers[j])
		}
		if hLen > 12 {
			minWidths[j] = 10
		} else if hLen < 4 {
			minWidths[j] = 4
		} else {
			minWidths[j] = hLen
		}
		totalMin += minWidths[j]
	}

	// In extremely tight layouts, enforce minimum readable floor
	if totalMin >= availableBudget {
		return minWidths
	}

	// Distribute remaining flexible budget proportionally
	remainingBudget := availableBudget - totalMin
	flexibleDemand := 0
	for j := 0; j < numCols; j++ {
		if naturalWidths[j] > minWidths[j] {
			flexibleDemand += (naturalWidths[j] - minWidths[j])
		}
	}

	allocated := make([]int, numCols)
	for j := 0; j < numCols; j++ {
		if flexibleDemand == 0 || naturalWidths[j] <= minWidths[j] {
			allocated[j] = minWidths[j]
		} else {
			extra := int(float64(remainingBudget) * float64(naturalWidths[j]-minWidths[j]) / float64(flexibleDemand))
			allocated[j] = minWidths[j] + extra
		}
	}

	return allocated
}

func truncateString(s string, maxWidth int) string {
	if maxWidth <= 1 {
		return "…"
	}
	targetWidth := maxWidth - 1
	var sb strings.Builder
	currentWidth := 0

	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if currentWidth+rw > targetWidth {
			break
		}
		sb.WriteRune(r)
		currentWidth += rw
	}
	sb.WriteString("…")
	return sb.String()
}

func truncateCellWithMatch(s string, maxWidth int, patterns []string, ignoreCase bool) string {
	if maxWidth <= 1 {
		return "…"
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}

	// Check if any search pattern matches inside this cell
	matchIdx := -1
	for _, p := range patterns {
		if colonIdx := strings.Index(p, ":"); colonIdx > 0 {
			p = p[colonIdx+1:]
		}
		if p == "" {
			continue
		}
		if ignoreCase {
			idx := strings.Index(strings.ToLower(s), strings.ToLower(p))
			if idx >= 0 && (matchIdx == -1 || idx < matchIdx) {
				matchIdx = idx
			}
		} else {
			idx := strings.Index(s, p)
			if idx >= 0 && (matchIdx == -1 || idx < matchIdx) {
				matchIdx = idx
			}
		}
	}

	// If match starts further into the string, shift visible window to keep match visible
	if matchIdx > maxWidth-5 {
		start := matchIdx - 3
		if start < 0 {
			start = 0
		}
		window := s[start:]
		return "…" + truncateString(window, maxWidth-1)
	}

	return truncateString(s, maxWidth)
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
