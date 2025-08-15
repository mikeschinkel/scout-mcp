// Package pipefuncs provides modular pipe functions for JSON testing and validation.
// Pipe functions transform and validate JSON values during assertion processing.
package pipefuncs

// Initialize ensures all pipe functions in this package are registered by triggering their init() functions.
// This function should be called when using pipe functions to guarantee proper registration.
func Initialize() error {
	// Call this to trigger all the init() funcs that register the pipe funcs in this package.
	// Reference all pipe function types to ensure their init() functions are called
	_ = JSONPipeFunc{}
	_ = ExistsPipeFunc{}
	_ = NotNullPipeFunc{}
	_ = NotEmptyPipeFunc{}
	_ = LenPipeFunc{}
	return nil
}
