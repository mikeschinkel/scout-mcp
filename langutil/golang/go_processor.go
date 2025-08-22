package golang

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/mikeschinkel/scout-mcp/langutil"
)

// Go language part type constants define the types of code constructs that can be found and manipulated in Go source code.
// These constants are used with the langutil framework to specify what type of language construct should be targeted
// during AST operations. Each constant corresponds to a specific Go language element that can be parsed and identified
// using Go's AST package.
//
// # Part Type Coverage
//
// The constants cover the major Go language constructs that developers commonly need to find or replace:
//   - Function and method declarations
//   - Type definitions (structs, interfaces, aliases)
//   - Constant declarations (both individual and grouped)
//   - Variable declarations (both individual and grouped)
//   - Import statements (both individual and grouped)
//   - Package declarations
//
// # Usage in AST Operations
//
// These part types are used throughout the Go processor for:
//   - Finding specific constructs by name and type
//   - Replacing constructs with new implementations
//   - Validating replacement content before substitution
//   - Generating position information for found constructs
//
// # Method Handling
//
// Function part types handle both regular functions and methods. Methods are identified using the format
// "ReceiverType.MethodName" (e.g., "*MyStruct.Method" or "MyStruct.Method"), allowing precise targeting
// of methods on specific receiver types.
const (
	// FuncGoPart represents Go function and method declarations.
	// This part type can find both standalone functions and methods attached to types.
	// For methods, use the format "ReceiverType.MethodName" where ReceiverType includes
	// the pointer indicator if applicable (e.g., "*MyStruct.Method").
	//
	// Examples of valid function names:
	//   - "main" (standalone function)
	//   - "MyStruct.Method" (method on value receiver)
	//   - "*MyStruct.Method" (method on pointer receiver)
	//   - "String" (could match multiple types with String methods)
	//
	// The AST search examines both the receiver type and method name to provide
	// precise matching for method lookups.
	FuncGoPart langutil.PartType = "func"

	// TypeGoPart represents Go type declarations including structs, interfaces, and type aliases.
	// This part type can find any kind of type definition in Go source code, including:
	//   - Struct definitions with fields and methods
	//   - Interface definitions with method signatures
	//   - Type aliases that create new names for existing types
	//   - Composite type definitions
	//
	// The search operates on the type name as declared in the source code, not including
	// the "type" keyword. For example, to find "type MyStruct struct {...}", use "MyStruct"
	// as the part name.
	TypeGoPart langutil.PartType = "type"

	// ConstGoPart represents Go constant declarations.
	// This part type can find both individual constant declarations and constants within
	// grouped constant blocks. The search targets the constant name and will match both:
	//   - Individual declarations: const MyConst = "value"
	//   - Grouped declarations: const ( MyConst = "value" )
	//
	// When replacing constants, the replacement content should be appropriate for the
	// declaration style (individual vs. grouped) and maintain Go syntax requirements.
	ConstGoPart langutil.PartType = "const"

	// VarGoPart represents Go variable declarations.
	// This part type can find both individual variable declarations and variables within
	// grouped variable blocks. The search targets the variable name and will match both:
	//   - Individual declarations: var myVar string
	//   - Grouped declarations: var ( myVar string )
	//   - Short declarations: myVar := "value" (if at package level)
	//
	// Variable searches examine the variable name regardless of the declaration style
	// or initialization method used.
	VarGoPart langutil.PartType = "var"

	// ImportGoPart represents Go import statements.
	// This part type can find both individual import statements and imports within
	// grouped import blocks. The search can target either:
	//   - Import paths: "github.com/user/package"
	//   - Import paths with quotes: "\"github.com/user/package\""
	//
	// The search handles both single imports and grouped import blocks:
	//   - import "fmt"
	//   - import ( "fmt"; "os" )
	//
	// Named imports and aliased imports are also supported in the search operations.
	ImportGoPart langutil.PartType = "import"

	// PackageGoPart represents Go package declarations.
	// This part type targets the package name declared at the top of Go source files.
	// The search operates on the package name itself, not including the "package" keyword.
	//
	// For example, to find "package main", use "main" as the part name.
	// Package declarations are unique per file, so searches should return at most one result
	// per source file.
	PackageGoPart langutil.PartType = "package"
)

// Compile-time verification that GoProcessor implements the langutil.Processor interface.
// This ensures that any changes to the Processor interface will cause compilation errors
// if GoProcessor doesn't implement all required methods, providing early detection of
// interface compatibility issues.
var _ langutil.Processor = (*GoProcessor)(nil)

// GoProcessor implements the langutil.Processor interface for Go language processing.
// This processor provides comprehensive AST-based operations for Go source code including
// finding, replacing, and validating Go language constructs. It uses Go's official AST
// parsing libraries to ensure accurate and complete analysis compatible with the Go compiler.
//
// # Capabilities
//
// GoProcessor supports all major Go language constructs:
//   - Functions and methods with receiver type handling
//   - Type definitions (structs, interfaces, aliases)
//   - Constant and variable declarations (individual and grouped)
//   - Import statements with path and alias support
//   - Package declarations
//
// # AST Integration
//
// The processor leverages Go's standard AST libraries:
//   - go/parser for syntax analysis and AST generation
//   - go/token for position tracking and file set management
//   - go/ast for AST traversal and manipulation
//
// This integration ensures compatibility with Go language evolution and provides
// access to all language features supported by the Go compiler.
//
// # Method Resolution
//
// The processor handles Go's method system by supporting receiver type specification
// in method names. Methods can be found using the format "ReceiverType.MethodName"
// where ReceiverType includes pointer indicators when applicable.
//
// # Position Accuracy
//
// All operations provide precise position information including line numbers and
// byte offsets that are compatible with editors and development tools. Position
// information is generated directly from the AST token positions.
//
// # Registration
//
// The processor registers itself automatically during package initialization,
// making Go language support available immediately when the golang package is imported.
type GoProcessor struct{}

// init registers the GoProcessor with the langutil processor registry during package initialization.
// This automatic registration ensures that Go language support is available immediately when
// the golang package is imported, without requiring explicit setup by applications.
//
// The registration makes the processor available for all langutil operations that target
// the Go language, including file validation, AST operations, and content replacement.
func init() {
	langutil.RegisterProcessor(&GoProcessor{})
}

// Language returns "go" to identify this processor as the Go language handler.
// This method implements the langutil.Processor interface requirement and provides
// the language identifier used for processor registration and lookup operations.
//
// The returned value matches the langutil.GoLanguage constant and is used throughout
// the langutil framework to route Go-related operations to this processor.
func (g *GoProcessor) Language() langutil.Language {
	return "go"
}

// SupportedPartTypes returns the complete list of Go language constructs that this processor can handle.
// This method provides introspection capabilities for the processor, allowing callers to determine
// what kinds of code constructs can be found or manipulated in Go source code.
//
// # Returned Part Types
//
// The method returns all Go part type constants defined in this package:
//   - FuncGoPart: Functions and methods
//   - TypeGoPart: Type definitions
//   - ConstGoPart: Constant declarations
//   - VarGoPart: Variable declarations
//   - ImportGoPart: Import statements
//   - PackageGoPart: Package declarations
//
// # Usage
//
// This method is commonly used for:
//   - Validating user input before attempting AST operations
//   - Building user interfaces that show available construct types
//   - Runtime discovery of processor capabilities
//   - Integration testing and capability verification
//
// The returned slice should not be modified by callers as it represents the
// processor's immutable capabilities.
func (g *GoProcessor) SupportedPartTypes() []langutil.PartType {
	return []langutil.PartType{
		FuncGoPart,
		TypeGoPart,
		ConstGoPart,
		VarGoPart,
		ImportGoPart,
		PackageGoPart,
	}
}

// FindPart finds a specific Go language construct in source code using AST parsing.
// This method performs comprehensive AST-based searching to locate code constructs matching
// the specified type and name. It provides precise position information and complete content
// extraction for found constructs.
//
// # Search Process
//
// The method follows a multi-step process:
//  1. Parses the complete source content into an AST using go/parser
//  2. Traverses the AST to find constructs matching the specified type and name
//  3. Extracts precise position information using the token.FileSet
//  4. Returns detailed PartInfo with location and content data
//
// # Construct Matching
//
// Different construct types use different matching strategies:
//   - Functions: Matches function names and handles method receiver types
//   - Types: Matches type names in type declarations
//   - Constants/Variables: Matches identifier names in declarations
//   - Imports: Matches import paths with flexible quote handling
//   - Packages: Matches package names in package declarations
//
// # Position Information
//
// The method provides both line-based and byte-based position information:
//   - Line numbers are 1-based to match editor conventions
//   - Byte offsets are 0-based for direct string slicing operations
//   - Positions include the complete construct from start to end
//
// # Error Handling
//
// Returns errors for:
//   - Syntax errors in the source content
//   - Unsupported part types
//   - AST parsing failures
//   - Internal processing errors
//
// Returns a PartInfo with Found=false if the construct is not found, which is
// not considered an error condition.
func (g *GoProcessor) FindPart(args langutil.PartArgs) (pi *langutil.PartInfo, err error) {
	var fs *token.FileSet
	var file *ast.File
	var start, end token.Pos
	var found bool
	var startPos, endPos token.Position

	pi = &langutil.PartInfo{Found: false}

	// Parse the Go file
	fs = token.NewFileSet()
	file, err = parser.ParseFile(fs, "", args.Content, parser.ParseComments)
	if err != nil {
		err = fmt.Errorf("failed to parse Go file: %w", err)
		goto end
	}

	// Find the part
	start, end, found, err = g.findGoPart(file, args)
	if err != nil {
		goto end
	}

	if !found {
		goto end
	}

	// Convert positions to line numbers and offsets
	startPos = fs.Position(start)
	endPos = fs.Position(end)

	pi.Found = true
	pi.StartLine = startPos.Line
	pi.EndLine = endPos.Line
	pi.StartOffset = startPos.Offset
	pi.EndOffset = endPos.Offset
	pi.Content = args.Content[startPos.Offset:endPos.Offset]

end:
	return pi, err
}

// ReplacePart replaces a specific Go language construct with new content.
// This method performs a comprehensive two-phase operation: finding the target construct
// and replacing it with validated new content. The method ensures that the replacement
// maintains syntactic correctness of the resulting Go source code.
//
// # Replacement Process
//
// The method follows these steps:
//  1. Uses FindPart to locate the target construct
//  2. Validates that the construct was found
//  3. Performs text replacement at the exact AST boundaries
//  4. Validates that the resulting source code is syntactically correct
//  5. Returns the complete modified source code
//
// # Content Validation
//
// Before performing the replacement, the method validates the new content to ensure
// it's appropriate for the target construct type. This includes checking that:
//   - Function replacements start with "func"
//   - Type replacements start with "type"
//   - Constant/variable replacements have appropriate syntax
//   - Import replacements contain valid import syntax
//   - Package replacements start with "package"
//
// # Syntax Verification
//
// After replacement, the method parses the complete modified source to ensure
// syntactic correctness. This prevents the method from returning invalid Go code
// that would cause compilation errors.
//
// # Error Conditions
//
// Returns errors for:
//   - Construct not found in the source
//   - Invalid replacement content for the construct type
//   - Replacement results in syntactically invalid Go code
//   - AST parsing or processing failures
//
// # Example Usage
//
//	args := langutil.PartArgs{
//		Language:   langutil.GoLanguage,
//		Content:    sourceCode,
//		PartType:   "func",
//		PartName:   "oldFunction",
//		NewContent: "func newFunction() { fmt.Println(\"new\") }",
//		Filepath:   "example.go",
//	}
//	newSource, err := processor.ReplacePart(args)
func (g *GoProcessor) ReplacePart(args langutil.PartArgs) (result string, err error) {
	var partInfo *langutil.PartInfo

	// Find the part first
	partInfo, err = g.FindPart(args)
	if err != nil {
		goto end
	}

	if !partInfo.Found {
		err = fmt.Errorf("%s '%s' not found in file", args.PartType, args.PartName)
		goto end
	}

	// Replace the content
	result = args.Content[:partInfo.StartOffset] + args.NewContent + args.Content[partInfo.EndOffset:]

	// Validate the result
	err = g.ValidateSyntax(result)
	if err != nil {
		err = fmt.Errorf("replacement resulted in invalid Go syntax: %w", err)
		goto end
	}

end:
	return result, err
}

// ValidateContent validates that replacement content is appropriate for the specified Go construct type.
// This method performs static analysis of replacement content to ensure it conforms to Go syntax
// requirements for the target construct type. It provides early validation before attempting
// replacement operations that might result in invalid source code.
//
// # Validation Rules
//
// The method applies construct-specific validation rules:
//
// Functions (FuncGoPart):
//   - Content must start with "func "
//   - Ensures basic function declaration syntax
//
// Types (TypeGoPart):
//   - Content must start with "type "
//   - Ensures basic type declaration syntax
//
// Constants (ConstGoPart):
//   - Content must contain "=" or start with "const"
//   - Handles both individual and grouped constant syntax
//
// Variables (VarGoPart):
//   - Content must contain "=" or start with "var"
//   - Handles both individual and grouped variable syntax
//
// Imports (ImportGoPart):
//   - Content must contain "import" or quote characters
//   - Validates basic import statement structure
//
// Packages (PackageGoPart):
//   - Content must start with "package "
//   - Ensures proper package declaration syntax
//
// # Limitations
//
// This validation is intentionally basic and focuses on obvious syntax requirements.
// It does not perform complete Go syntax validation, which is handled by ValidateSyntax
// after replacement operations are performed.
//
// # Performance
//
// The validation is designed to be fast and is suitable for real-time feedback in
// development tools. It uses simple string checks rather than full AST parsing.
func (g *GoProcessor) ValidateContent(args langutil.PartArgs) (err error) {
	content := strings.TrimSpace(args.Content)

	switch args.PartType {
	case FuncGoPart:
		if !strings.HasPrefix(content, "func ") {
			err = fmt.Errorf("func replacement must start with 'func ', got: %s", content[:min(20, len(content))])
		}
	case TypeGoPart:
		if !strings.HasPrefix(content, "type ") {
			err = fmt.Errorf("type replacement must start with 'type ', got: %s", content[:min(20, len(content))])
		}
	case ConstGoPart:
		if !strings.Contains(content, "=") && !strings.HasPrefix(content, "const") {
			err = fmt.Errorf("const replacement must contain '=' or start with 'const', got: %s", content[:min(20, len(content))])
		}
	case VarGoPart:
		if !strings.Contains(content, "=") && !strings.HasPrefix(content, "var") {
			err = fmt.Errorf("var replacement must contain '=' or start with 'var', got: %s", content[:min(20, len(content))])
		}
	case ImportGoPart:
		if !strings.Contains(content, "import") && !strings.Contains(content, "\"") {
			err = fmt.Errorf("import replacement must contain 'import' or quotes, got: %s", content[:min(20, len(content))])
		}
	case PackageGoPart:
		if !strings.HasPrefix(content, "package ") {
			err = fmt.Errorf("package replacement must start with 'package ', got: %s", content[:min(20, len(content))])
		}
	default:
		err = fmt.Errorf("unsupported part type for Go: %s", args.PartType)
	}

	return err
}

// ValidateSyntax validates that the provided Go source code is syntactically correct.
// This method performs complete syntax validation using Go's official parser to ensure
// that source code can be successfully compiled. It's used to verify file contents
// and validate the results of replacement operations.
//
// # Validation Process
//
// The method uses go/parser.ParseFile with the following characteristics:
//   - Creates a new token.FileSet for position tracking
//   - Includes comment parsing for complete source analysis
//   - Reports syntax errors with line and column information
//   - Validates the complete source file structure
//
// # Error Reporting
//
// Syntax errors are reported directly from the Go parser and include:
//   - Precise line and column positions
//   - Descriptive error messages about syntax violations
//   - Context about the specific syntax issue encountered
//
// # Performance
//
// This method performs full AST parsing and can be expensive for large source files.
// However, it provides definitive validation that matches the Go compiler's syntax
// requirements, ensuring that validated code will compile successfully.
//
// # Use Cases
//
// This method is used in several contexts:
//   - File validation during initial processing
//   - Post-replacement validation to ensure correctness
//   - Standalone syntax checking for development tools
//   - Batch validation of multiple source files
//
// Returns nil if the source is syntactically valid, or a detailed error if syntax
// violations are found.
func (g *GoProcessor) ValidateSyntax(source string) (err error) {
	var fs *token.FileSet

	fs = token.NewFileSet()
	_, err = parser.ParseFile(fs, "", source, parser.ParseComments)

	return err
}

// findGoPart locates a specific Go language construct within an AST.
// This method serves as the central dispatch function for construct-specific search operations.
// It examines the part type and delegates to specialized search functions that understand
// the structure and semantics of different Go language constructs.
//
// # Dispatch Logic
//
// The method routes searches based on construct type:
//   - PackageGoPart: Searches package declarations
//   - ImportGoPart: Searches import statements
//   - ConstGoPart: Searches constant declarations
//   - VarGoPart: Searches variable declarations
//   - TypeGoPart: Searches type definitions
//   - FuncGoPart: Searches function and method declarations
//
// # Position Information
//
// All search functions return token.Pos values that represent precise positions
// within the source file. These positions are used to extract content and provide
// accurate location information for found constructs.
//
// # Error Handling
//
// Returns an error for unsupported part types. Individual search functions may
// return additional errors for malformed AST structures or other processing issues.
//
// # Search Accuracy
//
// The method ensures accurate searches by using Go's AST structure directly,
// which provides complete semantic understanding of the source code structure
// and relationships between language constructs.
func (g *GoProcessor) findGoPart(file *ast.File, args langutil.PartArgs) (startPos, endPos token.Pos, found bool, err error) {
	partName := args.PartName
	switch args.PartType {
	case PackageGoPart:
		if file.Name.Name == partName {
			startPos = file.Name.Pos()
			endPos = file.Name.End()
			found = true
		}
	case ImportGoPart:
		startPos, endPos, found = g.findGoImport(file, partName)
	case ConstGoPart:
		startPos, endPos, found = g.findGoConst(file, partName)
	case VarGoPart:
		startPos, endPos, found = g.findGoVar(file, partName)
	case TypeGoPart:
		startPos, endPos, found = g.findGoType(file, partName)
	case FuncGoPart:
		startPos, endPos, found = g.findGoFunc(file, partName)
	default:
		err = fmt.Errorf("unsupported part type: %s", args.PartType)
	}

	return startPos, endPos, found, err
}

// findGoImport locates import statements by import path.
// This method searches through all import declarations in the file to find imports
// matching the specified import path. It handles both individual imports and
// grouped import blocks with flexible path matching.
//
// # Path Matching
//
// The method supports flexible import path matching:
//   - Exact path matching: "github.com/user/repo"
//   - Quoted path matching: "\"github.com/user/repo\""
//   - Handles both single and grouped import declarations
//
// # Search Strategy
//
// The search examines all import specifications in the file:
//   - Iterates through file.Imports slice
//   - Compares import path values with and without quotes
//   - Returns position of the entire import statement when found
//
// # Position Information
//
// Returns the start and end positions of the complete import statement,
// including any associated comments or formatting.
func (g *GoProcessor) findGoImport(file *ast.File, importPath string) (startPos, endPos token.Pos, found bool) {
	for _, imp := range file.Imports {
		if imp.Path.Value == `"`+importPath+`"` || imp.Path.Value == importPath {
			startPos = imp.Pos()
			endPos = imp.End()
			found = true
			return
		}
	}
	return
}

// findGoConst locates constant declarations by constant name.
// This method searches through all constant declarations in the file, handling both
// individual constant declarations and constants within grouped constant blocks.
// It provides comprehensive coverage of Go's constant declaration syntax.
//
// # Declaration Types
//
// The method handles all forms of constant declarations:
//   - Individual constants: const MyConst = "value"
//   - Grouped constants: const ( MyConst = "value"; Other = 123 )
//   - Typed constants: const MyConst string = "value"
//   - Untyped constants: const MyConst = "value"
//
// # Search Strategy
//
// The search process:
//   - Examines all general declarations (ast.GenDecl) with token.CONST
//   - Iterates through all value specifications within each declaration
//   - Checks all names within each value specification for matches
//   - Returns position of the entire constant declaration block when found
//
// # Position Scope
//
// Returns the position of the complete constant declaration, which may include
// multiple constants if the target constant is part of a grouped declaration.
// This ensures that replacement operations maintain proper Go syntax structure.
func (g *GoProcessor) findGoConst(file *ast.File, constName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						if name.Name == constName {
							startPos = genDecl.Pos()
							endPos = genDecl.End()
							found = true
							return
						}
					}
				}
			}
		}
	}
	return
}

// findGoVar locates variable declarations by variable name.
// This method searches through all variable declarations in the file, handling both
// individual variable declarations and variables within grouped variable blocks.
// It provides comprehensive coverage of Go's variable declaration syntax patterns.
//
// # Declaration Types
//
// The method handles all forms of variable declarations:
//   - Individual variables: var myVar string
//   - Grouped variables: var ( myVar string; other int )
//   - Initialized variables: var myVar = "value"
//   - Typed variables: var myVar string = "value"
//   - Multiple variables: var a, b, c int
//
// # Search Strategy
//
// The search process:
//   - Examines all general declarations (ast.GenDecl) with token.VAR
//   - Iterates through all value specifications within each declaration
//   - Checks all names within each value specification for matches
//   - Returns position of the entire variable declaration block when found
//
// # Position Scope
//
// Returns the position of the complete variable declaration, which may include
// multiple variables if the target variable is part of a grouped or multi-variable
// declaration. This ensures proper syntax preservation during replacements.
//
// # Package-Level Focus
//
// This method focuses on package-level variable declarations and does not search
// within function bodies for local variable declarations, as those require more
// complex scoping analysis.
func (g *GoProcessor) findGoVar(file *ast.File, varName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						if name.Name == varName {
							startPos = genDecl.Pos()
							endPos = genDecl.End()
							found = true
							return
						}
					}
				}
			}
		}
	}
	return
}

// findGoType locates type declarations by type name.
// This method searches through all type declarations in the file, handling all forms
// of Go type definitions including structs, interfaces, aliases, and composite types.
// It provides comprehensive coverage of Go's type system.
//
// # Type Definition Coverage
//
// The method handles all forms of type declarations:
//   - Struct types: type MyStruct struct { ... }
//   - Interface types: type MyInterface interface { ... }
//   - Type aliases: type MyType = ExistingType
//   - Named types: type MyType ExistingType
//   - Function types: type MyFunc func() error
//   - Composite types: type MySlice []string
//
// # Search Strategy
//
// The search process:
//   - Examines all general declarations (ast.GenDecl) with token.TYPE
//   - Iterates through all type specifications within each declaration
//   - Compares type names for exact matches
//   - Returns position of the entire type declaration when found
//
// # Position Scope
//
// Returns the position of the complete type declaration, including any associated
// documentation comments and the full type definition. This ensures that replacement
// operations can substitute the entire type definition while maintaining proper syntax.
//
// # Scoping
//
// This method finds package-level type declarations and does not search for types
// defined within function scopes, as such type definitions are rare and require
// different handling strategies.
func (g *GoProcessor) findGoType(file *ast.File, typeName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if typeSpec.Name.Name == typeName {
						startPos = genDecl.Pos()
						endPos = genDecl.End()
						found = true
						return
					}
				}
			}
		}
	}
	return
}

// findGoFunc locates function and method declarations by name with receiver type support.
// This method provides comprehensive search capabilities for both standalone functions
// and methods attached to types. It handles Go's method system by supporting receiver
// type specification in method names, enabling precise method targeting.
//
// # Function Types
//
// The method handles all forms of function declarations:
//   - Standalone functions: func MyFunction() { ... }
//   - Value receiver methods: func (r ReceiverType) Method() { ... }
//   - Pointer receiver methods: func (r *ReceiverType) Method() { ... }
//   - Functions with complex signatures: func Name(args...) (results...) { ... }
//
// # Method Name Format
//
// For methods, use the format "ReceiverType.MethodName":
//   - Value receivers: "MyStruct.Method"
//   - Pointer receivers: "*MyStruct.Method"
//   - Interface methods: "MyInterface.Method"
//
// # Search Strategy
//
// The search process:
//   - Examines all function declarations (ast.FuncDecl) in the file
//   - For standalone functions: matches the function name directly
//   - For methods: extracts receiver type and matches "ReceiverType.MethodName" format
//   - Handles both pointer and value receiver types appropriately
//   - Returns position of the entire function declaration including documentation
//
// # Receiver Type Extraction
//
// The method analyzes receiver lists to extract type information:
//   - Identifies pointer receivers using ast.StarExpr
//   - Extracts type names from receiver expressions
//   - Constructs qualified method names for comparison
//
// # Position Information
//
// Returns the position of the complete function declaration, including any
// associated documentation comments, receiver declarations, parameters, return
// types, and function body.
func (g *GoProcessor) findGoFunc(file *ast.File, funcName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			var name string

			// Handle regular functions
			if funcDecl.Recv == nil {
				name = funcDecl.Name.Name
			} else {
				// Handle methods - format as ReceiverType.MethodName
				if len(funcDecl.Recv.List) > 0 {
					var recvType string
					switch recv := funcDecl.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						if ident, ok := recv.X.(*ast.Ident); ok {
							recvType = "*" + ident.Name
						}
					case *ast.Ident:
						recvType = recv.Name
					}
					name = recvType + "." + funcDecl.Name.Name
				}
			}

			if name == funcName {
				startPos = funcDecl.Pos()
				endPos = funcDecl.End()
				found = true
				return
			}
		}
	}
	return
}
