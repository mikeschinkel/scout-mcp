package mcputil

import (
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// Property is the interface that all property types implement
type Property interface {
	propertyEmbed
	Clone() Property
	DefaultValue() any
	SetDefault(any) Property
	mcpToolOption([]mcp.PropertyOption) mcp.ToolOption
}

type propertyEmbed interface {
	GetName() string
	GetType() PropertyType
	IsRequired() bool
	Required() Property
	Name(string) Property
	Description(string) Property
	PropertyOptions() []PropertyOption
	mcpPropertyOptions() []mcp.PropertyOption
	String(request ToolRequest) (string, error)
	AnySlice(request ToolRequest) ([]any, error)
	StringSlice(request ToolRequest) ([]string, error)
	Bool(request ToolRequest) (bool, error)
	Number(request ToolRequest) (float64, error)
	Int(request ToolRequest) (int, error)
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

func (p *property) GetType() PropertyType {
	return p.dataType
}

func (p *property) IsRequired() bool {
	return p.required
}

func (p *property) PropertyOptions() []PropertyOption {
	opts := []PropertyOption{
		NameProperty{p.name},
		DescriptionProperty{p.description},
	}
	if p.required {
		opts = append(opts, RequiredProperty{true})
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
	np := *p
	np.required = true
	np.parent = p.parent.Clone()
	return np.parent
}

func (p *property) Name(name string) Property {
	np := *p
	np.name = name
	np.parent = p.parent.Clone()
	return np.parent
}

func (p *property) Default() any {
	return p.parent.DefaultValue()
}

// String return string value. If a default value it set we use that but if no
// default value is set and the property does not exist, we throw an error.
func (p *property) String(tr ToolRequest) (s string, err error) {
	var sp *string
	var value string
	var ok bool
	if p.required {
		s, err = tr.RequireString(p.name)
		goto end
	}
	sp, ok = p.Default().(*string)
	if !ok {
		err = fmt.Errorf("string expected for %s; got %T instead", p.name, p.Default())
	}
	if sp != nil {
		value = *sp
	}
	s = tr.GetString(p.name, value)
end:
	return s, err
}

// StringSlice return StringSlice value. If a default value it set we use that but if no
// default value is set and the property does not exist, we throw an error.
func (p *property) StringSlice(tr ToolRequest) (s []string, err error) {
	var a []any
	if p.required {
		a, err = tr.RequireArray(p.name)
	} else {
		a = tr.GetArray(p.name, convertContainedSlice(p.Default()))
	}
	if err != nil {
		goto end
	}
	s, err = convertSliceOfAny[string](a)
end:
	return s, err
}

// AnySlice return AnySlice value. If a default value it set we use that but if no
// default value is set and the property does not exist, we throw an error.
func (p *property) AnySlice(tr ToolRequest) (s []any, err error) {
	if p.required {
		s = tr.GetArray(p.name, convertContainedSlice(p.Default()))
		goto end
	}
	s, err = tr.RequireArray(p.name)
end:
	return s, err
}

// Bool return bool value. If a default value it set we use that but if no
// default value is set and the property does not exist, we throw an error.
func (p *property) Bool(tr ToolRequest) (b bool, err error) {
	var bp *bool
	var value, ok bool
	if p.parent.IsRequired() {
		b, err = tr.RequireBool(p.name)
		goto end
	}
	bp, ok = p.Default().(*bool)
	if !ok {
		err = fmt.Errorf("bool expected for %s; got %T instead", p.name, p.Default())
	}
	if bp != nil && *bp {
		value = true
	}
	b = tr.GetBool(p.name, value)
end:
	return b, err
}

// Number return float64 value. If a default value it set we use that but if no
// default value is set and the property does not exist, we throw an error.
func (p *property) Number(tr ToolRequest) (n float64, err error) {
	var sp *float64
	var value float64
	var ok bool
	if p.parent.IsRequired() {
		n, err = tr.RequireFloat(p.name)
		goto end
	}
	value, ok = p.Default().(float64)
	if !ok {
		err = fmt.Errorf("float64 expected for %s when calling Number(); got %T instead", p.name, p.Default())
	}
	sp, ok = p.Default().(*float64)
	if !ok {
		err = fmt.Errorf("number (float64) expected for %s when calling Number(); got %T instead", p.name, p.Default())
	}
	if sp != nil {
		value = *sp
	}
	n = tr.GetFloat(p.name, value)
end:
	return n, err
}

// Int return int value. If a default value it set we use that but if no
// default value is set and the property does not exist, we throw an error.
func (p *property) Int(tr ToolRequest) (n int, err error) {
	var sp *float64
	var value int
	var np *numberProperty
	var ok bool

	np, ok = p.parent.(*numberProperty)
	if !ok {
		err = fmt.Errorf("*numberProperty expected for %s when calling Int(); got %T instead", p.name, p.parent)
		goto end
	}

	if p.parent.IsRequired() {
		n, err = tr.RequireInt(p.name)
		goto end
	}

	sp, ok = p.Default().(*float64)
	if !ok {
		err = fmt.Errorf("number (float64) expected for %s when calling Int(); got %T instead", p.name, p.Default())
	}
	if sp != nil {
		value = int(*sp)
	}
	n = tr.GetInt(p.name, value)
end:
	if err == nil && n == 0 && np.ZeroOK() {
		err = fmt.Errorf("'%s' must be a valid number: %w", p.name, err)
	}
	return n, err
}

func (p *property) Description(desc string) Property {
	np := *p
	np.description = desc
	np.parent = p.parent.Clone()
	return np.parent
}

type RequiredProperty struct {
	Required bool
}

func (RequiredProperty) PropertyOption() {}

type NameProperty struct {
	Name string
}

func (NameProperty) PropertyOption() {}

type DescriptionProperty struct {
	Description string
}

func (DescriptionProperty) PropertyOption() {}

func TypedPropertySlice[T any](req ToolRequest, prop Property) (s []T, err error) {
	var items []any
	items, err = prop.AnySlice(req)
	if err != nil {
		goto end
	}
	s, err = convertSliceOfAny[T](items)
end:
	return s, err
}
