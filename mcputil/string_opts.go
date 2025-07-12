package mcputil

// StringOption handles option types and marker interface for MCP String properties
type StringOption interface {
	SetStringProperty(Property)
}

type Enum []string

func (Enum) PropertyOption() {}
func (e Enum) SetStringProperty(prop Property) {
	prop.(*stringProperty).Enum = e
}

type MinLen [1]int

func (MinLen) PropertyOption() {}
func (ml MinLen) SetStringProperty(prop Property) {
	prop.(*stringProperty).MinLen = &ml[0]
}

type MaxLen [1]int

func (MaxLen) PropertyOption() {}
func (ml MaxLen) SetStringProperty(prop Property) {
	prop.(*stringProperty).MaxLen = &ml[0]
}

type DefaultString [1]string

func (DefaultString) PropertyOption() {}
func (ds DefaultString) SetStringProperty(prop Property) {
	prop.(*stringProperty).Default = &ds[0]
}

type Pattern [1]string

func (Pattern) PropertyOption() {}
func (p Pattern) SetStringProperty(prop Property) {
	prop.(*stringProperty).Pattern = &p[0]
}

type StringOptions struct {
	Enum    []string
	MinLen  *int
	MaxLen  *int
	Default *string
	Pattern *string
}

func (so StringOptions) SetStringProperty(prop Property) {
	p := prop.(*stringProperty)
	if so.Enum != nil {
		p.Enum = so.Enum
	}
	if so.MinLen != nil {
		p.MinLen = so.MinLen
	}
	if so.MaxLen != nil {
		p.MaxLen = so.MaxLen
	}
	if so.Default != nil {
		p.Default = so.Default
	}
	if so.Pattern != nil {
		p.Pattern = so.Pattern
	}
}
