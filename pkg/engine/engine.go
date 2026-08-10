package engine

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/itchyny/gojq"
)

// Evaluate executes a JQ query string against a parsed JSON structure (or raw interface/JSON bytes).
func Evaluate(queryStr string, input any) (any, error) {
	queryStr = strings.TrimSpace(queryStr)
	if queryStr == "" || queryStr == "." {
		return input, nil
	}

	query, err := gojq.Parse(queryStr)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query %q: %w", queryStr, err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("compile error for jq query %q: %w", queryStr, err)
	}

	var results []any
	iter := code.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			return nil, fmt.Errorf("query execution error: %w", err)
		}
		results = append(results, v)
	}

	if len(results) == 0 {
		return nil, nil
	}
	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}

// EvaluateRaw executes a JQ query against raw JSON bytes.
func EvaluateRaw(queryStr string, rawJSON []byte) ([]byte, any, error) {
	var input any
	if err := json.Unmarshal(rawJSON, &input); err != nil {
		return nil, nil, fmt.Errorf("failed to parse json for evaluation: %w", err)
	}

	result, err := Evaluate(queryStr, input)
	if err != nil {
		return nil, nil, err
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode query output: %w", err)
	}

	return resultBytes, result, nil
}
