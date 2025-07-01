# scout-mcp
An MCP Server for enabling Claude UI to access local files using a Cloudflare Tunnel.

# Secure File Search MCP Server Setup Guide

This guide will help you set up a secure Model Context Protocol (MCP) server that allows Claude to search your whitelisted directories through a CloudFlare tunnel.

## Features

- **Explicit Directory Whitelisting**: Only specified directories are accessible
- **Two Main Tools**:
    - `search_files`: Search for files by name pattern in whitelisted directories
    - `read_file`: Read contents of files from whitelisted directories
- **Security**: Path validation prevents access outside whitelisted directories
- **CORS Support**: Configured for Claude.ai origins

## Setup Instructions

### 1. Initialize the Project

```bash
# Create project directory
mkdir mcp-file-search
cd mcp-file-search

# Initialize Go module
go mod init mcp-file-search

# Create the main.go file (use the code from the artifact above)

# Generate default configuration
go run main.go init
```

### 2. Configure Whitelisted Directories

Edit the generated `config.json` file:

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

**Important Security Notes:**
- Only add directories you want Claude to access
- Use absolute paths for clarity
- The server validates that all paths are directories and exist
- Subdirectories of whitelisted paths are automatically accessible

### 3. Set Up CloudFlare Tunnel

Install CloudFlare Tunnel (cloudflared):

```bash
# Download and install cloudflared
# Visit: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/

# Login to CloudFlare
cloudflared tunnel login

# Create a tunnel
cloudflared tunnel create mcp-file-search

# Configure the tunnel (create config.yml in ~/.cloudflared/)
```

Example CloudFlare tunnel configuration (`~/.cloudflared/config.yml`):

```yaml
tunnel: YOUR_TUNNEL_ID
credentials-file: /home/yourusername/.cloudflared/YOUR_TUNNEL_ID.json

ingress:
  - hostname: your-mcp-server.your-domain.com
    service: http://localhost:8080
  - service: http_status:404
```

### 4. Start the Services

Terminal 1 - Start the MCP server:
```bash
go run main.go
# Review the whitelisted directories
# Press Enter to start
```

Terminal 2 - Start the CloudFlare tunnel:
```bash
cloudflared tunnel --config ~/.cloudflared/config.yml run
```

### 5. Add Integration to Claude

1. Go to Claude.ai → Settings → Integrations
2. Click "Add Custom Integration"
3. Enter your tunnel URL: `https://your-mcp-server.your-domain.com/sse`
4. Complete the authentication process

## Usage Examples

Once connected, you can use Claude with commands like:

```
"Search for Go files in my Projects directory"
"Show me all README files in ~/Projects"
"Read the main.go file from ~/Projects/my-app"
"Find all files containing 'database' in the name in my Projects folder"
```

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

## Troubleshooting

### Common Issues

1. **"Access denied: path not whitelisted"**
    - Check that the requested path is within a whitelisted directory
    - Ensure paths in config.json are absolute and exist

2. **Connection issues**
    - Verify CloudFlare tunnel is running and accessible
    - Check that the port in config matches the tunnel configuration
    - Ensure CORS origins include Claude.ai domains

3. **File not found errors**
    - Verify the file exists and is readable
    - Check file permissions

### Testing the Server

You can test the server directly with curl:

```bash
# Test tools list
curl -X POST http://localhost:8080/sse \
  -H "Content-Type: application/json" \
  -d '{"id":"1","method":"tools/list","params":{}}'

# Test file search
curl -X POST http://localhost:8080/sse \
  -H "Content-Type: application/json" \
  -d '{"id":"2","method":"tools/call","params":{"name":"search_files","arguments":{"path":"/home/user/Projects","pattern":"*.go"}}}'
```

## Configuration Reference

### config.json Options

- `whitelisted_paths`: Array of directory paths that Claude can access
- `port`: Port number for the HTTP server (default: 8080)
- `allowed_origins`: Array of origins allowed for CORS (should include Claude.ai)

### Environment Considerations

- Ensure the user running the server has read access to whitelisted directories
- Consider running the server as a systemd service for production use
- Keep CloudFlare tunnel credentials secure

## Next Steps

Once running, you can:
- Add more whitelisted directories as needed
- Extend the server with additional tools (e.g., file writing, git operations)
- Set up monitoring and logging for production use
- Configure automatic startup as a system service