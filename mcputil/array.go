package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

type arrayProperty struct {
	*property
	Default  DefaultArray[any]
	MinItems *int
	MaxItems *int
}

func (p *arrayProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithArray(p.Name(), opts...)
}

func (*arrayProperty) Property() {}

func (p *arrayProperty) PropertyOptions() []PropertyOption {

	opts := p.property.PropertyOptions()

	if p.Default != nil {
		opts = append(opts, DefaultArray[any]{p.Default})
	}
	if p.MinItems != nil {
		opts = append(opts, MinItems{*p.MinItems})
	}
	if p.MaxItems != nil {
		opts = append(opts, MaxItems{*p.MaxItems})
	}

	return opts
}

func (p *arrayProperty) mcpPropertyOptions() []mcp.PropertyOption {

	opts := p.property.mcpPropertyOptions()

	if p.Default != nil {
		opts = append(opts, mcp.DefaultArray(p.Default))
	}
	if p.MinItems != nil {
		opts = append(opts, mcp.MinItems(*p.MinItems))
	}
	if p.MaxItems != nil {
		opts = append(opts, mcp.MaxItems(*p.MaxItems))
	}

	return opts
}

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
