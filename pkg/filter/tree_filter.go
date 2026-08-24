package filter

import (
	"fmt"
	"strings"
)

// PruneJSON recursively filters an arbitrary JSON structure down to only branches matching the patterns.
func PruneJSON(v any, opts MultiOptions) (any, bool) {
	if len(opts.Patterns) == 0 {
		return v, true
	}

	matchers, err := BuildPatternMatchers(opts.Patterns, opts.IgnoreCase)
	if err != nil {
		return v, false
	}

	numPatterns := len(matchers)
	fullMask := (uint64(1) << numPatterns) - 1

	result, _, matched := pruneRecursiveMulti(v, "", matchers, opts.Strict, fullMask)
	if opts.InvertMatch {
		matched = !matched
	}

	return result, matched
}

func pruneRecursiveMulti(v any, keyContext string, matchers []PatternMatcher, strict bool, fullMask uint64) (any, uint64, bool) {
	if v == nil {
		var mask uint64 = 0
		for i, m := range matchers {
			if (m.TargetCol == "" || strings.EqualFold(keyContext, m.TargetCol)) && m.MatcherFunc("null") {
				mask |= (1 << i)
			}
		}
		if strict {
			return nil, mask, (mask == fullMask)
		}
		return nil, mask, (mask != 0)
	}

	switch node := v.(type) {
	case map[string]any:
		filteredMap := make(map[string]any)
		var combinedMask uint64 = 0

		for k, child := range node {
			var keyMask uint64 = 0
			for i, m := range matchers {
				if (m.TargetCol == "" || strings.EqualFold(k, m.TargetCol)) && m.MatcherFunc(k) {
					keyMask |= (1 << i)
				}
			}

			// Recursively prune child
			prunedChild, childMask, childMatched := pruneRecursiveMulti(child, k, matchers, false, fullMask)

			totalBranchMask := keyMask | childMask
			if totalBranchMask != 0 {
				combinedMask |= totalBranchMask
				if childMatched {
					filteredMap[k] = prunedChild
				} else if keyMask != 0 {
					filteredMap[k] = child
				}
			}
		}

		if strict {
			if combinedMask == fullMask {
				return filteredMap, combinedMask, true
			}
			return nil, combinedMask, false
		}

		if combinedMask != 0 {
			return filteredMap, combinedMask, true
		}
		return nil, 0, false

	case []any:
		var filteredSlice []any
		var combinedMask uint64 = 0

		// Detect if slice contains logical records (maps)
		isListOfMaps := true
		for _, item := range node {
			if _, ok := item.(map[string]any); !ok {
				isListOfMaps = false
				break
			}
		}

		for i, item := range node {
			indexKey := fmt.Sprintf("[%d]", i)

			// If list of records and strict mode is active, each record must satisfy all patterns
			itemStrict := strict && isListOfMaps
			prunedItem, itemMask, itemMatched := pruneRecursiveMulti(item, indexKey, matchers, itemStrict, fullMask)

			if itemMatched {
				if isListOfMaps {
					filteredSlice = append(filteredSlice, item)
				} else {
					filteredSlice = append(filteredSlice, prunedItem)
				}
				combinedMask |= itemMask
			}
		}

		if len(filteredSlice) > 0 {
			return filteredSlice, combinedMask, true
		}
		return nil, combinedMask, false

	default:
		valStr := fmt.Sprintf("%v", node)
		var mask uint64 = 0
		for i, m := range matchers {
			colMatches := (m.TargetCol == "" || strings.EqualFold(keyContext, m.TargetCol))
			if colMatches && m.MatcherFunc(valStr) {
				mask |= (1 << i)
			}
		}

		if strict {
			return node, mask, (mask == fullMask)
		}
		return node, mask, (mask != 0)
	}
}
