# Scout MCP Specific Help Content

## File Reading Tools

### Best Practices for Scout MCP

#### For Editing Existing Code
- ✅ **PREFERRED:** Use granular editing tools (`update_file_lines`, `insert_file_lines`, etc.)
- ⚠️ **AVOID:** `update_file` unless you truly want to replace the entire file
- Use `read_files` first to understand the current content
- Make incremental changes rather than wholesale replacements

#### For Large Changes
- Break down into multiple granular operations
- Use the approval system for write operations
- Test changes incrementally

#### Common Patterns

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

## On Incorrect Usage

When you attempt to use a tool but learn that the way you attempted to use it is not the way the MCP Server works, add an "entry" in a file named `./MCP_USABILITY_CONCERNS.md` explaining:

1. Which MCP Server you used.
2. What tool for the MCP Server you called.
3. How you called the tool.
4. What error the MCP Server responded with.
5. Why you expected it to work as you called it.
6. How would you change the tool to improve its usability for your use.