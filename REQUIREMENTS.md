# Scout-MCP Enhancement: MCP Server Requirements and Design

## Project Overview

Convert Scout-MCP to a proper transport-based MCP implementation with initial stdio transport, enabling Claude Desktop integration for secure file operations, search capabilities, and command execution.

## Goals and Objectives

### Primary Goals
- **Transport-based MCP Implementation**: Replace HTTP with proper MCP using `github.com/mark3labs/mcp-go`
- **Claude Desktop Integration**: Enable direct connection with Claude Desktop via stdio
- **Secure File Operations**: Build on existing whitelist security model
- **Personal Productivity**: Tool for immediate personal use

## Completion Checklist

### Core MCP Implementation
- [ ] Remove broken HTTP server code
- [ ] Integrate `github.com/mark3labs/mcp-go` package
- [ ] Implement stdio transport MCP server
- [ ] Convert existing tools to MCP format (list_files, read_file)
- [ ] Basic MCP server responds to tool calls

### File Operations
- [ ] Enhanced list_files with recursive/extension filtering
- [ ] Working read_file tool
- [ ] New create_file tool with user approval
- [ ] New update_file tool with user approval  
- [ ] New delete_file tool with user approval

### Search and Commands
- [ ] New search_content tool with regex support
- [ ] Basic git command execution (status, log, diff)
- [ ] Determine read vs write operations (hardcoded classification)

### Integration and Polish
- [ ] Claude Desktop successfully connects via stdio
- [ ] User approval system working with conversational prompts
- [ ] All operations respect existing whitelist security
- [ ] Error handling and logging
- [ ] Documentation for setup and usage
- [ ] GitHub repository ready for publication

## Architecture

### Package Structure (Minimal)
```
scout-mcp/                  # Keep existing structure
├── cmd/                    # Keep existing
│   ├── main.go            # Convert to stdio MCP mode
│   ├── go.mod             
│   └── go.work            
└── scout/                  # Main package (existing files + new MCP code)
    ├── config.go          # Existing - enhance with MCP settings
    ├── mcp_server.go      # REPLACE - convert to MCP implementation  
    ├── types.go           # Existing - add MCP types
    ├── util.go            # Existing
    ├── logger.go          # Existing
    ├── const.go           # Existing
    ├── approval.go        # NEW - user approval system
    └── tools.go           # NEW - MCP tool implementations
```

### Command Line Interface (Simplified)
```bash
# Single mode - stdio MCP server
scout-mcp                        # Run as MCP server (stdio) with config paths
scout-mcp /path/to/dir           # MCP server with additional path
scout-mcp --only /path           # MCP server with only specified path
scout-mcp init                   # Create config (keep existing)
scout-mcp init /path             # Create config with initial path (keep existing)
```

### Configuration Enhancement

Extend existing config with MCP-specific settings:

```json
{
  "whitelisted_paths": [
    "/home/user/projects"
  ],
  "mcp": {
    "server_name": "scout-mcp",
    "server_version": "1.0.0",
    "require_approval_for": [
      "file_create",
      "file_update", 
      "file_delete",
      "destructive_git_commands"
    ],
    "approval_timeout": "60s",
    "max_file_size": "10MB",
    "allowed_git_commands": ["status", "log", "diff", "show", "branch"]
  }
}
```

## Tool Definitions

### File Operations

#### list_files (enhanced from search_files)
```json
{
  "name": "list_files",
  "description": "List files and directories in whitelisted paths",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {"type": "string", "description": "Directory path to list"},
      "recursive": {"type": "boolean", "description": "Recursive listing"},
      "extensions": {"type": "array", "items": {"type": "string"}, "description": "Filter by extensions"},
      "pattern": {"type": "string", "description": "Name pattern to match"}
    },
    "required": ["path"]
  }
}
```

#### create_file (NEW - requires approval)
```json
{
  "name": "create_file", 
  "description": "Create a new file (requires user approval)",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {"type": "string", "description": "File path to create"},
      "content": {"type": "string", "description": "File content"},
      "create_dirs": {"type": "boolean", "description": "Create parent directories"}
    },
    "required": ["path", "content"]
  }
}
```

#### update_file (NEW - requires approval)
```json
{
  "name": "update_file",
  "description": "Update existing file (requires user approval)", 
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {"type": "string", "description": "File path to update"},
      "content": {"type": "string", "description": "New content"},
      "mode": {"type": "string", "enum": ["replace", "append", "prepend"], "description": "Update mode"}
    },
    "required": ["path", "content"]
  }
}
```

#### delete_file (NEW - requires approval)
```json
{
  "name": "delete_file",
  "description": "Delete file or directory (requires user approval)",
  "inputSchema": {
    "type": "object", 
    "properties": {
      "path": {"type": "string", "description": "Path to delete"},
      "recursive": {"type": "boolean", "description": "Delete directory recursively"}
    },
    "required": ["path"]
  }
}
```

### Search Tools

#### search_content (NEW)
```json
{
  "name": "search_content",
  "description": "Search file contents using regex patterns",
  "inputSchema": {
    "type": "object",
    "properties": {
      "directory": {"type": "string", "description": "Root directory to search"},
      "pattern": {"type": "string", "description": "Regex pattern to match"},
      "file_extensions": {"type": "array", "items": {"type": "string"}, "description": "File types to search"},
      "case_sensitive": {"type": "boolean", "description": "Case sensitive matching"},
      "context_lines": {"type": "integer", "description": "Lines of context around matches", "default": 2},
      "max_results": {"type": "integer", "description": "Maximum results", "default": 100}
    },
    "required": ["directory", "pattern"]
  }
}
```

### Command Execution

#### execute_command (NEW)
```json
{
  "name": "execute_command", 
  "description": "Execute whitelisted commands in whitelisted directories",
  "inputSchema": {
    "type": "object",
    "properties": {
      "command": {"type": "string", "description": "Command to execute (git, jq, find, etc.)"},
      "args": {"type": "array", "items": {"type": "string"}, "description": "Command arguments"},
      "working_directory": {"type": "string", "description": "Working directory for command"}
    },
    "required": ["command", "working_directory"]
  }
}
```

## User Approval System

### Conversational Approval Workflow
User approval works through natural conversation prompts that explain the intent:

**Example Interaction:**
```
Claude: "I need to create 5 Go files in /project/src/ to implement the user authentication system. Is that okay?"
User: "yes" or "go ahead" or "sure"
```

### Approval Interface
```go
type ApprovalRequest struct {
    ID          string    `json:"id"`
    Operation   string    `json:"operation"`     // "create_files", "update_file", etc.
    Description string    `json:"description"`   // Human-readable intent
    Files       []string  `json:"files"`         // Affected file paths
    Risk        string    `json:"risk"`          // "low", "medium", "high"
}

func (am *ApprovalManager) RequestApproval(req *ApprovalRequest) bool {
    fmt.Printf("APPROVAL REQUIRED: %s\n", req.Description)
    fmt.Printf("Files: %v\n", req.Files)
    fmt.Print("Approve? (y/N): ")
    
    input, _ := bufio.NewReader(os.Stdin).ReadString('\n')
    return strings.ToLower(strings.TrimSpace(input)) == "y"
}
```

### Read vs Write Operation Classification

**Rule-based Classification:**
```go
type CommandClassification struct {
    Command         string   `json:"command"`
    ReadOperations  []string `json:"read_operations"`
    WriteOperations []string `json:"write_operations"`
    RequiresApproval bool    `json:"requires_approval"` // Default for unlisted operations
}

// Configuration-driven classification
var CommandClassifications = map[string]CommandClassification{
    "git": {
        Command: "git",
        ReadOperations: []string{"status", "log", "diff", "show", "branch", "ls-files"},
        WriteOperations: []string{"add", "commit", "push", "pull", "reset", "checkout"},
        RequiresApproval: true, // Default for unknown git commands
    },
    "jq": {
        Command: "jq",
        ReadOperations: []string{}, // All jq operations are read-only by nature
        WriteOperations: []string{},
        RequiresApproval: false, // jq is inherently safe
    },
    "find": {
        Command: "find",
        ReadOperations: []string{}, // All find operations are read-only
        WriteOperations: []string{},
        RequiresApproval: false,
    },
    "grep": {
        Command: "grep",
        ReadOperations: []string{}, // All grep operations are read-only
        WriteOperations: []string{},
        RequiresApproval: false,
    },
}

func (s *MCPServer) isWriteOperation(command string, operation string) bool {
    classification, exists := CommandClassifications[command]
    if !exists {
        return true // Unknown commands require approval by default
    }
    
    // Check if explicitly listed as write operation
    for _, writeOp := range classification.WriteOperations {
        if operation == writeOp {
            return true
        }
    }
    
    // Check if explicitly listed as read operation
    for _, readOp := range classification.ReadOperations {
        if operation == readOp {
            return false
        }
    }
    
    // Unknown operation for known command - use default
    return classification.RequiresApproval
}
```

**Configuration Enhancement:**
```json
{
  "whitelisted_paths": ["/home/user/projects"],
  "mcp": {
    "server_name": "scout-mcp",
    "server_version": "1.0.0",
    "allowed_commands": ["git", "jq", "find", "grep", "ls"],
    "command_classifications": {
      "git": {
        "read_operations": ["status", "log", "diff", "show", "branch", "ls-files"],
        "write_operations": ["add", "commit", "push", "pull", "reset", "checkout"],
        "requires_approval": true
      },
      "jq": {
        "requires_approval": false
      },
      "find": {
        "requires_approval": false  
      }
    }
  }
}
```

## Claude Desktop Integration

### Connection Method
- **stdio Transport**: Claude Desktop connects via command execution
- **Server Configuration**: Configure Claude Desktop to run scout-mcp as subprocess

### Claude Desktop Configuration
```json
{
  "mcpServers": {
    "scout-mcp": {
      "command": "/path/to/scout-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

## Security Model

### Enhanced Security (Building on Existing)
1. **Path Validation**: Use existing whitelist system
2. **User Approval**: All write operations require confirmation
3. **Command Restrictions**: Only whitelisted commands allowed
4. **Resource Limits**: File size limits
5. **Operation Classification**: Configuration-driven read/write classification
6. **Helpful Error Messages**: Exact instructions for whitelisting blocked operations

**Helpful Error Messages:**
When operations are blocked, provide exact instructions for whitelisting:

```go
type WhitelistError struct {
    Type        string `json:"type"`        // "path", "command", "operation"
    Blocked     string `json:"blocked"`     // What was blocked
    ConfigPath  string `json:"config_path"` // Path to config file
    Instructions string `json:"instructions"` // Exact steps to whitelist
}

func (s *MCPServer) createWhitelistError(errorType, blocked string) error {
    configPath, _ := getConfigPath()
    
    var instructions string
    switch errorType {
    case "path":
        instructions = fmt.Sprintf(`To whitelist this path, add it to your config:

1. Edit: %s
2. Add "%s" to the "whitelisted_paths" array
3. Restart scout-mcp

Example:
{
  "whitelisted_paths": [
    "/existing/path",
    "%s"
  ]
}`, configPath, blocked, blocked)

    case "command":
        instructions = fmt.Sprintf(`To whitelist this command, add it to your config:

1. Edit: %s  
2. Add "%s" to the "allowed_commands" array
3. Restart scout-mcp

Example:
{
  "mcp": {
    "allowed_commands": ["git", "jq", "%s"]
  }
}`, configPath, blocked, blocked)

    case "operation":
        parts := strings.Split(blocked, " ")
        command := parts[0]
        operation := parts[1]
        instructions = fmt.Sprintf(`To whitelist this operation, update your config:

1. Edit: %s
2. Add "%s" to the read_operations for "%s" command
3. Restart scout-mcp

Example:
{
  "mcp": {
    "command_classifications": {
      "%s": {
        "read_operations": ["status", "log", "%s"],
        "requires_approval": false
      }
    }
  }
}`, configPath, operation, command, command, operation)
    }

    return fmt.Errorf("Access denied: %s not whitelisted.\n\n%s", blocked, instructions)
}

// Usage in tool handlers:
func (s *MCPServer) executeCommand(command string, args []string, workingDir string) error {
    // Check if command is whitelisted
    if !s.isCommandAllowed(command) {
        return s.createWhitelistError("command", command)
    }
    
    // Check if path is whitelisted  
    allowed, err := s.isPathAllowed(workingDir)
    if err != nil || !allowed {
        return s.createWhitelistError("path", workingDir)
    }
    
    // Check if operation requires approval
    operation := ""
    if len(args) > 0 {
        operation = args[0]
    }
    
    if s.isWriteOperation(command, operation) {
        // Handle approval...
    }
    
    // Execute command...
}
```

**Claude-Friendly Error Format:**
```go
// When Claude gets these errors, they're actionable:
func (s *MCPServer) handleToolCall(toolName string, params map[string]any) (result any, err error) {
    result, err = s.callTool(toolName, params)
    if err != nil {
        // Format error to be helpful to Claude
        return nil, fmt.Errorf(`I encountered a restriction: %s

You can fix this by following the instructions above, then say "try again" and I'll retry the operation.`, err.Error())
    }
    return result, nil
}
```

## Implementation Notes

### MCP Server Implementation
```go
// main.go - simplified stdio mode
func main() {
    // Parse existing args but run MCP server instead of HTTP
    additionalPaths, opts, isInit, args, err := parseArgs()
    if err != nil {
        logger.Error("Error parsing arguments", "error", err)
        os.Exit(1)
    }

    if isInit {
        err = createDefaultConfig(args)
        // handle init as before
        return
    }

    // Always run MCP server now
    err = runMCPServer(additionalPaths, opts)
    if err != nil {
        logger.Error("MCP server failed", "error", err)
        os.Exit(1)
    }
}

func runMCPServer(additionalPaths []string, opts MCPServerOpts) error {
    // Create scout MCP server using existing config logic
    server, err := NewMCPServer(additionalPaths, opts)
    if err != nil {
        return err
    }
    
    // Start stdio MCP transport instead of HTTP
    return server.StartMCP()
}
```

### Tool Handler Example
```go
// scout/tools.go
func (s *MCPServer) createFileHandler(ctx context.Context, args map[string]any) (any, error) {
    path := args["path"].(string)
    content := args["content"].(string)
    
    // Check whitelist
    allowed, err := s.isPathAllowed(path)
    if err != nil || !allowed {
        return nil, fmt.Errorf("access denied: %s", path)
    }
    
    // Request approval
    req := &ApprovalRequest{
        Operation:   "create_file",
        Description: fmt.Sprintf("Create file: %s", path),
        Files:       []string{path},
        Risk:        "medium",
    }
    
    if !s.approvalManager.RequestApproval(req) {
        return nil, fmt.Errorf("operation denied by user")
    }
    
    // Create file
    err = os.WriteFile(path, []byte(content), 0644)
    if err != nil {
        return nil, err
    }
    
    return map[string]any{"success": true, "path": path}, nil
}
```

## Future Enhancements

### Medium Term
- **Background Service**: macOS service installation
- **SSE Transport**: Network-based operation with proper authentication
- **Cobra CLI**: Professional command structure
- **Enhanced Tools**: More git operations, project analysis

This focused design converts your existing HTTP-based implementation to a proper MCP server using stdio transport, building on your working configuration and security systems.