package langutil

import (
	"path/filepath"
	"strings"
)

type Language string

const (
	UnknownLanguage     Language = "unknown"
	NoLanguage          Language = "none"
	CLanguage           Language = "c"
	CPPLanguage         Language = "cpp"
	GoLanguage          Language = "go"
	JavaLanguage        Language = "java"
	JavasScriptLanguage Language = "javascript"
	PythonLanguage      Language = "python"
	RustLanguage        Language = "rust"
	TypeScriptLanguage  Language = "typescript"
	MarkdownLanguage    Language = "markdown"
)

func DetectLanguage(filePath string) Language {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".go":
		return GoLanguage
	case ".js", ".mjs":
		return JavasScriptLanguage
	case ".ts":
		return TypeScriptLanguage
	case ".py":
		return PythonLanguage
	case ".rs":
		return RustLanguage
	case ".java":
		return JavaLanguage
	case ".c":
		return CLanguage
	case ".cpp", ".cc", ".cxx":
		return CPPLanguage
	case ".md", ".markdown":
		return MarkdownLanguage
	case ".txt":
		return NoLanguage
	default:
		// TODO add logic to read languages from Config (which should itself access environment variables as an option)
	}
	return UnknownLanguage
}
