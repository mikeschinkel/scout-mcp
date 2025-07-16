# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building and Running
```bash
# Build the main binary
go build -o bin/scout-mcp ./cmd/main.go

# Run directly during development
go run ./cmd/main.go

# Initialize configuration with default path
go run ./cmd/main.go init ~/Projects

# Run server with additional paths
go run ./cmd/main.go ~/MyProjects

# Run server with only specified path (ignore config)
go run ./cmd/main.go --only /tmp/safe-dir
```

### Testing and Linting
```bash
# Run comprehensive test suite
./test/run_tests.sh

# Run individual tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Format code
go fmt ./...

# Vet code for issues
go vet ./...

# Tidy dependencies
go mod tidy
```

## Code Architecture

### High-Level Structure
Scout-MCP is a secure Model Context Protocol (MCP) server that provides Claude with file access through stdio transport. The architecture follows a "Clear Path" coding style with single return points and minimal nesting.

### Key Components

#### Core Packages
- **scout/**: Main package containing MCP server implementation and configuration
- **mcputil/**: MCP server utilities and abstractions over mark3labs/mcp-go
- **mcptools/**: Individual tool implementations with session management and approval system
- **langutil/**: Language-specific parsing utilities for AST-based operations

#### MCP Server Architecture
The server uses `github.com/mark3labs/mcp-go` for MCP protocol implementation:
- **mcp.go**: Main MCP server setup and tool registration
- **mcputil/mcp_server.go**: Server interface and wrapper around mark3labs library with framework-level session enforcement
- **mcptools/**: Tool implementations following session-based security model

#### Security Model
- **Session Management**: All tools require session tokens (except `start_session`)
- **Path Whitelisting**: Only explicitly allowed directories are accessible (config.go)
- **User Approval System**: Write operations require user confirmation via stdio prompts
- **Risk Assessment**: Tools are classified by risk level (mcptools/risk_level.go)
- **Framework-Level Enforcement**: Session validation happens at MCP server layer

#### Tool Implementation Pattern
All tools follow a consistent pattern in mcptools/:
- Inherit from `toolBase` for common functionality
- Automatic session validation through framework (via `EnsurePreconditions()`)
- Implement approval workflow for risky operations
- Use `FileAction` enum for operation classification
- Support both text-based and AST-based editing operations
- Language-aware tools use `langutil/` for syntax-aware parsing
- Follow "Clear Path" style with single return points

### Session-Based Architecture
- **Framework-Level Session Enforcement**: `mcputil/mcp_server.go` calls `EnsurePreconditions()` before `Handle()`
- **Session Management**: `mcptools/sessions.go` handles token creation, validation, and expiration
- **Instruction Delivery**: `start_session_tool.go` provides comprehensive instructions and coding guidelines
- **Token Expiration**: 24-hour sessions with server restart invalidation

### Configuration System
- **Config File**: `~/.config/scout-mcp/scout-mcp.json`
- **Path Management**: Combines config file paths with command-line arguments
- **Flexible Arguments**: Support for `--only` mode to override config

### Key Files to Understand
- **run_main.go**: Entry point and argument parsing (parseArgs function)
- **mcp.go**: Core MCP server setup and tool registration
- **config.go**: Configuration loading and path validation
- **const.go**: Application constants and default values
- **logger.go**: Logging utilities and configuration
- **mcptools/tool_base.go**: Base class for all tools with session validation and approval system
- **mcptools/sessions.go**: Session token management and validation
- **mcptools/start_session_tool.go**: Comprehensive onboarding tool with instructions
- **mcptools/types.go**: Common types and enums for tool system
- **mcputil/mcp_server.go**: Framework-level session enforcement

### Current Tool Implementation (19 tools)

#### Session Management
- **start_session**: Creates session tokens and delivers comprehensive instructions

#### Enhanced File Reading
- **read_files**: Efficiently read multiple files/directories with filtering (replaces read_file)
- **search_files**: Search for files with pattern matching and filtering

#### File Management (with approval)
- **create_file**: Create new files
- **update_file**: Replace entire file content (dangerous - granular tools preferred)
- **delete_files**: Delete files or directories

#### Granular Editing (with approval)
- **update_file_lines**: Update specific line ranges
- **delete_file_lines**: Delete specific lines
- **insert_file_lines**: Insert content at line numbers
- **insert_at_pattern**: Insert before/after patterns
- **replace_pattern**: Find/replace with regex support

#### Language-Aware (AST-based)
- **find_file_part**: Find language constructs (functions, types, etc.)
- **replace_file_part**: Replace language constructs (with approval)
- **validate_files**: Syntax validation

#### Analysis & System
- **analyze_files**: File analysis and insights
- **get_config**: Server configuration
- **tool_help**: Tool documentation

#### Approval System
- **request_approval**: User approval for risky operations
- **generate_approval_token**: Token generation after confirmation

### Coding Conventions
The codebase follows "Clear Path" style:
- Use `goto end` pattern instead of early returns
- Single return point per function with named return variables
- All variables declared before first `goto`
- Minimal nesting through helper functions
- No variable shadowing

### Session Flow
1. **User calls `start_session`**: Gets token + comprehensive instructions + tool docs + server config
2. **Framework validates sessions**: `mcputil/mcp_server.go` calls `EnsurePreconditions()` before each tool
3. **Tools focus on domain logic**: No session validation code in individual tools
4. **Approval for write operations**: Risk-based approval system for file modifications

### Claude Desktop Integration
The server communicates with Claude Desktop via stdio transport:
- No network configuration required
- Runs as subprocess of Claude Desktop
- Configuration via Claude Desktop's `claude_desktop_config.json`
- Session-based workflow: Must call `start_session` first in each conversation

### Recent Major Changes
- **Framework-Level Session Enforcement**: All session validation moved to MCP server layer
- **read_file → read_files**: New tool can read multiple files/directories efficiently
- **Comprehensive Instruction Delivery**: `start_session` provides complete guidance
- **Automatic Session Validation**: Tools no longer need manual session checks

When making changes, ensure they maintain the existing security model, follow the "Clear Path" coding style, and integrate properly with the session management system.
