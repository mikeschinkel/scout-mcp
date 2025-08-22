package mcputil

import (
	"fmt"
)

var _ Property = (*numberProperty)(nil)

// numberProperty implements Property for numeric-type MCP tool parameters.
type numberProperty struct {
	*property
	defaultValue *float64 // Default numeric value
	min          *float64 // Minimum allowed value
	max          *float64 // Maximum allowed value
	zeroOK       *bool    // Whether zero value is acceptable
}

// setBase sets the base property for this numeric property implementation.
// This method is part of the property interface implementation pattern.
func (p *numberProperty) setBase(prop *property) {
	p.property = prop
}

// ZeroOK returns whether zero values are acceptable for this numeric property.
// This method checks the zeroOK flag to determine if zero is a valid value.
func (p *numberProperty) ZeroOK() (zok bool) {
	if p.zeroOK == nil {
		goto end
	}
	zok = *p.zeroOK
end:
	return zok
}

// SetDefault sets the default value for this numeric property.
// The value can be int, int64, or float64, and will be converted to float64 internally.
func (p *numberProperty) SetDefault(n any) Property {
	switch v := n.(type) {
	case float64:
		p.defaultValue = &v
	case int:
		f := float64(v)
		p.defaultValue = &f
	case int64:
		f := float64(v)
		p.defaultValue = &f
	default:
		panic(fmt.Sprintf("Attempted to set default value for *numberProperty with %T value (int, int64 or float64 expected): %v", n, n))
	}
	return p
}

// Clone creates a deep copy of the numeric property with all its configuration.
// This method returns a new numeric property instance with the same settings.
func (p *numberProperty) Clone() Property {
	np := *p
	return &np
}

// DefaultValue returns the default numeric value for this property.
// Returns nil if no default value has been set.
func (p *numberProperty) DefaultValue() any {
	return p.defaultValue
}

// mcpToolOption creates an MCP tool option for numeric parameters.
// This method integrates with the underlying MCP library to define number properties.
func (p *numberProperty) mcpToolOption(opts []mcpPropertyOption) mcpToolOption {
	return mcpWithNumber(p.GetName(), opts...)
}

// Property implements the Property interface marker method.
func (*numberProperty) Property() {}

// PropertyOptions returns all property configuration options for this numeric property.
// This includes default values, minimum/maximum constraints, and base property options.
func (p *numberProperty) PropertyOptions() []PropertyOption {

	opts := p.property.PropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, DefaultNumber{*p.defaultValue})
	}
	if p.min != nil {
		opts = append(opts, MinNumber{*p.min})
	}
	if p.max != nil {
		opts = append(opts, MaxNumber{*p.max})
	}

	return opts
}

// mcpPropertyOptions returns MCP-specific property options for this numeric property.
// This method provides the underlying MCP library with validation constraints.
func (p *numberProperty) mcpPropertyOptions() []mcpPropertyOption {

	opts := p.property.mcpPropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, mcpDefaultNumber(*p.defaultValue))
	}
	if p.min != nil {
		opts = append(opts, mcpMin(*p.min))
	}
	if p.max != nil {
		opts = append(opts, mcpMax(*p.max))
	}

	return opts
}

// Number creates a new numeric property with the specified name, description, and options.
// This is the primary constructor for creating number-type tool parameters.
func Number(name, description string, opts ...NumberOption) Property {
	p := &numberProperty{
		property: &property{
			name:        name,
			dataType:    NumberType,
			description: description,
			required:    false,
		},
	}
	p.property.parent = p

	for _, opt := range opts {
		opt.SetNumberProperty(p)
	}

	return p
}
