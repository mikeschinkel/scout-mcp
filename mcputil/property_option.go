package mcputil

// PropertyOption is a marker interface for property configuration options.
// Implementations provide specific validation rules and constraints
// for MCP tool parameters.
type PropertyOption interface {
	PropertyOption() // Marker method
}
