package mcputil

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

var _ Property = (*stringProperty)(nil)

type stringProperty struct {
	*property
	defaultValue *string
	enum         []string
	minLen       *int
	maxLen       *int
	pattern      *string
}

func (p *stringProperty) SetDefault(s any) Property {
	value, ok := s.(string)
	if !ok {
		panic(fmt.Sprintf("Attempted to set default value for *stringProperty with %T value: %v", s, s))
	}
	p.defaultValue = &value
	return p
}

func (p *stringProperty) Clone() Property {
	np := *p
	return &np
}

func (p *stringProperty) DefaultValue() any {
	return p.defaultValue
}

func (p *stringProperty) mcpToolOption(opts []mcp.PropertyOption) mcp.ToolOption {
	return mcp.WithString(p.GetName(), opts...)
}

func (*stringProperty) Property() {}

func (p *stringProperty) PropertyOptions() []PropertyOption {

	opts := p.property.PropertyOptions()

	if p.defaultValue != nil {
		opts = append(opts, DefaultString{*p.defaultValue})
	}
	if p.enum != nil {
		opts = append(opts, Enum(p.enum))
	}
	if p.minLen != nil {
		opts = append(opts, MinLen{*p.minLen})
	}
	if p.maxLen != nil {
		opts = append(opts, MaxLen{*p.maxLen})
	}
	if p.pattern != nil {
		opts = append(opts, Pattern{*p.pattern})
	}

	return opts
}
func (p *stringProperty) mcpPropertyOptions() []mcp.PropertyOption {
	opts := p.property.mcpPropertyOptions()
	if p.defaultValue != nil {
		opts = append(opts, mcp.DefaultString(*p.defaultValue))
	}
	if p.enum != nil {
		opts = append(opts, mcp.Enum(p.enum...))
	}
	if p.minLen != nil {
		opts = append(opts, mcp.MinLength(*p.minLen))
	}
	if p.maxLen != nil {
		opts = append(opts, mcp.MaxLength(*p.maxLen))
	}
	if p.pattern != nil {
		opts = append(opts, mcp.Pattern(*p.pattern))
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
