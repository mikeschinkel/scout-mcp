# Scout-MCP: Secure File Search Server

Scout-MCP is a secure Model Context Protocol (MCP) server that allows Claude to search your whitelisted directories through stdio transport. Built with explicit directory whitelisting and robust security features.

## Features

- **Explicit Directory Whitelisting**: Only specified directories are accessible
- **Two Main Tools**:
  - `list_files`: List and search for files by name pattern in whitelisted directories  
  - `read_file`: Read contents of files from whitelisted directories
- **Security**: Path validation prevents access outside whitelisted directories
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
cd /path/to/scout-mcp/cmd
go build -o bin/scout-mcp
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
      "command": "/absolute/path/to/scout-mcp/cmd/bin/scout-mcp",
      "args": ["/Users/yourusername/Projects"]
    }
  }
}
```

**Configuration Notes:**
- Use **absolute paths** for both the command and arguments
- The `args` array should contain your whitelisted directories
- Scout-MCP runs as a subprocess of Claude Desktop
- No URL or network configuration needed

#### 4. Restart Claude Desktop

Completely quit and restart Claude Desktop for changes to take effect.

#### 5. Test Integration

In a new Claude conversation, try:
- "List files in my Projects directory"
- "Read the main.go file from scout-mcp"
- "Search for .go files in my current project"
- "Show me all README files"

### Multiple Directory Example

```json
{
  "mcpServers": {
    "scout-mcp": {
      "command": "/Users/mike/Projects/scout-mcp/cmd/bin/scout-mcp",
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
- Check that whitelisted directories exist
- Restart Claude Desktop after config changes
- Check Claude Desktop logs for error messages

**"Access denied" errors:**
- Ensure requested paths are within whitelisted directories
- Use absolute paths in configuration
- Verify directory permissions

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
  "whitelisted_paths": [
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
- Subdirectories of whitelisted paths are automatically accessible
- All paths are validated at startup

## Usage Examples

Once connected to Claude Desktop, you can use commands like:

```
"List files in my Projects directory"
"Show me all README files in ~/Projects"
"Read the main.go file from ~/Projects/scout-mcp"
"Find all files containing 'database' in the name in my Projects folder"
"Search for .go files recursively"
"List all .json files in my current project"
```

## API Tools

Scout-MCP provides two main tools to Claude:

### `list_files`
List and search for files by name pattern in whitelisted directories
- **Parameters**: 
  - `path` (required): Directory path to list
  - `recursive` (optional): Enable recursive listing (default: false)
  - `pattern` (optional): Name pattern to match
  - `extensions` (optional): Filter by file extensions
- **Returns**: JSON with file information (path, name, size, modified date, is_directory)

### `read_file`
Read contents of files from whitelisted directories
- **Parameters**: `path` (required): File path to read
- **Returns**: File contents as text

## Security Features

### Path Validation
- All file access requests are validated against the whitelist
- Absolute path resolution prevents directory traversal attacks
- Only directories (not individual files) can be whitelisted

### stdio Security
- No network exposure - communication only through Claude Desktop
- Process isolation - runs as subprocess with limited scope
- No authentication required for localhost stdio communication

### Access Logging
- Server logs all whitelisted directories on startup
- Invalid access attempts are logged with details

## Testing Your Setup

### Manual Testing with Claude Desktop

After configuring Claude Desktop integration:

1. **Start a new conversation** in Claude Desktop
2. **Test basic listing**: "List files in my Projects directory"
3. **Test recursive search**: "Search for .go files recursively in my scout-mcp project"
4. **Test file reading**: "Read the README.md file from my scout-mcp project"
5. **Test pattern matching**: "Show me all files containing 'config' in the name"

### Local Development Testing

Test the server standalone before Claude integration:

```bash
# Start Scout-MCP (will wait for stdio input)
./cmd/bin/scout-mcp ~/Projects

# Test with manual JSON (in another terminal, pipe input)
echo '{"id":1,"method":"tools/list","params":{}}' | ./cmd/bin/scout-mcp ~/Projects
```

### Integration Testing

Verify integration works correctly:

```
"List all files in my Projects directory"
"Search for .go files in my current project"
"Read the main.go file from my scout-mcp project"
"Find all package.json files recursively"
```

## Troubleshooting

### Command Line Issues

**"Error parsing arguments: path does not exist"**
- Verify the path exists: `ls -la /your/path`
- Use absolute paths for clarity
- Check permissions on the directory

**"No whitelisted directories specified"**
- Run `scout-mcp init` to create default config
- Or specify a path: `scout-mcp /your/project/path`
- Check config file exists: `cat ~/.config/scout-mcp/scout-mcp.json`

### Claude Desktop Integration Issues

**Tools not appearing in Claude:**
- Verify binary path in claude_desktop_config.json is correct and absolute
- Check that the binary is executable: `chmod +x /path/to/scout-mcp`
- Restart Claude Desktop completely after config changes
- Verify whitelisted directories exist and are readable

**"Access denied: path not whitelisted"**
- Check that the requested path is within a whitelisted directory
- Verify paths in config are absolute and exist
- Ensure Claude Desktop config args match your intended directories

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

- `whitelisted_paths`: Array of directory paths that Claude can access
- `port`: Port number (legacy - not used for stdio transport)
- `allowed_origins`: CORS origins (legacy - not used for stdio transport)

### Claude Desktop Configuration

**File Location**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)

- `command`: Absolute path to scout-mcp binary
- `args`: Array of whitelisted directory paths
- Server runs as subprocess of Claude Desktop

### Environment Considerations

- Ensure the user running Claude Desktop has read access to whitelisted directories
- Command line paths are not persisted to config file
- Use absolute paths to avoid working directory issues

## Future Development

This is an initial implementation. Planned enhancements include:

- **mcputil Integration**: Clean property-based API architecture
- **Enhanced tools**: File writing, git operations, project analysis  
- **User approval system**: For write operations with conversational prompts
- **Command execution**: Git commands with read/write classification
- **Logging and monitoring**: Structured logging and metrics
- **Security enhancements**: File size limits, operation timeouts

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

- Add capability to specify where the logs go, via environment variable, via CLI switch, and/or via config file. And I want it to be easy to say "In the current diretory" in in the `./log` directory.