package mcptools_test

import (
	"maps"
	"slices"

	"github.com/mikeschinkel/scout-mcp/mcptools"
)

var toolNames = slices.Collect(maps.Keys(toolNamesMap))

// Create set of expected tools
var toolNamesMap = map[string]struct{}{
	"start_session":           mcptools.NULL,
	"read_files":              mcptools.NULL,
	"search_files":            mcptools.NULL,
	"get_config":              mcptools.NULL,
	"tool_help":               mcptools.NULL,
	"create_file":             mcptools.NULL,
	"update_file":             mcptools.NULL,
	"delete_files":            mcptools.NULL,
	"update_file_lines":       mcptools.NULL,
	"delete_file_lines":       mcptools.NULL,
	"insert_file_lines":       mcptools.NULL,
	"insert_at_pattern":       mcptools.NULL,
	"replace_pattern":         mcptools.NULL,
	"find_file_part":          mcptools.NULL,
	"replace_file_part":       mcptools.NULL,
	"validate_files":          mcptools.NULL,
	"analyze_files":           mcptools.NULL,
	"request_approval":        mcptools.NULL,
	"generate_approval_token": mcptools.NULL,
}
