package mcputil

import (
	"encoding/json"
	"strings"
)

//type Config = Config

// TODO

type ToolBase struct {
	config  Config
	options ToolOptions
}

func NewToolBase(options ToolOptions) *ToolBase {
	options.Name = strings.ToLower(options.Name)
	return &ToolBase{
		options: options,
	}
}

func (b *ToolBase) IsAllowedPath(path string) bool {
	return b.config.IsAllowedPath(path)
}

func (b *ToolBase) Name() string {
	return b.options.Name
}

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

func (b *ToolBase) SetConfig(c Config) {
	b.config = c
}

func (b *ToolBase) Config() Config {
	return b.config
}

func (b *ToolBase) Options() ToolOptions {
	return b.options
}

// HasRequiredParams returns true if the tool has any required parameters beyond session_token
func (b *ToolBase) HasRequiredParams() bool {
	// Check individual required properties
	for _, prop := range b.options.Properties {
		// Skip session_token as it's handled automatically in tests
		if prop.IsRequired() && prop.GetName() != "session_token" {
			return true
		}
	}

	// Check complex requirements
	if len(b.options.Requires) > 0 {
		return true
	}

	return false
}

// EnsurePreconditions checks all shared preconditions for tools
func (b *ToolBase) EnsurePreconditions(req ToolRequest) (err error) {
	var sessionToken string

	// Session validation (skip for start_session tool)
	if b.options.Name == "start_session" {
		goto end
	}

	sessionToken, err = RequiredSessionTokenProperty.String(req)
	if err != nil {
		goto end
	}

	err = ValidateSession(sessionToken)
	if err != nil {
		goto end
	}

end:
	return err
}
