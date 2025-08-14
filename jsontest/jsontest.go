package jsontest

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tidwall/gjson"
)

/* ---------- Core struct ---------- */

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

/* ---------- Helpers: path + pipe engine ---------- */

func (jt *jsonTest) handlePiped(path string) (err error) {
	var val gjson.Result
	val, err = jt.evalBase(jt.pipes[0]) // supports [] in base
	if err != nil {
		// Special case: if next pipe is exists() we allow missing base and evaluate to false
		if len(jt.pipes) > 1 && jt.pipes[1] == "exists()" {
			// create scalar false result to feed the rest of the pipeline
			val = gjson.Parse("false")
			err = nil
		} else {
			goto end
		}
	}

	val, err = jt.applyPipes(val, jt.pipes[1:])
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

/* evalBase resolves the first path token. It supports map-over arrays via recursion. */
func (jt *jsonTest) evalBase(expr string) (res gjson.Result, err error) {
	var cur gjson.Result
	cur = gjson.ParseBytes(jt.data)
	res, err = jt.resolveRelative(cur, expr)
	if err != nil {
		goto end
	}
	if !res.Exists() {
		err = fmt.Errorf("missing path: %s", expr)
		goto end
	}
end:
	return res, err
}

/* resolveRelative resolves expr relative to cur (supports nested "[]." map-over). */
func (jt *jsonTest) resolveRelative(cur gjson.Result, expr string) (res gjson.Result, err error) {
	var out any
	var bytes []byte
	var idx int
	var prefix, suffix string
	var arr gjson.Result
	var items []gjson.Result
	var collected []any

	// Recurse over "[]."
	idx = strings.Index(expr, "[].")
	if idx < 0 {
		// No map-over; just relative get
		res = gjson.Get(cur.Raw, expr)
		goto end
	}

	prefix = expr[:idx]
	suffix = expr[idx+3:]

	arr = gjson.Get(cur.Raw, prefix)
	if !arr.Exists() {
		err = fmt.Errorf("missing path: %s", prefix)
		goto end
	}
	if !arr.IsArray() {
		err = fmt.Errorf("path is not an array: %s", prefix)
		goto end
	}

	items = arr.Array()
	collected = make([]any, 0, len(items))
	for _, it := range items {
		var sub gjson.Result
		sub, err = jt.resolveRelative(it, suffix)
		if err != nil {
			goto end
		}
		// Convert sub into interface{} so we can build a JSON array
		var v any
		if sub.Raw == "" {
			// treat as null
			v = nil
		} else {
			if err = json.Unmarshal([]byte(sub.Raw), &v); err != nil {
				// If Raw is a primitive without quotes (e.g., 123, true), Unmarshal works.
				// If it's a bare string without quotes (shouldn't happen), fall back to String().
				v = sub.Value()
			}
		}
		collected = append(collected, v)
	}
	out = collected

	bytes, _ = json.Marshal(out)
	res = gjson.ParseBytes(bytes)
end:
	return res, err
}

/*
applyPipes executes pipe tokens over the current gjson.Result.

	Supports: json(), exists(), len(), notNull(), notEmpty(), and relative subpaths (with recursion & "[].").
*/
func (jt *jsonTest) applyPipes(start gjson.Result, tokens []string) (res gjson.Result, err error) {
	res = start
	for _, p := range tokens {
		switch p {
		case "json()":
			// expects the current value to be a STRING containing JSON
			var inner any
			if err = json.Unmarshal([]byte(res.String()), &inner); err != nil {
				err = fmt.Errorf("json(): failed to parse string as JSON: %w", err)
				goto end
			}
			b, _ := json.Marshal(inner)
			res = gjson.ParseBytes(b)

		case "exists()":
			if res.Exists() {
				res = gjson.Parse("true")
			} else {
				res = gjson.Parse("false")
			}

		case "notNull()":
			if res.Exists() && strings.TrimSpace(res.Raw) != "null" {
				res = gjson.Parse("true")
			} else {
				res = gjson.Parse("false")
			}

		case "notEmpty()":
			var b bool
			b = jt.isNonEmpty(res)
			if b {
				res = gjson.Parse("true")
			} else {
				res = gjson.Parse("false")
			}

		case "len()":
			var n int
			switch {
			case res.IsArray():
				n = len(res.Array())
			default:
				// object? use Map(); string? use String(); primitives => 0
				m := res.Map()
				if len(m) > 0 {
					n = len(m)
				} else {
					s := res.String()
					n = len(s)
				}
			}
			res = gjson.Parse(fmt.Sprintf("%d", n))

		default:
			// Treat as a relative subpath (may contain "[].")
			// Subpaths after a scalar don’t make sense; detect and error early.
			if isScalar(res) && strings.ContainsAny(p, ".[]#") {
				err = fmt.Errorf("cannot apply subpath %q after scalar value", p)
				goto end
			}
			res, err = jt.resolveRelative(res, p)
			if err != nil {
				goto end
			}
			if !res.Exists() {
				err = fmt.Errorf("missing subpath after pipe: %s", p)
				goto end
			}
		}
	}
end:
	return res, err
}

func isScalar(r gjson.Result) bool {
	// Heuristic: object => Map() not empty or Raw starts with '{'
	// array => IsArray()
	if r.IsArray() {
		return false
	}
	if strings.HasPrefix(strings.TrimSpace(r.Raw), "{") {
		return false
	}
	// numbers, strings, bools, null => scalar
	return true
}

/* ---------- Array handlers (ordered vs any-order) ---------- */

type arrayArgs struct {
	path   string
	prefix string
	suffix string
}

// handleArray handles "arr.[].subpath" with ordered and AnyOrder comparisons.
func (jt *jsonTest) handleArray(path string) (err error) {
	var ok bool
	var args arrayArgs

	idx := strings.Index(path, "[].")
	args = arrayArgs{
		path:   path,
		prefix: path[:idx],
		suffix: path[idx+3:],
	}

	arr := gjson.GetBytes(jt.data, args.prefix)
	if !arr.Exists() {
		err = fmt.Errorf("missing path: %s", args.prefix)
		goto end
	}
	if !arr.IsArray() {
		err = fmt.Errorf("path is not an array: %s", args.prefix)
		goto end
	}
	// AnyOrder via reflection into the expected slice element type.
	_, ok = jt.expected.(anyOrderMarker)
	if ok {
		got := jt.collectArraySubpathAs(arr.Array(), args.suffix)
		if !jt.elementsMatch(got) {
			err = fmt.Errorf("path (any-order) %s: elements don't match - expected %v, got %v", path, jt.expected, got)
		}
		goto end
	}

	// Ordered comparisons for common slice types.

	err = jt.handleTypedArray(args, arr.Array())

end:
	return err
}

// handleTypedArray handles type-specific "arr.[].subpath"
func handleTypedArray[T any](exp []T, results []gjson.Result, args arrayArgs, convertFunc func(gjson.Result) T) (err error) {
	var errs []error
	got := make([]T, 0, len(results))
	for _, item := range results {
		sub := gjson.Get(item.Raw, args.suffix)
		if !sub.Exists() {
			errs = append(errs, fmt.Errorf("missing subpath %q inside %s", args.suffix, args.prefix))
			continue
		}
		got = append(got, convertFunc(sub))
	}
	err = errors.Join(errs...)
	if err != nil {
		goto end
	}
	if !reflect.DeepEqual(exp, got) {
		err = fmt.Errorf("path %s: expected %v, got %v", args.path, exp, got)
		goto end
	}
end:
	return err
}

func (jt *jsonTest) handleTypedArray(args arrayArgs, results []gjson.Result) (err error) {
	switch exp := jt.expected.(type) {
	case []string:
		err = handleTypedArray(exp, results, args, func(result gjson.Result) string {
			return result.String()
		})

	case []int:
		err = handleTypedArray(exp, results, args, func(result gjson.Result) int {
			return int(result.Int())
		})

	case []int64:
		err = handleTypedArray(exp, results, args, func(result gjson.Result) int64 {
			return result.Int()
		})

	case []float64:
		err = handleTypedArray(exp, results, args, func(result gjson.Result) float64 {
			return result.Float()
		})

	case []bool:
		err = handleTypedArray(exp, results, args, func(result gjson.Result) bool {
			return result.Bool()
		})
	default:
		err = fmt.Errorf("unsupported expected slice type for [] path=%s type=%T", args.path, jt.expected)
	}
	return err
}

/* ---------- Pipes: split & classify ---------- */

// classifyPath splits a raw key into (kind, basePath, pipes).
// - arr.[].subpath => map-over
// - base|... => piped
// - otherwise => plain
func (jt *jsonTest) classifyPath(path string, parts []string) pathKind {
	switch {
	case strings.Contains(path, "[]."):
		jt.kind = arrayPath
	case len(parts) > 1:
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

/* ---------- Comparison ---------- */

func (jt *jsonTest) compareResolvedValue(path string, val gjson.Result) (err error) {
	var actual any
	var items []gjson.Result

	// Marker: NotNull
	if _, ok := jt.expected.(NotNull); ok {
		if !val.Exists() || strings.TrimSpace(val.Raw) == "null" {
			err = fmt.Errorf("path %s: expected not null", path)
		}
		goto end
	}

	// Marker: NotEmpty
	if _, ok := jt.expected.(NotEmpty); ok {
		if !jt.isNonEmpty(val) {
			err = fmt.Errorf("path %s: expected not empty", path)
		}
		goto end
	}

	// Array handling
	if !val.IsArray() {
		// Scalar path: coerce only when necessary, then compare.
		actual = jt.coerceToTypeSmart(val)
		if !reflect.DeepEqual(jt.expected, actual) {
			err = fmt.Errorf("path %s: expected %v, got %v", path, jt.expected, actual)
		}
		goto end
	}

	items = val.Array()

	// AnyOrder? (expected is a named slice type with marker)
	if _, ok := jt.expected.(anyOrderMarker); ok {
		got := jt.collectArrayAs(items)
		if !jt.elementsMatch(got) {
			err = fmt.Errorf("path (any-order) %s: elements don't match - expected %v, got %v", path, jt.expected, got)
		}
		goto end
	}

	// Order-sensitive for common slice types.
	switch exp := jt.expected.(type) {
	case []string:
		err = compareTypedResolvedValue(exp, items, path, func(result gjson.Result) string {
			return result.String()
		})
	case []int:
		err = compareTypedResolvedValue(exp, items, path, func(result gjson.Result) int {
			return int(result.Int())
		})
	case []int64:
		err = compareTypedResolvedValue(exp, items, path, func(result gjson.Result) int64 {
			return result.Int()
		})
	case []float64:
		err = compareTypedResolvedValue(exp, items, path, func(result gjson.Result) float64 {
			return result.Float()
		})
	case []bool:
		err = compareTypedResolvedValue(exp, items, path, func(result gjson.Result) bool {
			return result.Bool()
		})
	default:
		err = fmt.Errorf("unsupported expected slice type for array path=%s type=%T", path, jt.expected)
		goto end
	}

end:
	return err
}

func compareTypedResolvedValue[T any](exp []T, items []gjson.Result, path string, convertFunc func(gjson.Result) T) (err error) {
	got := make([]T, 0, len(items))
	for _, it := range items {
		got = append(got, convertFunc(it))
	}
	if !reflect.DeepEqual(exp, got) {
		err = fmt.Errorf("path %s: expected %v, got %v", path, exp, got)
	}
	return err
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

func (jt *jsonTest) isNonEmpty(v gjson.Result) bool {
	switch {
	case !v.Exists():
		return false
	case v.IsArray():
		return len(v.Array()) > 0
	default:
		m := v.Map()
		if len(m) > 0 {
			return true
		}
		// string
		return len(v.String()) > 0 || strings.TrimSpace(v.Raw) == "true" || strings.TrimSpace(v.Raw) == "false" || strings.TrimSpace(v.Raw) == "0" || strings.TrimSpace(v.Raw) == "1"
	}
}
