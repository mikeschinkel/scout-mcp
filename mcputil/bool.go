package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

type boolProperty struct {
	*property
	Default *bool
}

func (p *boolProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithBoolean(p.Name(), opts...)
}

func (*boolProperty) Property() {}

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
