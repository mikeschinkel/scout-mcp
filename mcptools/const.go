package mcptools

import (
	"fmt"
)

type NULL = struct{}

type RelativePosition string

const (
	BeforePosition RelativePosition = "before"
	AfterPosition  RelativePosition = "after"
)

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

// ToolNamesMap is the set of expected tools
var ToolNamesMap = map[string]NULL{
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
