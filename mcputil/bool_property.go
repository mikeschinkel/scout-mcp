package mcputil

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

var _ Property = (*boolProperty)(nil)

type boolProperty struct {
	*property
	defaultValue *bool
}

func (p *boolProperty) SetDefault(b any) Property {
	value, ok := b.(bool)
	if !ok {
		panic(fmt.Sprintf("Attempted to set default value for *boolProperty with %T value: %v", b, b))
	}
	p.defaultValue = &value
	return p
}

func (p *boolProperty) Clone() Property {
	np := *p
	return &np
}

func (p *boolProperty) DefaultValue() any {
	return p.defaultValue
}

func (p *boolProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithBoolean(p.GetName(), opts...)
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
