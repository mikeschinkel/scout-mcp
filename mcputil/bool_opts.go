package mcputil

// BoolOption handles option types and marker interface for MCP Bool properties
type BoolOption interface {
	SetBoolProperty(Property)
}

type DefaultBool [1]bool

func (DefaultBool) BoolOpt() {}
func (db DefaultBool) SetBoolProperty(prop Property) {
	prop.(*boolProperty).defaultValue = &db[0]
}

type DefaultTrue struct{}

func (DefaultTrue) BoolOpt() {}
func (DefaultTrue) SetBoolProperty(prop Property) {
	b := true
	prop.(*boolProperty).defaultValue = &b
}

type DefaultFalse struct{}

func (DefaultFalse) BoolOpt() {}
func (DefaultFalse) SetBoolProperty(prop Property) {
	b := false
	prop.(*boolProperty).defaultValue = &b
}

type BoolOptions struct {
	Default *bool
}

func (BoolOptions) BoolOpt() {}
func (bo BoolOptions) SetBoolProperty(prop Property) {
	if bo.Default != nil {
		prop.(*boolProperty).defaultValue = bo.Default
	}
}
