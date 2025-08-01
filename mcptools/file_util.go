package mcptools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mikeschinkel/scout-mcp/langutil"
)

func writeFile(c Config, filePath string, content string) (err error) {
	var language string

	if !c.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	// Detect language from file extension and validate if supported
	language = detectLanguageFromExtension(filePath)
	if language != "" {
		err = langutil.ValidateSyntax(language, content)
		if err != nil {
			err = fmt.Errorf("validation failed - would result in invalid %s syntax: %w", language, err)
			goto end
		}
	}

	err = os.WriteFile(filePath, []byte(content), 0644)

end:
	return err
}

func readFile(c Config, filePath string) (content string, err error) {
	var fileData []byte

	if !c.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	fileData, err = os.ReadFile(filePath)
	if err != nil {
		goto end
	}

	content = string(fileData)

end:
	return content, err
}

func detectLanguageFromExtension(filePath string) (language string) {
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
