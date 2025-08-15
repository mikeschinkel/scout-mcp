package jsontest

// NotNull is a marker type used in JSON assertions to indicate
// that a value must not be null.
type NotNull struct{}

// NotEmpty is a marker type used in JSON assertions to indicate
// that a value must not be empty (non-zero, non-empty string, etc.).
type NotEmpty struct{}

// anyOrderMarker is an interface for types that support order-insensitive comparison.
type anyOrderMarker interface{ anyOrder() }

// AnyOrderSlice is a named slice type used to signal order-insensitive comparison.
type AnyOrderSlice[T comparable] []T

// anyOrder implements the anyOrderMarker interface for AnyOrderSlice.
func (AnyOrderSlice[T]) anyOrder() {}

// AnyOrder creates an AnyOrderSlice from the provided values,
// allowing order-insensitive comparison in JSON assertions.
//
//goland:noinspection GoUnusedExportedFunction
func AnyOrder[T comparable](vals ...T) AnyOrderSlice[T] { return vals }
