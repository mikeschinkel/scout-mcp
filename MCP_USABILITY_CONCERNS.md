# MCP Errors Encountered During Development

This document tracks errors encountered when calling MCP tools, to help improve the tools and understand their proper usage.

## Error #1: generate_approval_token - Invalid file_actions parameter type

**Date:** 2025-07-15

**Tool Called:** `generate_approval_token`

**How I Called It:**
```
scout-mcp:generate_approval_token
- file_actions: ["create", "create"]
- operations: ["create", "create"] 
- session_id: "session_enforcement_implementation"
```

**Error Received:**
```
MCP error -32603: error converting item 0: item a 'string', not a 'mcptools.FileAction'
error converting item 1: item a 'string', not a 'mcptools.FileAction'
```

**Why I Expected This To Work:**
I thought `file_actions` should be an array of strings representing the action types.

**How I Learned The Correct Way:**
I read `/Users/mikeschinkel/Projects/scout-mcp/mcptools/file_action.go` and found that `FileAction` is a struct with `Action`, `Path`, and `Purpose` fields.

**Attempted Correction #1:**
```
scout-mcp:generate_approval_token
- file_actions: [{"path": "/Users/mikeschinkel/Projects/scout-mcp/mcptools/sessions.go", "action": "create", "purpose": "Core session management logic and token validation"}, {"path": "/Users/mikeschinkel/Projects/scout-mcp/mcptools/start_session_tool.go", "action": "create", "purpose": "Tool that delivers comprehensive instructions and session token"}]
- operations: ["create", "create"]
- session_id: "session_enforcement_implementation"
```

**Error From Correction #1:**
```
MCP error -32603: error converting item 0: item a 'map[string]interface {}', not a 'mcptools.FileAction'
error converting item 1: item a 'map[string]interface {}', not a 'mcptools.FileAction'
```

**Resolution:**
Still investigating proper usage. Will attempt direct file creation instead.

---

## Error #2: update_file_lines - No result received

**Date:** 2025-07-15

**Tool Called:** `update_file_lines`

**How I Called It:**
```
scout-mcp:update_file_lines
- path: "/Users/mikeschinkel/Projects/scout-mcp/mcptools/start_session_tool.go"
- start_line: 6
- end_line: 6
- new_content: "\t\"strings\"\n\t\"time\""
```

**Error Received:**
```
No result received from client-side tool execution.
```

**Why I Expected This To Work:**
I was trying to add missing imports (`time` package) to the file after the existing `strings` import.

**How I Learned The Correct Way:**
This appeared to be a tool execution issue, not a parameter format problem. I switched to using `update_file` to replace the entire file content instead.

**Resolution:**
Used `update_file` to replace the entire file content with the correct imports included.

---

## Error #3: replace_pattern - Missing required pattern argument

**Date:** 2025-07-15

**Tool Called:** `replace_pattern`

**How I Called It:**
```
scout-mcp:replace_pattern
- path: "/Users/mikeschinkel/Projects/scout-mcp/mcptools/start_session_tool.go"
- old_text: "import (...)"
- new_text: "import (...)"
```

**Error Received:**
```
MCP error -32603: required argument "pattern" not found
```

**Why I Expected This To Work:**
I thought `old_text` and `new_text` were the main parameters, similar to other replace operations.

**How I Learned The Correct Way:**
The `replace_pattern` tool requires a `pattern` parameter (regex) to find what to replace, not `old_text`/`new_text` parameters.

**Resolution:**
Need to use a different approach or understand the correct parameter structure for `replace_pattern`.

---

## Error #4: update_file - No result received (second occurrence)

**Date:** 2025-07-15

**Tool Called:** `update_file`

**How I Called It:**
```
scout-mcp:update_file
- path: "/Users/mikeschinkel/Projects/scout-mcp/mcptools/start_session_tool.go"
- content: [large file content with import fix]
```

**Error Received:**
```
No result received from client-side tool execution.
```

**Why I Expected This To Work:**
I was trying to update the start_session_tool.go file to add the missing scout package import.

**How I Learned The Correct Way:**
This appears to be another tool execution issue. Multiple tools are returning "No result received" which suggests a broader system issue.

**Resolution:**
Will try breaking down the changes into smaller parts or using alternative approaches.

---

