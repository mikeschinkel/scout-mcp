# Scout-MCP: Secure File Operations Server

Scout-MCP is a comprehensive Model Context Protocol (MCP) server that provides Claude with secure file access and manipulation capabilities through stdio transport. Built with explicit directory whitelisting, session-based instruction enforcement, and advanced file operations including AST-based editing.

## Features

- **Session-Based Instruction Enforcement**: All interactions start with comprehensive instructions and coding guidelines
- **Explicit Directory Whitelisting**: Only specified directories are accessible
- **Comprehensive File Operations**: 20 tools for reading, writing, editing, and analyzing files
  - **Session Management**: `start_session` tool provides instructions and session tokens
  - **Efficient File Reading**: `read_files` tool can read multiple files/directories in one call
  - **Basic Operations**: create, update, delete files and search directories
  - **Advanced Editing**: Line-based operations, pattern replacement, AST-based editing
  - **Language-Aware**: Syntax-aware editing for Go, Python, JavaScript, and more
  - **Analysis Tools**: File validation, content analysis, and structure inspection
- **User Approval System**: Write operations require explicit user confirmation with risk assessment
- **Security**: Multi-layered security with path validation, session management, and approval workflows
- **stdio Transport**: Direct integration with Claude Desktop via subprocess

## Quick Start

```bash
# Clone/create your project
git clone your-repo/scout-mcp
cd scout-mcp

# Initialize Go module
go mod init scout-mcp

# Create default config with custom directory
go run cmd/main.go init ~/Code

# Start server with config file paths
go run cmd/main.go

# Or add additional path to config file paths
go run cmd/main.go ~/MyProjects

# Or use only a specific path (ignore config)
go run cmd/main.go --only /tmp/safe-dir
```

## Documentation

- **[Claude Desktop Integration](#claude-desktop-integration)** - Connect Scout-MCP to Claude Desktop
- **[Configuration Guide](#configuration)** - Local configuration options
- **[API Reference](#api-tools)** - Available tools and their parameters
- **[Troubleshooting](#troubleshooting)** - Common issues and solutions

## Claude Desktop Integration

Scout-MCP uses stdio transport for direct integration with Claude Desktop. This is **not** configured through Claude's web interface, but through Claude Desktop's configuration file.

### Setup Steps

#### 1. Build Scout-MCP
```bash
cd /path/to/scout-mcp
go build -o bin/scout-mcp ./cmd/main.go
```

#### 2. Find Claude Desktop Config File

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

#### 3. Configure Scout-MCP Server

Edit (or create) the Claude Desktop config file:

```json
{
  "mcpServers": {
    "scout-mcp": {
      "command": "/absolute/path/to/scout-mcp/bin/scout-mcp",
      "args": ["/Users/yourusername/Projects"]
    }
  }
}
```

**Configuration Notes:**
- Use **absolute paths** for both the command and arguments
- The `args` array should contain your allowed directories
- Scout-MCP runs as a subprocess of Claude Desktop
- No URL or network configuration needed

#### 4. Restart Claude Desktop

Completely quit and restart Claude Desktop for changes to take effect.

#### 5. Test Integration

In a new Claude conversation, you must **start with session setup**:

```
start_session
```

Then you can use commands like:
- "Read all Go files in my mcptools directory"
- "Show me the main files: main.go, README.md, and all files in ./mcptools"
- "Search for .go files containing 'tool' in my current project"
- "List all README files recursively"

### Multiple Directory Example

```json
{
  "mcpServers": {
    "scout-mcp": {
      "command": "/Users/mike/Projects/scout-mcp/bin/scout-mcp",
      "args": [
        "/Users/mike/Projects",
        "/Users/mike/Documents/Code"
      ]
    }
  }
}
```

### Integration Troubleshooting

**Tools not available in Claude:**
- Verify the binary path is correct and executable
- Check that allowed directories exist
- Restart Claude Desktop after config changes
- Check Claude Desktop logs for error messages

**"Access denied" errors:**
- Ensure requested paths are within allowed directories
- Use absolute paths in configuration
- Verify directory permissions

**"Invalid session token" errors:**
- Call `start_session` first to get a valid token
- Session tokens expire after 24 hours
- Tokens are invalidated when the server restarts

## Command Line Usage

Scout-MCP offers flexible path management through command line arguments:

**Basic Commands:**
- `scout-mcp <path>` - Add path to config file paths and start server
- `scout-mcp --only <path>` - Use only the specified path (ignore config file)
- `scout-mcp init` - Create empty config file (requires manual editing)
- `scout-mcp init <path>` - Create config with custom initial path
- `scout-mcp` - Start server with config file paths only

**Examples:**
```bash
# Start with ~/Projects (from config) + ~/MyCode (from command line)
scout-mcp ~/MyCode

# Use only /tmp/safe-dir, ignore any config file
scout-mcp --only /tmp/safe-dir

# Create config with ~/Development as initial directory
scout-mcp init ~/Development

# Start with just config file paths
scout-mcp
```

**Error Handling:**
- Running without arguments and no config file shows helpful usage information
- All paths are validated to exist and be directories before starting
- Clear error messages explain available options

## Configuration

### Configuration File

The configuration file is automatically created at `~/.config/scout-mcp/scout-mcp.json`:

```json
{
  "allowed_paths": [
    "/home/yourusername/Projects"
  ],
  "port": "8754",
  "allowed_origins": [
    "https://claude.ai",
    "https://*.anthropic.com"
  ]
}
```

### Path Management

**Config File Paths**: Persistent directories specified in the configuration file
**Command Line Paths**: Temporary paths added for a single session
**Combined Mode** (default): Command line paths are added to config file paths
**Only Mode** (`--only` flag): Uses only command line paths, ignoring config file

**Security Notes:**
- Only add directories you want Claude to access
- Use absolute paths for clarity
- Subdirectories of allowed paths are automatically accessible
- All paths are validated at startup

## Usage Examples

Once connected to Claude Desktop, you **must start every conversation** with:

```
start_session
```

This provides you with a session token and comprehensive instructions. Then you can use commands like:

```
"Read these files: main.go, README.md, and all .go files in ./mcptools"
"Search for all .go files containing 'tool' recursively"
"Show me all configuration files in my project"
"Read the entire ./test directory to understand the test structure"
```

## API Tools

Scout-MCP provides 20 comprehensive tools to Claude:

### Session Management
- **`start_session`**: ⭐ **START HERE** - Create session token and get comprehensive instructions (no session token required)

### Enhanced File Reading
- **`read_files`**: Read multiple files and/or directories efficiently with filtering options
- **`search_files`**: List and search for files by name pattern in allowed directories

### Basic File Operations (require approval)
- **`create_file`**: Create new files in allowed directories
- **`update_file`**: Replace entire file contents (⚠️ dangerous - use granular tools instead)
- **`delete_files`**: Delete files or directories

### Granular Editing Operations (require approval)
- **`update_file_lines`**: Replace specific lines in a file by line number range
- **`delete_file_lines`**: Delete specific line ranges from a file
- **`insert_file_lines`**: Insert content at specific line numbers
- **`insert_at_pattern`**: Insert content before/after pattern matches
- **`replace_pattern`**: Find and replace text patterns with regex support

### Language-Aware Operations (AST-based)
- **`find_file_part`**: Find specific language constructs (functions, types, etc.)
- **`replace_file_part`**: Replace language constructs using syntax-aware parsing (requires approval)
- **`validate_files`**: Validate syntax of source code files

### Analysis and System Tools
- **`analyze_files`**: Analyze file structure and provide insights
- **`get_config`**: Show current Scout-MCP configuration
- **`tool_help`**: Get detailed documentation for all tools
- **`detect_current_project`**: Detect the most recently active project by analyzing recent file modifications in Git repositories

### Approval System
- **`request_approval`**: Request user approval for risky operations
- **`generate_approval_token`**: Generate approval tokens after user confirmation

## Security Features

### Multi-Layered Security Model
Scout-MCP implements comprehensive security through multiple mechanisms:

### Session Management
- **Session tokens required**: All tools (except `start_session`) require valid session tokens
- **24-hour expiration**: Session tokens automatically expire
- **Server restart invalidation**: Tokens invalidated when server restarts
- **Instruction delivery**: Each session provides coding guidelines and tool documentation

### Path Validation
- All file access requests are validated against the whitelist
- Absolute path resolution prevents directory traversal attacks
- Only directories (not individual files) can be allowed

### User Approval System
- **Write operations require explicit user confirmation** via stdio prompts
- **Risk assessment**: Operations are classified as low, medium, or high risk
- **Operation preview**: Users see exactly what will be changed before approval
- **Approval tokens**: Generated after user confirmation for secure operation execution
- **Interactive prompts**: Clear descriptions of what each operation will do

### Operation Classification
- **Read operations**: Automatically allowed for files in whitelisted directories
- **Write operations**: Require user approval with risk-based warnings
- **Analysis operations**: Safe read-only analysis and validation tools

### stdio Security
- No network exposure - communication only through Claude Desktop
- Process isolation - runs as subprocess with limited scope
- No authentication required for localhost stdio communication

### Access Logging
- Server logs all allowed directories on startup
- Invalid access attempts are logged with details
- Approval requests and responses are logged for audit trails

## Testing Your Setup

### Manual Testing with Claude Desktop

After configuring Claude Desktop integration:

1. **Start a new conversation** in Claude Desktop
2. **Start session first**: Call `start_session` to get token and instructions
3. **Test basic reading**: "Read all files in my mcptools directory"
4. **Test recursive search**: "Search for .go files recursively in my scout-mcp project"
5. **Test multiple file reading**: "Read these files: main.go, README.md"
6. **Test pattern matching**: "Show me all files containing 'config' in the name"

### Local Development Testing

Test the server and run the comprehensive test suite:

```bash
# Run the comprehensive test suite
./test/run_tests.sh

# Run individual Go tests
go test ./...

# Test the server standalone before Claude integration
./bin/scout-mcp ~/Projects

# Test with manual JSON (in another terminal, pipe input)
echo '{"id":1,"method":"tools/list","params":{}}' | ./bin/scout-mcp ~/Projects
```

### Integration Testing

Verify integration works correctly:

```
start_session
"Read all files in my mcptools directory"
"Search for .go files in my current project"
"Read main.go and README.md together"
"Find all package.json files recursively"
```

## Troubleshooting

### Command Line Issues

**"Error parsing arguments: path does not exist"**
- Verify the path exists: `ls -la /your/path`
- Use absolute paths for clarity
- Check permissions on the directory

**"No allowed directories specified"**
- Run `scout-mcp init` to create default config
- Or specify a path: `scout-mcp /your/project/path`
- Check config file exists: `cat ~/.config/scout-mcp/scout-mcp.json`

### Claude Desktop Integration Issues

**Tools not appearing in Claude:**
- Verify binary path in claude_desktop_config.json is correct and absolute
- Check that the binary is executable: `chmod +x /path/to/scout-mcp`
- Restart Claude Desktop completely after config changes
- Verify allowed directories exist and are readable

**"Access denied: path not allowed"**
- Check that the requested path is within a allowed directory
- Verify paths in config are absolute and exist
- Ensure Claude Desktop config args match your intended directories

**"Invalid or expired session token"**
- Call `start_session` first to get a valid token
- Session tokens expire after 24 hours
- Tokens are invalidated when the server restarts
- Each new conversation should start with `start_session`

**Server not starting:**
- Test the binary manually: `./scout-mcp ~/Projects`
- Check for Go compilation errors
- Verify all dependencies are installed

### Connection Issues

**No response from tools:**
- Check Claude Desktop's developer console for error messages
- Verify the process is running: `ps aux | grep scout-mcp`
- Test with minimal configuration first

## Configuration Reference

### Configuration File Options

**File Location**: `~/.config/scout-mcp/scout-mcp.json`

- `allowed_paths`: Array of directory paths that Claude can access
- `port`: Port number (legacy - not used for stdio transport)
- `allowed_origins`: CORS origins (legacy - not used for stdio transport)

### Claude Desktop Configuration

**File Location**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)

- `command`: Absolute path to scout-mcp binary
- `args`: Array of allowed directory paths
- Server runs as subprocess of Claude Desktop

### Environment Considerations

- Ensure the user running Claude Desktop has read access to allowed directories
- Command line paths are not persisted to config file
- Use absolute paths to avoid working directory issues

## Architecture and Development

Scout-MCP follows a "Clear Path" coding style with comprehensive security and testing:

### Session-Based Architecture
- All tools require session tokens (except `start_session`)
- Framework-level session validation at MCP server layer
- Comprehensive instruction delivery with each session

### Tool Categories
1. **Session Management**: Token creation and instruction delivery
2. **File Reading**: Efficient multi-file reading with filtering
3. **File Management**: Create, update, delete operations with approval
4. **Granular Editing**: Line-based and pattern-based editing
5. **Language-Aware**: AST-based parsing for syntax-aware operations
6. **Analysis**: File validation and structure analysis
7. **System**: Configuration and help tools
8. **Approval**: Risk assessment and approval workflows

## Contributing

When contributing, please follow the "Clear Path" coding style used throughout the project:
- Use `goto end` pattern instead of early returns
- Single return point per function with named return variables
- Minimal nesting through helper functions
- All variables declared before first `goto`
- No variable shadowing

## License

MIT

## TODO (notes for me and Claude)

- Add capability to specify where the logs go, via environment variable, via CLI switch, and/or via config file. And I want it to be easy to say "In the current directory" or in the `./log` directory.

### JSON Schema
# TODO: JSON Schema Support for MCP Tools

## Requirements

### Core Implementation
- [ ] Add `Schema() map[string]any` method to Tool interface
- [ ] Generate JSON Schema from existing `Properties` structures (don't duplicate definitions)
- [ ] Return schema in MCP `list_tools` response for protocol compliance

### Schema Features
- [ ] **Type validation**: string, number, boolean, array, object types
- [ ] **Required vs optional parameters**: Mark session_token and core params as required
- [ ] **Default values**: Specify defaults for optional parameters (e.g., recursive: false)
- [ ] **Parameter descriptions**: Human-readable descriptions for each parameter
- [ ] **Array item types**: Specify types for array elements (e.g., paths array contains strings)

### Tool-Specific Schema Requirements
- [ ] **start_session**: No required parameters, returns session_token
- [ ] **read_files**: paths (required string array), session_token (required string)
- [ ] **search_files**: path (required string), pattern/extensions (optional), session_token (required)
- [ ] **File editing tools**: filepath + content + session_token (all required)
- [ ] **All tools except start_session**: session_token parameter marked as required

### Integration Benefits
- [ ] **Client validation**: Parameters validated before sending to server
- [ ] **Auto-completion**: IDEs and Claude Desktop can suggest parameters
- [ ] **Documentation**: tool_help can generate parameter docs from schema
- [ ] **Error messages**: Better validation errors for malformed requests

### Implementation Strategy
- [ ] Generate schema from existing Properties rather than hand-coding
- [ ] Keep Go-native Properties for internal use, schema for MCP protocol
- [ ] Add schema validation tests alongside existing unit tests
- [ ] Ensure backward compatibility with current tool implementations

### Future Considerations
- [ ] Schema versioning if tool parameters evolve
- [ ] Complex validation rules (e.g., path whitelist validation in schema)
- [ ] Conditional parameters (e.g., recursive only valid for directory operations)