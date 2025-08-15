package mcputil

import (
	"errors"
	"fmt"
	"reflect"
)

// ConvertContainedSlice converts a slice contained in an `any` into a slice of any — e.g. []any.
// It handles common slice types and falls back to reflection for unknown types.
func ConvertContainedSlice(input any) []any {
	switch v := input.(type) {
	case []string:
		return convertTypedSlice(v)
	case []int64:
		return convertTypedSlice(v)
	case []int:
		return convertTypedSlice(v)
	case []bool:
		return convertTypedSlice(v)
	case []any:
		return v
	default:
		return convertSliceByReflection(input)
	}
}

// convertTypedSlice converts a typed slice to []any.
func convertTypedSlice[T any](input []T) []any {
	output := make([]any, len(input))
	for i, val := range input {
		output[i] = val
	}
	return output
}

// convertSliceByReflection uses reflection to convert any slice type to []any.
func convertSliceByReflection(input any) (output []any) {
	var n int

	val := reflect.ValueOf(input)

	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		goto end
	}

	n = val.Len()
	output = make([]any, n)
	for i := 0; i < n; i++ {
		output[i] = val.Index(i).Interface()
	}
end:
	return output

}

// convertSliceOfAny converts a []any to a typed slice, validating each element.
// Returns an error if any elements cannot be converted to type T.
func convertSliceOfAny[T any](input []any) (output []T, err error) {
	var t T
	var errs []error

	output = make([]T, len(input))
	for i, item := range input {
		converted, ok := item.(T)
		if !ok {
			errs = append(errs, fmt.Errorf("error converting item %d: item a '%T', not a '%T'", i, item, t))
			continue
		}
		output[i] = converted
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return output, err
}

// empty returns true if an any value is nil or equals "".
func empty(value any) (empty bool) {
	if value == nil {
		empty = true
		goto end
	}
	if value == "" {
		empty = true
		goto end
	}
end:
	return empty
}
