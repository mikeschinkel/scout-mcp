package scoutcmds

import (
	"io"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

// Config holds Scout CLI configuration
type Config struct {
	RunArgs *scout.RunArgs

	Reader io.Reader
	Writer io.Writer

	// Global options
	ConfigPath *string
	Verbose    *bool

	// MCP server options
	OnlyMode        *bool
	AdditionalPaths []string

	// Session options
	SessionToken *string

	// Tool options
	ToolName *string
	ToolArgs []string
}

// Config implements the cliutil.Config interface
func (c *Config) Config() {}

var cfg = &Config{
	ConfigPath:   new(string),
	Verbose:      new(bool),
	OnlyMode:     new(bool),
	SessionToken: new(string),
	ToolName:     new(string),
}

// GetConfig returns the global config instance
func GetConfig() cliutil.Config {
	return cfg
}
