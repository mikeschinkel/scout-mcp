package mcputil

import (
	"fmt"
)

var _ Property = (*arrayProperty)(nil)

// arrayProperty implements Property for array-type MCP tool parameters.
type arrayProperty struct {
	*property
	defaultValue DefaultArray[any] // Default array value
	minItems     *int              // Minimum number of items allowed
	maxItems     *int              // Maximum number of items allowed
}

// setBase sets the base property for this array property implementation.
// This method is part of the property interface implementation pattern.
func (p *arrayProperty) setBase(prop *property) {
	p.property = prop
}

// Clone creates a deep copy of the array property with all its configuration.
// This method returns a new array property instance with the same settings.
func (p *arrayProperty) Clone() Property {
	np := *p
	return &np
}

// SetDefault sets the default value for this array property.
// The value must be of type DefaultArray[any] or the method will panic.
func (p *arrayProperty) SetDefault(d any) Property {
	dd, ok := d.(DefaultArray[any])
	if !ok {
		panic(fmt.Sprintf("Attempted to set default value for *arrayProperty of type []any with %T value: %v", d, d))
	}
	p.defaultValue = dd
	return p
}

// DefaultValue returns the default array value for this property.
// Returns nil if no default value has been set.
func (p *arrayProperty) DefaultValue() any {
	return p.defaultValue
}

// mcpToolOption creates an MCP tool option for array parameters.
// This method integrates with the underlying MCP library to define array properties.
func (p *arrayProperty) mcpToolOption(opts []mcpPropertyOption) mcpToolOption {
	return mcpWithArray(p.GetName(), opts...)
}

// Property implements the Property interface marker method.
func (*arrayProperty) Property() {}

// PropertyOptions returns all property configuration options for this array property.
// This includes default values, minimum/maximum item constraints, and base property options.
func (p *arrayProperty) PropertyOptions() []PropertyOption {

	opts := p.property.PropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, DefaultArray[any]{p.defaultValue})
	}
	if p.minItems != nil {
		opts = append(opts, MinItems{*p.minItems})
	}
	if p.maxItems != nil {
		opts = append(opts, MaxItems{*p.maxItems})
	}

	return opts
}

// mcpPropertyOptions returns MCP-specific property options for this array property.
// This method provides the underlying MCP library with validation constraints.
func (p *arrayProperty) mcpPropertyOptions() []mcpPropertyOption {

	opts := p.property.mcpPropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, mcpDefaultArray(p.defaultValue))
	}
	if p.minItems != nil {
		opts = append(opts, mcpMinItems(*p.minItems))
	}
	if p.maxItems != nil {
		opts = append(opts, mcpMaxItems(*p.maxItems))
	}

	return opts
}

// Array creates a new array property with the specified name, description, and options.
// This is the primary constructor for creating array-type tool parameters.
func Array(name, description string, opts ...ArrayOption) Property {
	p := &arrayProperty{
		property: &property{
			name:        name,
			dataType:    ArrayType,
			description: description,
			required:    false,
		},
	}
	p.property.parent = p
	for _, opt := range opts {
		opt.SetArrayProperty(p)
	}
	return p
}
