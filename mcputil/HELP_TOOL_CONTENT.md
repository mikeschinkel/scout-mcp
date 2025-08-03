# MCP Tools Documentation

This document provides comprehensive documentation for all available MCP tools in the MCP server.

## Best Practices

### Getting Started
1. **Always start with `start_session`** to get your session token and instructions
2. Use the session token with all subsequent tool calls

### Security Notes

- All tools (except `start_session`) require a valid session token
- Session tokens expire after 24 hours
- All operations respect the configured allowed paths
- Tool operations validate parameters before execution

### Error Handling

Tools will return descriptive error messages for common issues:
- Invalid or expired session tokens
- Path not in allowed directories
- File not found
- Invalid parameters
- Permission errors

Always check the error field in tool responses and handle failures gracefully.

### Session Management

- Session tokens are valid for 24 hours
- Tokens are invalidated when the server restarts
- If you get "invalid session token" errors, call `start_session` again
- Each session provides fresh instructions and tool documentation

## Available Tools