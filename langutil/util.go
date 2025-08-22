package langutil

import (
	"errors"
	"fmt"
)

// convertSlice performs type-safe conversion of a slice of interface{} values to a slice of type T.
// This generic utility function is used internally by the langutil package to safely convert
// untyped slice data (typically from JSON parsing or reflection operations) into strongly-typed
// slices that can be used with Go's type system.
//
// # Type Safety
//
// The function performs runtime type checking for each element in the input slice, ensuring
// that all values can be safely converted to type T. If any element cannot be converted,
// the function collects detailed error information about the failed conversion while
// continuing to process remaining elements.
//
// # Error Handling
//
// The function uses a comprehensive error collection strategy:
//   - Continues processing all elements even if some conversions fail
//   - Collects detailed error information including element index and type information
//   - Returns all errors joined together using errors.Join for complete error reporting
//   - Provides the successfully converted elements even when some conversions fail
//
// # Performance Characteristics
//
// The function pre-allocates the output slice with the correct capacity to avoid repeated
// memory allocations during conversion. Type assertions are performed using Go's built-in
// type assertion mechanism, which is optimized for performance.
//
// # Generic Implementation
//
// This function uses Go generics to provide type safety at compile time while maintaining
// flexibility for different target types. The generic constraint allows any type T to be
// used as the target type, making the function reusable across the langutil package.
//
// # Example Usage
//
//	// Convert []interface{} containing strings to []string
//	input := []interface{}{"hello", "world", 123, "test"}
//	output, err := convertSlice[string](input)
//	// output contains: ["hello", "world", "", "test"]
//	// err contains details about the failed conversion of 123
//
// # Internal Usage
//
// This function is primarily used internally by the langutil package for:
//   - Converting JSON array data to typed slices
//   - Processing reflection results from AST traversal
//   - Handling dynamic data from language processors
//   - Type-safe conversion in configuration processing
//
// # Error Message Format
//
// Error messages follow the format: "error converting item {index}: item a '{actual_type}', not a '{expected_type}'"
// This provides clear information about which element failed conversion and why.
func convertSlice[T any](input []any) (output []T, err error) {
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
