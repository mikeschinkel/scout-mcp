package mcptools

import (
	"context"
	_ "embed"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

//go:embed README.md
var readmeContent string

var _ mcputil.Tool = (*ToolHelpTool)(nil)

func init() {
	mcputil.RegisterTool(&ToolHelpTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "tool_help",
			Description: "Get detailed documentation for MCP tools and best practices",
		}),
	})
}

type ToolHelpTool struct {
	*toolBase
}

func (t *ToolHelpTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var toolName string
	var helpContent string

	logger.Info("Tool called", "tool", "tool_help")

	// TODO: Check error here
	toolName = req.GetString("tool", "")

	if toolName == "" {
		// Return full documentation
		helpContent = readmeContent
		logger.Info("Tool completed", "tool", "tool_help", "type", "full_documentation")
	} else {
		// Return tool-specific documentation
		helpContent = t.getToolSpecificHelp(toolName)
		logger.Info("Tool completed", "tool", "tool_help", "type", "tool_specific", "requested_tool", toolName)
	}

	result = mcputil.NewToolResultText(helpContent)

	//end:
	return result, err
}

func (t *ToolHelpTool) getToolSpecificHelp(toolName string) (helpText string) {
	var sections map[string]string
	var found bool

	// Parse the README into tool-specific sections
	sections = t.parseToolSections()

	helpText, found = sections[toolName]
	if !found {
		helpText = t.getToolNotFoundHelp(toolName)
	}

	return helpText
}

func (t *ToolHelpTool) parseToolSections() (sections map[string]string) {
	var lines []string
	var currentSection string
	var currentContent strings.Builder
	var inToolSection bool

	sections = make(map[string]string)
	lines = strings.Split(readmeContent, "\n")

	for _, line := range lines {
		// Check if this is a tool header (### `tool_name`)
		if strings.HasPrefix(line, "### `") && strings.HasSuffix(line, "`") {
			// Save previous section if we were in one
			if inToolSection && currentSection != "" {
				sections[currentSection] = strings.TrimSpace(currentContent.String())
			}

			// Start new section
			currentSection = t.extractToolName(line)
			currentContent.Reset()
			currentContent.WriteString(line + "\n")
			inToolSection = true
		} else if strings.HasPrefix(line, "## ") {
			// Save current section when we hit a major heading
			if inToolSection && currentSection != "" {
				sections[currentSection] = strings.TrimSpace(currentContent.String())
				inToolSection = false
			}
		} else if inToolSection {
			// Add line to current section
			currentContent.WriteString(line + "\n")
		}
	}

	// Save final section
	if inToolSection && currentSection != "" {
		sections[currentSection] = strings.TrimSpace(currentContent.String())
	}

	return sections
}

func (t *ToolHelpTool) extractToolName(line string) (toolName string) {
	// Extract tool name from "### `tool_name`"
	start := strings.Index(line, "`")
	end := strings.LastIndex(line, "`")

	if start != -1 && end != -1 && start < end {
		toolName = line[start+1 : end]
	}

	return toolName
}

func (t *ToolHelpTool) getToolNotFoundHelp(toolName string) (helpText string) {
	var availableTools []string

	availableTools = []string{
		"read_file", "create_file", "update_file", "delete_file",
		"search_files", "update_file_lines", "insert_file_lines",
		"insert_at_pattern", "delete_file_lines", "replace_pattern",
		"get_config", "tool_help",
	}

	helpText = "Tool '" + toolName + "' not found.\n\n"
	helpText += "Available tools:\n"

	for _, tool := range availableTools {
		helpText += "- " + tool + "\n"
	}

	helpText += "\nCall tool_help without parameters to see full documentation, or specify a tool name:\n"
	helpText += `{"tool": "tool_help", "parameters": {"tool": "read_file"}}`

	return helpText
}
