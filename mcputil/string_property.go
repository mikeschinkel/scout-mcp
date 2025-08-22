package mcputil

import (
	"fmt"
)

var _ Property = (*stringProperty)(nil)

// stringProperty implements Property for string-type MCP tool parameters.
type stringProperty struct {
	*property
	defaultValue *string  // Default string value
	enum         []string // Allowed enumeration values
	minLen       *int     // Minimum string length
	maxLen       *int     // Maximum string length
	pattern      *string  // Regular expression pattern
}

// setBase sets the base property for this string property implementation.
// This method is part of the property interface implementation pattern.
func (p *stringProperty) setBase(prop *property) {
	p.property = prop
}

// SetDefault sets the default value for this string property.
// The value must be of type string or the method will panic.
func (p *stringProperty) SetDefault(s any) Property {
	value, ok := s.(string)
	if !ok {
		panic(fmt.Sprintf("Attempted to set default value for *stringProperty with %T value: %v", s, s))
	}
	p.defaultValue = &value
	return p
}

// Clone creates a deep copy of the string property with all its configuration.
// This method returns a new string property instance with the same settings.
func (p *stringProperty) Clone() Property {
	np := *p
	return &np
}

// DefaultValue returns the default string value for this property.
// Returns nil if no default value has been set.
func (p *stringProperty) DefaultValue() any {
	return p.defaultValue
}

// mcpToolOption creates an MCP tool option for string parameters.
// This method integrates with the underlying MCP library to define string properties.
func (p *stringProperty) mcpToolOption(opts []mcpPropertyOption) mcpToolOption {
	return mcpWithString(p.GetName(), opts...)
}

// Property implements the Property interface marker method.
func (*stringProperty) Property() {}

// PropertyOptions returns all property configuration options for this string property.
// This includes default values, enumeration, length constraints, pattern, and base property options.
func (p *stringProperty) PropertyOptions() []PropertyOption {

	opts := p.property.PropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, DefaultString{*p.defaultValue})
	}
	if p.enum != nil {
		opts = append(opts, Enum(p.enum))
	}
	if p.minLen != nil {
		opts = append(opts, MinLen{*p.minLen})
	}
	if p.maxLen != nil {
		opts = append(opts, MaxLen{*p.maxLen})
	}
	if p.pattern != nil {
		opts = append(opts, Pattern{*p.pattern})
	}

	return opts
}

// mcpPropertyOptions returns MCP-specific property options for this string property.
// This method provides the underlying MCP library with validation constraints.
func (p *stringProperty) mcpPropertyOptions() []mcpPropertyOption {
	opts := p.property.mcpPropertyOptions()
	if p.defaultValue != nil {
		opts = append(opts, mcpDefaultString(*p.defaultValue))
	}
	if p.enum != nil {
		opts = append(opts, mcpEnum(p.enum...))
	}
	if p.minLen != nil {
		opts = append(opts, mcpMinLength(*p.minLen))
	}
	if p.maxLen != nil {
		opts = append(opts, mcpMaxLength(*p.maxLen))
	}
	if p.pattern != nil {
		opts = append(opts, mcpPattern(*p.pattern))
	}
	return opts
}

// String creates a new string property with the specified name, description, and options.
// This is the primary constructor for creating string-type tool parameters.
func String(name, description string, opts ...StringOption) Property {
	p := &stringProperty{
		property: &property{
			name:        name,
			dataType:    StringType,
			description: description,
		},
	}
	p.property.parent = p
	for _, opt := range opts {
		opt.SetStringProperty(p)
	}
	return p
}
