package mcptools

import (
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var (
	RequiredSessionTokenProperty = mcputil.String("session_token", "Session token from start_session").Required()
)

var (
	FilesProperty      = mcputil.Array("files", "List of files to process ro be affected")
	PathsProperty      = mcputil.Array("paths", "File or directory paths to use with this tool")
	RecursiveProperty  = mcputil.Bool("recursive", "Process directories recursively")
	StartLineProperty  = mcputil.Number("start_line", "First line to handle, inclusive")
	EndLineProperty    = mcputil.Number("end_line", "Last line to handle, inclusive")
	FilepathProperty   = mcputil.String("filepath", "File path to use for this tool")
	PathProperty       = mcputil.String("path", "File or directory path to use with this tool")
	NewContentProperty = mcputil.String("new_content", "New file content to use with this tool")
	LanguageProperty   = mcputil.String("language", "Programming language of file(s) to process")
	PartTypeProperty   = mcputil.String("part_type", "Type of the part of the programming language to process")
	PartNameProperty   = mcputil.String("part_name", "Name for the part to process")
)
