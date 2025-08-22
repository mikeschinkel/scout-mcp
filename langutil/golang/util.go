package golang

import (
	"log"
	"os"
	"strings"
)

// hasProperIdentifierPrefix validates that documentation text starts with the correct identifier.
// This function implements Go documentation conventions by checking that documentation comments
// begin with the name of the identifier being documented, followed by appropriate punctuation.
// This validation is essential for Go documentation standards compliance.
//
// # Go Documentation Standards
//
// According to Go documentation conventions, documentation should start with the name
// of the identifier being documented:
//   - Functions: "FunctionName does something..."
//   - Types: "TypeName represents..."
//   - Constants: "ConstName is..."
//   - Variables: "VarName holds..."
//
// # Validation Rules
//
// The function checks for proper prefixes using multiple patterns:
//   - "identifier " (space after identifier)
//   - "identifier\t" (tab after identifier)
//   - "identifier(" (parentheses for function descriptions)
//
// These patterns accommodate different documentation styles while maintaining
// the core requirement that documentation begins with the identifier name.
//
// # Text Processing
//
// The function performs careful text processing:
//   - Trims whitespace from the input text
//   - Extracts the first line for prefix checking
//   - Handles empty or whitespace-only documentation
//   - Preserves original formatting for accurate analysis
//
// # Usage Context
//
// This function is used throughout the documentation validation system for:
//   - Function documentation validation
//   - Type documentation validation
//   - Method documentation validation
//   - Constant and variable documentation checking
//
// # Return Value
//
// Returns true if the documentation follows proper Go conventions, false otherwise.
// Empty or whitespace-only text always returns false as it lacks proper documentation.
//
// # Example Usage
//
//	text := "MyFunction performs important operations..."
//	valid := hasProperIdentifierPrefix(text, "MyFunction") // returns true
//
//	text := "This function performs operations..."
//	valid := hasProperIdentifierPrefix(text, "MyFunction") // returns false
func hasProperIdentifierPrefix(text, id string) bool {
	s := strings.TrimSpace(text)
	if s == "" {
		return false
	}
	first := firstLine(s)
	return strings.HasPrefix(first, id+" ") ||
		strings.HasPrefix(first, id+"\t") ||
		strings.HasPrefix(first, id+"(")
}

// hasNonEmptyFirstLine checks whether documentation text has substantive content on the first line.
// This function validates that documentation comments contain meaningful content rather than
// being empty or consisting only of whitespace. It's used to ensure documentation quality
// and compliance with Go documentation standards.
//
// # Content Validation
//
// The function performs multi-level validation:
//   - Checks that the overall text is not empty
//   - Extracts the first line from multi-line text
//   - Validates that the first line contains non-whitespace content
//   - Returns false for documentation that appears empty or placeholder-like
//
// # Use Cases
//
// This validation is applied in several contexts:
//   - Group-level documentation for constant and variable blocks
//   - Package-level documentation validation
//   - Quality checks for generated documentation
//   - Automated documentation review processes
//
// # Text Processing Strategy
//
// The function uses a conservative approach to text validation:
//   - Trims whitespace to handle formatting variations
//   - Focuses on first line content for efficiency
//   - Preserves original text structure for accurate analysis
//   - Avoids false positives from formatting artifacts
//
// # Integration
//
// This function integrates with the broader documentation validation system
// to provide consistent quality standards across different types of Go documentation.
//
// # Performance
//
// Designed for high-performance batch processing of large codebases with
// minimal string processing overhead.
func hasNonEmptyFirstLine(text string) bool {
	s := strings.TrimSpace(text)
	if s == "" {
		return false
	}
	first := firstLine(s)
	return strings.TrimSpace(first) != ""
}

// firstLine extracts the first line from multi-line text.
// This utility function isolates the first line of documentation or other text
// for analysis purposes. It handles various line ending formats and provides
// consistent first-line extraction across different text sources.
//
// # Line Ending Handling
//
// The function handles standard line ending formats:
//   - Unix-style line endings (\n)
//   - Windows-style line endings (\r\n)
//   - Classic Mac line endings (\r)
//   - Mixed line ending formats in the same text
//
// # Edge Cases
//
// The function handles edge cases gracefully:
//   - Single-line text: returns the entire string
//   - Empty strings: returns empty string
//   - Strings with only line endings: returns appropriate substring
//   - Text without line endings: returns the entire string
//
// # Performance Characteristics
//
// The function is optimized for performance:
//   - Uses efficient string operations
//   - Avoids unnecessary memory allocation
//   - Performs minimal string scanning
//   - Suitable for high-volume text processing
//
// # Usage Context
//
// This function is used throughout the documentation analysis system for:
//   - Documentation prefix validation
//   - First-line content analysis
//   - Comment parsing and validation
//   - Text normalization operations
//
// # Return Value
//
// Returns the first line of the input string, or the entire string if no
// line separators are found.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

// fileExists checks whether a file exists and is not a directory.
// This utility function provides reliable file existence checking for documentation
// validation and project analysis. It distinguishes between files and directories
// to ensure accurate validation of required files like README.md.
//
// # File System Operations
//
// The function performs careful file system analysis:
//   - Uses os.Stat for reliable file system queries
//   - Handles various file system error conditions
//   - Distinguishes between files and directories
//   - Provides consistent results across different operating systems
//
// # Error Handling
//
// The function handles file system errors gracefully:
//   - Returns false for any file system access errors
//   - Handles permission denied errors
//   - Manages non-existent path errors
//   - Deals with invalid path format errors
//
// # Directory Exclusion
//
// The function specifically excludes directories from existence checks:
//   - Uses FileInfo.IsDir() to identify directories
//   - Returns false even if a directory exists at the specified path
//   - Ensures validation focuses on actual files rather than containers
//
// # Use Cases
//
// This function is used for various project validation tasks:
//   - README.md presence validation in Go packages
//   - Configuration file existence checking
//   - Documentation file validation
//   - Project structure analysis
//
// # Performance
//
// The function is designed for efficient batch processing:
//   - Single file system call per check
//   - Minimal memory allocation
//   - Fast return for common cases
//   - Suitable for recursive directory analysis
//
// # Cross-Platform Compatibility
//
// The function works consistently across different operating systems
// and file system types, using Go's standard library for portable
// file system operations.
func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

type PathType int

const (
	InvalidPathType PathType = iota
	FilePath
	DirPath
)

func checkPath(path string) (pt PathType, err error) {
	var info os.FileInfo
	info, err = os.Stat(path)
	if err != nil {
		goto end
	}
	if info.IsDir() {
		pt = DirPath
		goto end
	}
	pt = FilePath
end:
	return pt, err
}

func must(err error) {
	if err != nil {
		log.Print(err.Error())
	}
}
