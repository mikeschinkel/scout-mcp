package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

type numberProperty struct {
	*property
	Default *float64
	Min     *float64
	Max     *float64
}

func (p *numberProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithNumber(p.Name(), opts...)
}

func (*numberProperty) Property() {}
func (p *numberProperty) PropertyOptions() []PropertyOption {

	opts := p.property.PropertyOptions()

	if p.Default != nil {
		opts = append(opts, DefaultNumber{*p.Default})
	}
	if p.Min != nil {
		opts = append(opts, MinNumber{*p.Min})
	}
	if p.Max != nil {
		opts = append(opts, MaxNumber{*p.Max})
	}

	return opts
}

func (p *numberProperty) mcpPropertyOptions() []mcp.PropertyOption {

	opts := p.property.mcpPropertyOptions()

	if p.Default != nil {
		opts = append(opts, mcp.DefaultNumber(*p.Default))
	}
	if p.Min != nil {
		opts = append(opts, mcp.Min(*p.Min))
	}
	if p.Max != nil {
		opts = append(opts, mcp.Max(*p.Max))
	}

	return opts
}

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
