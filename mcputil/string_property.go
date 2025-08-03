package mcputil

import (
	"fmt"
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

func (p *stringProperty) setBase(prop *property) {
	p.property = prop
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

func (p *stringProperty) mcpToolOption(opts []mcpPropertyOption) mcpToolOption {
	return mcpWithString(p.GetName(), opts...)
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
func (p *stringProperty) mcpPropertyOptions() []mcpPropertyOption {
	opts := p.property.mcpPropertyOptions()
	if p.defaultValue != nil {
		opts = append(opts, mcpDefaultString(*p.defaultValue))
	}
	if p.enum != nil {
		opts = append(opts, mcpEnum(p.enum...))
	}
	if p.minLen != nil {
		opts = append(opts, mcpMinLength(*p.minLen))
	}
	if p.maxLen != nil {
		opts = append(opts, mcpMaxLength(*p.maxLen))
	}
	if p.pattern != nil {
		opts = append(opts, mcpPattern(*p.pattern))
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
