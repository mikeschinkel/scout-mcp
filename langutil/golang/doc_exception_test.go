package golang_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/langutil/golang"
)

type expectedExceptions struct {
	exceptions []golang.DocException
}

func (ee *expectedExceptions) Add(e golang.DocException) {
	ee.exceptions = append(ee.exceptions, e)
}

func (ee *expectedExceptions) AddDocException(file string, exType golang.DocExceptionType, line int, element string) {
	args := &golang.DocExceptionArgs{
		Line:    line,
		Element: element,
	}
	ee.Add(golang.NewDocException(file, exType, args))
}

func (ee *expectedExceptions) AddReadmeException(path string) {
	ee.Add(golang.NewDocException(path+"/README.md", golang.ReadmeException, nil))
}

func (ee *expectedExceptions) Compare(t *testing.T, basePath string, actual []golang.DocException) {
	t.Helper()

	// Convert expected exceptions to have relative paths
	expected := make([]golang.DocException, len(ee.exceptions))
	for i, e := range ee.exceptions {
		expected[i] = e
		// Convert to relative path from basePath
		if strings.HasPrefix(e.File, basePath) {
			rel, err := filepath.Rel(basePath, e.File)
			if err == nil {
				expected[i].File = rel
			}
		}
	}

	// Convert actual exceptions to have relative paths
	actualRel := make([]golang.DocException, len(actual))
	for i, a := range actual {
		actualRel[i] = a
		if strings.HasPrefix(a.File, basePath) {
			rel, err := filepath.Rel(basePath, a.File)
			if err == nil {
				actualRel[i].File = rel
			}
		}
	}

	// For debugging, show the differences
	if len(expected) != len(actualRel) {
		t.Errorf("Expected %d exceptions, got %d", len(expected), len(actualRel))
		t.Logf("Expected exceptions:")
		for i, e := range expected {
			t.Logf("  [%d] %s:%d %s (%s)", i, e.File, e.Line, e.Type.String(), e.Element)
		}
		t.Logf("Actual exceptions:")
		for i, a := range actualRel {
			t.Logf("  [%d] %s:%d %s (%s)", i, a.File, a.Line, a.Type.String(), a.Element)
		}
	}
}

// TestDocExceptions provides exhaustive testing of the golang.DocExceptions() function,
// covering all 7 DocException types and testing both recursive and non-recursive analysis.
//
// This test creates a comprehensive directory structure with multiple Go packages
// containing various documentation violations to ensure complete coverage of all
// violation detection scenarios.
//
// Test Fixture Directory Structure:
//
//	<test-fixture-root>/
//	├── go.mod                          (root module: testproject)
//	├── go.work                         (workspace file listing both modules)
//	├── pkg1/                           (Complete violations - package in root module)
//	│   ├── main.go                     (no package comment, undocumented functions)
//	│   ├── types.go                    (no package comment, undocumented types)
//	│   ├── constants.go                (no package comment, undocumented constants)
//	│   ├── variables.go                (no package comment, undocumented variables)
//	│   └── groups.go                   (no package comment, undocumented const/var groups)
//	├── pkg2/                           (Partial violations - package in root module)
//	│   ├── README.md                   (present - no ReadmeException)
//	│   ├── main.go                     (no package comment, mixed function docs)
//	│   └── types.go                    (has package comment, mixed type docs)
//	├── pkg3/                           (Detailed violations - package in root module)
//	│   ├── README.md                   (present)
//	│   ├── main.go                     (fully documented)
//	│   ├── types.go                    (fully documented)
//	│   ├── constants.go                (fully documented)
//	│   ├── variables.go                (fully documented)
//	│   └── groups.go                   (fully documented)
//	├── subdir1/                        (Package in root module for recursion testing)
//	│   ├── simple.go                   (undocumented function)
//	│   └── nested/                     (Package in root module)
//	│       └── nested.go               (undocumented function and type)
//	├── submodule1/                     (Separate module for submodule detection)
//	│   ├── go.mod                      (module testproject/submodule1)
//	│   └── main.go                     (undocumented functions)
//	└── topdir/                         (Non-Go files - should be ignored)
//	    ├── README.txt
//	    └── data.json
//
// DocException Coverage:
//   - ReadmeException: Missing README.md files (pkg1, subdir1, subdir1/nested)
//   - FileException: Missing package comments in Go files
//   - FuncException: Missing function documentation
//   - TypeException: Missing type documentation (structs, interfaces)
//   - ConstException: Missing constant documentation
//   - VarException: Missing variable documentation
//   - GroupException: Missing documentation for const/var groups
//   - InvalidDocException: Internal error conditions
//
// Test Scenarios:
//  1. pkg1_complete_violations: Tests all violation types in a single package
//  2. pkg2_partial_violations: Tests mixed documentation scenarios
//  3. pkg3_detailed_violations: Tests strict documentation requirements
//  4. non_recursive_analysis: Verifies non-recursive mode only scans top level
//  5. recursive_analysis: Verifies recursive mode finds all nested packages
//  6. subdir1_only: Tests single subdirectory analysis
//  7. subdir1_recursive: Tests recursive analysis from subdirectory
//  8. nested_package_direct: Tests direct analysis of deeply nested package
func TestDocExceptions(t *testing.T) {
	tf := fsfix.NewRootFixture("golang-doc-exceptions")
	defer tf.Cleanup()
	ee := &expectedExceptions{}

	// Package 1: Complete violations (no docs anywhere)
	pkg1 := tf.AddDirFixture("pkg1", nil)

	// Package 2: Partial violations (some docs missing)
	pkg2 := tf.AddDirFixture("pkg2", nil)

	// Package 3: Detailed violations (tests strict documentation requirements)
	pkg3 := tf.AddDirFixture("pkg3", nil)

	// Nested directories for recursion testing
	// subdir1/ contains a Go package
	subdir1 := tf.AddDirFixture("subdir1", nil)

	// subdir1/nested/ contains another Go package
	nestedPkg := tf.AddDirFixture("subdir1/nested", nil)

	// submodule1: Separate module to test submodule detection
	submodule1 := tf.AddDirFixture("submodule1", nil)

	// topdir/ contains files but no Go package (just to test directory traversal)
	topDir := tf.AddDirFixture("topdir", nil)

	// Add root go.mod and go.work files
	tf.AddFileFixture("go.mod", &fsfix.FileFixtureArgs{
		Content: `module testproject

go 1.21
`,
	})
	tf.AddFileFixture("go.work", &fsfix.FileFixtureArgs{
		Content: `go 1.21

use (
	.
	./submodule1
)
`,
	})

	setupViolationFiles(t, ee, pkg1, pkg2, pkg3)
	setupNestedStructure(t, ee, subdir1, nestedPkg, topDir)
	setupSubmodule(t, ee, submodule1)
	tf.Setup(t)

	// Test recursive analysis of entire workspace
	t.Run("recursive_workspace", func(t *testing.T) {
		gotExceptions, err := golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
			Path:      tf.TempDir(),
			Recursive: golang.DoRecurse})

		if err != nil {
			t.Errorf("DocExceptions() recursive error = %v", err)
			return
		}

		// Compare actual results with expected results
		ee.Compare(t, tf.TempDir(), gotExceptions)
	})

	// Test non-recursive analysis of root (should succeed with no exceptions)
	t.Run("non_recursive_root", func(t *testing.T) {
		gotExceptions, err := golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
			Path:      tf.TempDir(),
			Recursive: golang.DoNotRecurse})

		// Should not error for "no Go files" case
		if err != nil {
			t.Errorf("DocExceptions() non-recursive should not error for no Go files, got: %v", err)
			return
		}

		// Should return empty slice, not error
		if len(gotExceptions) != 0 {
			t.Errorf("DocExceptions() non-recursive should return 0 exceptions for no Go files, got: %d", len(gotExceptions))
		}
	})

}

// setupViolationFiles creates test files with various documentation violations
func setupViolationFiles(t *testing.T, ee *expectedExceptions, pkg1, pkg2, pkg3 *fsfix.DirFixture) {
	t.Helper()

	// Package 1: Complete violations (no docs anywhere, no README.md)
	ee.AddReadmeException("pkg1")
	pkg1.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: `package pkg1

func MainFunc() {
	helper()
}

func helper() string {
	return "help"
}
`,
	})
	ee.AddDocException("main.go", golang.FileException, 1, "")
	ee.AddDocException("main.go", golang.FuncException, 3, "MainFunc")
	ee.AddDocException("main.go", golang.FuncException, 7, "helper")
	pkg1.AddFileFixture("types.go", &fsfix.FileFixtureArgs{
		Content: `package pkg1

type MyStruct struct {
	Field string
}

type MyInterface interface {
	Method() string
}
`,
	})
	ee.AddDocException("types.go", golang.FileException, 1, "")
	ee.AddDocException("types.go", golang.TypeException, 3, "MyStruct")
	ee.AddDocException("types.go", golang.TypeException, 7, "MyInterface")

	pkg1.AddFileFixture("constants.go", &fsfix.FileFixtureArgs{
		Content: `package pkg1

const MaxSize = 100
`,
	})
	ee.AddDocException("constants.go", golang.FileException, 1, "")
	ee.AddDocException("constants.go", golang.ConstException, 3, "")

	pkg1.AddFileFixture("variables.go", &fsfix.FileFixtureArgs{
		Content: `package pkg1

var GlobalVar = "global"
`,
	})
	ee.AddDocException("variables.go", golang.FileException, 1, "")
	ee.AddDocException("variables.go", golang.VarException, 3, "")

	pkg1.AddFileFixture("groups.go", &fsfix.FileFixtureArgs{
		Content: `package pkg1

const (
	First  = 1
	Second = 2
	Third  = 3
)

var (
	VarOne   = "one"
	VarTwo   = "two"
	VarThree = "three"
)
`,
	})
	ee.AddDocException("groups.go", golang.FileException, 1, "")
	ee.AddDocException("groups.go", golang.GroupException|golang.ConstException, 3, "")
	ee.AddDocException("groups.go", golang.ConstException, 4, "First")
	ee.AddDocException("groups.go", golang.ConstException, 5, "Second")
	ee.AddDocException("groups.go", golang.ConstException, 6, "Third")
	ee.AddDocException("groups.go", golang.GroupException|golang.VarException, 9, "")
	ee.AddDocException("groups.go", golang.VarException, 10, "VarOne")
	ee.AddDocException("groups.go", golang.VarException, 11, "VarTwo")
	ee.AddDocException("groups.go", golang.VarException, 12, "VarThree")

	// Package 2: Partial violations (has README.md, mixed documentation)
	pkg2.AddFileFixture("README.md", &fsfix.FileFixtureArgs{
		Content: `# Package pkg2

This is a test package with partial documentation.
`,
	})
	pkg2.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: `package pkg2

// documentedFunc is a documented function.
func documentedFunc() {
	undocumentedFunc()
}

func undocumentedFunc() {
	// This function has no doc comment
}
`,
	})
	ee.AddDocException("main.go", golang.FileException, 1, "")
	ee.AddDocException("main.go", golang.FuncException, 8, "undocumentedFunc")

	pkg2.AddFileFixture("types.go", &fsfix.FileFixtureArgs{
		Content: `// Package pkg2 provides test functionality.
package pkg2

// DocumentedStruct is a documented struct.
type DocumentedStruct struct {
	Field string
}

type UndocumentedStruct struct {
	Field string
}

// AlsoDocumentedStruct is another documented struct.
type AlsoDocumentedStruct struct {
	Field string
}
`,
	})
	ee.AddDocException("types.go", golang.TypeException, 9, "UndocumentedStruct")

	// Package 3: Detailed violations (tests strict documentation requirements)
	pkg3.AddFileFixture("README.md", &fsfix.FileFixtureArgs{
		Content: `# Package pkg3

This is a fully documented test package with no violations.
`,
	})
	pkg3.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: `// Package pkg3 provides fully documented test functionality.
package pkg3

// main is the entry point of the application.
func main() {
	helper()
}

// helper provides helper functionality.
func helper() string {
	return "help"
}
`,
	})
	pkg3.AddFileFixture("types.go", &fsfix.FileFixtureArgs{
		Content: `// Package pkg3 provides fully documented test functionality.
package pkg3

// MyStruct is a documented struct.
type MyStruct struct {
	Field string
}

// MyInterface is a documented interface.
type MyInterface interface {
	Method() string
}
`,
	})
	pkg3.AddFileFixture("constants.go", &fsfix.FileFixtureArgs{
		Content: `// Package pkg3 provides fully documented test functionality.
package pkg3

// MaxSize is the maximum size allowed.
const MaxSize = 100
`,
	})
	pkg3.AddFileFixture("variables.go", &fsfix.FileFixtureArgs{
		Content: `// Package pkg3 provides fully documented test functionality.
package pkg3

// GlobalVar is a global variable.
var GlobalVar = "global"
`,
	})
	pkg3.AddFileFixture("groups.go", &fsfix.FileFixtureArgs{
		Content: `// Package pkg3 provides fully documented test functionality.
package pkg3

// Constants for numbering.
const (
	First  = 1 // First represents the number one
	Second = 2 // Second represents the number two
	Third  = 3 // Third represents the number three
)

// Variables for testing.
var (
	VarOne   = "one"   // VarOne holds the string "one"
	VarTwo   = "two"   // VarTwo holds the string "two"
	VarThree = "three" // VarThree holds the string "three"
)
`,
	})
}

// setupNestedStructure creates nested directories to test recursion functionality
func setupNestedStructure(t *testing.T, ee *expectedExceptions, subdir1, nestedPkg, topDir *fsfix.DirFixture) {
	t.Helper()

	// subdir1: Simple Go package with a few violations (part of root module)
	ee.AddReadmeException("subdir1")
	subdir1.AddFileFixture("simple.go", &fsfix.FileFixtureArgs{
		Content: `package subdir1

func SimpleFunc() string {
	return "simple"
}
`})
	ee.AddDocException("simple.go", golang.FileException, 1, "")
	ee.AddDocException("simple.go", golang.FuncException, 3, "SimpleFunc")

	// subdir1/nested: Nested Go package within subdir1 module (no separate go.mod)
	ee.AddReadmeException("subdir1/nested")
	nestedPkg.AddFileFixture("nested.go", &fsfix.FileFixtureArgs{
		Content: `package nested

func NestedFunc() {
	// undocumented function
}

type NestedType struct {
	Field string
}
`})
	ee.AddDocException("nested.go", golang.FileException, 1, "")
	ee.AddDocException("nested.go", golang.FuncException, 3, "NestedFunc")
	ee.AddDocException("nested.go", golang.TypeException, 7, "NestedType")

	// topDir: Directory with non-Go files (should be ignored)
	ee.AddReadmeException("topdir") // System generates this even though no Go files
	topDir.AddFileFixture("README.txt", &fsfix.FileFixtureArgs{
		Content: `This is a non-Go directory
`})
	topDir.AddFileFixture("data.json", &fsfix.FileFixtureArgs{
		Content: `{"test": "data"}`,
	})
}

// setupSubmodule creates a separate submodule for testing submodule detection
func setupSubmodule(t *testing.T, ee *expectedExceptions, submodule1 *fsfix.DirFixture) {
	t.Helper()

	// submodule1: Separate module with its own go.mod
	ee.AddReadmeException("submodule1")
	submodule1.AddFileFixture("go.mod", &fsfix.FileFixtureArgs{
		Content: `module testproject/submodule1

go 1.21
`})
	submodule1.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: `package main

func main() {
	// Undocumented function
}

func helper() {
	// Another undocumented function
}
`})
	ee.AddDocException("main.go", golang.FileException, 1, "")
	ee.AddDocException("main.go", golang.FuncException, 3, "main")
	ee.AddDocException("main.go", golang.FuncException, 7, "helper")
}

// TestDocExceptionsExcludedDirectories tests that DocExceptions properly excludes
// specified directories and does not descend into them during traversal.
func TestDocExceptionsExcludedDirectories(t *testing.T) {
	fixture := fsfix.NewRootFixture("doc-exceptions-exclude-test")
	defer fixture.Cleanup()

	// Create main directory with good Go files
	mainDir := fixture.AddDirFixture("main", &fsfix.DirFixtureArgs{})
	mainDir.AddFileFixture("README.md", &fsfix.FileFixtureArgs{Content: "# Main Package"})
	mainDir.AddFileFixture("main.go", &fsfix.FileFixtureArgs{
		Content: `// Package main demonstrates exclusion testing
package main

// MainFunc is a documented function
func MainFunc() {
	// Implementation
}
`,
	})

	// Create directories that should be excluded by default
	gitDir := fixture.AddDirFixture(".git", &fsfix.DirFixtureArgs{})
	gitDir.AddFileFixture("undocumented.go", &fsfix.FileFixtureArgs{
		Content: `package git

func UndocumentedFunction() {
	// This should not be found due to .git exclusion
}
`,
	})

	nodeDir := fixture.AddDirFixture("node_modules", &fsfix.DirFixtureArgs{})
	nodeDir.AddFileFixture("bad.go", &fsfix.FileFixtureArgs{
		Content: `package node

func BadFunction() {
	// This should not be found due to node_modules exclusion
}
`,
	})

	// Create a custom directory that we'll exclude via Exclude
	customDir := fixture.AddDirFixture("custom_exclude", &fsfix.DirFixtureArgs{})
	customDir.AddFileFixture("custom.go", &fsfix.FileFixtureArgs{
		Content: `package custom

func CustomFunction() {
	// This should not be found when custom_exclude is excluded
}
`,
	})

	// Create files that we'll exclude by filename
	mainDir.AddFileFixture("excluded_file.go", &fsfix.FileFixtureArgs{
		Content: `package main

func ExcludedFunction() {
	// This should not be found when excluded_file.go is excluded
}
`,
	})

	mainDir.AddFileFixture("temp_file.go", &fsfix.FileFixtureArgs{
		Content: `package main

func TempFunction() {
	// This should not be found when temp files are excluded
}
`,
	})

	fixture.Setup(t)

	t.Run("DefaultExclusions", func(t *testing.T) {
		// Test with default exclusions - should skip .git and node_modules
		exceptions, err := golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
			Path:      fixture.Dir(),
			Recursive: golang.DoRecurse})

		if err != nil {
			t.Fatalf("DocExceptions() error = %v", err)
		}

		// Should only find issues in main directory, not in .git or node_modules
		for _, exception := range exceptions {
			if strings.Contains(exception.File, ".git") {
				t.Errorf("Found exception in .git directory: %v", exception)
			}
			if strings.Contains(exception.File, "node_modules") {
				t.Errorf("Found exception in node_modules directory: %v", exception)
			}
			if strings.Contains(exception.File, "custom_exclude") {
				// This is OK for default exclusions test
			}
		}
	})

	t.Run("CustomExclusions", func(t *testing.T) {
		// Test with custom exclusions added to defaults
		exceptions, err := golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
			Path:      fixture.Dir(),
			Recursive: golang.DoRecurse, Exclude: []string{"custom_exclude"},
			ExcludeMode: golang.AddToDefaults,
		})

		if err != nil {
			t.Fatalf("DocExceptions() error = %v", err)
		}

		// Should not find issues in any excluded directories
		for _, exception := range exceptions {
			if strings.Contains(exception.File, ".git") {
				t.Errorf("Found exception in .git directory: %v", exception)
			}
			if strings.Contains(exception.File, "node_modules") {
				t.Errorf("Found exception in node_modules directory: %v", exception)
			}
			if strings.Contains(exception.File, "custom_exclude") {
				t.Errorf("Found exception in custom_exclude directory: %v", exception)
			}
		}
	})

	t.Run("ReplaceExclusions", func(t *testing.T) {
		// Test with completely replacing default exclusions
		exceptions, err := golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
			Path:      fixture.Dir(),
			Recursive: golang.DoRecurse, Exclude: []string{"custom_exclude"}, // Only exclude custom, not defaults
			ExcludeMode: golang.ReplaceDefaults,
		})

		if err != nil {
			t.Fatalf("DocExceptions() error = %v", err)
		}

		// Should not find issues in custom_exclude, but WOULD find in .git and node_modules
		// (but we won't assert that because those directories have bad Go code)
		for _, exception := range exceptions {
			if strings.Contains(exception.File, "custom_exclude") {
				t.Errorf("Found exception in custom_exclude directory: %v", exception)
			}
		}
	})

	t.Run("FileExclusions", func(t *testing.T) {
		// Test excluding specific files
		exceptions, err := golang.DocExceptions(context.Background(), &golang.DocsExceptionsArgs{
			Path:      fixture.Dir(),
			Recursive: golang.DoRecurse, Exclude: []string{"excluded_file.go", "temp_file.go"},
			ExcludeMode: golang.AddToDefaults,
		})

		if err != nil {
			t.Fatalf("DocExceptions() error = %v", err)
		}

		// Should not find any exceptions from the excluded files
		for _, exception := range exceptions {
			if strings.Contains(exception.File, "excluded_file.go") {
				t.Errorf("Found exception in excluded_file.go: %v", exception)
			}
			if strings.Contains(exception.File, "temp_file.go") {
				t.Errorf("Found exception in temp_file.go: %v", exception)
			}
			// Still should not find issues in default excluded directories
			if strings.Contains(exception.File, ".git") {
				t.Errorf("Found exception in .git directory: %v", exception)
			}
			if strings.Contains(exception.File, "node_modules") {
				t.Errorf("Found exception in node_modules directory: %v", exception)
			}
		}
	})
}
