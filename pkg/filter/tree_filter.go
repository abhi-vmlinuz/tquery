package filter

import (
	"fmt"
	"strings"
)

// PruneJSON recursively filters an arbitrary JSON structure down to only branches matching the pattern.
func PruneJSON(v any, pattern string, ignoreCase bool, invert bool) (any, bool) {
	if pattern == "" {
		return v, true
	}

	targetCol := ""
	if colonIdx := strings.Index(pattern, ":"); colonIdx > 0 {
		targetCol = pattern[:colonIdx]
		pattern = pattern[colonIdx+1:]
	}

	matcher, err := buildMatcher(pattern, ignoreCase)
	if err != nil {
		return v, false
	}

	result, matched := pruneRecursive(v, "", matcher, targetCol)
	if invert {
		matched = !matched
	}

	return result, matched
}

func pruneRecursive(v any, keyContext string, matcher func(string) bool, targetCol string) (any, bool) {
	if v == nil {
		matches := (targetCol == "" || strings.EqualFold(keyContext, targetCol)) && matcher("null")
		return nil, matches
	}

	switch node := v.(type) {
	case map[string]any:
		filteredMap := make(map[string]any)
		anyMatched := false

		for k, child := range node {
			// If key itself matches pattern
			keyMatches := (targetCol == "" || strings.EqualFold(k, targetCol)) && matcher(k)
			if keyMatches && targetCol == "" {
				filteredMap[k] = child
				anyMatched = true
				continue
			}

			// Check children
			prunedChild, childMatched := pruneRecursive(child, k, matcher, targetCol)
			if childMatched {
				filteredMap[k] = prunedChild
				anyMatched = true
			}
		}

		if anyMatched {
			return filteredMap, true
		}
		return nil, false

	case []any:
		var filteredSlice []any
		anyMatched := false

		for i, item := range node {
			indexKey := fmt.Sprintf("[%d]", i)
			prunedItem, itemMatched := pruneRecursive(item, indexKey, matcher, targetCol)
			if itemMatched {
				filteredSlice = append(filteredSlice, prunedItem)
				anyMatched = true
			}
		}

		if anyMatched {
			return filteredSlice, true
		}
		return nil, false

	default:
		valStr := fmt.Sprintf("%v", node)
		colMatches := targetCol == "" || strings.EqualFold(keyContext, targetCol)
		if colMatches && matcher(valStr) {
			return node, true
		}
		return nil, false
	}
}
