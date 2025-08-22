package langutil

import (
	"path/filepath"
	"strings"
)

// Language represents a programming language type.
// This type is used throughout the langutil package to identify which language
// processor should handle a particular file or operation. Languages are
// represented as string constants to allow for easy extension and comparison.
//
// The language detection system uses file extensions as the primary mechanism
// for determining the appropriate language, though explicit language specification
// is also supported for cases where the extension may be ambiguous or non-standard.
type Language string

// Language constants define all supported programming languages.
// These constants should be used when specifying languages explicitly
// rather than relying on automatic detection.
const (
	// UnknownLanguage indicates that the language could not be determined automatically.
	// This is returned when file extension detection fails and no explicit language
	// is provided. Files with unknown languages cannot be processed by language-specific
	// processors and will result in errors if AST operations are attempted.
	UnknownLanguage Language = "unknown"

	// NoLanguage indicates that the file has no programming language structure.
	// This is used for plain text files, documentation, and other non-code files.
	// The NoProcessor handles files of this type with minimal processing capabilities.
	NoLanguage Language = "none"

	// CLanguage represents the C programming language.
	// Files with .c and .h extensions are typically detected as C language.
	// Full AST support for C is not yet implemented.
	CLanguage Language = "c"

	// CPPLanguage represents the C++ programming language.
	// Files with .cpp, .cc, .cxx, and .hpp extensions are detected as C++.
	// Full AST support for C++ is not yet implemented.
	CPPLanguage Language = "cpp"

	// GoLanguage represents the Go programming language.
	// Files with .go extensions are detected as Go language.
	// This language has full AST support including functions, types, constants,
	// variables, imports, and package declarations.
	GoLanguage Language = "go"

	// JavaLanguage represents the Java programming language.
	// Files with .java extensions are detected as Java language.
	// Full AST support for Java is not yet implemented.
	JavaLanguage Language = "java"

	// JavasScriptLanguage represents the JavaScript programming language.
	// Files with .js and .mjs extensions are detected as JavaScript.
	// Full AST support for JavaScript is not yet implemented.
	JavasScriptLanguage Language = "javascript"

	// PythonLanguage represents the Python programming language.
	// Files with .py extensions are detected as Python language.
	// Full AST support for Python is not yet implemented.
	PythonLanguage Language = "python"

	// RustLanguage represents the Rust programming language.
	// Files with .rs extensions are detected as Rust language.
	// Full AST support for Rust is not yet implemented.
	RustLanguage Language = "rust"

	// TypeScriptLanguage represents the TypeScript programming language.
	// Files with .ts extensions are detected as TypeScript.
	// Full AST support for TypeScript is not yet implemented.
	TypeScriptLanguage Language = "typescript"

	// MarkdownLanguage represents Markdown markup language.
	// Files with .md and .markdown extensions are detected as Markdown.
	// This is primarily used for documentation files and has minimal processing support.
	MarkdownLanguage Language = "markdown"
)

// DetectLanguage determines the programming language of a file based on its extension.
// This function performs automatic language detection by examining the file extension
// and mapping it to the appropriate Language constant. The detection is case-insensitive
// and handles multiple extensions that may map to the same language.
//
// The function uses a comprehensive mapping of file extensions to languages, covering
// common file types for each supported language. For languages with multiple valid
// extensions (like C++ with .cpp, .cc, .cxx), all variants are recognized.
//
// If the file extension is not recognized, the function returns UnknownLanguage.
// Callers should handle this case appropriately, either by prompting for explicit
// language specification or by treating the file as plain text.
//
// # Extension Mapping
//
// The following extensions are currently supported:
//   - .go → GoLanguage
//   - .js, .mjs → JavasScriptLanguage
//   - .ts → TypeScriptLanguage
//   - .py → PythonLanguage
//   - .rs → RustLanguage
//   - .java → JavaLanguage
//   - .c → CLanguage
//   - .cpp, .cc, .cxx → CPPLanguage
//   - .md, .markdown → MarkdownLanguage
//   - .txt → NoLanguage
//
// # Future Extensions
//
// Additional language mappings can be loaded from configuration files or environment
// variables in future versions. The TODO comment indicates this feature is planned
// but not yet implemented.
//
// # Performance
//
// This function performs a simple switch statement lookup and is very fast.
// It can be called repeatedly without performance concerns.
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
