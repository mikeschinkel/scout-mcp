package mcputil

// PropertyType represents the type of a tool property
type PropertyType string

const (
	StringType PropertyType = "string"
	NumberType PropertyType = "number"
	BoolType   PropertyType = "bool"
	ArrayType  PropertyType = "array"
	ObjectType PropertyType = "object"
)
