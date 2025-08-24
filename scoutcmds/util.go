package scoutcmds

import (
	"fmt"
	"io"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/cliutil"
)

// convertConfig converts CLI config to Scout domain config
func convertConfig(config cliutil.Config, args []string) (opts *scout.Opts, err error) {
	cfg, ok := config.(*Config)
	if !ok {
		err = fmt.Errorf("invalid config type '%T'; %v", config, config)
		goto end
	}

	opts = &scout.Opts{
		OnlyMode:        *cfg.OnlyMode,
		AdditionalPaths: append(cfg.AdditionalPaths, args...),
		MCPReader:       scout.NewNormalizingReader(cfg.Reader),
		MCPWriter:       cfg.Writer,
	}
end:
	return opts, err
}

func fprintf(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}
