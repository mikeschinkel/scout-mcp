package mcptools

import (
	_ "embed"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

//go:embed HELP_TOOL_CONTENT.md
var helpContent string

func init() {
	mcputil.RegisterTool(mcputil.NewHelpTool(&HelpPayload{}))
}

var _ mcputil.Payload = (*HelpPayload)(nil)

// HelpPayload contains Scout-specific help content
type HelpPayload struct {
	ServerSpecificHelp string `json:"server_specific_help"`
}

func (h *HelpPayload) Payload() {}

func (h *HelpPayload) Initialize(_ mcputil.Tool, _ mcputil.ToolRequest) (err error) {
	// Load Scout-specific help content
	h.ServerSpecificHelp = helpContent
	return nil
}
