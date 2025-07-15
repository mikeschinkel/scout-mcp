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
- **mcptools/**: Individual tool implementations with approval system
- **langutil/**: Language-specific parsing utilities for AST-based operations

#### MCP Server Architecture
The server uses `github.com/mark3labs/mcp-go` for MCP protocol implementation:
- **mcp.go**: Main MCP server setup and tool registration
- **mcputil/mcp_server.go**: Server interface and wrapper around mark3labs library
- **mcptools/**: Tool implementations following approval-based security model

#### Security Model
- **Path Whitelisting**: Only explicitly allowed directories are accessible (config.go)
- **User Approval System**: Write operations require user confirmation via stdio prompts
- **Risk Assessment**: Tools are classified by risk level (mcptools/risk_level.go)
- **Operation Classification**: Read vs write operations are automatically determined

#### Tool Implementation Pattern
All tools follow a consistent pattern in mcptools/:
- Inherit from `ToolBase` for common functionality
- Implement approval workflow for risky operations
- Use `FileAction` enum for operation classification
- Support both text-based and AST-based editing operations
- Language-aware tools use `langutil/` for syntax-aware parsing
- Follow "Clear Path" style with single return points

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
- **mcptools/tool_base.go**: Base class for all tools with approval system
- **mcptools/types.go**: Common types and enums for tool system

### Coding Conventions
The codebase follows "Clear Path" style:
- Use `goto end` pattern instead of early returns
- Single return point per function with named return variables
- All variables declared before first `goto`
- Minimal nesting through helper functions
- No variable shadowing

### Claude Desktop Integration
The server communicates with Claude Desktop via stdio transport:
- No network configuration required
- Runs as subprocess of Claude Desktop
- Configuration via Claude Desktop's `claude_desktop_config.json`
- Tools available: 18 comprehensive tools including:
  - **Basic file operations**: read_file, create_file, update_file, delete_file, search_files
  - **Advanced editing**: update_file_lines, delete_file_lines, insert_file_lines, insert_at_pattern, replace_pattern
  - **Language-aware editing**: find_file_part, replace_file_part (AST-based)
  - **Analysis and validation**: analyze_files, validate_files
  - **System tools**: get_config, tool_help
  - **Approval system**: request_approval, generate_approval_token

When making changes, ensure they maintain the existing security model and follow the "Clear Path" coding style established throughout the codebase.