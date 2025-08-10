package scout

type Args struct {
	AdditionalPaths []string
	ServerOpts      MCPServerOpts
	IsInit          bool
	InitialPath     string
}

func ParseArgs(osArgs []string) (args Args, err error) {
	var i int
	var arg string

	if len(osArgs) == 0 {
		goto end
	}

	// Check for init command
	if osArgs[0] == "init" {
		args.IsInit = true
		if len(osArgs) > 1 {
			args.InitialPath = osArgs[1]
		}
		goto end
	}

	// Parse flags and paths
	for i = 0; i < len(osArgs); i++ {
		arg = osArgs[i]

		if arg == "--only" {
			args.ServerOpts.OnlyMode = true
			continue
		}

		// Validate and add path
		err = validatePath(arg)
		if err != nil {
			goto end
		}

		args.AdditionalPaths = append(args.AdditionalPaths, arg)
	}

end:
	return args, err
}
