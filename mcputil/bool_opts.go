package mcputil

// BoolOption handles option types and marker interface for MCP Bool properties
type BoolOption interface {
	SetBoolProperty(Property)
}

// DefaultBool sets a specific default boolean value for a boolean property.
type DefaultBool [1]bool

// BoolOpt implements the BoolOption interface marker method.
func (DefaultBool) BoolOpt() {}

// SetBoolProperty configures the default boolean value for a boolean property.
// This method sets the boolean property's default value to the specified bool.
func (db DefaultBool) SetBoolProperty(prop Property) {
	prop.(*boolProperty).defaultValue = &db[0]
}

// DefaultTrue sets the default value to true for a boolean property.
type DefaultTrue struct{}

// BoolOpt implements the BoolOption interface marker method.
func (DefaultTrue) BoolOpt() {}

// SetBoolProperty configures the boolean property to default to true.
// This is a convenience method for setting a true default value.
func (DefaultTrue) SetBoolProperty(prop Property) {
	b := true
	prop.(*boolProperty).defaultValue = &b
}

// DefaultFalse sets the default value to false for a boolean property.
type DefaultFalse struct{}

// BoolOpt implements the BoolOption interface marker method.
func (DefaultFalse) BoolOpt() {}

// SetBoolProperty configures the boolean property to default to false.
// This is a convenience method for setting a false default value.
func (DefaultFalse) SetBoolProperty(prop Property) {
	b := false
	prop.(*boolProperty).defaultValue = &b
}

// BoolOptions provides a convenient way to set boolean property options.
type BoolOptions struct {
	Default *bool // Default boolean value
}

// PropertyOption implements PropertyOption interface for BoolOptions.
func (BoolOptions) PropertyOption() {}

// BoolOpt implements the BoolOption interface marker method.
func (BoolOptions) BoolOpt() {}

// SetBoolProperty configures boolean property options in a single operation.
// This method allows setting the default value if specified in the options.
func (bo BoolOptions) SetBoolProperty(prop Property) {
	if bo.Default != nil {
		prop.(*boolProperty).defaultValue = bo.Default
	}
}
