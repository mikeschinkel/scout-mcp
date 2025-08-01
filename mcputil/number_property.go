package mcputil

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

var _ Property = (*numberProperty)(nil)

type numberProperty struct {
	*property
	defaultValue *float64
	min          *float64
	max          *float64
	zeroOK       *bool
}

func (p *numberProperty) ZeroOK() (zok bool) {
	if p.zeroOK == nil {
		goto end
	}
	zok = *p.zeroOK
end:
	return zok
}

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

func (p *numberProperty) Clone() Property {
	np := *p
	return &np
}
func (p *numberProperty) DefaultValue() any {
	return p.defaultValue
}

func (p *numberProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithNumber(p.GetName(), opts...)
}

func (*numberProperty) Property() {}
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

func (p *numberProperty) mcpPropertyOptions() []mcp.PropertyOption {

	opts := p.property.mcpPropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, mcp.DefaultNumber(*p.defaultValue))
	}
	if p.min != nil {
		opts = append(opts, mcp.Min(*p.min))
	}
	if p.max != nil {
		opts = append(opts, mcp.Max(*p.max))
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
