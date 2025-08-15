package jsontest

import (
	"errors"
	"fmt"
)

// ---------- Public API ----------

// pathKind represents the type of JSON path being processed.
type pathKind int

const (
	plainPath pathKind = iota // Plain GJSON path without special syntax
	pipedPath                 // Path with pipe functions (e.g., "path|func()")
	arrayPath                 // Array iteration path (e.g., "arr.[].subpath")
)

// TestJSON asserts JSON content against declarative checks and returns an aggregated error.
// Keep it small: classify the path, dispatch to a focused handler, accumulate errors.
func TestJSON(data []byte, checks map[string]any) (err error) {
	var errs []error

	for path, expected := range checks {
		jt := newJSONTest(path, jtArgs{
			data:     data,
			expected: expected,
		})

		switch jt.kind {
		case arrayPath: // "arr.[].subpath"
			errs = append(errs, jt.handleArray(path))

		case pipedPath: // "base|json()|sub.path|..."
			errs = append(errs, jt.handlePiped(path))

		case plainPath: // plain GJSON path
			errs = append(errs, jt.handlePlain(path))

		default:
			errs = append(errs, fmt.Errorf("unhandled path kind for %q", path))
		}
	}

	return errors.Join(errs...)
}
