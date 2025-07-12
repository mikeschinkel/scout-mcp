package mcputil

// NumberOption handles option types and marker interface for MCP Number properties
type NumberOption interface {
	SetNumberProperty(Property)
}

type MinMax = MinMaxFloat // Alias for convenience with literals

type MinMaxFloat [2]float64

func (MinMaxFloat) PropertyOption() {}
func (mmf MinMaxFloat) SetNumberProperty(prop Property) {
	p := prop.(*numberProperty)
	p.Min = &mmf[0]
	p.Max = &mmf[1]
}

type MinMaxInt [2]int64

func (MinMaxInt) PropertyOption() {}
func (mmi MinMaxInt) SetNumberProperty(prop Property) {
	minVal := float64(mmi[0])
	maxVal := float64(mmi[1])
	p := prop.(*numberProperty)
	p.Min = &minVal
	p.Max = &maxVal
}

type MinFloat [1]float64

func (mf MinFloat) SetNumberProperty(prop Property) {
	prop.(*numberProperty).Min = &mf[0]
}
func (MinFloat) PropertyOption() {}

type MinInt [1]int64

func (mi MinInt) SetNumberProperty(prop Property) {
	mf := float64(mi[0])
	prop.(*numberProperty).Min = &mf
}
func (MinInt) PropertyOption() {}

type MinNumber = MinFloat // Alias for convenience with literals

type MaxFloat [1]float64

func (mf MaxFloat) SetNumberProperty(prop Property) {
	prop.(*numberProperty).Max = &mf[0]
}
func (MaxFloat) PropertyOption() {}

type MaxInt [1]int64

func (mi MaxInt) SetNumberProperty(prop Property) {
	mf := float64(mi[0])
	prop.(*numberProperty).Max = &mf
}
func (MaxInt) PropertyOption() {}

type MaxNumber = MaxFloat // Alias for convenience with literals

type DefaultFloat [1]float64

func (df DefaultFloat) SetNumberProperty(prop Property) {
	prop.(*numberProperty).Default = &df[0]
}
func (DefaultFloat) PropertyOption() {}

type DefaultInt [1]int64

func (di DefaultInt) SetNumberProperty(prop Property) {
	df := float64(di[0])
	prop.(*numberProperty).Default = &df
}
func (DefaultInt) PropertyOption() {}

type DefaultNumber = DefaultFloat // Alias for convenience with literals

type NumberOptions struct {
	Min     *float64
	Max     *float64
	Default *float64
}

func (no NumberOptions) SetNumberProperty(prop Property) {
	p := prop.(*numberProperty)
	if no.Min != nil {
		p.Min = no.Min
	}
	if no.Max != nil {
		p.Max = no.Max
	}
	if no.Default != nil {
		p.Default = no.Default
	}
}
