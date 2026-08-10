package parser

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
)

// DataStructure represents the parsed dataset categorized for rendering.
type DataStructure struct {
	Raw       any
	Unwrapped any
	Type      Kind
	Headers   []string
	Rows      [][]string
}

type Kind string

const (
	KindTable    Kind = "table"
	KindKeyValue Kind = "key_value"
	KindList     Kind = "list"
	KindValue    Kind = "primitive"
)

// Parse parses raw JSON input into a DataStructure.
func Parse(data []byte, autoUnwrap bool) (*DataStructure, error) {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid json input: %w", err)
	}

	target := raw
	if autoUnwrap {
		target = UnwrapRoot(raw)
	}

	ds := &DataStructure{
		Raw:       raw,
		Unwrapped: target,
	}

	switch v := target.(type) {
	case []any:
		if len(v) == 0 {
			ds.Type = KindList
			ds.Headers = []string{}
			ds.Rows = [][]string{}
			return ds, nil
		}

		// Check if slice of maps (Table)
		isListOfMaps := true
		for _, item := range v {
			if _, ok := item.(map[string]any); !ok {
				isListOfMaps = false
				break
			}
		}

		if isListOfMaps {
			ds.Type = KindTable
			ds.Headers, ds.Rows = mapsToTable(v)
		} else {
			ds.Type = KindList
			ds.Headers = []string{"Index", "Value"}
			ds.Rows = sliceToRows(v)
		}

	case map[string]any:
		ds.Type = KindKeyValue
		ds.Headers = []string{"Key", "Value"}
		ds.Rows = mapToRows(v)

	default:
		ds.Type = KindValue
		ds.Headers = []string{"Value"}
		ds.Rows = [][]string{{formatValue(v)}}
	}

	return ds, nil
}

// UnwrapRoot intelligently unwraps wrapper objects like {"data": [...]}, {"items": [...]}, etc.
func UnwrapRoot(v any) any {
	m, ok := v.(map[string]any)
	if !ok {
		return v
	}

	// Priority wrapper key names commonly used in APIs
	priorityKeys := []string{"data", "items", "results", "models", "rows", "records", "list", "value"}
	for _, key := range priorityKeys {
		if val, exists := m[key]; exists {
			if _, isSlice := val.([]any); isSlice {
				return val
			}
		}
	}

	// If map has only 1 key and its value is a slice, unwrap it
	if len(m) == 1 {
		for _, val := range m {
			if _, isSlice := val.([]any); isSlice {
				return val
			}
		}
	}

	return v
}

func mapsToTable(items []any) ([]string, [][]string) {
	// Discover all unique keys maintaining order of appearance
	keySeen := make(map[string]bool)
	var keys []string

	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		// Sort keys within item for consistent ordering
		itemKeys := make([]string, 0, len(m))
		for k := range m {
			itemKeys = append(itemKeys, k)
		}
		sort.Strings(itemKeys)

		for _, k := range itemKeys {
			if !keySeen[k] {
				keySeen[k] = true
				keys = append(keys, k)
			}
		}
	}

	rows := make([][]string, len(items))
	for i, item := range items {
		m, _ := item.(map[string]any)
		row := make([]string, len(keys))
		for j, k := range keys {
			val, exists := m[k]
			if !exists {
				row[j] = ""
			} else {
				row[j] = formatValue(val)
			}
		}
		rows[i] = row
	}

	return keys, rows
}

func mapToRows(m map[string]any) [][]string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	rows := make([][]string, len(keys))
	for i, k := range keys {
		rows[i] = []string{k, formatValue(m[k])}
	}
	return rows
}

func sliceToRows(items []any) [][]string {
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = []string{strconv.Itoa(i), formatValue(item)}
	}
	return rows
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case float64:
		// Check if whole number
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case map[string]any, []any:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf("%v", val)
		}
		return string(b)
	default:
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.String {
			return rv.String()
		}
		return fmt.Sprintf("%v", v)
	}
}

// MarshalAny encodes any interface into JSON bytes.
func MarshalAny(v any) ([]byte, error) {
	return json.Marshal(v)
}

