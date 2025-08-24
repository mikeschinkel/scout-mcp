package cliutil

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// CLAUDE: I renamed to globalFlags because "Handler"  is GARBAJE.
// CLAUDE: Also, I restructured and renamed interface because gmover-specific flags should not be encoded into generic cliutil

type GlobalFlagDefGetter interface {
	GlobalFlagDefs() []FlagDef
}

type CmdRunner struct {
	config        Config
	globalFlagSet *FlagSet
	args          []string
}
type CmdRunnerArgs struct {
	Config        Config
	GlobalFlagSet *FlagSet
	Args          []string
}

func NewCmdRunner(args CmdRunnerArgs) *CmdRunner {
	return &CmdRunner{
		config:        args.Config,
		globalFlagSet: args.GlobalFlagSet,
		args:          args.Args,
	}
}

func (cr CmdRunner) Run(ctx context.Context) (err error) {
	var cmd Command
	var path string
	var args []string
	var handler CommandHandler
	var ok bool

	// Validate commands first
	err = ValidateCmds()
	if err != nil {
		goto end
	}

	if len(cr.args) == 0 {
		err = ShowMainHelp()
		goto end
	}

	// Parse global flags and extract remaining args
	args, err = cr.globalFlagSet.Parse(cr.args)
	if err != nil {
		goto end
	}

	// Try to find the most specific command match
	path, args = findBestCmdMatch(args)
	if path == "" {
		err = fmt.Errorf("unknown command: %s\nRun 'scout help' for usage", args[0])
		goto end
	}

	cmd, path = GetDefaultCommand(path, args)
	if cmd == nil {
		err = fmt.Errorf("command not found: %s", path)
		goto end
	}

	args, err = cmd.ParseFlagSets(args, cr.config)
	if err != nil {
		goto end
	}

	err = cmd.AssignArgs(args)
	if err != nil {
		goto end
	}

	// Command resolution should ensure we only get CommandHandler implementations
	handler, ok = cmd.(CommandHandler)
	if !ok {
		err = fmt.Errorf("command '%s' does not implement handler logic", cmd.Name())
		goto end
	}

	err = handler.Handle(ctx, cr.config, args)

end:
	return err
}

// findBestCmdMatch finds the longest matching command path
func findBestCmdMatch(args []string) (path string, remainingArgs []string) {
	var cmd Command
	var tryPath string
	var n int
	tryPaths := make([]string, len(args))

	for i, arg := range args {
		if strings.HasPrefix(arg, "-") {
			break
		}
		tryPath = fmt.Sprintf("%s.%s", tryPath, arg)
		if i == 0 {
			tryPath = strings.TrimLeft(tryPath, ".")
		}
		n++
		tryPaths[len(tryPaths)-i-1] = tryPath
	}
	if n < len(args) {
		tryPaths = tryPaths[len(tryPaths)-n:]
	}

	// Try progressively longer paths
	for _, p := range tryPaths {
		cmd, p = GetDefaultCommand(p, args)
		if cmd != nil {
			path = p
			remainingArgs = args[n:]
			break
		}
		n--
	}

	// If no match found, return empty path with original args
	if path == "" {
		remainingArgs = args
	}

	return path, remainingArgs
}

// ShowMainHelp displays the main help screen
func ShowMainHelp() (err error) {
	Printf(`Scout - Secure file operations for Claude Desktop

USAGE:
    scout <command> [subcommand] [options]

COMMANDS:
`)

	// Show all top-level commands
	topCmds := GetTopLevelCmds()
	// Sort commands by name for deterministic output
	sort.Slice(topCmds, func(i, j int) bool {
		return topCmds[i].Name() < topCmds[j].Name()
	})
	for _, cmd := range topCmds {
		subCmds := GetSubCmds(cmd.Name())
		subCmdText := ""
		if len(subCmds) > 0 {
			// Sort subcommands for deterministic output
			sort.Slice(subCmds, func(i, j int) bool {
				return subCmds[i].Name() < subCmds[j].Name()
			})
			subCmdText = fmt.Sprintf(" [%s]", subCmds[0].Name()) // Show first subcommand as example
		}
		Printf("    %-20s %s\n", cmd.Name()+subCmdText, cmd.Description())
	}

	Printf(`
EXAMPLES:
    # Start MCP server
    scout mcp
    scout mcp --only /safe/directory
    
    # Session management
    scout session new
    scout session list
    scout session clear all         # Clear all sessions
    scout session clear <token>     # Clear specific session
    
    # Tool operations
    scout tool list
    scout tool run read_files /path/to/file

For more information, visit: https://github.com/mikeschinkel/scout-mcp
`)
	return err
}

// ShowCmdHelp displays help for a specific command
func ShowCmdHelp(cmdName string) (err error) {
	var cmd Command
	var subCmds []Command

	cmd = GetExactCommand(cmdName)
	if cmd == nil {
		err = fmt.Errorf("unknown command: %s", cmdName)
		goto end
	}

	Printf("Usage: %s\n\n%s\n", cmd.Usage(), cmd.Description())

	subCmds = GetSubCmds(cmdName)
	if len(subCmds) > 0 {
		Printf("\nSubcommands:\n")
		for _, subCmd := range subCmds {
			Printf("    %-12s %s\n", subCmd.Name(), subCmd.Description())
		}
	}

end:
	return err
}
