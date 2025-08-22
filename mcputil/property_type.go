package mcputil

// PropertyType represents the data type of a tool property for MCP parameter validation.
// These types correspond to JSON Schema types and determine how parameters are validated
// and converted during tool execution.
type PropertyType string

const (
	// StringType represents text-based parameters that accept string values.
	StringType PropertyType = "string"

	// NumberType represents numeric parameters that accept float64 or integer values.
	NumberType PropertyType = "number"

	// BoolType represents boolean parameters that accept true/false values.
	BoolType PropertyType = "bool"

	// ArrayType represents array parameters that accept slice/array values.
	ArrayType PropertyType = "array"

	// ObjectType represents object parameters that accept map/struct values.
	ObjectType PropertyType = "object"
)
