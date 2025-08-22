package langutil

// PartInfo represents information about a found language construct in a source file.
// This structure contains complete location and content information for a code construct
// that was located through AST parsing. It provides both line-based and byte-based
// positioning information to support different use cases for code manipulation.
//
// PartInfo is returned by AST search operations and contains everything needed to
// either extract the found construct or replace it with new content. The dual
// addressing scheme (line numbers and byte offsets) allows for efficient editing
// operations while also providing human-readable position information.
//
// # Position Information
//
// Line numbers are 1-based to match standard editor conventions, while byte offsets
// are 0-based for direct use with string slicing operations. The Start* and End*
// fields define an inclusive range that encompasses the entire construct including
// any associated documentation comments or annotations.
//
// # Content Extraction
//
// The Content field contains the exact text of the found construct as it appears
// in the source file, including formatting and comments. This can be used for
// analysis, display, or as a baseline for generating replacement content.
type PartInfo struct {
	StartLine   int    `json:"start_line"`   // Line number where the construct starts (1-based, inclusive)
	EndLine     int    `json:"end_line"`     // Line number where the construct ends (1-based, inclusive)
	StartOffset int    `json:"start_offset"` // Byte offset where the construct starts (0-based, inclusive)
	EndOffset   int    `json:"end_offset"`   // Byte offset where the construct ends (0-based, exclusive)
	Content     string `json:"content"`      // The actual content of the construct including formatting
	Found       bool   `json:"found"`        // Whether the construct was found (false indicates search failure)
}

// PartArgs contains arguments for finding or replacing language constructs.
// This structure encapsulates all the information needed to perform AST operations
// on source code, including the target language, source content, and operation parameters.
//
// PartArgs is used by both search and replacement operations, with some fields being
// optional depending on the specific operation being performed.
//
// # Field Usage by Operation
//
// For FindPart operations:
//   - Language, Content, PartType, PartName, and Filepath are required
//   - NewContent is ignored
//
// For ReplacePart operations:
//   - All fields are required
//   - NewContent specifies the replacement text
//
// # Content Requirements
//
// The Content field should contain the complete source file content, not just
// a fragment. AST parsers typically require complete, syntactically valid source
// code to properly identify construct boundaries and relationships.
//
// The NewContent field (for replacement operations) should contain valid source
// code that can be substituted for the found construct without breaking syntax.
// The replacement content is validated before the replacement is performed.
type PartArgs struct {
	Language   Language // Programming language of the source code (required for processor selection)
	Content    string   // Source code content to search in (must be complete file content)
	PartType   PartType // Type of construct to find (function, type, const, var, etc.)
	PartName   string   // Name of the specific construct to find (identifier name)
	NewContent string   // New content to replace with (for replacement operations only)
	Filepath   string   // Path to the source file (for error reporting and context)
}

// FindPart finds a language construct in source code using the appropriate language processor.
// This function performs AST-based searching to locate a specific code construct within
// source file content. It delegates to the registered processor for the specified language
// to handle language-specific parsing and searching logic.
//
// The search is performed by parsing the entire source content into an AST and then
// traversing the tree to find constructs matching the specified type and name. This
// approach provides precise location information and handles complex language features
// like nested scopes, overloaded names, and language-specific syntax.
//
// # Search Accuracy
//
// AST-based searching is more accurate than text-based searching because it understands
// language semantics. For example, it can distinguish between a function named "Test"
// and a variable named "Test", or find the correct method when multiple types have
// methods with the same name.
//
// # Performance Considerations
//
// AST parsing can be expensive for large files. If you need to find multiple constructs
// in the same file, consider using the processor directly to reuse the parsed AST.
//
// # Error Handling
//
// Returns ErrLanguageNotSupported if no processor is registered for the specified language.
// Returns syntax errors if the source content cannot be parsed.
// Returns domain-specific errors if the construct type is not supported by the language.
//
// # Example Usage
//
//	args := PartArgs{
//		Language: GoLanguage,
//		Content:  sourceCode,
//		PartType: "func",
//		PartName: "MyFunction",
//		Filepath: "example.go",
//	}
//	partInfo, err := FindPart(args)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if partInfo.Found {
//		fmt.Printf("Found function at lines %d-%d\n", partInfo.StartLine, partInfo.EndLine)
//	}
func FindPart(args PartArgs) (*PartInfo, error) {
	p, err := GetProcessor(args.Language)
	if err != nil {
		return nil, err
	}

	return p.FindPart(args)
}
