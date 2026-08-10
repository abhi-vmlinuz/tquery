package filter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tquery/tquery/pkg/parser"
)

var (
	// MatchHighlightStyle renders search matches in bold hot-magenta badge
	MatchHighlightStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#E11D48"))
)

type MultiOptions struct {
	Patterns    []string
	InvertMatch bool
	IgnoreCase  bool
	Strict      bool
}

type PatternMatcher struct {
	Original    string
	Pattern     string
	TargetCol   string
	MatcherFunc func(string) bool
	IsRegex     bool
}

// Filter applies single or multi-pattern filtering across a DataStructure.
func Filter(ds *parser.DataStructure, opts MultiOptions) (*parser.DataStructure, error) {
	if len(opts.Patterns) == 0 || ds == nil || len(ds.Rows) == 0 {
		return ds, nil
	}

	matchers, err := BuildPatternMatchers(opts.Patterns, opts.IgnoreCase)
	if err != nil {
		return nil, err
	}

	numPatterns := len(matchers)
	fullMask := (uint64(1) << numPatterns) - 1

	// Map column names to header indices
	colIndices := make([]int, numPatterns)
	for pIdx, m := range matchers {
		colIndices[pIdx] = -1
		if m.TargetCol != "" {
			for i, h := range ds.Headers {
				if strings.EqualFold(h, m.TargetCol) {
					colIndices[pIdx] = i
					break
				}
			}
		}
	}

	var filteredRows [][]string
	var matchedIndices []int

	for i, row := range ds.Rows {
		var rowMask uint64 = 0

		for pIdx, m := range matchers {
			colIdx := colIndices[pIdx]
			if colIdx >= 0 && colIdx < len(row) {
				if m.MatcherFunc(row[colIdx]) {
					rowMask |= (1 << pIdx)
				}
			} else {
				for _, cell := range row {
					if m.MatcherFunc(cell) {
						rowMask |= (1 << pIdx)
						break
					}
				}
			}
		}

		matched := false
		if opts.Strict {
			matched = (rowMask == fullMask)
		} else {
			matched = (rowMask != 0)
		}

		if opts.InvertMatch {
			matched = !matched
		}

		if matched {
			filteredRows = append(filteredRows, row)
			matchedIndices = append(matchedIndices, i)
		}
	}

	newDS := &parser.DataStructure{
		Raw:     ds.Raw,
		Type:    ds.Type,
		Headers: ds.Headers,
		Rows:    filteredRows,
	}

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

// BuildPatternMatchers parses pattern strings into independent matchers.
func BuildPatternMatchers(patterns []string, ignoreCase bool) ([]PatternMatcher, error) {
	matchers := make([]PatternMatcher, len(patterns))
	for i, p := range patterns {
		original := p
		targetCol := ""
		pattern := p

		if colonIdx := strings.Index(p, ":"); colonIdx > 0 {
			targetCol = p[:colonIdx]
			pattern = p[colonIdx+1:]
		}

		matcherFunc, err := buildMatcher(pattern, ignoreCase)
		if err != nil {
			return nil, fmt.Errorf("invalid grep pattern %q: %w", p, err)
		}

		isRegex := strings.ContainsAny(pattern, `^$.*+?()[]{}|\`)
		matchers[i] = PatternMatcher{
			Original:    original,
			Pattern:     pattern,
			TargetCol:   targetCol,
			MatcherFunc: matcherFunc,
			IsRegex:     isRegex,
		}
	}
	return matchers, nil
}

// HighlightMulti highlights all patterns in text using ANSI colors, preserving original casing.
func HighlightMulti(val string, patterns []string, ignoreCase bool) string {
	if len(patterns) == 0 || val == "" {
		return val
	}

	result := val
	for _, p := range patterns {
		result = Highlight(result, p, ignoreCase)
	}
	return result
}

// Highlight highlights occurrences of a single pattern in text using ANSI colors.
func Highlight(val string, pattern string, ignoreCase bool) string {
	if pattern == "" || val == "" {
		return val
	}

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
