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
- `path` (required): Full path where the file should be created
- `content` (required): Content to write to the file
- `create_dirs` (optional): Create parent directories if they don't exist

**Example:**
```json
{
  "tool": "create_file",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/new_file.go",
    "content": "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}",
    "create_dirs": true
  }
}
```

### `update_file`
**⚠️ DANGEROUS: Replaces entire file content. Use granular editing tools for safer changes.**

Completely replaces the content of an existing file. Use this ONLY when you intend to replace the entire file.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the file to update
- `content` (required): New content that will replace ALL existing content

**Example:**
```json
{
  "tool": "update_file",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/config.json",
    "content": "{\"version\": \"2.0\", \"name\": \"updated-config\"}"
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
- `path` (required): Full path to the file
- `start_line` (required): Starting line number (1-based)
- `end_line` (required): Ending line number (1-based, inclusive)
- `content` (required): New content to replace the specified line range

**Example:**
```json
{
  "tool": "update_file_lines",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "start_line": "15",
    "end_line": "18",
    "content": "\tfmt.Println(\"Updated function\")\n\treturn nil"
  }
}
```

### `insert_file_lines`
Insert content at a specific line number.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the file
- `line_number` (required): Line number where to insert (1-based)
- `content` (required): Content to insert
- `position` (optional): "before" or "after" the specified line (default: "after")

**Example:**
```json
{
  "tool": "insert_file_lines",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "line_number": "1",
    "content": "import \"fmt\"",
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
- `content` (required): Content to insert
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
    "content": "\nimport \"fmt\"",
    "position": "after"
  }
}
```

### `delete_file_lines`
Delete specific lines from a file.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the file
- `start_line` (required): Starting line number to delete (1-based)
- `end_line` (optional): Ending line number to delete (defaults to start_line for single line)

**Example:**
```json
{
  "tool": "delete_file_lines",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "start_line": "25",
    "end_line": "30"
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

### `find_file_part`
Find specific language constructs (functions, types, constants) by name using AST parsing.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the source code file
- `part_type` (required): Type of construct to find ("function", "type", "const", "var")
- `part_name` (required): Name of the construct to find
- `language` (optional): Programming language (auto-detected from file extension)

**Example:**
```json
{
  "tool": "find_file_part",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "part_type": "function",
    "part_name": "main"
  }
}
```

### `replace_file_part`
Replace specific language constructs using syntax-aware parsing. Requires user approval.

**Parameters:**
- `session_token` (required): Session token from start_session
- `path` (required): Full path to the source code file
- `part_type` (required): Type of construct to replace ("function", "type", "const", "var")
- `part_name` (required): Name of the construct to replace
- `new_content` (required): New implementation content
- `language` (optional): Programming language (auto-detected from file extension)

**Example:**
```json
{
  "tool": "replace_file_part",
  "parameters": {
    "session_token": "your-session-token",
    "path": "/Users/mike/project/main.go",
    "part_type": "function",
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
- `language` (optional): Programming language (auto-detected from file extensions)

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

### `tool_help`
Get detailed documentation for all available tools.

**Parameters:**
- `session_token` (required): Session token from start_session

**Example:**
```json
{
  "tool": "tool_help",
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

### `generate_approval_token`
Generate approval tokens after user confirmation for secure operation execution.

**Parameters:**
- `session_token` (required): Session token from start_session
- `file_actions` (required): Array of file actions that were approved
- `operations` (required): Array of operations approved ("create", "update", "delete")
- `session_id` (optional): Session identifier for this approval

**Example:**
```json
{
  "tool": "generate_approval_token",
  "parameters": {
    "session_token": "your-session-token",
    "file_actions": ["update_config.json"],
    "operations": ["update"]
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
    "content": "\nimport \"fmt\"",
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
