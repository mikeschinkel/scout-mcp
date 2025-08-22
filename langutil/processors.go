package langutil

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrLanguageNotSupported is returned when attempting to use a language that has no registered processor.
	// This error indicates that the requested language is not supported by the current langutil configuration.
	// It typically occurs when trying to perform AST operations on files with unrecognized extensions or
	// when explicitly requesting a language that hasn't been registered with RegisterProcessor.
	//
	// To resolve this error, either:
	//   - Register a processor for the language using RegisterProcessor
	//   - Use a supported language instead
	//   - Check that the language constant is spelled correctly
	//
	// The error message includes details about the unsupported language and lists available alternatives.
	ErrLanguageNotSupported = errors.New("language/file type not supported")
)

// Processor represents a programming language handler that provides AST-based operations for a specific language.
// This interface defines the contract that all language processors must implement to integrate with the
// langutil package's pluggable architecture. Each processor is responsible for understanding the syntax
// and semantics of one programming language and providing operations like finding, replacing, and validating
// code constructs within that language.
//
// # Processor Architecture
//
// The Processor interface follows a plugin-style architecture where each language is handled by a dedicated
// implementation. This design allows for:
//   - Language-specific optimizations and features
//   - Independent development of language support
//   - Easy addition of new languages without modifying core langutil code
//   - Consistent API across all supported languages
//
// # Implementation Requirements
//
// Processor implementations must:
//   - Handle their language's complete syntax through proper AST parsing
//   - Provide accurate position information for found constructs
//   - Validate replacement content before performing substitutions
//   - Return consistent error messages for unsupported operations
//   - Be thread-safe for concurrent usage
//
// # Lifecycle
//
// Processors are typically registered during package initialization using RegisterProcessor and remain
// available for the lifetime of the application. They should be stateless and reusable across multiple
// operations.
//
// # Error Handling
//
// Processors should return descriptive errors that include:
//   - File paths and line numbers for syntax errors
//   - Context about what was being searched for when operations fail
//   - Suggestions for correct usage when validation fails
//   - Language-specific error details when appropriate
type Processor interface {
	// Language returns the language name that this processor handles.
	// The returned Language constant is used as the key for processor registration
	// and lookup operations. It should match one of the predefined Language constants
	// (e.g., GoLanguage, JavasScriptLanguage, PythonLanguage).
	//
	// This method should always return the same value for a given processor instance
	// and should never return an empty string or UnknownLanguage.
	Language() Language

	// SupportedPartTypes returns the part types this language supports.
	// This method provides introspection capabilities for the processor, allowing
	// callers to determine what kinds of code constructs can be found or manipulated.
	// The returned slice should include all PartType values that the processor can
	// handle in its FindPart and ReplacePart methods.
	//
	// Common part types include "func", "type", "const", "var", "import", "package",
	// though the exact set depends on the language's features and the processor's
	// implementation capabilities.
	//
	// The returned slice should not be modified by callers and may be cached by
	// the processor for performance.
	SupportedPartTypes() []PartType

	// FindPart finds a specific part in the source code.
	// This method performs AST-based searching to locate a code construct matching
	// the specified type and name within the provided source content. It returns
	// detailed position and content information about the found construct.
	//
	// The method should:
	//   - Parse the source content into an AST
	//   - Search for constructs matching args.PartType and args.PartName
	//   - Return accurate position information including line numbers and byte offsets
	//   - Include the complete construct content in the result
	//   - Handle language-specific scoping and naming rules
	//
	// Returns a PartInfo with Found=false if the construct is not found.
	// Returns an error for syntax errors, unsupported part types, or other failures.
	FindPart(PartArgs) (*PartInfo, error)

	// ReplacePart replaces a specific part in the source code.
	// This method performs a two-phase operation: first finding the target construct
	// using the same logic as FindPart, then replacing it with the provided new content.
	// The method validates both the search operation and the replacement content before
	// performing the substitution.
	//
	// The replacement process:
	//   - Finds the target construct using args.PartType and args.PartName
	//   - Validates that args.NewContent is appropriate for the part type
	//   - Performs the text replacement at the exact AST boundaries
	//   - Validates that the resulting source code is syntactically correct
	//   - Returns the complete modified source code
	//
	// Returns an error if the construct is not found, the replacement content is invalid,
	// or the resulting code would be syntactically incorrect.
	ReplacePart(PartArgs) (string, error)

	// ValidateContent validates that content is appropriate for the part type.
	// This method checks that replacement content conforms to the language's syntax
	// requirements for the specified part type. It performs static analysis without
	// requiring the content to be inserted into a complete source file.
	//
	// Validation includes:
	//   - Syntax correctness for the part type (e.g., "func" content should start with "func")
	//   - Basic structural requirements (e.g., functions should have proper signatures)
	//   - Language-specific constraints and conventions
	//
	// This method is called internally by ReplacePart but is also available for
	// standalone validation of replacement content before performing replacements.
	ValidateContent(PartArgs) error

	// ValidateSyntax validates that the entire source code is syntactically correct.
	// This method performs complete syntax validation of the provided source code
	// without requiring any specific constructs to be present. It's used to verify
	// that source files are well-formed and can be successfully parsed.
	//
	// The validation:
	//   - Parses the complete source into an AST
	//   - Reports syntax errors with line and column information
	//   - Handles language-specific syntax rules and requirements
	//   - Does not perform semantic analysis or type checking
	//
	// This method is used by file validation operations and after replacement
	// operations to ensure the resulting code remains syntactically valid.
	ValidateSyntax(source string) error
}

// processors holds the registry of registered language handlers.
// This map stores all available processors indexed by their language name in lowercase.
// The registry is used by GetProcessor to look up the appropriate processor for a given language.
//
// The map is populated by calls to RegisterProcessor during package initialization.
// Access to this map is not explicitly synchronized, as registration typically occurs
// during startup before concurrent access begins.
var processors = make(map[string]Processor)

// RegisterProcessor registers a language handler in the global processor registry.
// This function adds a new processor to the available set of language handlers,
// making it available for use by GetProcessor and related functions. Registration
// is typically performed during package initialization by language-specific processors.
//
// The processor's Language() method is called to determine the registration key,
// which is normalized to lowercase for case-insensitive lookup. If a processor
// for the same language is already registered, it will be replaced by the new processor.
//
// # Registration Pattern
//
// Language processors typically register themselves during package initialization:
//
//	package golang
//
//	func init() {
//		langutil.RegisterProcessor(&GoProcessor{})
//	}
//
// This ensures that processors are available as soon as their packages are imported.
//
// # Thread Safety
//
// This function is not thread-safe and should only be called during package initialization
// before any concurrent access to the processor registry begins. Calling this function
// after concurrent operations have started may result in race conditions.
func RegisterProcessor(p Processor) {
	name := strings.ToLower(string(p.Language()))
	processors[name] = p
}

// GetProcessor retrieves a registered language handler for the specified language.
// This function performs a case-insensitive lookup in the processor registry to find
// the appropriate processor for the given language. It's the primary mechanism for
// accessing language-specific functionality within the langutil package.
//
// The lookup process:
//   - Normalizes the language name to lowercase for consistent matching
//   - Searches the processor registry for a matching processor
//   - Returns the processor if found, or an error with details if not found
//
// # Error Information
//
// If no processor is found, the function returns ErrLanguageNotSupported with additional
// context including:
//   - The requested language name that could not be found
//   - A list of all available/registered languages
//   - Suggestions for resolving the issue
//
// This detailed error information helps with debugging and provides users with
// actionable information about available alternatives.
//
// # Performance
//
// This function performs a simple map lookup and is very fast. Processors can be
// retrieved repeatedly without performance concerns, and callers may choose to
// cache the returned processor for multiple operations.
//
// # Example Usage
//
//	processor, err := langutil.GetProcessor(langutil.GoLanguage)
//	if err != nil {
//		log.Fatal(err)
//	}
//	partInfo, err := processor.FindPart(args)
func GetProcessor(name Language) (p Processor, err error) {
	p, exists := processors[strings.ToLower(string(name))]
	if !exists {
		err = errors.Join(ErrLanguageNotSupported,
			fmt.Errorf("language=%s", name),
			fmt.Errorf("available=[%v]", GetLanguages()),
		)
	}
	return p, err
}

// GetLanguages returns a list of all registered processors' languages.
// This function provides introspection capabilities for the processor registry,
// allowing callers to discover what languages are currently supported. It's useful
// for building user interfaces, validating input, and debugging registration issues.
//
// # Return Value
//
// The returned slice contains the Language constant for each registered processor.
// The order of languages in the slice is not guaranteed and may change between calls
// if processors are registered or unregistered.
//
// The returned slice is a new slice created for each call, so it can be safely
// modified by callers without affecting the internal registry state.
//
// # Performance
//
// This function iterates through all registered processors and creates a new slice
// for each call. For performance-critical code that needs to check language support
// frequently, consider caching the result or using GetProcessor directly with
// error handling.
//
// # Example Usage
//
//	languages := langutil.GetLanguages()
//	fmt.Printf("Supported languages: %v\n", languages)
//
//	for _, lang := range languages {
//		fmt.Printf("Language %s is supported\n", lang)
//	}
func GetLanguages() (langs []Language) {
	langs = make([]Language, 0, len(processors))
	for _, p := range processors {
		langs = append(langs, p.Language())
	}
	return langs
}
