package mcputil

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/scout-mcp/langutil"
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

// ReadFile reads files and returns a string
// TODO Does the functionality of this really carry its own weight?  Seems like it does not.
func (b *ToolBase) ReadFile(path string) (content string, err error) {
	var data []byte

	data, err = os.ReadFile(path)
	if err != nil {
		goto end
	}

	content = string(data)

end:
	return content, err
}

// WriteFile writes to files with a default 0644 permissions
// TODO Does the functionality of this really carry its own weight?  Seems like it does not.
func (b *ToolBase) WriteFile(path, content string) (err error) {
	err = os.WriteFile(path, []byte(content), 0644)
	return err
}

func (b *ToolBase) writeFileWithValidation(path, content string) (err error) {
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
	err = b.WriteFile(path, content)

end:
	return err
}

func (b *ToolBase) SearchForFiles(path string, recursive bool, extensions []string) (files []string, err error) {
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

func (b *ToolBase) walkDirectory(path string, extensions []string, files *[]string) (err error) {
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

func (b *ToolBase) listDirectory(path string, extensions []string, files *[]string) (err error) {
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

func (b *ToolBase) matchesExtensions(filePath string, extensions []string) bool {
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

func (b *ToolBase) detectLanguageFromExtension(filePath string) (language string) {
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
