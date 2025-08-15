package jsontest

import (
	"context"
	"fmt"
	"strings"
)

// PipeFunc represents a function that can be used in JSON test pipe operations.
// Pipe functions validate or transform JSON values during assertion processing.
type PipeFunc interface {
	Name() string
	Handle(context.Context, *PipeState) error
	PipeFunc()
}

// BasePipeFunc provides common functionality for all pipe functions.
type BasePipeFunc struct {
	name string
}

// PipeFunc implements the PipeFunc interface marker method.
func (pf BasePipeFunc) PipeFunc() {}

// NewBasePipeFunc creates a new BasePipeFunc with the given name.
// The name is normalized to lowercase and trimmed of whitespace.
func NewBasePipeFunc(name string) BasePipeFunc {
	return BasePipeFunc{
		name: strings.ToLower(strings.TrimSpace(name)),
	}
}

// Name returns the normalized name of the pipe function
// Name returns the normalized name of the pipe function.
func (pf BasePipeFunc) Name() string {
	return pf.name
}

// registeredPipeFunc holds all pipe functions that have been registered.
var registeredPipeFunc []PipeFunc

// RegisteredPipeFunc returns a slice of all registered pipe functions.
func RegisteredPipeFunc() []PipeFunc {
	return registeredPipeFunc
n
}
// RegisteredPipeFuncMap returns a map of pipe function names to PipeFunc instances
// for efficient lookup by name.
func RegisteredPipeFuncMap() (m map[string]PipeFunc) {
	m = make(map[string]PipeFunc, len(registeredPipeFunc))
	for _, pf := range registeredPipeFunc {
		m[pf.Name()] = pf
	}
	return m
}

// RegisterPipeFunc adds a pipe function to the global registry.
// The function name must end with "()" to indicate it's callable.
	if !strings.HasSuffix(pf.Name(), "()") {
		panic(fmt.Sprintf("PipeFunc '%s' must end in '()'", pf.Name()))
		panic(fmt.Sprintf("PipeFunc '%s' must end in '()'",pf.Name()))
	}
	registeredPipeFunc = append(registeredPipeFunc, pf)
}

// GetRegisteredPipeFunc finds a registered pipe function by name (case-insensitive).
// Returns nil if no pipe function with the given name is found.
func GetRegisteredPipeFunc(name string) (pf PipeFunc) {
	if registeredPipeFunc == nil {
		panic("No jsontest pipe funcs have been registered.\nDid you forget to import github.com/mikeschinkel/scout-mcp/jsontest/pipefuncs for side effects (by prefixing it with '_')?")
	}
	name = strings.ToLower(name)
	for _, rpf := range registeredPipeFunc {
		if strings.ToLower(rpf.Name()) != name {
			continue
		}
		pf = rpf
		goto end
	}
end:
	return pf
}

// GetRegisteredPipeFuncNames returns the names of all registered PipeFunc
func GetRegisteredPipeFuncNames() []string {
	names := make([]string, 0, len(registeredPipeFunc))
	for _, PipeFunc := range registeredPipeFunc {
		names = append(names, PipeFunc.Name())
	}
	return names
}
