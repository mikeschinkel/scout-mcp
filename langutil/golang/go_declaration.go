package golang

import (
	"context"
	"go/ast"
	"go/token"
)

// GoDeclaration represents a single declaration within a Go source file with documentation analysis capabilities.
// This structure wraps Go's ast.Decl interface with additional functionality for analyzing documentation
// compliance and coding standards. It provides methods for examining declarations and identifying
// documentation violations according to Go community standards.
//
// # Declaration Types
//
// GoDeclaration handles all types of Go declarations:
//   - Function declarations (ast.FuncDecl) - standalone functions and methods
//   - General declarations (ast.GenDecl) - types, constants, variables, and imports
//   - Bad declarations (ast.BadDecl) - malformed declarations for error analysis
//
// # Documentation Analysis
//
// The structure provides comprehensive documentation analysis including:
//   - Function documentation validation with proper naming conventions
//   - Type documentation following Go standards
//   - Constant and variable documentation formatting checks
//   - Group vs. individual declaration documentation requirements
//
// # File Context
//
// Each GoDeclaration maintains a reference to its containing GoFile, providing access to:
//   - File-level information for error reporting
//   - Token position resolution for accurate line numbers
//   - Comment analysis for documentation validation
//   - Package context for qualified name resolution
//
// # Integration
//
// GoDeclaration integrates with the broader documentation validation system by:
//   - Generating DocException instances for violations
//   - Providing detailed position information for issues
//   - Supporting batch analysis across multiple declarations
//   - Enabling fine-grained documentation compliance reporting
type GoDeclaration struct {
	File     *GoFile // Reference to the containing Go source file for context and position resolution
	ast.Decl         // Embedded Go AST declaration interface providing access to declaration details
}

// NewGoDeclaration creates a new GoDeclaration instance wrapping the provided AST declaration.
// This constructor function establishes the bidirectional relationship between the declaration
// and its containing file, enabling context-aware analysis and accurate position reporting.
//
// # Parameter Relationships
//
// The function establishes important relationships:
//   - decl: The AST declaration to be wrapped and analyzed
//   - f: The containing file providing context for analysis
//
// # Usage Context
//
// This constructor is typically called during file parsing and analysis:
//   - When processing declarations during AST traversal
//   - During file-level documentation analysis
//   - When building declaration collections for batch processing
//
// # Return Value
//
// Returns a fully initialized GoDeclaration ready for documentation analysis
// and violation detection operations.
func NewGoDeclaration(decl ast.Decl, f *GoFile) GoDeclaration {
	return GoDeclaration{
		File: f,
		Decl: decl,
	}
}

// Filename returns the file path containing this declaration.
// This method provides convenient access to the file path for error reporting,
// logging, and violation tracking. The returned path matches the path used
// during file loading and analysis.
//
// # Usage
//
// Commonly used for:
//   - Error message generation with file context
//   - Violation reporting systems
//   - IDE integration for problem highlighting
//   - Build system integration for issue tracking
//
// # Path Format
//
// The returned path maintains the same format as provided during file analysis,
// which may be absolute or relative depending on how the analysis was initiated.
func (decl *GoDeclaration) Filename() string {
	return decl.File.Name()
}

// Exceptions analyzes the declaration for documentation violations and returns a list of issues.
// This method performs comprehensive documentation analysis specific to the declaration type,
// applying Go community standards and coding conventions to identify areas where documentation
// is missing, incomplete, or improperly formatted.
//
// # Analysis Scope
//
// The method performs different types of analysis based on declaration type:
//
// Function Declarations (ast.FuncDecl):
//   - Validates presence of function documentation
//   - Checks that documentation starts with the function name
//   - Applies to both standalone functions and methods
//
// Type Declarations (ast.GenDecl with token.TYPE):
//   - Validates presence of type documentation
//   - Checks documentation formatting conventions
//   - Applies to structs, interfaces, and type aliases
//
// Constant/Variable Declarations (ast.GenDecl with token.CONST/VAR):
//   - Validates documentation for grouped vs. individual declarations
//   - Checks for appropriate comment placement (leading vs. end-of-line)
//   - Handles multi-name declarations appropriately
//
// # Context Awareness
//
// The analysis is context-aware and considers:
//   - File-level documentation requirements
//   - Package-level conventions and patterns
//   - Go community standards and best practices
//   - Specific formatting requirements for different declaration types
//
// # Return Value
//
// Returns a slice of DocException instances, each representing a specific
// documentation violation with detailed position and type information.
// An empty slice indicates no violations were found.
//
// # Performance
//
// The method is designed for efficient batch processing and can be called
// repeatedly across large codebases without performance concerns.
func (decl *GoDeclaration) Exceptions(ctx context.Context) (exceptions []DocException) {
	f := decl.File
	switch d := decl.Decl.(type) {
	case *ast.FuncDecl:
		missing := f.FuncException(d)
		if missing != nil {
			exceptions = append(exceptions, *missing)
		}
	case *ast.GenDecl:
		switch d.Tok {
		case token.TYPE:
			exceptions = f.TypeExceptions(d)
		case token.CONST, token.VAR:
			exceptions = f.ConstVarExceptions(d)
		default:
			// Not other tokens matter here
		}
	}
	return exceptions
}
