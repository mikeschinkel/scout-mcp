# Scout MCP Tools Documentation

This document provides comprehensive documentation for all available MCP tools in the Scout MCP server.

## Session Management

### `start_session`
**⭐ START HERE:** Creates a session token and provides comprehensive instructions for using Scout-MCP effectively. **This must be called first.**

**Parameters:**
- None required

**Returns:**
- Session token (valid for 24 hours)
- Complete tool documentation
- Server configuration
- Language-specific coding instructions
- Usage guidelines

**Example:**
```json
{
  "tool": "start_session",
  "parameters": {}
}
```

**⚠️ IMPORTANT:** All other tools require the `session_token` parameter returned by this tool.

## File Reading Tools

### `read_files`
Read contents of multiple files and/or directories with filtering options. Much more efficient than reading files individually.

**Parameters:**
- `session_token` (required): Session token from start_session
- `paths` (required): Array of file paths and/or directory paths to read
- `extensions` (optional): Filter by file extensions (e.g., [".go", ".txt"]) - applies to directories only
- `recursive` (optional): Include subdirectories (default: false) - applies to directories only
- `pattern` (optional): Filename pattern to match (case-insensitive substring) - applies to directories only
- `max_files` (optional): Maximum number of files to read (default: 100)

**Usage Examples:**
```json
{
  "tool": "read_files",
  "parameters": {
    "session_token": "your-session-token",
    "paths": ["main.go", "config.go", "README.md"]
  }
}
```

```json
{
  "tool": "read_files", 
  "parameters": {
    "session_token": "your-session-token",
    "paths": ["./mcptools", "./langutil"],
    "extensions": [".go"],
    "recursive": true
  }
}
```

```json
{
  "tool": "read_files",
  "parameters": {
    "session_token": "your-session-token", 
    "paths": ["README.md", "./src", "package.json"],
    "recursive": true,
    "max_files": 50
  }
}
```

**Response Format:**
```json
{
  "files": [
    {
      "path": "/full/path/to/file.go",
      "name": "file.go", 
      "content": "package main...",
      "size": 1234,
      "error": null
    }
  ],
  "total_files": 3,
  "total_size": 5678,
  "errors": ["could not read protected.txt: permission denied"],
  "truncated": false
}
```

### `search_files`
Search for files and directories with various filtering options.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Directory path to search in
- `recursive` (optional): Search subdirectories recursively (default: false)
- `pattern` (optional): Case-insensitive substring to match in filenames
- `name_pattern` (optional): Exact filename pattern with wildcards (e.g., "*.go", "test_*")
- `extensions` (optional): Array of file extensions to filter by (e.g., [".go", ".txt"])
- `files_only` (optional): Return only files, not directories
- `dirs_only` (optional): Return only directories, not files
- `max_results` (optional): Maximum number of results to return (default: 1000)

**Example:**
```json
{
  "tool": "search_files",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project",
    "recursive": true,
    "extensions": [".go"],
    "pattern": "test"
  }
}
```

## File Management Tools

### `create_file`
Creates a new file with specified content. Requires user approval.

**Parameters:**
- `session_token` (required): Session token from start_session
- `filepath` (required): Full path where the file should be created
- `new_content` (required): Content to write to the file
- `create_dirs` (optional): Create parent directories if they don't exist

**Example:**
```json
{
  "tool": "create_file",
  "parameters": {
    "session_token": "your-session-token",
    "filepath": "/Users/mike/project/new_file.go",
    "new_content": "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
    "create_dirs": true
  }
}
```

### `update_file`
**⚠️ DANGEROUS: Replaces entire file content. Use granular editing tools for safer changes.**

Completely replaces the content of an existing file. Use this ONLY when you intend to replace the entire file.

**Parameters:**
- `session_token` (required): Session token from start_session
- `filepath` (required): Full path to the file to update
- `new_content` (required): New content that will replace ALL existing content

**Example:**
```json
{
  "tool": "update_file",
  "parameters": {
    "session_token": "your-session-token",
    "filepath": "/Users/mike/project/config.json",
    "new_content": "{\"version\": \"2.0\", \"name\": \"updated-config\"}"
  }
}
```

**⚠️ WARNING:** This tool replaces the ENTIRE file. For code changes, use granular editing tools instead:
- `update_file_lines` - Update specific line ranges
- `insert_file_lines` - Insert content at specific lines
- `replace_pattern` - Find and replace patterns

### `delete_files`
Deletes a file or directory. Requires user approval.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the file or directory to delete
- `recursive` (optional): Delete directory recursively if it's a directory

**Example:**
```json
{
  "tool": "delete_files",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/old_file.txt"
  }
}
```

## Granular File Editing Tools

**🎯 RECOMMENDED: Use these tools for precise code editing instead of `update_file`**

### `update_file_lines`
Update specific lines in a file by line number range. Much safer than `update_file`.

**Parameters:**
- `session_token` (required): Session token from start_session
- `filepath` (required): Full path to the file
- `start_line` (required): Starting line number (1-based)
- `end_line` (required): Ending line number (1-based, inclusive)
- `new_content` (required): New content to replace the specified line range

**Example:**
```json
{
  "tool": "update_file_lines",
  "parameters": {
    "session_token": "your-session-token",
    "filepath": "/Users/mike/project/main.go",
    "start_line": 15,
    "end_line": 18,
    "new_content": "\tfmt.Println(\"Updated function\")\n\treturn nil"
  }
}
```

### `insert_file_lines`
Insert content at a specific line number.

**Parameters:**
- `session_token` (required): Session token from start_session
- `filepath` (required): Full path to the file
- `line_number` (required): Line number where to insert (1-based)
- `new_content` (required): Content to insert
- `position` (required): "before" or "after" the specified line

**Example:**
```json
{
  "tool": "insert_file_lines",
  "parameters": {
    "session_token": "your-session-token",
    "filepath": "/Users/mike/project/main.go",
    "line_number": 1,
    "new_content": "import \"fmt\"",
    "position": "after"
  }
}
```

### `insert_at_pattern`
Insert content before or after a pattern match in the file.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the file
- `before_pattern` OR `after_pattern` (required): Pattern to search for
- `new_content` (required): Content to insert
- `position` (optional): "before" or "after" the pattern (default: "before")
- `regex` (optional): Use regex pattern matching (default: false)

**Example:**
```json
{
  "tool": "insert_at_pattern",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "after_pattern": "package main",
    "new_content": "\nimport \"fmt\"",
    "position": "after"
  }
}
```

### `delete_file_lines`
Delete specific lines from a file.

**Parameters:**
- `session_token` (required): Session token from start_session
- `filepath` (required): Full path to the file
- `start_line` (required): Starting line number to delete (1-based)
- `end_line` (optional): Ending line number to delete (defaults to start_line for single line)

**Example:**
```json
{
  "tool": "delete_file_lines",
  "parameters": {
    "session_token": "your-session-token",
    "filepath": "/Users/mike/project/main.go",
    "start_line": 25,
    "end_line": 30
  }
}
```

### `replace_pattern`
Find and replace text patterns with support for regex.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the file
- `pattern` (required): Text pattern to find
- `replacement` (required): Text to replace the pattern with
- `regex` (optional): Use regex pattern matching (default: false)
- `all_occurrences` (optional): Replace all occurrences or just the first (default: true)

**Example:**
```json
{
  "tool": "replace_pattern",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "pattern": "oldFunctionName",
    "replacement": "newFunctionName",
    "all_occurrences": true
  }
}
```

**Regex Example:**
```json
{
  "tool": "replace_pattern",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "pattern": "func (\\w+)\\(",
    "replacement": "// $1 function\nfunc $1(",
    "regex": true,
    "all_occurrences": true
  }
}
```

## Language-Aware Tools (AST-Based)

### `check_docs`
Find all types/funcs/var/consts/etc without conforming comment, files without a top comment, and subdirectories withouth a README.md file.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the source code directory to check
- `language` (required): Programming language ("go" currently supported)
- `offset`: Number of items to skip for pagination (default: 0)
- `recursive`: Check only the path (false) or check path and all its subdirectories (true) (default: true)

**Example:**
```json
{
  "tool": "check_docs",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/myproject",
    "language": "go", 
    "offset": 100,
    "recursive": true 
  }
}
```

### `find_file_part`
Find specific language constructs (functions, types, constants) by name using AST parsing.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the source code file
- `language` (required): Programming language ("go" currently supported)
- `part_type` (required): Type of construct to find ("func", "type", "const", "var")
- `part_name` (required): Name of the construct to find

**Example:**
```json
{
  "tool": "find_file_part",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "language": "go",
    "part_type": "func",
    "part_name": "main"
  }
}
```

### `replace_file_part`
Replace specific language constructs using syntax-aware parsing. Requires user approval.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the source code file
- `language` (required): Programming language ("go" currently supported)
- `part_type` (required): Type of construct to replace ("func", "type", "const", "var")
- `part_name` (required): Name of the construct to replace
- `new_content` (required): New implementation content

**Example:**
```json
{
  "tool": "replace_file_part",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "language": "go",
    "part_type": "func",
    "part_name": "processData",
    "new_content": "func processData(input string) (string, error) {\n\treturn strings.ToUpper(input), nil\n}"
  }
}
```

### `validate_files`
Validate syntax of source code files using language-specific parsers.

**Parameters:**
- `session_token` (required): Session token from start_session
- `files` (required): Array of file paths to validate
- `paths` (required): Array of file or directory paths to validate
- `language` (required): Programming language ("go" currently supported)

**Example:**
```json
{
  "tool": "validate_files",
  "parameters": {
    "session_token": "your-session-token",
    "files": ["/Users/mike/project/main.go", "/Users/mike/project/config.go"]
  }
}
```

## Analysis Tools

### `analyze_files`
Analyze file structure and provide insights about the codebase.

**Parameters:**
- `session_token` (required): Session token from start_session
- `files` (required): Array of file paths to analyze

**Example:**
```json
{
  "tool": "analyze_files",
  "parameters": {
    "session_token": "your-session-token",
    "files": ["/Users/mike/project/main.go", "/Users/mike/project/utils.go"]
  }
}
```

### `detect_current_project`
Detect the most recently active project by analyzing recent file modifications in allowed paths and their immediate subdirectories.

**Parameters:**
- `session_token` (required): Session token from start_session
- `list_recent` (optional): If true, lists the 5 most recent projects instead of detecting current (default: false)
- `max_projects` (optional): Maximum number of recent projects to track (default: 5)
- `ignore_git_requirement` (optional): If true, don't require .git directory to consider a directory a project (default: false)

**Examples:**

**Detect current project (Git repositories only):**
```json
{
  "tool": "detect_current_project",
  "parameters": {
    "session_token": "your-session-token"
  }
}
```

**List recent projects including non-Git directories:**
```json
{
  "tool": "detect_current_project",
  "parameters": {
    "session_token": "your-session-token",
    "list_recent": true,
    "ignore_git_requirement": true,
    "max_projects": 10
  }
}
```

**Response Format:**
```json
{
  "current_project": {
    "name": "my-active-project",
    "path": "/Users/mike/Projects/my-active-project",
    "last_modified": "2024-01-15T10:30:00Z",
    "relative_age": "2 hours ago"
  },
  "recent_projects": [
    {
      "name": "my-active-project", 
      "path": "/Users/mike/Projects/my-active-project",
      "last_modified": "2024-01-15T10:30:00Z",
      "relative_age": "2 hours ago"
    }
  ],
  "requires_choice": false,
  "summary": "Current project: my-active-project (last modified 2 hours ago)"
}
```

**Project Detection Logic:**
1. **Checks allowed paths themselves** - if they contain recent file modifications and meet project criteria
2. **Checks immediate subdirectories** of allowed paths - only one level deep
3. **Project criteria**:
   - By default: Must contain a `.git` directory (Git repository) AND have at least 5 files
   - With `ignore_git_requirement: true`: Any directory with at least 5 files is considered a project
4. **Activity detection**: Uses most recent file modification time (recursively within each project, but excludes hidden files/directories like `.git`)
5. **Filtering**: Skips hidden directories (starting with '.') when scanning for projects
6. **File counting**: Counts all non-hidden files recursively to determine if directory qualifies as a project
7. **Current project logic**: If one project is 24+ hours newer than others, it's identified as current
8. **User choice**: If multiple projects are modified within 24 hours, user choice is required

## Configuration and Help Tools

### `get_config`
Get information about the Scout MCP server configuration.

**Parameters:**
- `session_token` (required): Session token from start_session

**Example:**
```json
{
  "tool": "get_config",
  "parameters": {
    "session_token": "your-session-token"
  }
}
```

### `help`
Get detailed documentation for all available tools.

**Parameters:**
- `session_token` (required): Session token from start_session

**Example:**
```json
{
  "tool": "help",
  "parameters": {
    "session_token": "your-session-token"
  }
}
```

## Approval System Tools

### `request_approval`
Request user approval for risky operations with risk assessment.

**Parameters:**
- `session_token` (required): Session token from start_session
- `operation` (required): Brief description of the operation
- `files` (required): Array of files that will be affected
- `impact_summary` (required): Summary of what will change
- `risk_level` (required): Risk level ("low", "medium", "high")
- `preview_content` (optional): Code preview or diff content

**Example:**
```json
{
  "tool": "request_approval",
  "parameters": {
    "session_token": "your-session-token",
    "operation": "Update configuration file",
    "files": ["/Users/mike/project/config.json"],
    "impact_summary": "Change database connection settings",
    "risk_level": "medium"
  }
}
```

## Best Practices

### Getting Started
1. **Always start with `start_session`** to get your session token and instructions
2. **Use `read_files`** to examine multiple files efficiently before making changes
3. **Use `search_files`** to find files matching criteria

### For Editing Existing Code
- ✅ **PREFERRED:** Use granular editing tools (`update_file_lines`, `insert_file_lines`, etc.)
- ⚠️ **AVOID:** `update_file` unless you truly want to replace the entire file
- Use `read_files` first to understand the current content
- Make incremental changes rather than wholesale replacements

### For Large Changes
- Break down into multiple granular operations
- Use the approval system for write operations
- Test changes incrementally

### Common Patterns

**Adding an import to a Go file:**
```json
{
  "tool": "insert_at_pattern",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/path/to/file.go",
    "after_pattern": "package main",
    "new_content": "\nimport \"fmt\"",
    "position": "after"
  }
}
```

**Reading multiple related files:**
```json
{
  "tool": "read_files",
  "parameters": {
    "session_token": "your-session-token",
    "paths": ["./mcptools"],
    "extensions": [".go"],
    "pattern": "tool",
    "recursive": true
  }
}
```

**Updating a function implementation:**
1. Use `find_file_part` to locate the function
2. Use `replace_file_part` to replace the entire function
3. Or use `update_file_lines` to replace just the function body

**Refactoring variable names:**
```json
{
  "tool": "replace_pattern",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/path/to/file.go",
    "pattern": "oldVariableName",
    "replacement": "newVariableName",
    "all_occurrences": true
  }
}
```

## Security Notes

- All tools (except `start_session`) require a valid session token
- Session tokens expire after 24 hours
- Write operations require user approval with risk assessment
- All operations respect the configured allowed paths
- File operations validate parameters before execution

## Error Handling

Tools will return descriptive error messages for common issues:
- Invalid or expired session tokens
- Path not in allowed directories
- File not found
- Invalid line numbers
- Invalid regex patterns
- Permission errors

Always check the error field in tool responses and handle failures gracefully.

## On Incorrect Usage

When you attempt to use a tool but learn that the way you attempted to use it is not the way the MCP Server works, add an "entry" in a file named `./MCP_USABILITY_CONCERNS.md` explaining:

1. Which MCP Server you used.
2. What tool for the MCP Server you called.
3. How you called the tool.
4. What error the MCP Server responded with.
5. Why you expected it to work as you called it.
6. How would you change the tool to improve its usability for your use.

## Session Management

- Session tokens are valid for 24 hours
- Tokens are invalidated when the server restarts
- If you get "invalid session token" errors, call `start_session` again
- Each session provides fresh instructions and tool documentation
