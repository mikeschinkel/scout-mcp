package testutil

import (
	"reflect"
)

func convertSlice(input any) []any {
	switch v := input.(type) {
	case []string:
		return convertTypedSlice(v)
	case []int64:
		return convertTypedSlice(v)
	case []int:
		return convertTypedSlice(v)
	case []bool:
		return convertTypedSlice(v)
	default:
		return convertSliceByReflection(input)
	}
}
func convertTypedSlice[T any](input []T) []any {
	output := make([]any, len(input))
	for i, val := range input {
		output[i] = val
	}
	return output
}

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
