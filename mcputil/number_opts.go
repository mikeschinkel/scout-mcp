package mcputil

// NumberOption handles option types and marker interface for MCP Number properties
type NumberOption interface {
	SetNumberProperty(Property)
}

// MinMax is an alias for MinMaxFloat for convenience with literals.
type MinMax = MinMaxFloat

// MinMaxFloat sets both minimum and maximum values for a number property.
type MinMaxFloat [2]float64

// PropertyOption implements PropertyOption interface for MinMaxFloat.
func (MinMaxFloat) PropertyOption() {}

// SetNumberProperty configures both minimum and maximum float values for a number property.
// This method sets both the min and max constraints in a single operation.
func (mmf MinMaxFloat) SetNumberProperty(prop Property) {
	p := prop.(*numberProperty)
	p.min = &mmf[0]
	p.max = &mmf[1]
}

// MinMaxInt sets both minimum and maximum integer values for a number property.
type MinMaxInt [2]int64

// PropertyOption implements PropertyOption interface for MinMaxInt.
func (MinMaxInt) PropertyOption() {}

// SetNumberProperty configures both minimum and maximum integer values for a number property.
// This method converts integer values to float64 and sets both min and max constraints.
func (mmi MinMaxInt) SetNumberProperty(prop Property) {
	minVal := float64(mmi[0])
	maxVal := float64(mmi[1])
	p := prop.(*numberProperty)
	p.min = &minVal
	p.max = &maxVal
}

// MinFloat sets the minimum float value for a number property.
type MinFloat [1]float64

// SetNumberProperty configures the minimum float value for a number property.
// This method sets only the minimum constraint, leaving maximum unconstrained.
func (mf MinFloat) SetNumberProperty(prop Property) {
	prop.(*numberProperty).min = &mf[0]
}

// PropertyOption implements PropertyOption interface for MinFloat.
func (MinFloat) PropertyOption() {}

// MinInt sets the minimum integer value for a number property.
type MinInt [1]int64

// SetNumberProperty configures the minimum integer value for a number property.
// This method converts the integer to float64 and sets only the minimum constraint.
func (mi MinInt) SetNumberProperty(prop Property) {
	mf := float64(mi[0])
	prop.(*numberProperty).min = &mf
}

// PropertyOption implements PropertyOption interface for MinInt.
func (MinInt) PropertyOption() {}

// MinNumber is an alias for MinFloat for convenience with literals.
type MinNumber = MinFloat

// MaxFloat sets the maximum float value for a number property.
type MaxFloat [1]float64

// SetNumberProperty configures the maximum float value for a number property.
// This method sets only the maximum constraint, leaving minimum unconstrained.
func (mf MaxFloat) SetNumberProperty(prop Property) {
	prop.(*numberProperty).max = &mf[0]
}

// PropertyOption implements PropertyOption interface for MaxFloat.
func (MaxFloat) PropertyOption() {}

// MaxInt sets the maximum integer value for a number property.
type MaxInt [1]int64

// SetNumberProperty configures the maximum integer value for a number property.
// This method converts the integer to float64 and sets only the maximum constraint.
func (mi MaxInt) SetNumberProperty(prop Property) {
	mf := float64(mi[0])
	prop.(*numberProperty).max = &mf
}

// PropertyOption implements PropertyOption interface for MaxInt.
func (MaxInt) PropertyOption() {}

// MaxNumber is an alias for MaxFloat for convenience with literals.
type MaxNumber = MaxFloat

// DefaultFloat sets the default float value for a number property.
type DefaultFloat [1]float64

// SetNumberProperty configures the default float value for a number property.
// This method sets the property's default value to the specified float64.
func (df DefaultFloat) SetNumberProperty(prop Property) {
	prop.(*numberProperty).defaultValue = &df[0]
}

// PropertyOption implements PropertyOption interface for DefaultFloat.
func (DefaultFloat) PropertyOption() {}

// DefaultInt sets the default integer value for a number property.
type DefaultInt [1]int64

// SetNumberProperty configures the default integer value for a number property.
// This method converts the integer to float64 and sets it as the default value.
func (di DefaultInt) SetNumberProperty(prop Property) {
	df := float64(di[0])
	prop.(*numberProperty).defaultValue = &df
}

// PropertyOption implements PropertyOption interface for DefaultInt.
func (DefaultInt) PropertyOption() {}

// ZeroOK allows zero values for a number property.
type ZeroOK [1]bool

// SetNumberProperty configures whether zero values are acceptable for a number property.
// This method sets the zeroOK flag to control zero value validation.
func (di ZeroOK) SetNumberProperty(prop Property) {
	zok := di[0]
	prop.(*numberProperty).zeroOK = &zok
}

// PropertyOption implements PropertyOption interface for ZeroOK.
func (ZeroOK) PropertyOption() {}

// DefaultNumber is an alias for DefaultFloat for convenience with literals.
type DefaultNumber = DefaultFloat

// NumberOptions provides a convenient way to set multiple number property options at once.
type NumberOptions struct {
	Min     *float64 // Minimum allowed value
	Max     *float64 // Maximum allowed value
	Default *float64 // Default numeric value
}

// PropertyOption implements PropertyOption interface for NumberOptions.
func (NumberOptions) PropertyOption() {}

// DefaultInt returns the default value as an integer.
// This method converts the float64 default value to int64 for convenience.
func (no NumberOptions) DefaultInt() int64 {
	return int64(*no.Default)
}

// SetNumberProperty configures multiple number constraints and default values in a single operation.
// This method allows setting minimum, maximum, and default values together.
func (no NumberOptions) SetNumberProperty(prop Property) {
	p := prop.(*numberProperty)
	if no.Min != nil {
		p.min = no.Min
	}
	if no.Max != nil {
		p.max = no.Max
	}
	if no.Default != nil {
		p.defaultValue = no.Default
	}
}
