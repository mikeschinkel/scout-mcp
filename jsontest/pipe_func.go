package jsontest

import (
	"context"
	"fmt"
	"strings"
)

type PipeFunc interface {
	Name() string
	Handle(context.Context, *PipeState) error
	PipeFunc()
}

type BasePipeFunc struct {
	name string
}

func (pf BasePipeFunc) PipeFunc() {}

func NewBasePipeFunc(name string) BasePipeFunc {
	return BasePipeFunc{
		name: strings.ToLower(strings.TrimSpace(name)),
	}
}
func (pf BasePipeFunc) Name() string {
	return pf.name
}

var registeredPipeFunc []PipeFunc

func RegisteredPipeFunc() []PipeFunc {
	return registeredPipeFunc
}
func RegisteredPipeFuncMap() (m map[string]PipeFunc) {
	m = make(map[string]PipeFunc, len(registeredPipeFunc))
	for _, pf := range registeredPipeFunc {
		m[pf.Name()] = pf
	}
	return m
}

func RegisterPipeFunc(pf PipeFunc) {
	if !strings.HasSuffix(pf.Name(),"()") {
		panic(fmt.Sprintf("PipeFunc '%s' must end in '()'",pf.Name()))
	}
	registeredPipeFunc = append(registeredPipeFunc, pf)
}

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
