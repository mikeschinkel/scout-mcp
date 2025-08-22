package langutil

// Compile-time verification that NoProcessor implements the Processor interface.
// This ensures that any changes to the Processor interface will cause a compilation
// error if NoProcessor doesn't implement all required methods, providing early
// detection of interface compatibility issues.
var _ Processor = (*NoProcessor)(nil)

// NoProcessor implements the Processor interface for files with no programming language structure.
// This processor handles plain text files, documentation files, and other non-code content
// that doesn't require language-specific parsing or AST operations. It provides minimal
// processing capabilities while maintaining compatibility with the langutil processing framework.
//
// # Use Cases
//
// NoProcessor is used for:
//   - Plain text files (.txt)
//   - Documentation files that don't have code structure
//   - Configuration files without programming language syntax
//   - Binary files that need basic validation
//   - Unknown file types that should be treated as plain text
//
// # Limitations
//
// As a minimal processor, NoProcessor has significant limitations:
//   - No AST parsing capabilities (FindPart and ReplacePart will panic)
//   - No support for language constructs like functions, types, or variables
//   - No syntax validation beyond basic text checks
//   - No support for language-specific features or optimizations
//
// # Design Philosophy
//
// NoProcessor follows the "do no harm" principle for non-code files. It accepts all
// content as valid and provides safe defaults for all operations. This ensures that
// the langutil framework can handle mixed file types in projects without failing
// on non-code files.
//
// # Registration
//
// NoProcessor registers itself automatically during package initialization, making it
// available for NoLanguage files without requiring explicit setup by applications.
type NoProcessor struct{}

// init registers the NoProcessor with the global processor registry during package initialization.
// This ensures that NoLanguage files can be processed without requiring explicit processor
// registration by applications. The registration happens automatically when the langutil
// package is imported.
func init() {
	RegisterProcessor(&NoProcessor{})
}

// Language returns NoLanguage to indicate this processor handles non-programming language files.
// This method implements the Processor interface requirement and provides the language
// identifier used for processor registration and lookup operations.
//
// The NoLanguage constant is used to identify files that contain plain text or other
// non-code content that doesn't require language-specific processing capabilities.
func (n NoProcessor) Language() Language {
	return NoLanguage
}

// SupportedPartTypes returns an empty slice indicating no part types are supported.
// Since plain text files don't have programming language constructs like functions
// or types, NoProcessor cannot support any AST-based part operations.
//
// This method returns an empty slice to indicate that operations like FindPart and
// ReplacePart are not available for NoLanguage files. Callers should check the
// supported part types before attempting AST operations.
//
// # Design Rationale
//
// Returning an empty slice rather than nil provides clearer semantics and allows
// callers to safely iterate over the result without nil checks. It also maintains
// consistency with other processors that may have varying numbers of supported types.
func (n NoProcessor) SupportedPartTypes() []PartType {
	return make([]PartType, 0)
}

// FindPart panics because plain text files don't have language constructs to find.
// This method implements the Processor interface but is not functional for NoProcessor
// since plain text content doesn't have parseable language constructs like functions,
// types, or variables.
//
// # Panic Behavior
//
// The method panics with a clear message indicating that the operation is not supported.
// This provides immediate feedback to developers who accidentally attempt AST operations
// on plain text files, rather than returning confusing error messages.
//
// # Alternative Approaches
//
// Callers should check SupportedPartTypes() before calling this method to avoid panics.
// For text-based searching in plain text files, consider using regular expressions
// or string search functions instead of AST-based operations.
//
// # Future Enhancements
//
// Future versions might support basic text searching capabilities for plain text files,
// though this would require extending the PartType system to include text-based constructs.
func (n NoProcessor) FindPart(PartArgs) (*PartInfo, error) {
	panic("FindPart operation not supported for plain text files - use SupportedPartTypes() to check availability")
	return nil, nil
}

// ReplacePart panics because plain text files don't have language constructs to replace.
// This method implements the Processor interface but is not functional for NoProcessor
// since plain text content doesn't have parseable language constructs that can be
// replaced using AST-based operations.
//
// # Panic Behavior
//
// The method panics with a clear message indicating that the operation is not supported.
// This provides immediate feedback to developers who accidentally attempt AST replacement
// operations on plain text files.
//
// # Alternative Approaches
//
// For text replacement in plain text files, consider using:
//   - strings.Replace() for simple text substitution
//   - regexp.ReplaceAllString() for pattern-based replacement
//   - Text editing libraries for more complex transformations
//
// These approaches are more appropriate for plain text content than AST-based operations.
func (n NoProcessor) ReplacePart(PartArgs) (string, error) {
	panic("ReplacePart operation not supported for plain text files - use SupportedPartTypes() to check availability")
	return "", nil
}

// ValidateContent always returns nil since plain text has no specific content requirements.
// This method implements the Processor interface and provides a permissive validation
// policy for plain text content. Since plain text files don't have syntax rules or
// structural requirements, all content is considered valid.
//
// # Validation Philosophy
//
// NoProcessor follows a "accept everything" approach for content validation. This ensures
// that operations involving plain text files don't fail due to content restrictions,
// maintaining compatibility with the broader langutil processing framework.
//
// # Use Cases
//
// This permissive validation is appropriate for:
//   - User-generated text content
//   - Configuration files with flexible formats
//   - Documentation files with arbitrary content
//   - Binary data that passes through as text
//
// # Future Enhancements
//
// Future versions might add optional validation for specific plain text formats
// like CSV, TSV, or structured text, while maintaining backward compatibility
// with the current permissive approach.
func (n NoProcessor) ValidateContent(PartArgs) error {
	return nil
}

// ValidateSyntax always returns nil since plain text has no syntax to validate.
// This method implements the Processor interface and provides unconditional acceptance
// of all text content. Plain text files, by definition, don't have syntax rules that
// can be violated, so all content is considered syntactically valid.
//
// # Validation Scope
//
// The method accepts any string content without validation, including:
//   - Empty strings
//   - Binary data represented as text
//   - Malformed data from other file types
//   - Multi-line content with any line ending format
//   - Content with any character encoding (within Go string limitations)
//
// # Performance
//
// This method is extremely fast since it performs no actual validation work.
// It simply returns nil immediately, making it suitable for high-volume processing
// of plain text files without performance concerns.
//
// # Integration
//
// This permissive validation integrates well with file validation workflows that
// process mixed file types, ensuring that plain text files don't cause validation
// failures in projects that contain both code and documentation files.
func (n NoProcessor) ValidateSyntax(string) error {
	return nil
}
