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

// ---------- Public API ----------

// TestJSON asserts JSON content against declarative checks and returns an aggregated error.
// Keep it small: classify the path, dispatch to a focused handler, accumulate errors.
func TestJSON(t *testing.T, data []byte, checks map[string]any) (err error) {
	t.Helper()
	var errs []error

	for rawPath, expected := range checks {
		kind, base, pipes := classifyPath(rawPath)

		switch kind {
		case mapOverPath: // "arr.[].subpath"
			errs = append(errs, handleMapOverArray(data, rawPath, expected))

		case pipedPath: // "base|json()|sub.path|..."
			errs = append(errs, handlePiped(data, base, pipes, rawPath, expected))

		case plainPath: // plain GJSON path
			errs = append(errs, handlePlain(data, base, pipes, rawPath, expected))

		default:
			errs = append(errs, fmt.Errorf("unhandled path kind for %q", rawPath))
		}
	}

	return errors.Join(errs...)
}

// ---------- Classification ----------

type pathKind int

const (
	plainPath pathKind = iota
	pipedPath
	mapOverPath
)

// classifyPath splits a raw key into (kind, basePath, pipes).
// - arr.[].subpath => map-over
// - base|... => piped
// - otherwise => plain
func classifyPath(raw string) (pathKind, string, []string) {
	if strings.Contains(raw, "[].") {
		return mapOverPath, raw, nil
	}
	base, pipes := splitPipes(raw)
	if len(pipes) > 0 {
		return pipedPath, base, pipes
	}
	return plainPath, raw, nil
}

// splitPipes splits "base|f1()|sub.path|f2()" into base and []pipes.
// We only support json() as a function; other tokens are treated as relative subpaths.
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

// ---------- Pipe execution (json only) ----------

// applyPipesJSON applies json() and relative subpaths on a gjson.Result.
// json() expects the current value to be a STRING containing JSON.
func applyPipesJSON(start gjson.Result, pipes []string) (gjson.Result, error) {
	cur := start
	for _, p := range pipes {
		switch p {
		case "json()":
			var inner any
			if err := json.Unmarshal([]byte(cur.String()), &inner); err != nil {
				return gjson.Result{}, fmt.Errorf("json(): failed to parse string as JSON: %w", err)
			}
			b, _ := json.Marshal(inner)
			cur = gjson.ParseBytes(b)
		default:
			next := gjson.Get(cur.Raw, p)
			if !next.Exists() {
				return gjson.Result{}, fmt.Errorf("missing subpath after pipe: %s", p)
			}
			cur = next
		}
	}
	return cur, nil
}

// ---------- Dispatch after value is resolved ----------

// compareResolvedValue routes to array vs scalar handling (and AnyOrder), reusing helpers.
func compareResolvedValue(rawPath string, expected any, val gjson.Result) error {
	// If an array is returned at this path, handle slice comparisons (AnyOrder or ordered).
	if val.IsArray() {
		items := val.Array()

		// AnyOrder? (expected is a named slice type with marker)
		if _, ok := expected.(anyOrderMarker); ok {
			got := collectArrayAs(expected, items)
			if !isElementsMatch(expected, got) {
				return fmt.Errorf("path (any-order) %s: elements don't match - expected %v, got %v", rawPath, expected, got)
			}
			return nil
		}

		// Order-sensitive for common slice types.
		switch exp := expected.(type) {
		case []string:
			got := make([]string, 0, len(items))
			for _, it := range items {
				got = append(got, it.String())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
			}
			return nil

		case []int:
			got := make([]int, 0, len(items))
			for _, it := range items {
				got = append(got, int(it.Int()))
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
			}
			return nil

		case []int64:
			got := make([]int64, 0, len(items))
			for _, it := range items {
				got = append(got, it.Int())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
			}
			return nil

		case []float64:
			got := make([]float64, 0, len(items))
			for _, it := range items {
				got = append(got, it.Float())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
			}
			return nil

		case []bool:
			got := make([]bool, 0, len(items))
			for _, it := range items {
				got = append(got, it.Bool())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
			}
			return nil

		default:
			return fmt.Errorf("unsupported expected slice type for array path=%s type=%T", rawPath, expected)
		}
	}

	// Scalar path: coerce only when necessary, then compare.
	actual := coerceToTypeSmart(expected, val)
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf("path %s: expected %v, got %v", rawPath, expected, actual)
	}
	return nil
}

// ---------- Handlers for specific kinds ----------

func handlePiped(data []byte, base string, pipes []string, rawPath string, expected any) (err error) {
	val := gjson.GetBytes(data, base)
	if !val.Exists() {
		err = fmt.Errorf("missing path: %s", base)
		goto end
	}
	val, err = applyPipesJSON(val, pipes)
	if err != nil {
		err = fmt.Errorf("pipe error at %s: %v", rawPath, err)
		goto end
	}
	err = compareResolvedValue(rawPath, expected, val)
end:
	return err
}

func handlePlain(data []byte, base string, pipes []string, rawPath string, expected any) (err error) {
	val := gjson.GetBytes(data, base)
	if !val.Exists() {
		err = fmt.Errorf("missing path: %s", base)
		goto end
	}
	err = compareResolvedValue(rawPath, expected, val)
end:
	return err
}

// handleMapOverArray handles "arr.[].subpath" with ordered and AnyOrder comparisons.
func handleMapOverArray(data []byte, rawPath string, expected any) error {
	idx := strings.Index(rawPath, "[].")
	prefix := rawPath[:idx]
	suffix := rawPath[idx+3:]

	arr := gjson.GetBytes(data, prefix)
	if !arr.Exists() {
		return fmt.Errorf("missing path: %s", prefix)
	}
	if !arr.IsArray() {
		return fmt.Errorf("path is not an array: %s", prefix)
	}

	// AnyOrder via reflection into the expected slice element type.
	if _, ok := expected.(anyOrderMarker); ok {
		got := collectArraySubpathAs(expected, arr.Array(), suffix)
		if !isElementsMatch(expected, got) {
			return fmt.Errorf("path (any-order) %s: elements don't match - expected %v, got %v", rawPath, expected, got)
		}
		return nil
	}

	// Ordered comparisons for common slice types.
	results := arr.Array()
	switch exp := expected.(type) {
	case []string:
		got := make([]string, 0, len(results))
		for _, item := range results {
			sub := gjson.Get(item.Raw, suffix)
			if !sub.Exists() {
				return fmt.Errorf("missing subpath %q inside %s", suffix, prefix)
			}
			got = append(got, sub.String())
		}
		if !reflect.DeepEqual(exp, got) {
			return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
		}
		return nil

	case []int:
		got := make([]int, 0, len(results))
		for _, item := range results {
			sub := gjson.Get(item.Raw, suffix)
			if !sub.Exists() {
				return fmt.Errorf("missing subpath %q inside %s", suffix, prefix)
			}
			got = append(got, int(sub.Int()))
		}
		if !reflect.DeepEqual(exp, got) {
			return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
		}
		return nil

	case []int64:
		got := make([]int64, 0, len(results))
		for _, item := range results {
			sub := gjson.Get(item.Raw, suffix)
			if !sub.Exists() {
				return fmt.Errorf("missing subpath %q inside %s", suffix, prefix)
			}
			got = append(got, sub.Int())
		}
		if !reflect.DeepEqual(exp, got) {
			return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
		}
		return nil

	case []float64:
		got := make([]float64, 0, len(results))
		for _, item := range results {
			sub := gjson.Get(item.Raw, suffix)
			if !sub.Exists() {
				return fmt.Errorf("missing subpath %q inside %s", suffix, prefix)
			}
			got = append(got, sub.Float())
		}
		if !reflect.DeepEqual(exp, got) {
			return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
		}
		return nil

	case []bool:
		got := make([]bool, 0, len(results))
		for _, item := range results {
			sub := gjson.Get(item.Raw, suffix)
			if !sub.Exists() {
				return fmt.Errorf("missing subpath %q inside %s", suffix, prefix)
			}
			got = append(got, sub.Bool())
		}
		if !reflect.DeepEqual(exp, got) {
			return fmt.Errorf("path %s: expected %v, got %v", rawPath, exp, got)
		}
		return nil

	default:
		return fmt.Errorf("unsupported expected slice type for [] path=%s type=%T", rawPath, expected)
	}
}

// ---------- Conversions & collectors ----------

type anyOrderMarker interface{ anyOrder() }

// AnyOrderSlice is a named slice type used to signal order-insensitive comparison.
type AnyOrderSlice[T comparable] []T

func (AnyOrderSlice[T]) anyOrder() {}

func AnyOrder[T comparable](vals ...T) AnyOrderSlice[T] { return AnyOrderSlice[T](vals) }

// collectArrayAs converts gjson array items into the same element type as `expected` (AnyOrderSlice[T]).
func collectArrayAs(expected any, items []gjson.Result) any {
	expT := reflect.TypeOf(expected) // AnyOrderSlice[T]
	elemT := expT.Elem()
	gotV := reflect.MakeSlice(expT, 0, len(items))

	for _, it := range items {
		gotV = reflect.Append(gotV, convertGJSONTo(it, elemT))
	}
	return gotV.Interface()
}

// collectArraySubpathAs converts gjson array items' subpath into the same element type as `expected`.
func collectArraySubpathAs(expected any, items []gjson.Result, subpath string) any {
	expT := reflect.TypeOf(expected) // AnyOrderSlice[T]
	elemT := expT.Elem()
	gotV := reflect.MakeSlice(expT, 0, len(items))

	for _, it := range items {
		sub := gjson.Get(it.Raw, subpath)
		gotV = reflect.Append(gotV, convertGJSONTo(sub, elemT))
	}
	return gotV.Interface()
}

// convertGJSONTo converts a gjson.Result into a reflect.Value of elemT.
// Avoid identity conversions: if the native kind already matches elemT, use it directly.
func convertGJSONTo(v gjson.Result, elemT reflect.Type) reflect.Value {
	switch elemT.Kind() {
	case reflect.String:
		// gjson already holds native string; no extra conversion needed beyond type cast
		return reflect.ValueOf(v.String()).Convert(elemT)
	case reflect.Bool:
		return reflect.ValueOf(v.Bool()).Convert(elemT)
	case reflect.Int:
		return reflect.ValueOf(int(v.Int())).Convert(elemT)
	case reflect.Int8:
		return reflect.ValueOf(int8(v.Int())).Convert(elemT)
	case reflect.Int16:
		return reflect.ValueOf(int16(v.Int())).Convert(elemT)
	case reflect.Int32:
		return reflect.ValueOf(int32(v.Int())).Convert(elemT)
	case reflect.Int64:
		return reflect.ValueOf(v.Int()).Convert(elemT)
	case reflect.Float32:
		return reflect.ValueOf(float32(v.Float())).Convert(elemT)
	case reflect.Float64:
		return reflect.ValueOf(v.Float()).Convert(elemT)
	default:
		panic("unsupported element kind in AnyOrder slice: " + elemT.Kind().String())
	}
}

// coerceToTypeSmart converts gjson value into the same Go type as expected only when necessary.
// (Avoids redundant string->string / bool->bool / float64->float64 conversions.)
func coerceToTypeSmart(expected any, val gjson.Result) any {
	switch exp := expected.(type) {
	case int:
		return int(val.Int())
	case int64:
		return val.Int()
	case float64:
		return val.Float()
	case bool:
		return val.Bool()
	case string:
		// return exactly val.String(), not an extra conversion
		return val.String()
	default:
		// Unknown expected type; fall back to string representation for deterministic comparison.
		_ = exp
		return val.String()
	}
}

// ---------- Comparators ----------

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

	expCounts := make(map[any]int, expVal.Len())
	actCounts := make(map[any]int, actVal.Len())

	for i := 0; i < expVal.Len(); i++ {
		expCounts[expVal.Index(i).Interface()]++
	}
	for i := 0; i < actVal.Len(); i++ {
		actCounts[actVal.Index(i).Interface()]++
	}
	return reflect.DeepEqual(expCounts, actCounts)
}
