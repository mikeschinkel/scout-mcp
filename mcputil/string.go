package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

var _ Property = (*stringProperty)(nil)

type stringProperty struct {
	*property
	Default *string
	Enum    []string
	MinLen  *int
	MaxLen  *int
	Pattern *string
}

func (p *stringProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithString(p.Name(), opts...)
}

func (*stringProperty) Property() {}

func (p *stringProperty) PropertyOptions() []PropertyOption {

	opts := p.property.PropertyOptions()

	if p.Default != nil {
		opts = append(opts, DefaultString{*p.Default})
	}
	if p.Enum != nil {
		opts = append(opts, Enum(p.Enum))
	}
	if p.MinLen != nil {
		opts = append(opts, MinLen{*p.MinLen})
	}
	if p.MaxLen != nil {
		opts = append(opts, MaxLen{*p.MaxLen})
	}
	if p.Pattern != nil {
		opts = append(opts, Pattern{*p.Pattern})
	}

	return opts
}
func (p *stringProperty) mcpPropertyOptions() []mcp.PropertyOption {
	opts := p.property.mcpPropertyOptions()
	if p.Default != nil {
		opts = append(opts, mcp.DefaultString(*p.Default))
	}
	if p.Enum != nil {
		opts = append(opts, mcp.Enum(p.Enum...))
	}
	if p.MinLen != nil {
		opts = append(opts, mcp.MinLength(*p.MinLen))
	}
	if p.MaxLen != nil {
		opts = append(opts, mcp.MaxLength(*p.MaxLen))
	}
	if p.Pattern != nil {
		opts = append(opts, mcp.Pattern(*p.Pattern))
	}
	return opts
}

// String creates a Property of type String (StringOption)
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
