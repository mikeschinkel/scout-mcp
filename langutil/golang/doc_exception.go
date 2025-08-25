package golang

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

const (
	InvalidDocException DocExceptionType = 0
	ReadmeException     DocExceptionType = 1 << iota
	FileException
	FuncException
	TypeException
	ConstException
	VarException
	GroupException
)

type DocExceptionType int

func (et DocExceptionType) String() (s string) {
	switch et {
	case ReadmeException:
		s = "Missing README.md file"
	case FileException:
		s = "Missing file comment"
	case FuncException:
		s = "Missing func comment"
	case TypeException:
		s = "Missing type comment"
	case ConstException:
		s = "Missing const comment"
	case VarException:
		s = "Missing var comment"
	case GroupException | ConstException:
		s = "Missing const group comment"
	case GroupException | VarException:
		s = "Missing var group comment"
	case GroupException:
		s = "Missing group comment"
	case InvalidDocException:
		fallthrough
	default:
		s = fmt.Sprintf("Invalid DocExceptionType '%d'", et)
		logger.Error(s)
	}
	return
}

// DocException represents a documentation or coding standard violation found in a Go source file.
// This structure captures detailed information about documentation issues, coding standard violations,
// and other quality concerns discovered during Go code analysis. It provides precise location
// information and categorizes different types of violations for automated processing and reporting.
//
// # Violation Categories
//
// DocException is used to track various types of violations:
//   - Missing package documentation
//   - Missing function/method documentation
//   - Missing type documentation
//   - Missing constant/variable documentation
//   - Incorrect documentation formatting
//   - Missing README.md files in packages
//   - Coding standard violations
//
// # Position Information
//
// The structure provides precise position information to help developers locate and fix issues:
//   - Line numbers are 1-based to match editor conventions
//   - EndLine is optional for single-line violations
//   - MultiLine flag indicates violations spanning multiple lines
//
// # JSON Serialization
//
// The structure is designed for JSON serialization to support tooling integration:
//   - Can be serialized to JSON for IDE plugins and CI/CD systems
//   - Supports omitempty tags to reduce JSON output size for optional fields
//   - Compatible with standard Go JSON marshaling conventions
//
// # Integration
//
// DocException integrates with various development tools:
//   - IDE plugins for real-time feedback
//   - CI/CD systems for automated quality checks
//   - Documentation generation tools
//   - Code review systems for violation reporting
type DocException struct {
	File      string           `json:"file"`               // Path to the file containing the violation
	Type      DocExceptionType `json:"type,omitempty"`     // Category of violation (func, type, const, var, package_file, etc.)
	Line      int              `json:"line"`               // Line number where the violation occurs (1-based)
	EndLine   *int             `json:"end_line,omitempty"` // Optional end line for multi-line violations
	MultiLine bool             `json:"multi_line"`         // Whether the violation spans multiple lines
	Element   string           `json:"element"`            // The name of the source element needing documentation
}

func (e DocException) Issue() (issue string) {
	issue = e.Type.String()
	if strings.HasPrefix(issue, "Invalid ") {
		issue = fmt.Sprintf("%s;exception=%v", issue, e)
		logger.Error(issue)
	}
	return issue
}

// DocExceptionArgs contains arguments for creating DocException instances.
// This structure provides a convenient way to pass location information when creating
// DocException instances, allowing for flexible specification of violation boundaries
// and characteristics.
//
// # Usage Patterns
//
// Common usage patterns include:
//   - Single-line violations: Set Line only, leave EndLine nil and MultiLine false
//   - Multi-line violations: Set Line, EndLine, and MultiLine true
//   - Line range violations: Set Line and EndLine for specific ranges
//
// # Validation
//
// The structure does not perform validation of the provided values, allowing for
// flexible usage while requiring callers to ensure consistency between the fields.
// For example, if MultiLine is true, EndLine should typically be provided.
type DocExceptionArgs struct {
	Line      int    // Starting line number for the violation (1-based)
	EndLine   *int   // Optional ending line number for multi-line violations
	MultiLine bool   // Whether the violation spans multiple lines or constructs
	Element   string // The name of the source element needing documentation
}

// NewDocException creates a new DocException with the specified details.
// This constructor function provides a convenient way to create DocException instances
// with consistent initialization and proper handling of optional arguments.
//
// # Parameter Handling
//
// The function handles the arguments structure gracefully:
//   - If args is nil, creates a DocException with minimal information
//   - Copies all provided fields from args to the new DocException
//   - Preserves pointer semantics for optional fields like EndLine
//
// # Type Categorization
//
// The Type parameter should specify the category of violation using standard names:
//   - "func" for function/method documentation issues
//   - "type" for type documentation issues
//   - "const" for constant documentation issues
//   - "var" for variable documentation issues
//   - "package_file" for package-level documentation issues
//   - "readme_missing" for missing README.md files
//   - Custom types for specialized validation rules
//
// # Example Usage
//
//	exception := NewDocException(
//		"main.go",
//		"func",
//		&DocExceptionArgs{
//			Line:      42,
//			EndLine:   nil,
//			MultiLine: false,
//		},
//	)
//
// # Return Value
//
// Returns a fully initialized DocException ready for use in violation reporting,
// JSON serialization, or further processing by analysis tools.
func NewDocException(file string, Type DocExceptionType, args *DocExceptionArgs) DocException {
	fe := DocException{
		File: file,
		Type: Type,
	}
	if args != nil {
		fe.Line = args.Line
		fe.EndLine = args.EndLine
		fe.MultiLine = args.MultiLine
		fe.Element = args.Element
	}
	return fe
}

// DocsExceptionsArgs contains arguments for Go documentation validation operations.
// This structure encapsulates the parameters needed to perform comprehensive documentation
// analysis across Go packages and source files. It supports both single-file and
// multi-package analysis with flexible filtering and recursion options.
//
// # Usage Patterns
//
// The structure supports several common analysis patterns:
//   - Single file analysis: Provide a specific .go file path
//   - Package analysis: Provide a package directory path
//   - Recursive analysis: Enable recursive processing for multi-package projects
//   - Filtered analysis: Use extensions to limit analysis to specific file types
//   - Excluded directories: Skip irrelevant directories during traversal
//
// # Performance Tuning
//
// The MaxFiles parameter allows for performance tuning in large codebases by limiting
// the number of files processed in a single operation. This helps prevent memory
// exhaustion and provides predictable processing times.
//
// # File and Directory Exclusion
//
// The Exclude and ExcludeMode parameters provide flexible control over which
// files and directories are skipped during traversal, improving performance and relevance.
type DocsExceptionsArgs struct {
	Path        string           // File or directory path to analyze (supports "..." for recursive)
	Recursive   RecurseDirective // Whether to process directories recursively
	Exclude     []string         // File and directory names to exclude (used with ExcludeMode)
	ExcludeMode ExcludeMode      // How to interpret Exclude (default: UseDefaults)
}

// parse parses and update both Path and Recursive, although recursion is
// handle via /... at the end of the Path.
func (args *DocsExceptionsArgs) parse() (err error) {
	var pt PathType
	path := args.Path
	if args.Recursive == NonSpecifiedRecurse {
		args.Recursive = DoRecurse
	}
	if strings.HasSuffix(path, "...") {
		args.Recursive = DoRecurse
		path = path[:len(path)-3]
	}
	pt, err = checkPath(path)
	if err != nil {
		goto end
	}
	switch pt {
	case FilePath:
		args.Recursive = DoNotRecurse
	case DirPath:
		path = filepath.Clean(path)
	default:
		err = fmt.Errorf("invalid path: %s", args.Path)
	}
	args.Path = path
end:
	return err
}

// DocExceptions performs comprehensive Go documentation validation and returns a list of violations.
// This function analyzes Go source files to identify documentation issues and coding
// standard violations. It uses Go's AST parser to directly parse individual files,
// eliminating the need for valid Go modules or workspace configuration.
//
// # Analysis Scope
//
// The function performs several types of validation:
//   - Package-level documentation requirements
//   - Function and method documentation standards
//   - Type documentation following Go conventions
//   - Constant and variable documentation formatting
//   - README.md presence validation for directories with Go files
//
// # Path Resolution
//
// The function intelligently handles different path types:
//   - File paths: Analyzes the single .go file
//   - Directory paths: Finds all .go files in the directory
//   - Recursive paths: Recursively finds all .go files in subdirectories
//
// # File Discovery
//
// Uses direct file system operations to discover Go files:
//   - Individual file parsing with go/parser.ParseFile
//   - No dependency on Go modules or workspace configuration
//   - Parse errors in individual files don't stop analysis of other files
//   - Supports both test and non-test Go files
//
// # Error Handling
//
// Returns detailed error information for:
//   - File system access issues
//   - Path resolution problems
//   - Individual file parse errors are logged but don't stop analysis
//
// # Example Usage
//
//	args := DocsExceptionsArgs{
//		Path:      "/path/to/project",
//		Recursive: true,
//		MaxFiles:  1000,
//	}
//	exceptions, err := DocExceptions(ctx, args)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, exc := range exceptions {
//		fmt.Printf("Documentation issue in %s:%d - %s\n", exc.File, exc.Line, exc.Type)
//	}
//
// # Integration
//
// This function integrates with broader documentation tooling and can be used for:
//   - Continuous integration documentation checks
//   - Pre-commit hooks for documentation validation
//   - IDE integration for real-time documentation feedback
//   - Documentation coverage reporting and metrics
func DocExceptions(ctx context.Context, args *DocsExceptionsArgs) (exceptions []DocException, err error) {
	var dir *GoDirectory
	var traverseArgs *TraverseArgs

	ensureLogger()

	err = args.parse()
	if err != nil {
		goto end
	}

	// Use default exclusions if not specified
	if args.ExcludeMode == 0 { // Zero value means not set
		args.ExcludeMode = UseDefaults
	}

	dir = NewGoDirectory(args.Path, nil)
	traverseArgs = &TraverseArgs{
		RecurseDirectory: args.Recursive,
		Exclude:          args.Exclude,
		ExcludeMode:      args.ExcludeMode,
	}
	err = dir.Traverse(ctx, traverseArgs)
	if err != nil {
		goto end
	}
	exceptions = dir.Exceptions(ctx, args.Path, args.Recursive)

end:
	return exceptions, err
}
