package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test file names for granular editing tests
const (
	UpdateTestFile     = "update_test.go"
	InsertTestFile     = "insert_test.go"
	PatternTestFile    = "pattern_test.go"
	DeleteTestFile     = "delete_test.txt"
	ReplaceTestFile    = "replace_test.go"
	ReplaceFirstFile   = "replace_first_test.go"
	RegexReplaceFile   = "regex_replace_test.go"
	RegexErrorTestFile = "regex_error_test.txt"
)

// Test content templates for granular editing
const (
	UpdateTestContent = `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("This is line 7")
	fmt.Println("This is line 8")
	fmt.Println("Goodbye!")
}
`

	InsertTestContent = `package main

func main() {
	println("Hello")
	println("World")
}
`

	PatternTestContent = `package main

func main() {
	println("Hello, World!")
}

func helper() {
	println("Helper function")
}
`

	DeleteTestContent = `Line 1
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7
Line 8
Line 9
Line 10
`

	ReplaceTestContent = `package main

import "fmt"

func oldFunctionName() {
	fmt.Println("Hello from oldFunctionName")
	oldVariableName := "old value"
	fmt.Println(oldVariableName)
	anotherOldVariable := "another old value"
	fmt.Println(anotherOldVariable)
}

func anotherOldFunctionName() {
	fmt.Println("Another old function")
}
`

	ReplaceFirstTestContent = `package main

func test() {
	value := "test"
	anotherValue := "test"
	finalValue := "test"
}
`

	RegexReplaceTestContent = `package main

func functionOne() {
	return
}

func functionTwo(param string) {
	return
}

func functionThree(a int, b string) {
	return
}
`
)

// TestGranularFileEditingDirect tests all the granular file editing tools using direct server access
func TestGranularFileEditingDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("UpdateFileLines", func(t *testing.T) {
		testUpdateFileLinesDirect(t, env)
	})

	t.Run("InsertFileLines", func(t *testing.T) {
		testInsertFileLinesDirect(t, env)
	})

	t.Run("InsertAtPattern", func(t *testing.T) {
		testInsertAtPatternDirect(t, env)
	})

	t.Run("DeleteFileLines", func(t *testing.T) {
		testDeleteFileLinesDirect(t, env)
	})

	t.Run("ReplacePattern", func(t *testing.T) {
		testReplacePatternDirect(t, env)
	})
}

func testUpdateFileLinesDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create a test file with multiple lines
	testFilePath := filepath.Join(env.GetTestDir(), UpdateTestFile)
	err := os.WriteFile(testFilePath, []byte(UpdateTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Test updating lines 7-8
	result := env.CallTool(t, "update_file_lines", map[string]interface{}{
		"filepath":    testFilePath,
		"start_line":  7,
		"end_line":    8,
		"new_content": "\tfmt.Println(\"Updated line 7\")\n\tfmt.Println(\"Updated line 8\")",
	})
	assert.NotNil(t, result, "update_file_lines should return result")

	// Verify the update
	updatedContent, err := os.ReadFile(testFilePath)
	require.NoError(t, err, "Failed to read updated file")

	expectedContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	fmt.Println("Updated line 7")
	fmt.Println("Updated line 8")
	fmt.Println("Goodbye!")
}
`
	assert.Equal(t, expectedContent, string(updatedContent), "File content should match expected")
}

func testInsertFileLinesDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create a test file
	testFilePath := filepath.Join(env.GetTestDir(), InsertTestFile)
	err := os.WriteFile(testFilePath, []byte(InsertTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	t.Run("InsertAfterLine", func(t *testing.T) {
		// Insert after line 1 (package declaration)
		result := env.CallTool(t, "insert_file_lines", map[string]interface{}{
			"filepath":    testFilePath,
			"position":    "after",
			"line_number": 1,
			"new_content": "\nimport \"fmt\"",
		})
		assert.NotNil(t, result, "insert_file_lines should return result")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "package main\n\nimport \"fmt\"", "Should contain inserted import")
	})

	t.Run("InsertBeforeLine", func(t *testing.T) {
		// Insert before the last line
		result := env.CallTool(t, "insert_file_lines", map[string]interface{}{
			"filepath":    testFilePath,
			"position":    "before",
			"line_number": 7,
			"new_content": "\t// Added comment",
		})
		assert.NotNil(t, result, "insert_file_lines should return result")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "\t// Added comment\n\tprintln(\"World\")", "Should contain inserted comment before println World")
	})
}

func testInsertAtPatternDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create a test Go file
	testFilePath := filepath.Join(env.GetTestDir(), PatternTestFile)
	err := os.WriteFile(testFilePath, []byte(PatternTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	t.Run("InsertAfterPattern", func(t *testing.T) {
		// Insert import after package declaration
		result := env.CallTool(t, "insert_at_pattern", map[string]interface{}{
			"path":          testFilePath,
			"after_pattern": "package main",
			"position":      "after",
			"content":       "\nimport \"fmt\"",
		})
		assert.NotNil(t, result, "insert_at_pattern should return result")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "package main\n\nimport \"fmt\"", "Should contain inserted import")
	})

	t.Run("InsertBeforePattern", func(t *testing.T) {
		// Insert comment before helper function
		result := env.CallTool(t, "insert_at_pattern", map[string]interface{}{
			"path":           testFilePath,
			"before_pattern": "func helper()",
			"position":       "before",
			"content":        "// Helper function comment",
		})
		assert.NotNil(t, result, "insert_at_pattern should return result")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "// Helper function comment\nfunc helper()", "Should contain inserted comment before function")
	})

	t.Run("RegexPattern", func(t *testing.T) {
		// Use regex to find function declaration
		result := env.CallTool(t, "insert_at_pattern", map[string]interface{}{
			"path":           testFilePath,
			"before_pattern": "func \\w+\\(\\)",
			"position":       "before",
			"content":        "// Function found by regex",
			"regex":          true,
		})
		assert.NotNil(t, result, "insert_at_pattern with regex should return result")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "// Function found by regex", "Should contain regex-inserted comment")
	})
}

func testDeleteFileLinesDirect(t *testing.T, env *DirectServerTestEnv) {
	t.Run("DeleteSingleLine", func(t *testing.T) {
		// Create a test file with numbered lines
		testFilePath := filepath.Join(env.GetTestDir(), "single_"+DeleteTestFile)
		err := os.WriteFile(testFilePath, []byte(DeleteTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		// Delete line 5
		result := env.CallTool(t, "delete_file_lines", map[string]interface{}{
			"filepath":   testFilePath,
			"start_line": 5,
			"end_line":   5,
		})
		assert.NotNil(t, result, "delete_file_lines should return result")

		// Verify the deletion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.NotContains(t, string(updatedContent), "Line 5", "Line 5 should be deleted")
		assert.Contains(t, string(updatedContent), "Line 4", "Line 4 should still exist")
		assert.Contains(t, string(updatedContent), "Line 6", "Line 6 should still exist")
	})

	t.Run("DeleteLineRange", func(t *testing.T) {
		// Create a fresh test file with numbered lines
		testFilePath := filepath.Join(env.GetTestDir(), "range_"+DeleteTestFile)
		err := os.WriteFile(testFilePath, []byte(DeleteTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		// Delete lines 7-9
		result := env.CallTool(t, "delete_file_lines", map[string]interface{}{
			"filepath":   testFilePath,
			"start_line": 7,
			"end_line":   9,
		})
		assert.NotNil(t, result, "delete_file_lines should return result")

		// Verify the deletion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		lines := strings.Split(strings.TrimSpace(string(updatedContent)), "\n")

		// Should not contain lines 7, 8, 9
		for _, line := range lines {
			assert.NotContains(t, line, "Line 7", "Line 7 should be deleted")
			assert.NotContains(t, line, "Line 8", "Line 8 should be deleted")
			assert.NotContains(t, line, "Line 9", "Line 9 should be deleted")
		}

		// Should still contain line 6 and 10
		assert.Contains(t, string(updatedContent), "Line 6", "Line 6 should still exist")
		assert.Contains(t, string(updatedContent), "Line 10", "Line 10 should still exist")
	})
}

func testReplacePatternDirect(t *testing.T, env *DirectServerTestEnv) {
	// Create a test file with patterns to replace
	testFilePath := filepath.Join(env.GetTestDir(), ReplaceTestFile)
	err := os.WriteFile(testFilePath, []byte(ReplaceTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	t.Run("ReplaceAllOccurrences", func(t *testing.T) {
		// Replace all occurrences of "old" with "new"
		result := env.CallTool(t, "replace_pattern", map[string]interface{}{
			"path":            testFilePath,
			"pattern":         "old",
			"replacement":     "new",
			"all_occurrences": true,
		})
		assert.NotNil(t, result, "replace_pattern should return result")

		// Verify the replacements
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "newFunctionName", "Should contain newFunctionName")
		assert.Contains(t, string(updatedContent), "newVariableName", "Should contain newVariableName")
		assert.NotContains(t, string(updatedContent), "oldFunctionName", "Should not contain oldFunctionName")
	})

	t.Run("ReplaceFirstOccurrence", func(t *testing.T) {
		// Reset file content for this test
		testFilePath2 := filepath.Join(env.GetTestDir(), ReplaceFirstFile)
		err := os.WriteFile(testFilePath2, []byte(ReplaceFirstTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		// Replace only first occurrence
		result := env.CallTool(t, "replace_pattern", map[string]interface{}{
			"path":            testFilePath2,
			"pattern":         "test",
			"replacement":     "demo",
			"all_occurrences": false,
		})
		assert.NotNil(t, result, "replace_pattern should return result")

		// Verify only first occurrence was replaced
		updatedContent, err := os.ReadFile(testFilePath2)
		require.NoError(t, err, "Failed to read updated file")

		// Should contain both "demo" and "test"
		assert.Contains(t, string(updatedContent), "func demo()", "Should contain replaced function name")
		assert.Contains(t, string(updatedContent), `"test"`, "Should still contain test strings")
	})

	t.Run("RegexReplace", func(t *testing.T) {
		// Create test content with function declarations
		testFilePath3 := filepath.Join(env.GetTestDir(), RegexReplaceFile)
		err := os.WriteFile(testFilePath3, []byte(RegexReplaceTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		// Use regex to add comments to all function declarations
		result := env.CallTool(t, "replace_pattern", map[string]interface{}{
			"path":            testFilePath3,
			"pattern":         "func (\\w+)\\(",
			"replacement":     "// $1 function\nfunc $1(",
			"regex":           true,
			"all_occurrences": true,
		})
		assert.NotNil(t, result, "replace_pattern with regex should return result")

		// Verify regex replacements
		updatedContent, err := os.ReadFile(testFilePath3)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "// functionOne function", "Should contain comment for functionOne")
		assert.Contains(t, string(updatedContent), "// functionTwo function", "Should contain comment for functionTwo")
		assert.Contains(t, string(updatedContent), "// functionThree function", "Should contain comment for functionThree")
	})
}

// TestGranularEditingErrorCasesDirect tests error handling for granular editing tools
func TestGranularEditingErrorCasesDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("InvalidLineNumbers", func(t *testing.T) {
		err := env.CallToolExpectError(t, "update_file_lines", map[string]interface{}{
			"filepath":    "/nonexistent/file.txt",
			"start_line":  0, // Invalid line number
			"end_line":    5,
			"new_content": "test",
		})
		assert.Error(t, err, "Should return error for invalid line number")
	})

	t.Run("NonAllowedPath", func(t *testing.T) {
		err := env.CallToolExpectError(t, "create_file", map[string]interface{}{
			"filepath":    "/etc/should_not_work.txt",
			"new_content": "This should fail",
		})
		assert.Error(t, err, "Should return error for non-allowed path")
	})

	t.Run("InvalidRegexPattern", func(t *testing.T) {
		testFilePath := filepath.Join(env.GetTestDir(), RegexErrorTestFile)
		err := os.WriteFile(testFilePath, []byte("test content"), 0644)
		require.NoError(t, err, "Failed to create test file")

		err = env.CallToolExpectError(t, "replace_pattern", map[string]interface{}{
			"path":        testFilePath,
			"pattern":     "[invalid regex", // Invalid regex
			"replacement": "test",
			"regex":       true,
		})
		assert.Error(t, err, "Should return error for invalid regex")
	})
}
