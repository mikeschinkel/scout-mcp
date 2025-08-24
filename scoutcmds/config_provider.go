package scoutcmds

import (
	"io"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

var _ scout.ConfigProvider = (*configProvider)(nil)

// configProvider implements the scout.ConfigProvider interface
type configProvider struct {
}

func (cp *configProvider) config() *Config {
	sc, ok := cp.GetConfig().(*Config)
	if !ok {
		panic("Can't get config from ConfigProvider")
	}
	return sc
}

func (cp *configProvider) SetIO(r io.Reader, w io.Writer) {
	cfg := cp.config()
	cfg.Reader = r
	cfg.Writer = w
}

func (cp *configProvider) GetIO() (io.Reader, io.Writer) {
	cfg := cp.config()
	return cfg.Reader, cfg.Writer
}

// GetConfig returns the global config instance
func (cp *configProvider) GetConfig() cliutil.Config {
	return GetConfig()
}

// GlobalFlagSet returns the global flag set
func (cp *configProvider) GlobalFlagSet() *cliutil.FlagSet {
	return GlobalFlagSet
}

// NewConfigProvider creates a new configProvider instance
func NewConfigProvider() scout.ConfigProvider {
	return &configProvider{}
}
