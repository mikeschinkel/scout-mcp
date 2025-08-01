package mcputil

// ArrayOption handles option types and marker interface for MCP Array properties
type ArrayOption interface {
	SetArrayProperty(Property)
}

type MinMaxItems [2]int

func (MinMaxItems) PropertyOption() {}
func (mmi MinMaxItems) SetArrayProperty(prop Property) {
	p := prop.(*arrayProperty)
	p.minItems = &mmi[0]
	p.maxItems = &mmi[1]
}

type MinItems [1]int

func (MinItems) PropertyOption() {}
func (mi MinItems) SetArrayProperty(prop Property) {
	prop.(*arrayProperty).minItems = &mi[0]
}

type MaxItems [1]int

func (MaxItems) PropertyOption() {}
func (mi MaxItems) SetArrayProperty(prop Property) {
	prop.(*arrayProperty).maxItems = &mi[0]
}

type DefaultArray[T any] []T

func (DefaultArray[T]) PropertyOption() {}
func (da DefaultArray[T]) SetArrayProperty(prop Property) {
	defaultArray := make([]any, len(da))
	for i, p := range da {
		defaultArray[i] = p
	}
	prop.(*arrayProperty).defaultValue = defaultArray
}

type ArrayOptions struct {
	MinItems *int
	MaxItems *int
	Default  DefaultArray[any]
}

func (ao ArrayOptions) SetArrayProperty(prop Property) {
	p := prop.(*arrayProperty)
	if ao.MinItems != nil {
		p.minItems = ao.MinItems
	}
	if ao.MaxItems != nil {
		p.maxItems = ao.MaxItems
	}
	if ao.Default != nil {
		p.defaultValue = ao.Default
	}
}
