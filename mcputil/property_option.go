package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// propertyOption wraps mcp property options to isolate mcp imports
type propertyOption struct {
	option PropertyOption
}

type PropertyOption interface {
	PropertyOption()
}

// PropertyOptionsGetter is the interface for getting property options
type PropertyOptionsGetter interface {
	PropertyOptions() []PropertyOption
}

// PropertyOptionsGetter is the interface for getting property options
type mcpPropertyOptionsGetter interface {
	mcpPropertyOptions() []mcp.PropertyOption
}
