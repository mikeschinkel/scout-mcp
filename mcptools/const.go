// Package mcptools provides MCP tool implementations with session management and approval workflows.
// All tools follow a consistent pattern with session validation and risk-based approval for file operations.
package mcptools

import (
	"fmt"
)

// NULL is a zero-memory type used as a value for existence maps.
type NULL = struct{}

// RelativePosition specifies where to insert content relative to a pattern.
type RelativePosition string

const (
	BeforePosition RelativePosition = "before" // Insert content before the pattern
	AfterPosition  RelativePosition = "after"  // Insert content after the pattern
)

// Validate checks if the RelativePosition has a valid value.
func (rp RelativePosition) Validate() (err error) {
	switch rp {
	case BeforePosition:
	case AfterPosition:
	default:
		err = fmt.Errorf("position must be '%s' or '%s', got '%s'",
			BeforePosition,
			AfterPosition,
			rp,
		)
	}
	return err
}

// ToolNamesMap contains all supported MCP tool names for validation purposes.
var ToolNamesMap = map[string]NULL{
	"start_session":          {},
	"read_files":             {},
	"search_files":           {},
	"get_config":             {},
	"help":                   {},
	"create_file":            {},
	"update_file":            {},
	"delete_files":           {},
	"update_file_lines":      {},
	"delete_file_lines":      {},
	"insert_file_lines":      {},
	"insert_at_pattern":      {},
	"replace_pattern":        {},
	"find_file_part":         {},
	"replace_file_part":      {},
	"validate_files":         {},
	"analyze_files":          {},
	"request_approval":       {},
	"detect_current_project": {},
}
