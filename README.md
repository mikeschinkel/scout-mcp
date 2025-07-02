# Scout-MCP: Secure File Search Server

Scout-MCP is a secure Model Context Protocol (MCP) server that allows Claude to search your whitelisted directories through a CloudFlare tunnel. Built with explicit directory whitelisting and robust security features.

## Features

- **Explicit Directory Whitelisting**: Only specified directories are accessible
- **Two Main Tools**:
  - `search_files`: Search for files by name pattern in whitelisted directories
  - `read_file`: Read contents of files from whitelisted directories
- **Security**: Path validation prevents access outside whitelisted directories
- **CORS Support**: Configured for Claude.ai origins

## Quick Start

```bash
# Clone/create your project
git clone your-repo/scout-mcp
cd scout-mcp

# Initialize Go module
go mod init scout-mcp

# Create default config with custom directory
go run main.go init ~/Code

# Start server with config file paths
go run main.go

# Or add additional path to config file paths
go run main.go ~/MyProjects

# Or use only a specific path (ignore config)
go run main.go --only /tmp/safe-dir
```

## Documentation

- **[CloudFlare Tunnel Setup](docs/cloudflare-tunnel-setup.md)** - Complete guide to setting up secure tunnel access
- **[Configuration Guide](#configuration)** - Local configuration options
- **[API Reference](#api-tools)** - Available tools and their parameters
- **[Troubleshooting](#troubleshooting)** - Common issues and solutions

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
  "port": "8080",
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

## Setup Process

1. **Configure Scout-MCP**: Create config file and whitelist directories
2. **Set up CloudFlare Tunnel**: Follow the [CloudFlare Tunnel Setup Guide](docs/cloudflare-tunnel-setup.md)
3. **Connect to Claude**: Add the tunnel URL as a custom integration
4. **Start using**: Search and read files through Claude conversations

## Usage Examples

Once connected to Claude, you can use commands like:

```
"Search for Go files in my Projects directory"
"Show me all README files in ~/Projects"
"Read the main.go file from ~/Projects/scout-mcp"
"Find all files containing 'database' in the name in my Projects folder"
"List all .json files in my current project"
```

## API Tools

Scout-MCP provides two main tools to Claude:

### `search_files`
Search for files by name pattern in whitelisted directories
- **Parameters**: `path` (required), `pattern` (optional)
- **Returns**: Array of file information (path, name, size, modified date, is_directory)

### `read_file`
Read contents of files from whitelisted directories
- **Parameters**: `path` (required)
- **Returns**: File contents as text

## Security Features

### Path Validation
- All file access requests are validated against the whitelist
- Absolute path resolution prevents directory traversal attacks
- Only directories (not individual files) can be whitelisted

### CORS Protection
- Configured to only allow requests from Claude.ai domains
- Can be customized in the configuration file

### Access Logging
- Server logs all whitelisted directories on startup
- Invalid access attempts are logged with details

## Testing Your Setup

### Local Testing

Before setting up the tunnel, test Scout-MCP locally:

```bash
# Start Scout-MCP
scout-mcp ~/MyProjects

# In another terminal, test the API
curl -X POST http://localhost:8080/sse \
  -H "Content-Type: application/json" \
  -d '{"id":"1","method":"tools/list","params":{}}'

# Test file search
curl -X POST http://localhost:8080/sse \
  -H "Content-Type: application/json" \
  -d '{"id":"2","method":"tools/call","params":{"name":"search_files","arguments":{"path":"/your/path","pattern":"*.go"}}}'
```

### Integration Testing

Once tunnel is set up and connected to Claude:

```
"List all files in my Projects directory"
"Search for .go files in my current project"
"Read the README.md file from my scout-mcp project"
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

### Server Issues

**"Access denied: path not whitelisted"**
- Check that the requested path is within a whitelisted directory
- Verify paths in config are absolute and exist
- Restart server after config changes

**Connection issues**
- See the [CloudFlare Tunnel Setup Guide](docs/cloudflare-tunnel-setup.md) for tunnel-specific troubleshooting
- Verify Scout-MCP server is running on the correct port
- Ensure CORS origins include Claude.ai domains

## Configuration Reference

### Configuration File Options

**File Location**: `~/.config/scout-mcp/scout-mcp.json`

- `whitelisted_paths`: Array of directory paths that Claude can access
- `port`: Port number for the HTTP server (default: 8080)
- `allowed_origins`: Array of origins allowed for CORS (should include Claude.ai)

### Environment Considerations

- Ensure the user running the server has read access to whitelisted directories
- Consider running as a systemd service for production use
- Keep CloudFlare tunnel credentials secure
- Command line paths are not persisted to config file

## Future Development

This is an initial implementation. Planned enhancements include:

- **Cobra CLI framework**: `start`, `stop`, `status`, `add-path`, `remove-path` commands
- **Service management**: Systemd integration and daemon mode
- **Path management**: Commands to permanently add/remove paths from config
- **Enhanced tools**: File writing, git operations, project analysis
- **Logging and monitoring**: Structured logging and metrics
- **Security enhancements**: Authentication, rate limiting, audit trails

## Contributing

When contributing, please follow the "Clear Path" coding style used throughout the project:
- Use `goto end` pattern instead of early returns
- Single return point per function with named return variables
- Minimal nesting through helper functions
- All variables declared before first `goto`
- No variable shadowing

## License

MIT