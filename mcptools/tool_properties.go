package mcptools

import (
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// Required property aliases for commonly used tool parameters.
var (
	RequiredSessionTokenProperty = mcputil.RequiredSessionTokenProperty
	RequiredFilesProperty        = FilesProperty.Required()
	RequiredPathProperty         = PathProperty.Required()
	RequiredNewContentProperty   = NewContentProperty.Required()
	RequiredPathsProperty        = PathsProperty.Required()
)

// Property definitions for MCP tool parameters with descriptions and defaults.
var (
	FilesProperty          = mcputil.Array("files", "List of files to process")
	PathsProperty          = mcputil.Array("paths", "File or directory paths to use with this tool")
	RecursiveProperty      = mcputil.Bool("recursive", "Process directories recursively", mcputil.DefaultTrue{})
	StartLineProperty      = mcputil.Number("start_line", "First line to handle, inclusive")
	EndLineProperty        = mcputil.Number("end_line", "Last line to handle, inclusive")
	FilepathProperty       = mcputil.String("filepath", "File path to use for this tool")
	PathProperty           = mcputil.String("path", "File or directory path to use with this tool")
	NewContentProperty     = mcputil.String("new_content", "New file content to use with this tool")
	LanguageProperty       = mcputil.String("language", "Programming language of file(s) to process")
	PartTypeProperty       = mcputil.String("part_type", "Type of the part of the programming language to process")
	PartNameProperty       = mcputil.String("part_name", "Name for the part to process")
	PositionProperty       = mcputil.String("position", "Position to use with this tool")
	LineNumberProperty     = mcputil.Number("line_number", "Line number to use with this tool")
	SessionIdProperty      = mcputil.String("session_id", "Session ID returned by start_session")
	PatternProperty        = mcputil.String("pattern", "Text pattern to find")
	ReplacementProperty    = mcputil.String("replacement", "Text to replace the pattern with")
	RegexProperty          = mcputil.Bool("regex", "Whether to treat pattern as regular expression")
	AllOccurrencesProperty = mcputil.Bool("all_occurrences", "Whether to replace all occurrences (default: true)", mcputil.DefaultBool{true})
	ExtensionsProperty     = mcputil.Array("extensions", "Filter by file extensions (e.g., ['.go', '.txt'])")
	NamePatternProperty    = mcputil.String("name_pattern", "Exact filename pattern to match")
	FilesOnlyProperty      = mcputil.Bool("files_only", "Return only files, not directories")
	DirsOnlyProperty       = mcputil.Bool("dirs_only", "Return only directories, not files")
	MaxResultsProperty     = mcputil.Number("max_results", "Maximum number of results to return")
	MaxFilesProperty       = mcputil.Number("max_files", "Maximum number of files to read (default: 100)", mcputil.DefaultInt{100})
	CreateDirsProperty     = mcputil.Bool("create_dirs", "Create parent directories if needed")
	MaxProjectsProperty    = mcputil.Number("max_projects", "Maximum number of recent projects to track (default: 5)", mcputil.DefaultInt{5})
	IgnoreGitProperty      = mcputil.Bool("ignore_git_requirement", "If true, don't require .git directory to consider a directory a project (default: false)")

	OffsetProperty = mcputil.Number("offset", "Number of items to skip for pagination (default: 0)", mcputil.DefaultInt{0})
)
