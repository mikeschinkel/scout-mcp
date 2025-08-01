package mcptools_test

import (
	"github.com/mikeschinkel/scout-mcp/mcptools"
)

// Create set of expected tools
var toolNamesMap = map[string]mcptools.NULL{
	"start_session":           {},
	"read_files":              {},
	"search_files":            {},
	"get_config":              {},
	"help":                    {},
	"create_file":             {},
	"update_file":             {},
	"delete_files":            {},
	"update_file_lines":       {},
	"delete_file_lines":       {},
	"insert_file_lines":       {},
	"insert_at_pattern":       {},
	"replace_pattern":         {},
	"find_file_part":          {},
	"replace_file_part":       {},
	"validate_files":          {},
	"analyze_files":           {},
	"request_approval":        {},
	"generate_approval_token": {},
	"detect_current_project":  {},
}
