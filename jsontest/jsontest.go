package jsontest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"
)

type jsonTest struct {
	data     []byte
	kind     pathKind
	path     string
	pipes    []string
	expected any
}
type jtArgs struct {
	data     []byte
	expected any
}

func newJSONTest(path string, args jtArgs) *jsonTest {
	jt := &jsonTest{
		path:     path,
		data:     args.data,
		expected: args.expected,
	}
	jt.pipes = jt.splitPipes(path)
	jt.kind = jt.classifyPath(path, jt.pipes)
	return jt
}

func (jt *jsonTest) handlePiped(path string) (err error) {
	val := gjson.GetBytes(jt.data, jt.pipes[0])
	if !val.Exists() {
		err = fmt.Errorf("missing path: %s", jt.pipes[0])
		goto end
	}
	val, err = jt.applyPipesJSON(val)
	if err != nil {
		err = fmt.Errorf("pipe error at %s: %v", path, err)
		goto end
	}
	err = jt.compareResolvedValue(path, val)
end:
	return err
}

func (jt *jsonTest) handlePlain(path string) (err error) {
	val := gjson.GetBytes(jt.data, path)
	if !val.Exists() {
		err = fmt.Errorf("missing path: %s", path)
		goto end
	}
	err = jt.compareResolvedValue(path, val)
end:
	return err
}

// handleArray handles "arr.[].subpath" with ordered and AnyOrder comparisons.
func (jt *jsonTest) handleArray(rawPath string) (err error) {
	var results []gjson.Result
	var ok bool

	idx := strings.Index(rawPath, "[].")
	prefix := rawPath[:idx]
	suffix := rawPath[idx+3:]

	arr := gjson.GetBytes(jt.data, prefix)
	if !arr.Exists() {
		err = fmt.Errorf("missing path: %s", prefix)
		goto end
	}
	if !arr.IsArray() {
		err = fmt.Errorf("path is not an array: %s", prefix)
		goto end
	}

	// AnyOrder via reflection into the expected slice element type.
	_, ok = jt.expected.(anyOrderMarker)
	if ok {
		got := jt.collectArraySubpathAs(arr.Array(), suffix)
		if !jt.elementsMatch(got) {
			return fmt.Errorf("path (any-order) %s: elements don't match - expected %v, got %v", rawPath, jt.expected, got)
		}
		return nil
	}

	// Ordered comparisons for common slice types.
	results = arr.Array()
	switch exp := jt.expected.(type) {
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
		return fmt.Errorf("unsupported expected slice type for [] path=%s type=%T", rawPath, jt.expected)
	}
end:
	return err
}

// classifyPath splits a raw key into (kind, basePath, pipes).
// - arr.[].subpath => map-over
// - base|... => piped
// - otherwise => plain
func (jt *jsonTest) classifyPath(path string, parts []string) pathKind {
	switch {
	case strings.Contains(path, "[]."):
		// TODO This does not support pipes and arrays; we need recursion
		jt.kind = arrayPath
	case len(parts) > 1:
		// TODO This does not support arrays and pipes; we need recursion
		jt.kind = pipedPath
	default:
		jt.kind = plainPath
	}
	return jt.kind
}

// splitPipes splits "base|f1()|sub.path|f2()" into base and []pipes.
// We only support json() as a function; other tokens are treated as relative subpaths.
func (jt *jsonTest) splitPipes(s string) (parts []string) {
	parts = strings.Split(s, "|")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}

// applyPipesJSON applies json() and relative subpaths on a gjson.Result.
// json() expects the current value to be a STRING containing JSON.
func (jt *jsonTest) applyPipesJSON(start gjson.Result) (gjson.Result, error) {
	cur := start
	for _, p := range jt.pipes[1:] {
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

// compareResolvedValue routes to array vs scalar handling (and AnyOrder), reusing helpers.
func (jt *jsonTest) compareResolvedValue(path string, val gjson.Result) error {
	// If an array is returned at this path, handle slice comparisons (AnyOrder or ordered).
	if val.IsArray() {
		items := val.Array()

		// AnyOrder? (expected is a named slice type with marker)
		if _, ok := jt.expected.(anyOrderMarker); ok {
			got := jt.collectArrayAs(items)
			if !jt.elementsMatch(got) {
				return fmt.Errorf("path (any-order) %s: elements don't match - expected %v, got %v", path, jt.expected, got)
			}
			return nil
		}

		// Order-sensitive for common slice types.
		switch exp := jt.expected.(type) {
		case []string:
			got := make([]string, 0, len(items))
			for _, it := range items {
				got = append(got, it.String())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", path, exp, got)
			}
			return nil

		case []int:
			got := make([]int, 0, len(items))
			for _, it := range items {
				got = append(got, int(it.Int()))
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", path, exp, got)
			}
			return nil

		case []int64:
			got := make([]int64, 0, len(items))
			for _, it := range items {
				got = append(got, it.Int())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", path, exp, got)
			}
			return nil

		case []float64:
			got := make([]float64, 0, len(items))
			for _, it := range items {
				got = append(got, it.Float())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", path, exp, got)
			}
			return nil

		case []bool:
			got := make([]bool, 0, len(items))
			for _, it := range items {
				got = append(got, it.Bool())
			}
			if !reflect.DeepEqual(exp, got) {
				return fmt.Errorf("path %s: expected %v, got %v", path, exp, got)
			}
			return nil

		default:
			return fmt.Errorf("unsupported expected slice type for array path=%s type=%T", path, jt.expected)
		}
	}

	// Scalar path: coerce only when necessary, then compare.
	actual := jt.coerceToTypeSmart(val)
	if !reflect.DeepEqual(jt.expected, actual) {
		return fmt.Errorf("path %s: expected %v, got %v", path, jt.expected, actual)
	}
	return nil
}

// collectArrayAs converts gjson array items into the same element type as `expected` (AnyOrderSlice[T]).
func (jt *jsonTest) collectArrayAs(items []gjson.Result) any {
	expT := reflect.TypeOf(jt.expected) // AnyOrderSlice[T]
	elemT := expT.Elem()
	gotV := reflect.MakeSlice(expT, 0, len(items))

	for _, it := range items {
		gotV = reflect.Append(gotV, jt.convertGJSONTo(it, elemT))
	}
	return gotV.Interface()
}

// collectArraySubpathAs converts gjson array items' subpath into the same element type as `expected`.
func (jt *jsonTest) collectArraySubpathAs(items []gjson.Result, subpath string) any {
	expT := reflect.TypeOf(jt.expected) // AnyOrderSlice[T]
	elemT := expT.Elem()
	gotV := reflect.MakeSlice(expT, 0, len(items))

	for _, it := range items {
		sub := gjson.Get(it.Raw, subpath)
		gotV = reflect.Append(gotV, jt.convertGJSONTo(sub, elemT))
	}
	return gotV.Interface()
}

// convertGJSONTo converts a gjson.Result into a reflect.Value of elemT.
// Avoid identity conversions: if the native kind already matches elemT, use it directly.
func (jt *jsonTest) convertGJSONTo(v gjson.Result, elemT reflect.Type) reflect.Value {
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
func (jt *jsonTest) coerceToTypeSmart(val gjson.Result) any {
	switch exp := jt.expected.(type) {
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

// elementsMatch checks if two slices contain the same elements regardless of order
func (jt *jsonTest) elementsMatch(actual any) bool {
	expVal := reflect.ValueOf(jt.expected)
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
