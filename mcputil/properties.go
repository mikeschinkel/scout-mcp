package mcputil

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// Property is the interface that all property types implement
type Property interface {
	propertyEmbed
	mcpToolOption([]mcp.PropertyOption) mcp.ToolOption
}

type propertyEmbed interface {
	GetName() string
	Required() Property
	Name(string) Property
	Description(string) Property
	mcpPropertyOptions() []mcp.PropertyOption
}

var _ propertyEmbed = (*property)(nil)

// property is the base type with only truly shared fields
type property struct {
	name        string
	dataType    PropertyType
	required    bool
	description string
	parent      Property
}

func (p *property) Property() {}

func (p *property) GetName() string {
	return p.name
}

func (p *property) PropertyOptions() []PropertyOption {
	opts := []PropertyOption{
		nameProperty{p.name},
		descriptionProperty{p.description},
	}
	if p.required {
		opts = append(opts, requiredProperty{true})
	}
	return opts
}

func (p *property) mcpPropertyOptions() []mcp.PropertyOption {
	opts := []mcp.PropertyOption{mcp.Description(p.description)}
	if p.required {
		opts = append(opts, mcp.Required())
	}
	return opts
}

func (p *property) Required() Property {
	p.required = true
	return p.parent
}

func (p *property) Name(name string) Property {
	p.name = name
	return p.parent
}

func (p *property) Description(desc string) Property {
	p.description = desc
	return p.parent
}

type requiredProperty struct {
	required bool
}

func (requiredProperty) PropertyOption() {}

type nameProperty struct {
	name string
}

func (nameProperty) PropertyOption() {}

type descriptionProperty struct {
	description string
}

func (descriptionProperty) PropertyOption() {}
