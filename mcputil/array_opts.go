package mcputil

// ArrayOption handles option types and marker interface for MCP Array properties
type ArrayOption interface {
	SetArrayProperty(Property)
}

// MinMaxItems sets both minimum and maximum item counts for an array property.
type MinMaxItems [2]int

// PropertyOption implements PropertyOption interface for MinMaxItems.
func (MinMaxItems) PropertyOption() {}

// SetArrayProperty configures the MinMaxItems constraint on an array property.
// This method sets both the minimum and maximum allowed item counts.
func (mmi MinMaxItems) SetArrayProperty(prop Property) {
	p := prop.(*arrayProperty)
	p.minItems = &mmi[0]
	p.maxItems = &mmi[1]
}

// MinItems sets the minimum number of items allowed in an array property.
type MinItems [1]int

// PropertyOption implements PropertyOption interface for MinItems.
func (MinItems) PropertyOption() {}

// SetArrayProperty configures the MinItems constraint on an array property.
// This method sets the minimum allowed number of items in the array.
func (mi MinItems) SetArrayProperty(prop Property) {
	prop.(*arrayProperty).minItems = &mi[0]
}

// MaxItems sets the maximum number of items allowed in an array property.
type MaxItems [1]int

// PropertyOption implements PropertyOption interface for MaxItems.
func (MaxItems) PropertyOption() {}

// SetArrayProperty configures the MaxItems constraint on an array property.
// This method sets the maximum allowed number of items in the array.
func (mi MaxItems) SetArrayProperty(prop Property) {
	prop.(*arrayProperty).maxItems = &mi[0]
}

// DefaultArray sets the default value for an array property.
type DefaultArray[T any] []T

// PropertyOption implements PropertyOption interface for DefaultArray.
func (DefaultArray[T]) PropertyOption() {}

// SetArrayProperty configures the default value for an array property.
// This method converts the typed default array to []any and sets it as the default value.
func (da DefaultArray[T]) SetArrayProperty(prop Property) {
	defaultArray := make([]any, len(da))
	for i, p := range da {
		defaultArray[i] = p
	}
	prop.(*arrayProperty).defaultValue = defaultArray
}

// ArrayOptions provides a convenient way to set multiple array property options at once.
type ArrayOptions struct {
	MinItems *int              // Minimum number of items
	MaxItems *int              // Maximum number of items
	Default  DefaultArray[any] // Default array value
}

// PropertyOption implements PropertyOption interface for ArrayOptions.
func (ArrayOptions) PropertyOption() {}

// SetArrayProperty configures multiple array constraints and default values in a single operation.
// This method allows setting minimum items, maximum items, and default values together.
func (ao ArrayOptions) SetArrayProperty(prop Property) {
	p := prop.(*arrayProperty)
	if ao.MinItems != nil {
		p.minItems = ao.MinItems
	}
	if ao.MaxItems != nil {
		p.maxItems = ao.MaxItems
	}
	if ao.Default != nil {
		p.defaultValue = ao.Default
	}
}
