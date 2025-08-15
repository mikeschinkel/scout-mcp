package mcptools

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

// homeDir stores the user's home directory path for session initialization.
var homeDir string

// instructionsMsg contains additional instructions displayed during session creation.
var instructionsMsg = `
MORE IMPORTANT INSTRUCTIONS:
1. **Language Instructions**: Review the language-specific instructions for proper coding style
2. **Validation**: If you see instruction files with non-standard names like 'golang.md', 'js.md', or 'py.md', warn the user that these should be renamed to 'go.md', 'javascript.md', and 'python.md' respectively for best compatibility
`

func init() {
	mcputil.RegisterTool(mcputil.NewStartSessionTool(&StartSessionResult{}))
}

var _ mcputil.Payload = (*StartSessionResult)(nil)

// ExtensionMappings maps file extensions to their corresponding language names.
type ExtensionMappings map[string]string

// StartSessionResult contains the response structure for session creation with instructions and configuration.
type StartSessionResult struct {
	MoreInstructions     string                  `json:"more_instructions,omitempty"` // Generic instruction payload
	Message              string                  `json:"message"`
	QuickStart           []string                `json:"quick_start,omitempty"` // Generic quick start payload
	LanguageInstructions []LanguageInstructions  `json:"language_instructions"`
	ExtensionMappings    ExtensionMappings       `json:"extension_mappings"`
	ServerConfig         map[string]any          `json:"server_config,omitempty"`   // Generic config payload
	CurrentProject       *ProjectDetectionResult `json:"current_project,omitempty"` // Generic project info
}

// Payload implements the mcputil.Payload interface.
func (sr *StartSessionResult) Payload() {}
func (sr *StartSessionResult) ensureInstructionsDirectory() (err error) {
	var dir string

	// Get instructions directory
	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}
	dir = getInstructionsDir()
	err = os.MkdirAll(dir, 0755)
end:
	return err
}

func errAndLog(err error, msg string, args ...any) error {
	logger.Error(msg, append(args, "error", err)...)
	return fmt.Errorf("%s: %w", msg, err)
}

func (sr *StartSessionResult) Initialize(t mcputil.Tool, _ mcputil.ToolRequest) (err error) {
	var userInstructions string

	sst, ok := t.(*mcputil.StartSessionTool)
	if !ok {
		err = fmt.Errorf("tool is %T, expected *mcputil.StartSessionTool", t)
		goto end
	}

	err = sr.ensureInstructionsDirectory()
	if err != nil {
		err = errAndLog(err, "unable to ensure instructions directory", []any{
			"directory", getInstructionsDir(),
		})
		goto end
	}

	// Detect current project (optional - don't fail session creation if this fails)
	sr.CurrentProject, err = sr.detectCurrentProject(sst)
	if err != nil {
		logger.Warn("unable to detect current project", "error", err)
		err = nil // Reset error so session creation continues
	}

	// Get server config (from existing get_config functionality)
	sr.ServerConfig, err = sst.Config().ToMap()
	if err != nil {
		err = errAndLog(err, "failed to generate server config", nil)
		goto end
	}

	// Load instructions
	sr.LanguageInstructions, err = sr.loadLanguageInstructions()
	if err != nil {
		err = errAndLog(err, "failed to load language instructions", nil)
		goto end
	}

	userInstructions, err = sr.loadInstructions()
	if err != nil {
		err = errAndLog(err, "failed to load general instructions", nil)
		goto end
	}
	sr.MoreInstructions = fmt.Sprintf("%s\n\nCUSTOM INSTRUCTIONS\n%s", instructionsMsg, userInstructions)

	// Generate quick start list
	sr.QuickStart = sr.generateQuickStartList()

	// Load instructions
	sr.ExtensionMappings = sr.loadExtensionMappings()

end:
	if err != nil {
		sr.CurrentProject = nil
		sr.ServerConfig = nil
		sr.LanguageInstructions = nil
	}
	return err
}

// generateQuickStartList creates a list of essential tools with their quick help descriptions
func (sr *StartSessionResult) generateQuickStartList() []string {
	var tools []mcputil.Tool
	var tool mcputil.Tool
	var options mcputil.ToolOptions
	var quickStartList []string

	tools = mcputil.RegisteredTools()
	quickStartList = make([]string, 0)

	for _, tool = range tools {
		options = tool.Options()
		if options.QuickHelp != "" {
			quickStartList = append(quickStartList, fmt.Sprintf("%s - %s", options.Name, options.QuickHelp))
		}
	}

	return quickStartList
}

// instructionsDir returns the directory where instructions can be found
// TODO Centralize this knowledge
// loadLanguageInstructions loads instruction files from the config directory
func (sr *StartSessionResult) loadLanguageInstructions() (instructions []LanguageInstructions, err error) {
	var entries []os.DirEntry
	var entry os.DirEntry
	var filename string
	var language string
	var version string
	var content []byte
	var hasGo bool

	// Load language-specific instructions
	instructions = make([]LanguageInstructions, 0)

	entries, err = os.ReadDir(getInstructionsDir())
	if err != nil {
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

		if !hasGo && language == "go" {
			hasGo = true
		}

		// Read file content
		content, err = os.ReadFile(filepath.Join(getInstructionsDir(), filename))
		if err != nil {
			continue // Skip files we can't read
		}

		instructions = append(instructions, LanguageInstructions{
			Language:   language,
			Version:    version,
			Content:    string(content),
			Extensions: getExtensionsForLanguage(language), //TODO Make this extensible via config
		})
	}

	if !hasGo {
		instructions = append(instructions, getGoLanguageInstructions())
	}

end:
	return instructions, err
}

// loadInstructions loads instruction files from the config directory
func (sr *StartSessionResult) loadExtensionMappings() ExtensionMappings {
	return getDefaultExtensionMappings()
}

// loadInstructions loads instruction files from the config directory
func (sr *StartSessionResult) loadInstructions() (instructions string, err error) {
	var content []byte

	// Load general instructions
	content, err = os.ReadFile(filepath.Join(getInstructionsDir(), "general.md"))
	if err != nil {
		// If general.md doesn't exist, provide default
		instructions = getDefaultGeneralInstructions()
		err = nil
		goto end
	}
	instructions = string(content)

end:
	return instructions, err
}

// detectCurrentProject runs the same logic as the detect_current_project tool
func (sr *StartSessionResult) detectCurrentProject(sst *mcputil.StartSessionTool) (result *ProjectDetectionResult, err error) {
	var detectTool *DetectCurrentProjectTool
	var allowedPaths []string
	var detectionResult ProjectDetectionResult

	// Get allowed paths from config
	allowedPaths = sst.Config().AllowedPaths()
	if len(allowedPaths) == 0 {
		// No allowed paths configured, return nil without error
		return nil, nil
	}

	// Create a DetectCurrentProjectTool instance to reuse its logic
	detectTool = &DetectCurrentProjectTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name: "detect_current_project_internal",
		}),
	}
	detectTool.SetConfig(sst.Config())

	// Use default parameters: use default max_projects (5), require git
	detectionResult, err = detectTool.detectCurrentProject(5, false)
	if err != nil {
		goto end
	}

	result = &detectionResult

end:
	return result, err
}

func getInstructionsDir() string {
	// TOOD: Get this from a single source of truth
	return filepath.Join(homeDir, ".config", "scout-mcp", "instructions")
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
	var mappings ExtensionMappings
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

func getGoLanguageInstructions() LanguageInstructions {
	return LanguageInstructions{
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
	}
}

func getDefaultExtensionMappings() ExtensionMappings {
	return ExtensionMappings{
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
