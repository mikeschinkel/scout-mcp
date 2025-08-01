package mcptools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

type Config = mcputil.Config

type toolBase struct {
	config  Config
	options mcputil.ToolOptions
}

func newToolBase(options mcputil.ToolOptions) *toolBase {
	options.Name = strings.ToLower(options.Name)
	return &toolBase{
		options: options,
	}
}

func (b *toolBase) IsAllowedPath(path string) bool {
	return b.config.IsAllowedPath(path)
}

func (b *toolBase) Name() string {
	return b.options.Name
}

func (b *toolBase) ToMap() (m map[string]any, err error) {
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

func (b *toolBase) SetConfig(c Config) {
	b.config = c
}

func (b *toolBase) Config() Config {
	return b.config
}

func (b *toolBase) Options() mcputil.ToolOptions {
	return b.options
}

// EnsurePreconditions checks all shared preconditions for tools
func (b *toolBase) EnsurePreconditions(req mcputil.ToolRequest) (err error) {
	var sessionToken string

	// Session validation (skip for start_session tool)
	if b.options.Name != "start_session" {
		sessionToken = req.GetString("session_token", "")
		err = RequireValidSession(sessionToken)
		if err != nil {
			goto end
		}
	}

	// Future preconditions can be added here:
	// - Rate limiting checks
	// - User permission validation
	// - Feature flag checks
	// - etc.

end:
	return err
}

// File system operations
func (b *toolBase) readFile(path string) (content string, err error) {
	var data []byte

	data, err = os.ReadFile(path)
	if err != nil {
		goto end
	}

	content = string(data)

end:
	return content, err
}

func (b *toolBase) writeFile(path, content string) (err error) {
	err = os.WriteFile(path, []byte(content), 0644)
	return err
}

func (b *toolBase) writeFileWithValidation(path, content string) (err error) {
	var language string

	// Detect language from file extension
	language = b.detectLanguageFromExtension(path)

	// If we can detect the language, validate syntax before writing
	if language != "" {
		err = langutil.ValidateSyntax(language, content)
		if err != nil {
			err = fmt.Errorf("validation failed - would result in invalid %s syntax: %w", language, err)
			goto end
		}
	}

	// Write file
	err = b.writeFile(path, content)

end:
	return err
}

func (b *toolBase) searchForFiles(path string, recursive bool, extensions []string) (files []string, err error) {
	var info os.FileInfo

	info, err = os.Stat(path)
	if err != nil {
		goto end
	}

	if !info.IsDir() {
		// Single file
		if b.matchesExtensions(path, extensions) {
			files = []string{path}
		}
		goto end
	}

	// Directory - collect files
	if recursive {
		err = b.walkDirectory(path, extensions, &files)
	} else {
		err = b.listDirectory(path, extensions, &files)
	}

end:
	return files, err
}

func (b *toolBase) walkDirectory(path string, extensions []string, files *[]string) (err error) {
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if !info.IsDir() && b.matchesExtensions(filePath, extensions) {
			*files = append(*files, filePath)
		}

		return nil
	})

	return err
}

func (b *toolBase) listDirectory(path string, extensions []string, files *[]string) (err error) {
	var entries []os.DirEntry

	entries, err = os.ReadDir(path)
	if err != nil {
		goto end
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(path, entry.Name())
			if b.matchesExtensions(filePath, extensions) {
				*files = append(*files, filePath)
			}
		}
	}

end:
	return err
}

func (b *toolBase) matchesExtensions(filePath string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	for _, allowedExt := range extensions {
		if strings.ToLower(allowedExt) == ext {
			return true
		}
	}

	return false
}

func (b *toolBase) detectLanguageFromExtension(filePath string) (language string) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		language = "go"
	case ".js", ".mjs":
		language = "javascript"
	case ".ts":
		language = "typescript"
	case ".py":
		language = "python"
	case ".rs":
		language = "rust"
	case ".java":
		language = "java"
	case ".c":
		language = "c"
	case ".cpp", ".cc", ".cxx":
		language = "cpp"
	default:
		language = ""
	}

	return language
}
