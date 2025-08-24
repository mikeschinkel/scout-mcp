# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## INCLUDE ../CLAUDE-*.md

**IMPORTANT**: Before working on this project, you MUST:

1. **Check the parent directory** for files named `CLAUDE-*.md` (e.g., `CLAUDE-golang.md`, `CLAUDE-mcp-server-usage.md`)
2. **Read all applicable files** based on your current tasks:
  - For Go code: Read `CLAUDE-golang.md` for coding style requirements
  - For Go code: Read `CLAUDE-golang-package-design-guidelines` for Go package design guidelines,
  - For project exploration: Read `CLAUDE-mcp-server-usage.md` for file system tool usage
  - For other languages/topics: Read the relevant `CLAUDE-*.md` files
3. **Follow those guidelines** in addition to the project-specific guidance below

These parent-level files contain critical coding standards and tool usage patterns that apply across multiple projects.

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

#### Direct Binary Commands
```bash
# Build the main binary
make build

# Show comprehensive help
./bin/scout help

# Initialize configuration with allowed path
./bin/scout init ~/Projects

# Start MCP server with additional paths
./bin/scout mcp ~/MyProjects

# Start server with only specified path (ignore config)  
./bin/scout mcp --only /tmp/safe-dir

# Session management
./bin/scout session new
./bin/scout session list
./bin/scout session clear all

# Tool management
./bin/scout tool list
./bin/scout tool run read_files
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
- **scoutcmds/**: CLI command implementations with I/O separation and help system
- **cliutil/**: CLI framework with command routing, help system, and output abstraction
- **testutil/**: Testing utilities including `TestOutputWriter` for CLI output capture
- **langutil/**: Language-specific parsing utilities for AST-based operations
- **jsontest/**: JSON testing framework with sophisticated path-based assertions and pipe functions
- **jsontest/pipefuncs/**: Modular pipe function implementations (exists, notNull, notEmpty, len, json)

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

### JSON Testing Architecture
The project includes a sophisticated JSON testing framework for validating JSON-RPC responses and complex JSON structures:

#### Core Framework (`jsontest/`)
- **JSON Path Testing**: Test JSON responses using path-based assertions (e.g., `"result.content.0.type": "text"`)
- **Pipe Functions**: Transform and validate values using pipe syntax (e.g., `"error.message|notEmpty()": true`)
- **Collection Testing**: Support for ordered and unordered array comparisons
- **Type Coercion**: Smart type conversion for accurate comparisons

#### Modular Pipe Functions (`jsontest/pipefuncs/`)
- **exists()**: Check if a path exists in JSON
- **notNull()**: Verify value is not null
- **notEmpty()**: Validate non-empty values (arrays, objects, strings, numbers, booleans)
- **len()**: Get length of arrays, objects, or strings  
- **json()**: Parse JSON strings and access nested properties
- **Extensible**: Easy to add new pipe functions with `Handle(ctx, *PipeState)` interface

#### Testing Integration
- **Integration Tests**: All `jsonrpc_*_test.go` files use `RunJSONRPCTest()` framework
- **Protocol Testing**: JSON-RPC error handling with proper error code validation (-32602, etc.)
- **Consistent Patterns**: Standardized testing approach across all MCP tools

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
- **test/jsonrpc_test.go**: JSON-RPC testing framework using jsontest for response validation
- **test/jsonrpc_protocol_test.go**: Protocol-level error testing with JSON-RPC error codes
- **jsontest/**: JSON testing framework with path-based assertions and pipe functions
- **jsontest/pipefuncs/**: Modular pipe function implementations for value transformation
- **testutil/**: Mock configurations, requests, and test utilities
  - **testutil/output.go**: `TestOutputWriter` for capturing CLI output in tests

### Current Tool Implementation (20 tools)

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
- **detect_current_project**: Detect most recently active project by modification time

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

#### CLI Architecture Refactoring (Latest)
- **I/O Separation**: Complete separation of MCP protocol I/O (`MCPReader`/`MCPWriter`) from CLI user output (`CLIWriter`)
- **RunArgs Refactoring**: Updated `RunArgs` structure with proper field naming and dependency injection
- **Error Handling Improvements**: Unknown commands now display helpful errors directing users to `scout help`
- **Help Command Implementation**: Added `scout help` command with consistent, deterministic output
- **TestOutputWriter Infrastructure**: New testutil package with `TestOutputWriter` for capturing CLI output in tests
- **Framework Testability**: CLI commands now fully testable through dependency injection
- **Consistent Help System**: Both `scout` and `scout help` show identical output using shared `ShowMainHelp()` function
- **Makefile Updates**: Fixed all development commands to use correct `mcp` subcommand structure

#### Pipe Function Architecture Refactoring
- **Modular Pipe Functions**: Extracted all pipe functions (exists, notNull, notEmpty, len, json) from main jsontest package into dedicated `jsontest/pipefuncs` package
- **Clean Switch Replacement**: Replaced large switch statement with pipe function registry using `err = pf.Handle(ctx, &out)` pattern
- **Circular Import Resolution**: Fixed import cycle by moving tests to `jsontest_test` package with side-effect import
- **Enhanced Framework**: Added `PipeFunc()` marker method and improved error handling with panic messages for missing registrations
- **JSON-RPC Protocol Testing**: Converted `jsonrpc_protocol_test.go` to use established `RunJSONRPCTest()` framework with proper JSON-RPC error code validation
- **Code Quality**: Added JSON-RPC 2.0 error code constants in `mcputil` package, replacing magic numbers with named constants

#### Integration Test Refactoring
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
- **CLI Architecture Refactoring**: Complete separation of MCP protocol I/O from user-facing CLI output
- **Help System Implementation**: Consistent, deterministic help output with `scout help` command
- **Testable CLI Framework**: Dependency injection allows full testing of CLI commands
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

### Current Known Bugs

### Current TODOs
- [ ] Update golang.CheckDocumentation() to limit the results returned to an estimate of 50% of the max size for a response, starting with symbol comments (first priority) file comments (second priority) and subdirectory README.md files (third priority)  
- [ ] Update all comments in the project using the check_docs tool. 
- [ ] Add logic to track what files have been read and require reading files before updating them.
- [ ] Add test to ensure that "/tmp" is always an allowed path but that we don't have it duplicated.
- [ ] Add unit tests for the CLI to `scoutcmds` that call `scout.RunMain()` and pass in CLI args from a table driven test and compares with expected output using buffered CLIWriter.
- [ ] Update integration tests in ./test to cover more of the use-cases based upon the values logged in ./log/test_responses.jsonl.
