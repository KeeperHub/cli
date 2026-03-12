package output

import (
	"fmt"

	"github.com/itchyny/gojq"
)

// ApplyJQFilter runs the given jq expression against data in-process (no external binary).
// If the expression produces exactly one result, it is returned directly.
// If it produces multiple results, they are returned as []any.
func ApplyJQFilter(expr string, data any) (any, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("jq parse error: %w", err)
	}

	code, err := gojq.Compile(query)
	if err != nil {
		return nil, fmt.Errorf("jq compile error: %w", err)
	}

	iter := code.Run(data)
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
