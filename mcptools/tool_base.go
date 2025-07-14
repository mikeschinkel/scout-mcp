package mcptools

import (
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

type Config = mcputil.Config

type toolBase struct {
	config  Config
	options mcputil.ToolOptions
}

func newToolBase(options mcputil.ToolOptions) *toolBase {
	return &toolBase{
		options: options,
	}
}

func (b *toolBase) IsAllowedPath(path string) (bool, error) {
	return b.config.IsAllowedPath(path)
}
func (b *toolBase) Name() string {
	return b.options.Name
}

func (b *toolBase) SetConfig(c Config) {
	b.config = c
}
func (b *toolBase) Config() Config {
	return b.config
}

func (b *toolBase) Options() mcputil.ToolOptions {
	return b.options
}
