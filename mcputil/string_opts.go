package mcputil

// StringOption handles option types and marker interface for MCP String properties
type StringOption interface {
	SetStringProperty(Property)
}

// Enum defines allowed string values for a string property.
type Enum []string

// PropertyOption implements PropertyOption interface for Enum.
func (Enum) PropertyOption() {}

// SetStringProperty configures the enumeration constraint for a string property.
// This method sets the allowed values that the string property can accept.
func (e Enum) SetStringProperty(prop Property) {
	prop.(*stringProperty).enum = e
}

// MinLen sets the minimum length constraint for a string property.
type MinLen [1]int

// PropertyOption implements PropertyOption interface for MinLen.
func (MinLen) PropertyOption() {}

// SetStringProperty configures the minimum length constraint for a string property.
// This method sets the minimum number of characters required in the string value.
func (ml MinLen) SetStringProperty(prop Property) {
	prop.(*stringProperty).minLen = &ml[0]
}

// MaxLen sets the maximum length constraint for a string property.
type MaxLen [1]int

// PropertyOption implements PropertyOption interface for MaxLen.
func (MaxLen) PropertyOption() {}

// SetStringProperty configures the maximum length constraint for a string property.
// This method sets the maximum number of characters allowed in the string value.
func (ml MaxLen) SetStringProperty(prop Property) {
	prop.(*stringProperty).maxLen = &ml[0]
}

// DefaultString sets the default value for a string property.
type DefaultString [1]string

// PropertyOption implements PropertyOption interface for DefaultString.
func (DefaultString) PropertyOption() {}

// SetStringProperty configures the default value for a string property.
// This method sets the property's default string value.
func (ds DefaultString) SetStringProperty(prop Property) {
	prop.(*stringProperty).defaultValue = &ds[0]
}

// Pattern sets a regular expression validation pattern for a string property.
type Pattern [1]string

// PropertyOption implements PropertyOption interface for Pattern.
func (Pattern) PropertyOption() {}

// SetStringProperty configures the regex pattern constraint for a string property.
// This method sets a regular expression that the string value must match.
func (p Pattern) SetStringProperty(prop Property) {
	prop.(*stringProperty).pattern = &p[0]
}

// StringOptions provides a convenient way to set multiple string property options at once.
type StringOptions struct {
	Enum    []string // Allowed enumeration values
	MinLen  *int     // Minimum string length
	MaxLen  *int     // Maximum string length
	Default *string  // Default string value
	Pattern *string  // Regular expression pattern
}

// PropertyOption implements PropertyOption interface for StringOptions.
func (StringOptions) PropertyOption() {}

// SetStringProperty configures multiple string constraints and default values in a single operation.
// This method allows setting enumeration, length limits, default value, and regex pattern together.
func (so StringOptions) SetStringProperty(prop Property) {
	p := prop.(*stringProperty)
	if so.Enum != nil {
		p.enum = so.Enum
	}
	if so.MinLen != nil {
		p.minLen = so.MinLen
	}
	if so.MaxLen != nil {
		p.maxLen = so.MaxLen
	}
	if so.Default != nil {
		p.defaultValue = so.Default
	}
	if so.Pattern != nil {
		p.pattern = so.Pattern
	}
}
