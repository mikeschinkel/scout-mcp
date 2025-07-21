package mcptools

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*StartSessionTool)(nil)

func init() {
	mcputil.RegisterTool(&StartSessionTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "start_session",
			Description: "Start an MCP session and get comprehensive instructions for using Scout-MCP effectively",
			Properties:  []mcputil.Property{},
		}),
	})
}

type StartSessionTool struct {
	*toolBase
}

// EnsurePreconditions bypasses session validation for start_session but runs other preconditions
func (t *StartSessionTool) EnsurePreconditions(req mcputil.ToolRequest) (err error) {
	// start_session tool doesn't require any preconditions
	// Future non-session preconditions could be added here if needed
	return nil
}

//go:embed INSTRUCTIONS_MESSAGE.md
var instructionsMsg string

func (t *StartSessionTool) Handle(_ context.Context, _ mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var ss *Sessions
	var token string
	var expiresAt time.Time
	var response SessionResponse
	var toolHelp string
	var serverConfig map[string]any
	var instructions InstructionsConfig

	logger.Info("Tool called", "tool", "start_session")

	// Get session manager and create new session
	ss = GetSessions()
	token, expiresAt, err = ss.NewSession()
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to create session: %v", err))
		goto end
	}

	// Get tool help (from existing tool_help functionality)
	toolHelp, err = t.generateToolHelp()
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to generate tool help: %v", err))
		goto end
	}

	// Get server config (from existing get_config functionality)
	serverConfig, err = t.config.ToMap()
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to generate server config: %v", err))
		goto end
	}

	// Load instructions
	instructions, err = t.loadInstructions()
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to load instructions: %v", err))
		goto end
	}

	// Build response
	response = SessionResponse{
		SessionToken:   token,
		TokenExpiresAt: expiresAt,
		ToolHelp:       toolHelp,
		ServerConfig:   serverConfig,
		Instructions:   instructions,
		Message:        instructionsMsg,
	}

	logger.Info("Tool completed", "tool", "start_session", "success", true, "token_length", len(token))
	result = mcputil.NewToolResultJSON(response)

end:
	return result, err
}

// generateToolHelp creates the same output as the tool_help tool
func (t *StartSessionTool) generateToolHelp() (helpText string, err error) {
	var tools []mcputil.Tool
	var tool mcputil.Tool
	var options mcputil.ToolOptions
	var prop mcputil.Property
	var propOptions []mcputil.PropertyOption

	tools = mcputil.RegisteredTools()
	helpText = "# Scout-MCP Tools Documentation\n\n"

	for _, tool = range tools {
		options = tool.Options()
		helpText += fmt.Sprintf("## %s\n\n%s\n\n", options.Name, options.Description)

		if len(options.Properties) > 0 {
			helpText += "### Parameters\n\n"
			for _, prop = range options.Properties {
				propOptions = prop.PropertyOptions()
				helpText += fmt.Sprintf("- **%s** (%s)", prop.GetName(), getPropertyType(prop))

				// Check if required
				for _, opt := range propOptions {
					p, ok := opt.(mcputil.RequiredProperty)
					if ok && p.Required {
						helpText += " *[required]*"
						break
					}
				}

				helpText += fmt.Sprintf(": %s\n", getPropertyDescription(prop))
			}
			helpText += "\n"
		}
	}

	return helpText, err
}

// loadInstructions loads instruction files from the config directory
func (t *StartSessionTool) loadInstructions() (instructions InstructionsConfig, err error) {
	var homeDir string
	var instructionsDir string
	var generalPath string
	var generalContent []byte
	var entries []os.DirEntry
	var entry os.DirEntry
	var filename string
	var language string
	var version string
	var content []byte
	var langInstr LanguageInstructions
	var extensionMappings map[string]string

	// Get instructions directory
	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	instructionsDir = filepath.Join(homeDir, ".config", "scout-mcp", "instructions")

	// Load general instructions
	generalPath = filepath.Join(instructionsDir, "general.md")
	generalContent, err = os.ReadFile(generalPath)
	if err != nil {
		// If general.md doesn't exist, provide default
		instructions.General = getDefaultGeneralInstructions()
	} else {
		instructions.General = string(generalContent)
	}

	// Load language-specific instructions
	instructions.Languages = make([]LanguageInstructions, 0)

	entries, err = os.ReadDir(instructionsDir)
	if err != nil {
		// If instructions directory doesn't exist, use defaults
		instructions.Languages = getDefaultLanguageInstructions()
		instructions.ExtensionMappings = getDefaultExtensionMappings()
		err = nil
		goto end
	}

	for _, entry = range entries {
		if entry.IsDir() {
			continue
		}

		filename = entry.Name()
		if !strings.HasSuffix(filename, ".md") {
			continue
		}

		if filename == "general.md" {
			continue // Already processed
		}

		// Parse language and version from filename
		language, version = parseLanguageFilename(filename)
		if language == "" {
			continue
		}

		// Read file content
		content, err = os.ReadFile(filepath.Join(instructionsDir, filename))
		if err != nil {
			continue // Skip files we can't read
		}

		langInstr = LanguageInstructions{
			Language:   language,
			Version:    version,
			Content:    string(content),
			Extensions: getExtensionsForLanguage(language),
		}

		instructions.Languages = append(instructions.Languages, langInstr)
	}

	// Load extension mappings from config or use defaults
	extensionMappings = getDefaultExtensionMappings()
	// TODO: Load from config file when we add instructions config
	instructions.ExtensionMappings = extensionMappings

end:
	return instructions, err
}

// parseLanguageFilename parses "python-3.md" into language="python", version="3"
func parseLanguageFilename(filename string) (language string, version string) {
	var name string
	var parts []string

	name = strings.TrimSuffix(filename, ".md")
	parts = strings.Split(name, "-")

	language = parts[0]
	if len(parts) > 1 {
		version = strings.Join(parts[1:], "-")
	}

	return language, version
}

// getExtensionsForLanguage returns file extensions for a language
func getExtensionsForLanguage(language string) []string {
	var mappings map[string]string
	var extensions []string
	var ext string
	var lang string

	mappings = getDefaultExtensionMappings()
	extensions = make([]string, 0)

	for ext, lang = range mappings {
		if lang == language {
			extensions = append(extensions, ext)
		}
	}

	return extensions
}

// Default configurations
func getDefaultGeneralInstructions() string {
	return `# General Instructions for AI Assistants

## Core Problem
AI assistants often pattern-match to familiar solutions instead of understanding your specific architectural goals. They rush to implementation before grasping the design philosophy.

## Guidance Strategies

### 1. Stop Veering Early
When the AI starts suggesting familiar patterns that don't match your vision:
- **Say:** "Stop. That's not the pattern I want. Let me re-explain the architecture."
- **Don't:** Politely correct the direction and hope they figure it out

### 2. Force Understanding Before Implementation
Before any code gets written:
- **Say:** "Before you write any code, tell me back what architectural problem I'm trying to solve and why."
- **Require:** The AI to demonstrate understanding of your goals, not just the implementation

### 3. Be Direct About Pattern-Matching
When you see the AI forcing your design into familiar patterns:
- **Say:** "You're trying to force this into a familiar pattern instead of understanding what I need. Step back."
- **Be explicit:** Call out when they're pattern-matching instead of listening

### 4. Demand Architecture-First Discussions
- **Say:** "Don't write any code yet. First, explain back to me why [specific approach] is problematic for my use case."
- **Focus:** On design philosophy before implementation details

### 5. Use Strong Language When Needed
Instead of gentle corrections:
- **Say:** "You're not listening" or "You're making assumptions"
- **Be blunt:** "No, you're missing the point entirely. I'm not looking for [their suggestion]."

### 6. Treat AI Like Eager Junior Developer
- Assume they want to code before understanding requirements
- Force them to slow down and understand the problem
- Don't let them implement until they can articulate your architectural goals`
}

func getDefaultLanguageInstructions() []LanguageInstructions {
	return []LanguageInstructions{
		{
			Language: "go",
			Content: `# Clear Path Go Coding Style

## Core Principles
- **Minimize nesting wherever possible**
- **Avoid using 'else' wherever possible**
- **Use 'goto end' instead of early 'return'**
- **Put 'end:' label before the only return**
- **Place the sole return on the last line of the function**
- **Do not use more than the one 'end:' label**

## Go-Specific Rules
- Declare all vars prior to first 'goto' (Go team requirement)
- Use named returns variable in the 'func' signature for most funcs
- Do not use compound expressions in control flow statements like 'if'
- Use the named return variables on the final 'return'
- Always handle errors, even in 'defer'
- Do not shadow any variables
- Do not use ':=' after the first 'goto end'
- Leverage Go's 'zero' values with 'return' variables where possible`,
			Extensions: []string{".go"},
		},
	}
}

func getDefaultExtensionMappings() map[string]string {
	return map[string]string{
		".go":   "go",
		".py":   "python",
		".js":   "javascript",
		".jsx":  "javascript",
		".mjs":  "javascript",
		".ts":   "typescript",
		".tsx":  "typescript",
		".r":    "r",
		".php":  "php",
		".sh":   "bash",
		".bash": "bash",
		".zsh":  "bash",
		".rs":   "rust",
		".c":    "c",
		".cpp":  "cpp",
		".h":    "c",
		".hpp":  "cpp",
	}
}

// Helper functions for property information (simplified versions from tool_help)
func getPropertyType(_ mcputil.Property) string {
	panic("IMPLEMENT ME")
	// This is a simplified version - you may need to implement based on your property types
	return "string" // Default, should be enhanced based on actual property type
}

func getPropertyDescription(prop mcputil.Property) string {
	// Extract description from property options
	var options []mcputil.PropertyOption
	var opt mcputil.PropertyOption

	options = prop.PropertyOptions()
	for _, opt = range options {
		p, ok := opt.(mcputil.DescriptionProperty)
		if ok {
			return p.Description
		}
	}
	return "No description available"
}
