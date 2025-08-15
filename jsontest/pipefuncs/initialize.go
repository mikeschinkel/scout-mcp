package pipefuncs

func Initialize() error {
	// Call this to trigger all the init() funcs that register the pipe funcs in this package.
	// Reference all pipe function types to ensure their init() functions are called
	_ = JSONPipeFunc{}
	_ = ExistsPipeFunc{}
	_ = NotNullPipeFunc{}
	_ = NotEmptyPipeFunc{}
	_ = LenPipeFunc{}
	return nil
}
