package mcptools

import (
	_ "embed"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// helpContent contains embedded help documentation content for the Scout MCP server.
//
//go:embed HELP_TOOL_CONTENT.md
var helpContent string

func init() {
	mcputil.RegisterTool(mcputil.NewHelpTool(&HelpPayload{}))
}

var _ mcputil.Payload = (*HelpPayload)(nil)

// HelpPayload contains Scout-specific help content for the help tool.
type HelpPayload struct {
	ServerSpecificHelp string `json:"server_specific_help"` // Scout MCP server help content
}

// Payload implements the mcputil.Payload interface.
func (h *HelpPayload) Payload() {}

// Initialize loads the Scout-specific help content from embedded resources.
func (h *HelpPayload) Initialize(_ mcputil.Tool, _ mcputil.ToolRequest) (err error) {
	// Load Scout-specific help content
	h.ServerSpecificHelp = helpContent
	return nil
}
