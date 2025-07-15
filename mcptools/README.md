# Scout MCP Tools Documentation

This document provides comprehensive documentation for all available MCP tools in the Scout MCP server.

## File Management Tools

### `read_file`
Reads the contents of a file from an allowed directory.

**Parameters:**
- `path` (required): Full path to the file to read

**Example:**
```json
{
  "tool": "read_file",
  "parameters": {
    "path": "/Users/mike/project/src/main.go"
  }
}
```

### `create_file`
Creates a new file with specified content. Requires user approval.

**Parameters:**
- `path` (required): Full path where the file should be created
- `content` (required): Content to write to the file
- `create_dirs` (optional): Create parent directories if they don't exist

**Example:**
```json
{
  "tool": "create_file",
  "parameters": {
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
- `path` (required): Full path to the file to update
- `content` (required): New content that will replace ALL existing content

**Example:**
```json
{
  "tool": "update_file",
  "parameters": {
    "path": "/Users/mike/project/config.json",
    "content": "{\"version\": \"2.0\", \"name\": \"updated-config\"}"
  }
}
```

**⚠️ WARNING:** This tool replaces the ENTIRE file. For code changes, use granular editing tools instead:
- `update_file_lines` - Update specific line ranges
- `insert_file_lines` - Insert content at specific lines
- `replace_pattern` - Find and replace patterns

### `delete_file`
Deletes a file or directory. Requires user approval.

**Parameters:**
- `path` (required): Full path to the file or directory to delete
- `recursive` (optional): Delete directory recursively if it's a directory

**Example:**
```json
{
  "tool": "delete_file",
  "parameters": {
    "path": "/Users/mike/project/old_file.txt"
  }
}
```

## File Search Tools

### `search_files`
Search for files and directories with various filtering options.

**Parameters:**
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
    "path": "/Users/mike/project",
    "recursive": true,
    "extensions": [".go"],
    "pattern": "test"
  }
}
```

## Granular File Editing Tools

**🎯 RECOMMENDED: Use these tools for precise code editing instead of `update_file`**

### `update_file_lines`
Update specific lines in a file by line number range. Much safer than `update_file`.

**Parameters:**
- `path` (required): Full path to the file
- `start_line` (required): Starting line number (1-based)
- `end_line` (required): Ending line number (1-based, inclusive)
- `content` (required): New content to replace the specified line range

**Example:**
```json
{
  "tool": "update_file_lines",
  "parameters": {
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
- `path` (required): Full path to the file
- `line_number` (required): Line number where to insert (1-based)
- `content` (required): Content to insert
- `position` (optional): "before" or "after" the specified line (default: "after")

**Example:**
```json
{
  "tool": "insert_file_lines",
  "parameters": {
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
- `path` (required): Full path to the file
- `start_line` (required): Starting line number to delete (1-based)
- `end_line` (optional): Ending line number to delete (defaults to start_line for single line)

**Example:**
```json
{
  "tool": "delete_file_lines",
  "parameters": {
    "path": "/Users/mike/project/main.go",
    "start_line": "25",
    "end_line": "30"
  }
}
```

### `replace_pattern`
Find and replace text patterns with support for regex.

**Parameters:**
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
    "path": "/Users/mike/project/main.go",
    "pattern": "func (\\w+)\\(",
    "replacement": "// $1 function\nfunc $1(",
    "regex": true,
    "all_occurrences": true
  }
}
```

## Configuration Tools

### `get_config`
Get information about the Scout MCP server configuration.

**Parameters:**
- `show_relative` (optional): Show paths relative to home directory (default: false)

**Example:**
```json
{
  "tool": "get_config",
  "parameters": {
    "show_relative": true
  }
}
```

## Best Practices

### When to Use Each Tool

**For Reading Files:**
- Use `read_file` to examine file contents before making changes
- Use `search_files` to find files matching criteria

**For Creating New Content:**
- Use `create_file` for entirely new files

**For Editing Existing Code:**
- ✅ **PREFERRED:** Use granular editing tools (`update_file_lines`, `insert_file_lines`, etc.)
- ⚠️ **AVOID:** `update_file` unless you truly want to replace the entire file

**For Large Changes:**
- Break down into multiple granular operations
- Use `read_file` first to understand the current content
- Make incremental changes rather than wholesale replacements

### Common Patterns

**Adding an import to a Go file:**
```json
{
  "tool": "insert_at_pattern",
  "parameters": {
    "path": "/path/to/file.go",
    "after_pattern": "package main",
    "content": "\nimport \"fmt\"",
    "position": "after"
  }
}
```

**Updating a function implementation:**
1. Use `read_file` to see the current function
2. Use `update_file_lines` to replace just the function lines
3. Or use `replace_pattern` to replace the entire function

**Refactoring variable names:**
```json
{
  "tool": "replace_pattern",
  "parameters": {
    "path": "/path/to/file.go",
    "pattern": "oldVariableName",
    "replacement": "newVariableName",
    "all_occurrences": true
  }
}
```

## Safety Notes

- All write operations (create, update, delete) require user approval
- All operations respect the configured allowed paths
- File operations validate line numbers and patterns before execution
- Large changes may require explicit confirmation

## Error Handling

Tools will return descriptive error messages for common issues:
- Path not in allowed directories
- File not found
- Invalid line numbers
- Invalid regex patterns
- Permission errors

Always check the error field in tool responses and handle failures gracefully.
