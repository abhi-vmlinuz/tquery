package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tquery/tquery/pkg/parser"
)

var (
	// MatchHighlightStyle renders search matches in bold hot-magenta with background
	MatchHighlightStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#E11D48")) // Vibrant red/pink highlight badge
)

type Options struct {
	Pattern     string
	InvertMatch bool
	IgnoreCase  bool
}

// Filter applies regex or literal pattern filtering across a DataStructure's rows and unwrapped data.
func Filter(ds *parser.DataStructure, opts Options) (*parser.DataStructure, error) {
	if opts.Pattern == "" || ds == nil || len(ds.Rows) == 0 {
		return ds, nil
	}

	pattern := opts.Pattern
	targetCol := ""

	// Check for column-scoped search like "status:Running" or "id:MiniMax"
	if colonIdx := strings.Index(pattern, ":"); colonIdx > 0 {
		potentialCol := pattern[:colonIdx]
		potentialPattern := pattern[colonIdx+1:]

		for _, h := range ds.Headers {
			if strings.EqualFold(h, potentialCol) {
				targetCol = h
				pattern = potentialPattern
				break
			}
		}
	}

	// Prepare Matcher: regex or fast literal substring
	matcher, err := buildMatcher(pattern, opts.IgnoreCase)
	if err != nil {
		return nil, fmt.Errorf("invalid grep pattern %q: %w", pattern, err)
	}

	colIdx := -1
	if targetCol != "" {
		for i, h := range ds.Headers {
			if h == targetCol {
				colIdx = i
				break
			}
		}
	}

	var filteredRows [][]string
	var matchedIndices []int

	for i, row := range ds.Rows {
		matched := false

		if colIdx >= 0 && colIdx < len(row) {
			matched = matcher(row[colIdx])
		} else {
			// Scan all cells in the row
			for _, cell := range row {
				if matcher(cell) {
					matched = true
					break
				}
			}
		}

		if opts.InvertMatch {
			matched = !matched
		}

		if matched {
			filteredRows = append(filteredRows, row)
			matchedIndices = append(matchedIndices, i)
		}
	}

	// Build new DataStructure with filtered rows
	newDS := &parser.DataStructure{
		Raw:     ds.Raw,
		Type:    ds.Type,
		Headers: ds.Headers,
		Rows:    filteredRows,
	}

	// Filter Unwrapped slice if applicable
	if slice, ok := ds.Unwrapped.([]any); ok {
		var newSlice []any
		for _, idx := range matchedIndices {
			if idx < len(slice) {
				newSlice = append(newSlice, slice[idx])
			}
		}
		newDS.Unwrapped = newSlice
	} else {
		newDS.Unwrapped = ds.Unwrapped
	}

	return newDS, nil
}

// Highlight highlights occurrences of pattern in text using ANSI colors.
func Highlight(val string, pattern string, ignoreCase bool) string {
	if pattern == "" || val == "" {
		return val
	}

	// Strip column prefix if any (e.g. "id:google" -> "google")
	if colonIdx := strings.Index(pattern, ":"); colonIdx > 0 {
		pattern = pattern[colonIdx+1:]
	}

	regexPattern := pattern
	if !strings.ContainsAny(pattern, `^$.*+?()[]{}|\`) {
		regexPattern = regexp.QuoteMeta(pattern)
	}

	if ignoreCase && !strings.HasPrefix(regexPattern, "(?i)") {
		regexPattern = "(?i)" + regexPattern
	}

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return val
	}

	return re.ReplaceAllStringFunc(val, func(match string) string {
		return MatchHighlightStyle.Render(match)
	})
}

func buildMatcher(pattern string, ignoreCase bool) (func(string) bool, error) {
	isRegex := strings.ContainsAny(pattern, `^$.*+?()[]{}|\`)

	if !isRegex {
		// Fast path: literal substring match
		if ignoreCase {
			lowerPattern := strings.ToLower(pattern)
			return func(s string) bool {
				return strings.Contains(strings.ToLower(s), lowerPattern)
			}, nil
		}
		return func(s string) bool {
			return strings.Contains(s, pattern)
		}, nil
	}

	// Regex path with RE2 linear-time engine
	regexPattern := pattern
	if ignoreCase && !strings.HasPrefix(pattern, "(?i)") {
		regexPattern = "(?i)" + regexPattern
	}

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}

	return func(s string) bool {
		return re.MatchString(s)
	}, nil
}
