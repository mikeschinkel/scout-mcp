package cliutil

// ArgDef defines a positional command argument
type ArgDef struct {
	Name     string
	Usage    string
	Required bool
	String   *string // Where to assign the argument value
}
