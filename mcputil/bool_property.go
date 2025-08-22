package mcputil

import (
	"fmt"
)

var _ Property = (*boolProperty)(nil)

// boolProperty implements Property for boolean-type MCP tool parameters.
type boolProperty struct {
	*property
	defaultValue *bool // Default boolean value
}

// setBase sets the base property for this boolean property implementation.
// This method is part of the property interface implementation pattern.
func (p *boolProperty) setBase(prop *property) {
	p.property = prop
}

// SetDefault sets the default value for this boolean property.
// The value must be of type bool or the method will panic.
func (p *boolProperty) SetDefault(b any) Property {
	value, ok := b.(bool)
	if !ok {
		panic(fmt.Sprintf("Attempted to set default value for *boolProperty with %T value: %v", b, b))
	}
	p.defaultValue = &value
	return p
}

// Clone creates a deep copy of the boolean property with all its configuration.
// This method returns a new boolean property instance with the same settings.
func (p *boolProperty) Clone() Property {
	np := *p
	return &np
}

// DefaultValue returns the default boolean value for this property.
// Returns nil if no default value has been set.
func (p *boolProperty) DefaultValue() any {
	return p.defaultValue
}

// mcpToolOption creates an MCP tool option for boolean parameters.
// This method integrates with the underlying MCP library to define boolean properties.
func (p *boolProperty) mcpToolOption(opts []mcpPropertyOption) mcpToolOption {
	return mcpWithBoolean(p.GetName(), opts...)
}

// Property implements the Property interface marker method.
func (*boolProperty) Property() {}

// Bool creates a new boolean property with the specified name, description, and options.
// This is the primary constructor for creating boolean-type tool parameters.
func Bool(name, description string, opts ...BoolOption) Property {
	p := &boolProperty{
		property: &property{
			name:        name,
			dataType:    BoolType,
			description: description,
			required:    false,
		},
	}
	p.property.parent = p

	for _, opt := range opts {
		opt.SetBoolProperty(p)
	}

	return p
}
