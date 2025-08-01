package mcputil

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

var _ Property = (*arrayProperty)(nil)

type arrayProperty struct {
	*property
	defaultValue DefaultArray[any]
	minItems     *int
	maxItems     *int
}

func (p *arrayProperty) Clone() Property {
	np := *p
	return &np
}

func (p *arrayProperty) SetDefault(d any) Property {
	dd, ok := d.(DefaultArray[any])
	if !ok {
		panic(fmt.Sprintf("Attempted to set default value for *arrayProperty of type []any with %T value: %v", d, d))
	}
	p.defaultValue = dd
	return p
}

func (p *arrayProperty) DefaultValue() any {
	return p.defaultValue
}

func (p *arrayProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithArray(p.GetName(), opts...)
}

func (*arrayProperty) Property() {}

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

func (p *arrayProperty) mcpPropertyOptions() []mcp.PropertyOption {

	opts := p.property.mcpPropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, mcp.DefaultArray(p.defaultValue))
	}
	if p.minItems != nil {
		opts = append(opts, mcp.MinItems(*p.minItems))
	}
	if p.maxItems != nil {
		opts = append(opts, mcp.MaxItems(*p.maxItems))
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
