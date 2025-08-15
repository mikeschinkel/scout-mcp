package mcptools

import (
	"fmt"

	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// ReadFile is an alias for mcputil.ReadFile for convenience.
var ReadFile = mcputil.ReadFile

// WriteFile writes content to a file with syntax validation for supported languages.
func WriteFile(c mcputil.Config, filePath string, content string) (err error) {
	var language string

	lf := langutil.NewFile(filePath)
	err = lf.Initialize()
	if err != nil {
		goto end
	}
	err = lf.ValidateSyntax(content)
	if err != nil {
		err = fmt.Errorf("validation failed - would result in invalid %s syntax: %w", language, err)
		goto end
	}

	err = mcputil.WriteFile(c, filePath, content)

end:
	return err
}
