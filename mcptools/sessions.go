package mcptools

// LanguageInstructions contains instructions for a specific language
type LanguageInstructions struct {
	Language   string   `json:"language"`          // "python", "go", "javascript"
	Version    string   `json:"version,omitempty"` // "2", "3", "5.7", "8.4", etc.
	Content    string   `json:"content"`           // Markdown content
	Extensions []string `json:"extensions"`        // File extensions this applies to
}
