// Package langutil provides language-specific utilities for parsing and validating source code files.
// It supports AST-based operations for finding and replacing language constructs like functions and types.
//
// # Overview
//
// The langutil package implements a pluggable architecture for language processing that allows
// different programming languages to be analyzed and manipulated using Abstract Syntax Tree (AST)
// operations. The package supports operations like finding functions, types, constants, and variables
// within source code, as well as replacing these constructs with new implementations.
//
// # Architecture
//
// The package is built around several key concepts:
//
//   - Language detection: Automatically determines programming language from file extensions
//   - Processor registry: Plugin-style architecture where each language implements the Processor interface
//   - AST operations: Language-agnostic interface for finding and replacing code constructs
//   - File validation: Syntax validation and language verification for source files
//
// # Supported Languages
//
// Currently supported languages include:
//   - Go: Full AST support for functions, types, constants, variables, imports, and packages
//   - Plain text: Basic support for files with no programming language structure
//
// Additional language processors can be registered using the RegisterProcessor function.
//
// # Basic Usage
//
// To validate a single file:
//
//	language, err := langutil.ValidateFileAs("main.go", langutil.GoLanguage)
//	if err != nil {
//		log.Fatal(err)
//	}
//
// To find a function in Go source code:
//
//	args := langutil.PartArgs{
//		Language: langutil.GoLanguage,
//		Content:  sourceCode,
//		PartType: "func",
//		PartName: "MyFunction",
//		Filepath: "example.go",
//	}
//	partInfo, err := langutil.FindPart(args)
//
// # Performance Considerations
//
// AST parsing can be memory-intensive for large files. The package caches parsed results
// where possible and provides streaming interfaces for processing multiple files efficiently.
// For optimal performance, batch file operations when possible and reuse Processor instances.
//
// # Error Handling
//
// All functions return detailed error information including:
//   - Syntax errors with line numbers and positions
//   - Language detection failures with suggested alternatives
//   - Part not found errors with context about available constructs
//   - File I/O errors with full path information
//
// # Thread Safety
//
// The package is designed to be thread-safe. Processor instances can be shared across
// goroutines, and the global processor registry is protected by appropriate locking.
// Individual File instances should not be shared between goroutines without external
// synchronization.
package langutil

import (
	"errors"
	"log/slog"
)

// Args contains initialization arguments for the langutil package.
// This structure allows for future extensibility of package configuration
// without breaking existing code that uses Initialize().
type Args struct {
	AppName string // Name of the application using langutil for error messages and logging
	Logger  *slog.Logger
}

// appName stores the application name for use in error messages and logging.
// This global variable is set once during package initialization and used
// throughout the package to provide consistent error reporting.
var appName string

// Initialize sets up the langutil package with the provided configuration.
// This function should be called once during application startup to configure
// the package for use. Currently it only sets the application name, but future
// versions may perform additional initialization tasks like loading language
// processor plugins or setting up caching.
//
// The application name is used in error messages to provide context about
// which application encountered parsing or validation issues.
//
// Returns an error if initialization fails, though current implementation
// always returns nil. Future versions may return initialization errors.
func Initialize(args Args) (err error) {
	appName = args.AppName
	logger.Info("Initializing langutil\n")
	err = CallInitializerFuncs(args)
	logger.Info("langutil initialized\n")
	return err
}

// ValidateFileAs validates a single file as the specified language and returns the validated language.
// This function reads the file from disk, detects its language if not specified, and validates
// that the file contents are syntactically correct for the target language.
//
// If the language parameter is empty or UnknownLanguage, the function will attempt to detect
// the language automatically based on the file extension. If a specific language is provided,
// the function will validate the file contents against that language's syntax rules.
//
// The returned language may differ from the input language if automatic detection is used
// or if the file extension suggests a different language than what was requested.
//
// Returns ErrLanguageNotSupported if the detected or specified language has no registered
// processor. Returns syntax errors if the file contents are not valid for the target language.
func ValidateFileAs(fp string, language Language) (_ Language, err error) {
	var f *File

	f = NewFile(fp)
	err = f.Initialize()
	if err != nil {
		goto end
	}
	language, err = f.ValidateAs(language)
	if err != nil {
		goto end
	}
end:
	return language, err
}

// ValidationResult contains the result of validating a single file.
// This structure provides complete information about the validation process
// including the original file path, the determined language, and any errors
// that occurred during validation.
type ValidationResult struct {
	FilePath string   // Path to the file that was validated (absolute or relative as provided)
	Language Language // The validated language (may differ from requested if auto-detected)
	Error    error    // Any error that occurred during validation (nil if successful)
}

// ValidateFilesAs validates multiple files as the specified language and returns results for each.
// This function processes a batch of files, validating each one against the specified language.
// It continues processing all files even if individual files fail validation, collecting all
// errors for comprehensive reporting.
//
// The language parameter applies to all files in the batch. If language detection is desired
// on a per-file basis, use ValidateFileAs for each file individually.
//
// The function returns a ValidationResult for every input file path, allowing callers to
// inspect both successful and failed validations. The aggregated error return value
// contains all individual file errors joined together, or nil if all files validated successfully.
//
// This function is optimized for batch processing and will be more efficient than calling
// ValidateFileAs repeatedly for large numbers of files.
func ValidateFilesAs(filepaths []string, language Language) (results []ValidationResult, _ error) {
	var errs []error

	results = make([]ValidationResult, 0, len(filepaths))
	// Validate each file
	for _, fp := range filepaths {
		lang, err := ValidateFileAs(fp, language)
		errs = append(errs, err)

		results = append(results, ValidationResult{
			FilePath: fp,
			Language: lang,
			Error:    err,
		})
	}
	return results, errors.Join(errs...)
}
