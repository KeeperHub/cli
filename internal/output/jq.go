package output

import (
	"encoding/json"
	"fmt"

	"github.com/itchyny/gojq"
)

// normalizeForJQ converts data to gojq-compatible form via JSON round-trip.
// gojq panics on typed Go structs; JSON round-trip produces only maps, slices, and primitives.
func normalizeForJQ(data any) (any, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("marshalling data for jq: %w", err)
	}
	var normalized any
	if err := json.Unmarshal(b, &normalized); err != nil {
		return nil, fmt.Errorf("normalising data for jq: %w", err)
	}
	return normalized, nil
}

// ApplyJQFilter runs the given jq expression against data in-process (no external binary).
// If the expression produces exactly one result, it is returned directly.
// If it produces multiple results, they are returned as []any.
func ApplyJQFilter(expr string, data any) (any, error) {
	normalized, err := normalizeForJQ(data)
	if err != nil {
		return nil, err
	}

	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("jq parse error: %w", err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("jq compile error: %w", err)
	}

	iter := code.Run(normalized)
	var results []any
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if iterErr, isErr := v.(error); isErr {
			return nil, iterErr
		}
		results = append(results, v)
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return results, nil
}
