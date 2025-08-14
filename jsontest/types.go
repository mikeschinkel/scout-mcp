package jsontest

type anyOrderMarker interface{ anyOrder() }

// AnyOrderSlice is a named slice type used to signal order-insensitive comparison.
type AnyOrderSlice[T comparable] []T

func (AnyOrderSlice[T]) anyOrder() {}

//goland:noinspection GoUnusedExportedFunction
func AnyOrder[T comparable](vals ...T) AnyOrderSlice[T] { return vals }
