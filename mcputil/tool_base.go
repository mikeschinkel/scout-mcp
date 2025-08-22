package mcputil

import (
	"context"
	"encoding/json"
	"strings"
)

// ToolBase provides common functionality for all MCP tools including
// configuration management, path validation, and session handling.
type ToolBase struct {
	config  Config      // Server configuration for path validation
	options ToolOptions // Tool-specific options and metadata
}

// NewToolBase creates a new ToolBase instance with the specified options.
// This constructor initializes the tool with common functionality for MCP tools.
func NewToolBase(options ToolOptions) *ToolBase {
	options.Name = strings.ToLower(options.Name)
	return &ToolBase{
		options: options,
	}
}

// IsAllowedPath checks if the specified path is allowed based on server configuration.
// This method provides path validation security for file system operations.
func (b *ToolBase) IsAllowedPath(path string) bool {
	return b.config.IsAllowedPath(path)
}

// Name returns the tool's name.
// This method implements the Tool interface Name requirement.
func (b *ToolBase) Name() string {
	return b.options.Name
}

// ToMap converts the ToolBase instance to a map representation.
// This method provides JSON serialization support for tool introspection.
func (b *ToolBase) ToMap() (m map[string]any, err error) {
	var bytes []byte
	bytes, err = json.Marshal(b)
	if err != nil {
		goto end
	}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		goto end
	}
end:
	return m, err
}

// SetConfig sets the server configuration for this tool.
// This method provides the tool with access to server settings and security policies.
func (b *ToolBase) SetConfig(c Config) {
	b.config = c
}

// Config returns the current server configuration.
// This method provides access to server settings for path validation and other security checks.
func (b *ToolBase) Config() Config {
	return b.config
}

// Options returns the tool's options and metadata.
// This method implements the Tool interface Options requirement.
func (b *ToolBase) Options() ToolOptions {
	return b.options
}

// HasRequiredParams returns true if the tool has any required parameters beyond session_token.
// This method is used by testing frameworks to determine if mock parameters are needed.
func (b *ToolBase) HasRequiredParams() (hasParams bool) {
	// Check individual required properties
	for _, prop := range b.options.Properties {
		if !prop.IsRequired() {
			continue
		}
		// Skip session_token as it's handled automatically in tests
		if prop.GetName() == "session_token" {
			continue
		}
		hasParams = true
		goto end
	}

	// Check complex requirements
	if len(b.options.Requires) == 0 {
		goto end
	}
	hasParams = true

end:
	return hasParams
}

// EnsurePreconditions checks all shared preconditions for tools.
// This method implements session validation and other framework-level security checks.
func (b *ToolBase) EnsurePreconditions(ctx context.Context, req ToolRequest) (err error) {
	var sessionToken string
	var testing, ok bool

	// Session validation (skip for start_session tool)
	if b.options.Name == "start_session" {
		goto end
	}

	sessionToken, err = RequiredSessionTokenProperty.String(req)
	if err != nil {
		goto end
	}

	// TODO This is a short-term hack. Replace with dependency injection.
	testing, ok = ctx.Value("testing").(bool)
	if ok && testing {
		goto end
	}
	err = ValidateSession(sessionToken)
	if err != nil {
		goto end
	}

end:
	return err
}
