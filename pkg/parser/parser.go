package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

// DecodeStream decodes single JSON documents, JSON arrays, and multi-document streams (JSONL / NDJSON).
func DecodeStream(data []byte) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var docs []any

	for {
		var doc any
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid json stream: %w", err)
		}
		docs = append(docs, doc)
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("empty json input")
	}

	if len(docs) == 1 {
		return docs[0], nil
	}

	// Multiple JSON documents: normalize into a unified slice of records
	return docs, nil
}

// Parse parses raw JSON input (single object, array, or NDJSON stream) into a DataStructure.
func Parse(data []byte, autoUnwrap bool) (*DataStructure, error) {
	raw, err := DecodeStream(data)
	if err != nil {
		return nil, err
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
		ds.Rows = [][]string{{FormatValue(v)}}
	}

	return ds, nil
}

// IsComplexStructure detects if JSON data is a single complex object better suited for Tree view.
func IsComplexStructure(v any) bool {
	switch val := v.(type) {
	case map[string]any:
		return len(val) > 4 || hasNestedContainers(val)
	case []any:
		if len(val) == 1 {
			if m, ok := val[0].(map[string]any); ok {
				return len(m) > 4 || hasNestedContainers(m)
			}
		}
		return false
	default:
		return false
	}
}

func hasNestedContainers(m map[string]any) bool {
	for _, child := range m {
		switch child.(type) {
		case map[string]any, []any:
			return true
		}
	}
	return false
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
				row[j] = FormatValue(val)
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
		rows[i] = []string{k, FormatValue(m[k])}
	}
	return rows
}

func sliceToRows(items []any) [][]string {
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = []string{strconv.Itoa(i), FormatValue(item)}
	}
	return rows
}

// FormatValue renders any decoded JSON value as a display string,
// preserving integer representation for whole-number float64s.
func FormatValue(v any) string {
	if v == nil {
		return "null"
	}

	switch val := v.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case float64:
		if val == float64(int64(val)) {
			return strconv.FormatInt(int64(val), 10)
		}
		return strconv.FormatFloat(val, 'f', -1, 64)
	case map[string]any:
		if len(val) == 0 {
			return "{}"
		}
		b, err := json.Marshal(val)
		if err == nil && len(b) <= 30 {
			return string(b)
		}
		return fmt.Sprintf("[object: %d keys]", len(val))
	case []any:
		if len(val) == 0 {
			return "[]"
		}
		b, err := json.Marshal(val)
		if err == nil && len(b) <= 30 {
			return string(b)
		}
		return fmt.Sprintf("[%d items]", len(val))
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

// ResolveColumnIndex resolves a column name or 1-based index to a 0-based header index.
// It also provides smart alias fallback (e.g. "name" matches "id", "model", "title" if "name" does not exist).
func ResolveColumnIndex(headers []string, col string) int {
	colClean := strings.TrimSpace(col)
	if colClean == "" || len(headers) == 0 {
		return -1
	}

	// 1. Exact case-insensitive match
	for i, h := range headers {
		if strings.EqualFold(h, colClean) {
			return i
		}
	}

	// 2. 1-based numeric column index (e.g. "1", "2", "3")
	if num, err := strconv.Atoi(colClean); err == nil {
		if num >= 1 && num <= len(headers) {
			return num - 1
		}
		if num == 0 {
			return 0
		}
	}

	return -1
}

// ProjectColumns filters DataStructure headers and rows to only include specified column names or 1-based indices.
func ProjectColumns(ds *DataStructure, columns []string) *DataStructure {
	if ds == nil || len(columns) == 0 || len(ds.Headers) == 0 {
		return ds
	}

	var selectedIndices []int
	var newHeaders []string

	for _, col := range columns {
		idx := ResolveColumnIndex(ds.Headers, col)
		if idx >= 0 {
			alreadySelected := false
			for _, prevIdx := range selectedIndices {
				if prevIdx == idx {
					alreadySelected = true
					break
				}
			}
			if !alreadySelected {
				selectedIndices = append(selectedIndices, idx)
				newHeaders = append(newHeaders, ds.Headers[idx])
			}
		}
	}

	if len(selectedIndices) == 0 {
		return ds
	}

	newRows := make([][]string, len(ds.Rows))
	for rIdx, row := range ds.Rows {
		newRow := make([]string, len(selectedIndices))
		for i, colIdx := range selectedIndices {
			if colIdx < len(row) {
				newRow[i] = row[colIdx]
			}
		}
		newRows[rIdx] = newRow
	}

	newDS := &DataStructure{
		Raw:     ds.Raw,
		Type:    ds.Type,
		Headers: newHeaders,
		Rows:    newRows,
	}

	// Project unwrapped data if it is a slice of maps
	if slice, ok := ds.Unwrapped.([]any); ok {
		newSlice := make([]any, len(slice))
		for i, item := range slice {
			if m, isMap := item.(map[string]any); isMap {
				projMap := make(map[string]any, len(newHeaders))
				for _, h := range newHeaders {
					for k, v := range m {
						if strings.EqualFold(k, h) {
							projMap[k] = v
							break
						}
					}
				}
				newSlice[i] = projMap
			} else {
				newSlice[i] = item
			}
		}
		newDS.Unwrapped = newSlice
	} else if m, isMap := ds.Unwrapped.(map[string]any); isMap {
		projMap := make(map[string]any)
		for _, h := range newHeaders {
			for k, v := range m {
				if strings.EqualFold(k, h) {
					projMap[k] = v
					break
				}
			}
		}
		newDS.Unwrapped = projMap
	} else {
		newDS.Unwrapped = ds.Unwrapped
	}

	return newDS
}

// SortDataStructure sorts DataStructure rows (and Unwrapped records) by specified column name or index.
func SortDataStructure(ds *DataStructure, colName string, desc bool) *DataStructure {
	if ds == nil || len(ds.Rows) <= 1 {
		return ds
	}

	colIdx := ResolveColumnIndex(ds.Headers, colName)
	if colIdx == -1 {
		return ds
	}

	indices := make([]int, len(ds.Rows))
	for i := range indices {
		indices[i] = i
	}

	sort.SliceStable(indices, func(i, j int) bool {
		idxA := indices[i]
		idxB := indices[j]

		var valA, valB string
		if colIdx < len(ds.Rows[idxA]) {
			valA = ds.Rows[idxA][colIdx]
		}
		if colIdx < len(ds.Rows[idxB]) {
			valB = ds.Rows[idxB][colIdx]
		}

		numA, errA := strconv.ParseFloat(valA, 64)
		numB, errB := strconv.ParseFloat(valB, 64)

		var less bool
		if errA == nil && errB == nil {
			less = numA < numB
		} else {
			less = strings.ToLower(valA) < strings.ToLower(valB)
		}

		if desc {
			return !less && (valA != valB)
		}
		return less
	})

	newRows := make([][]string, len(ds.Rows))
	for i, origIdx := range indices {
		newRows[i] = ds.Rows[origIdx]
	}

	newDS := &DataStructure{
		Raw:     ds.Raw,
		Type:    ds.Type,
		Headers: ds.Headers,
		Rows:    newRows,
	}

	if slice, ok := ds.Unwrapped.([]any); ok && len(slice) == len(ds.Rows) {
		newSlice := make([]any, len(slice))
		for i, origIdx := range indices {
			newSlice[i] = slice[origIdx]
		}
		newDS.Unwrapped = newSlice
	} else {
		newDS.Unwrapped = ds.Unwrapped
	}

	return newDS
}


