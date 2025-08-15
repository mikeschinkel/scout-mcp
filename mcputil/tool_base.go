package mcputil

import (
	"context"
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
func (b *ToolBase) HasRequiredParams() (hasParams bool) {
	// Check individual required properties
	for _, prop := range b.options.Properties {
		if ! prop.IsRequired(){
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
	hasParams= true

end:
	return hasParams
}

// EnsurePreconditions checks all shared preconditions for tools
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
