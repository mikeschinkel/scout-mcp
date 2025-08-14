package jsontest

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
)

// ---- Any-order generic wrapper ----

type anyOrderMarker interface{ anyOrder() }

// AnyOrderSlice is a named slice type used to signal order-insensitive comparison.
type AnyOrderSlice[T comparable] []T

func (AnyOrderSlice[T]) anyOrder() {}

// AnyOrder constructs an AnyOrderSlice[T] from values.
// Use in CheckJSON expected map to mean "order doesn't matter".
func AnyOrder[T comparable](vals ...T) AnyOrderSlice[T] { return AnyOrderSlice[T](vals) }

// ---- Main helper ----

// TestJSON runs a set of path -> expected assertions against a JSON payload.
//
// Supported path features:
// - GJSON paths (e.g. "result.content.0.type")
// - Length operator: "#", e.g. "result.content.#" == 2
// - Pipe to re-parse JSON stored in a string field: "outer.path|inner.path"
// - Map-over-array: "arrayPath.[].subpath" collects subpath for each array element
//
// Comparison rules:
// - Scalars: assert.Equal
// - Slices collected via "[]":
//   - []T: order-sensitive assert.Equal
//   - AnyOrderSlice[T]: order-insensitive assert.ElementsMatch
func TestJSON(t *testing.T, data []byte, checks map[string]any) (err error) {
	var errs []error
	t.Helper()

	for rawPath, expected := range checks {
		// Case 1: pipe "outer|inner" — parse JSON stored as string
		// NEW: multi-pipe handling with json() and subpaths
		basePath, pipes := splitPipes(rawPath)
		if len(pipes) > 0 {
			val := gjson.GetBytes(data, basePath)
			if !val.Exists() {
				errs = append(errs, fmt.Errorf("missing path: %s", basePath))
				continue
			}
			val, err = applyPipesJSON(val, pipes)
			if err != nil {
				errs = append(errs, fmt.Errorf("pipe error at %s: %v", rawPath, err))
				continue
			}

			// From here, behave as if rawPath resolved to val2.
			// If your code has a common “compare scalar/array” branch,
			// call into it using val2. Example:

			// Arrays:
			if val.IsArray() {
				// (reuse your existing array comparison logic here)
				// e.g., AnyOrderSlice handling or []string/[]int/… comparisons
				// if mismatch: errs = append(errs, fmt.Errorf("..."))
				continue
			}

			// Scalars:
			actual := coerceToType(expected, val)
			if !isEqual(expected, actual) {
				errs = append(errs, fmt.Errorf("path %s: expected %v, got %v", rawPath, expected, actual))
			}
			continue
		}

		// Case 2: "arrayPath.[].subpath" — map-over-array
		if idx := strings.Index(rawPath, "[]."); idx != -1 {
			prefix := rawPath[:idx]   // array
			suffix := rawPath[idx+3:] // subpath inside each element

			arr := gjson.GetBytes(data, prefix)
			if !arr.Exists() {
				errs = append(errs, fmt.Errorf("missing path: %s", prefix))
				continue
			}
			if !arr.IsArray() {
				errs = append(errs, fmt.Errorf("path is not an array: %s", prefix))
				continue
			}

			// If expected is AnyOrderSlice[T], do order-insensitive via reflection.
			if _, ok := expected.(anyOrderMarker); ok {
				expV := reflect.ValueOf(expected)
				expT := expV.Type() // AnyOrderSlice[T]
				elemT := expT.Elem()

				gotV := reflect.MakeSlice(expT, 0, len(arr.Array()))
				for _, item := range arr.Array() {
					sub := gjson.Get(item.Raw, suffix)
					if !sub.Exists() {
						errs = append(errs, fmt.Errorf("missing subpath %q inside %s", suffix, prefix))
						continue
					}

					var converted reflect.Value
					switch elemT.Kind() {
					case reflect.String:
						converted = reflect.ValueOf(sub.String())
					case reflect.Bool:
						converted = reflect.ValueOf(sub.Bool())
					case reflect.Int:
						converted = reflect.ValueOf(int(sub.Int()))
					case reflect.Int8:
						converted = reflect.ValueOf(int8(sub.Int()))
					case reflect.Int16:
						converted = reflect.ValueOf(int16(sub.Int()))
					case reflect.Int32:
						converted = reflect.ValueOf(int32(sub.Int()))
					case reflect.Int64:
						converted = reflect.ValueOf(sub.Int())
					case reflect.Float32:
						converted = reflect.ValueOf(float32(sub.Float()))
					case reflect.Float64:
						converted = reflect.ValueOf(sub.Float())
					default:
						errs = append(errs, fmt.Errorf("unsupported AnyOrder element kind=%s path=%s", elemT.Kind(), rawPath))
						continue
					}
					gotV = reflect.Append(gotV, converted.Convert(elemT))
				}

				if !isElementsMatch(expected, gotV.Interface()) {
					errs = append(errs, fmt.Errorf("path (any-order) %s: elements don't match - expected %v, got %v", rawPath, expected, gotV.Interface()))
				}
				continue
			}

			// Otherwise, do order-sensitive comparisons for common slice types.
			results := arr.Array()
			switch exp := expected.(type) {
			case []string:
				got := make([]string, 0, len(results))
				hasError := false
				for _, item := range results {
					sub := gjson.Get(item.Raw, suffix)
					if !sub.Exists() {
						errs = append(errs, fmt.Errorf("missing subpath %q inside %s", suffix, prefix))
						hasError = true
						continue
					}
					got = append(got, sub.String())
				}
				if !hasError && !isEqual(exp, got) {
					errs = append(errs, fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got))
				}

			case []int:
				got := make([]int, 0, len(results))
				hasError := false
				for _, item := range results {
					sub := gjson.Get(item.Raw, suffix)
					if !sub.Exists() {
						errs = append(errs, fmt.Errorf("missing subpath %q inside %s", suffix, prefix))
						hasError = true
						continue
					}
					got = append(got, int(sub.Int()))
				}
				if !hasError && !isEqual(exp, got) {
					errs = append(errs, fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got))
				}

			case []int64:
				got := make([]int64, 0, len(results))
				hasError := false
				for _, item := range results {
					sub := gjson.Get(item.Raw, suffix)
					if !sub.Exists() {
						errs = append(errs, fmt.Errorf("missing subpath %q inside %s", suffix, prefix))
						hasError = true
						continue
					}
					got = append(got, sub.Int())
				}
				if !hasError && !isEqual(exp, got) {
					errs = append(errs, fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got))
				}

			case []float64:
				got := make([]float64, 0, len(results))
				hasError := false
				for _, item := range results {
					sub := gjson.Get(item.Raw, suffix)
					if !sub.Exists() {
						errs = append(errs, fmt.Errorf("missing subpath %q inside %s", suffix, prefix))
						hasError = true
						continue
					}
					got = append(got, sub.Float())
				}
				if !hasError && !isEqual(exp, got) {
					errs = append(errs, fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got))
				}

			case []bool:
				got := make([]bool, 0, len(results))
				hasError := false
				for _, item := range results {
					sub := gjson.Get(item.Raw, suffix)
					if !sub.Exists() {
						errs = append(errs, fmt.Errorf("missing subpath %q inside %s", suffix, prefix))
						hasError = true
						continue
					}
					got = append(got, sub.Bool())
				}
				if !hasError && !isEqual(exp, got) {
					errs = append(errs, fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got))
				}

			default:
				errs = append(errs, fmt.Errorf("unsupported expected slice type for [] path=%s type=%T", rawPath, expected))
			}
			continue
		}

		// Case 3: regular scalar path
		val := gjson.GetBytes(data, rawPath)
		if !val.Exists() {
			errs = append(errs, fmt.Errorf("missing path: %s", rawPath))
			continue
		}

		if actual := coerceToType(expected, val); !isEqual(expected, actual) {
			errs = append(errs, fmt.Errorf("path %s: expected %v, got %v", rawPath, expected, actual))
		}
	}
	return errors.Join(errs...)
}

// coerceToType converts gjson value into the same Go type as expected (for clean assert.Equal diffs).
func coerceToType(expected any, val gjson.Result) any {
	switch expected.(type) {
	case int:
		return int(val.Int())
	case int64:
		return val.Int()
	case float64:
		return val.Float()
	case bool:
		return val.Bool()
	default:
		return val.String()
	}
}

// isEqual performs deep equality check between two values
func isEqual(expected, actual any) bool {
	return reflect.DeepEqual(expected, actual)
}

// isElementsMatch checks if two slices contain the same elements regardless of order
func isElementsMatch(expected, actual any) bool {
	expVal := reflect.ValueOf(expected)
	actVal := reflect.ValueOf(actual)

	if expVal.Kind() != reflect.Slice || actVal.Kind() != reflect.Slice {
		return false
	}

	if expVal.Len() != actVal.Len() {
		return false
	}

	// Create maps to count occurrences
	expCounts := make(map[interface{}]int)
	actCounts := make(map[interface{}]int)

	for i := 0; i < expVal.Len(); i++ {
		elem := expVal.Index(i).Interface()
		expCounts[elem]++
	}

	for i := 0; i < actVal.Len(); i++ {
		elem := actVal.Index(i).Interface()
		actCounts[elem]++
	}

	return reflect.DeepEqual(expCounts, actCounts)
}

// splitPipes splits "base|f1()|sub.path|f2()" into base and []pipes.
func splitPipes(s string) (base string, pipes []string) {
	parts := strings.Split(s, "|")
	base = strings.TrimSpace(parts[0])
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p != "" {
			pipes = append(pipes, p)
		}
	}
	return
}

// applyPipesJSON applies only json()-pipe and relative subpaths (no scalars).
// - "json()" expects the current value to be a STRING containing JSON.
// - other tokens are treated as GJSON subpaths relative to current.
// Returns the final gjson.Result or an error.
func applyPipesJSON(start gjson.Result, pipes []string) (gjson.Result, error) {
	cur := start
	for _, p := range pipes {
		switch p {
		case "json()":
			// parse current string as JSON and make it the new context
			var inner any
			if err := json.Unmarshal([]byte(cur.String()), &inner); err != nil {
				return gjson.Result{}, fmt.Errorf("json(): failed to parse string as JSON: %w", err)
			}
			b, _ := json.Marshal(inner)
			cur = gjson.ParseBytes(b)
		default:
			// treat as a relative subpath
			next := gjson.Get(cur.Raw, p)
			if !next.Exists() {
				return gjson.Result{}, fmt.Errorf("missing subpath after pipe: %s", p)
			}
			cur = next
		}
	}
	return cur, nil
}
