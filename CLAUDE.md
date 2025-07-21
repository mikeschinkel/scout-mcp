# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Building and Running

#### Using Makefile (Recommended)
```bash
# Build the main binary
make build

# Run in development mode
make dev

# Initialize configuration with path
make init PATH_ARG=~/Projects

# Run server with specific path
make run PATH_ARG=~/MyProjects

# Run server with only specified path (ignore config)
make run-only PATH_ARG=/tmp/safe-dir

# Show all available commands
make help
```

#### Direct Go Commands
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
# Run all tests (unit + integration)
make test

# Run only integration tests
make test-integration

# Run only unit tests
make test-unit

# Run tests with coverage
make test-coverage

# Format code
make fmt

# Vet code for issues
make vet

# Tidy dependencies
make tidy

# Run all code quality checks
make check

# See all available commands
make help
```

## Development Infrastructure

### Makefile Build System
The project uses a comprehensive Makefile in the root directory for all development tasks:

- **Build Commands**: `make build`, `make install`, `make clean`
- **Testing**: `make test`, `make test-unit`, `make test-integration`, `make test-coverage`
- **Code Quality**: `make fmt`, `make vet`, `make tidy`, `make check`
- **Development**: `make dev`, `make run`, `make init`, `make run-only`
- **Release**: `make release VERSION=v1.0.0`

The Makefile provides colorized output, proper error handling, and comprehensive help via `make help`.

### Testing Architecture
The project has two distinct test suites:

#### Unit Tests (`./mcptools/`)
- Test individual tool implementations in isolation
- Use mock configurations and requests
- Focus on tool logic and parameter validation
- Run with: `make test-unit`

#### Integration Tests (`./test/`)
- Test complete MCP server functionality end-to-end
- Use direct in-process server testing (not external processes)
- Test session management, tool interactions, and error handling
- Run with: `make test-integration`
- **Recent Improvement**: Refactored from `exec.Command()` external process testing to direct in-process MCP server testing for better performance and reliability

### Continuous Integration
- All tests must pass before merging changes
- Code quality checks (`fmt`, `vet`, `tidy`) are enforced
- Integration tests provide comprehensive coverage of MCP server functionality

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

#### Core Application
- **Makefile**: Root build system with all development commands
- **run_main.go**: Entry point and argument parsing (parseArgs function)
- **mcp.go**: Core MCP server setup and tool registration
- **config.go**: Configuration loading and path validation
- **const.go**: Application constants and default values
- **logger.go**: Logging utilities and configuration

#### Framework Layer  
- **mcputil/mcp_server.go**: Framework-level session enforcement and MCP server wrapper
- **mcputil/tools.go**: Tool interfaces and result types
- **mcputil/registered_tools.go**: Tool registration system

#### Tool Implementation
- **mcptools/tool_base.go**: Base class for all tools with session validation and approval system
- **mcptools/sessions.go**: Session token management and validation
- **mcptools/start_session_tool.go**: Comprehensive onboarding tool with instructions
- **mcptools/properties.go**: Shared property definitions for consistent parameter naming
- **mcptools/types.go**: Common types and enums for tool system

#### Testing Infrastructure
- **test/direct_server_test.go**: Direct in-process MCP server testing infrastructure
- **testutil/**: Mock configurations, requests, and test utilities

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

#### Integration Test Refactoring (Latest)
- **Modernized Test Infrastructure**: Replaced `exec.Command()` external process testing with direct in-process MCP server testing
- **Enhanced Error Handling**: Implemented reflection-based `ToolResultError` detection for comprehensive error testing
- **Framework-Level Validation**: Tests now use the same validation flow as the real MCP server via `EnsurePreconditions()`
- **Bug Fixes**: Resolved tool implementation issues discovered during testing:
  - `insert_file_lines` missing `line_number` parameter extraction
  - `replace_pattern` property definitions mismatched with actual parameters
  - `insert_at_pattern` missing property declarations
  - Parameter naming inconsistencies across tools

#### Build System Improvements
- **Professional Makefile**: Replaced ad-hoc `./test/run_tests.sh` with comprehensive root Makefile
- **Comprehensive Commands**: Build, test, code quality, development, and release commands
- **Enhanced Developer Experience**: Colorized output, proper error handling, and help system

#### Core Architecture
- **Framework-Level Session Enforcement**: All session validation moved to MCP server layer
- **read_file → read_files**: New tool can read multiple files/directories efficiently  
- **Comprehensive Instruction Delivery**: `start_session` provides complete guidance
- **Automatic Session Validation**: Tools no longer need manual session checks

## Development Workflow

### Getting Started
1. **Clone and Build**: `git clone` → `make build`
2. **Run Tests**: `make test` to ensure everything works
3. **Development**: Use `make dev` for running server during development
4. **Code Quality**: Always run `make check` before committing

### Making Changes
1. **Follow "Clear Path" Style**: Single return points, minimal nesting, `goto end` pattern
2. **Maintain Security Model**: All tools must use session validation and path restrictions
3. **Update Tests**: Add/update both unit and integration tests for changes
4. **Run Full Test Suite**: `make test` must pass before submitting changes
5. **Documentation**: Update relevant documentation (README, tool help, etc.)

### Testing Strategy
- **Unit Tests First**: Test individual tool logic in isolation
- **Integration Tests**: Verify end-to-end functionality with MCP server
- **Parameter Validation**: Ensure all tool parameters are properly validated
- **Error Cases**: Test error handling and edge cases thoroughly

### Tool Development
When adding new MCP tools:
1. **Inherit from `toolBase`**: Provides session validation and common utilities
2. **Define Properties**: Use property definitions from `mcptools/properties.go`
3. **Implement Handle Method**: Follow "Clear Path" style with single return
4. **Add Session Validation**: Framework handles this automatically via `EnsurePreconditions()`
5. **Register Tool**: Add to init() function with proper registration
6. **Add Tests**: Both unit tests (mcptools/) and integration tests (test/)
7. **Update Documentation**: Add to mcptools/README.md and tool_help

### Parameter Naming Conventions
- Use `filepath` for single file paths (not `path`)
- Use `new_content` for content parameters (not `content`) 
- Use `session_token` for all session tokens
- Be consistent across similar tools

When making changes, ensure they maintain the existing security model, follow the "Clear Path" coding style, and integrate properly with the session management system.
