package scout

import (
	"io"

	"github.com/mikeschinkel/scout-mcp/cliutil"
)

type Opts struct {
	OnlyMode        bool
	AdditionalPaths []string
	MCPReader       io.Reader
	MCPWriter       io.Writer
	CLIWriter       cliutil.OutputWriter
}
