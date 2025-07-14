package test

import (
	"context"
	"encoding/json"
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

	ReplaceFirstTestContent = `func test() {
	value := "test"
	anotherValue := "test"
	finalValue := "test"
}
`

	RegexReplaceTestContent = `func functionOne() {
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

// TestGranularFileEditing tests all the granular file editing tools
func TestGranularFileEditing(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()
	testDir := GetTestDir()

	t.Run("UpdateFileLines", func(t *testing.T) {
		testUpdateFileLines(t, client, ctx, testDir)
	})

	t.Run("InsertFileLines", func(t *testing.T) {
		testInsertFileLines(t, client, ctx, testDir)
	})

	t.Run("InsertAtPattern", func(t *testing.T) {
		testInsertAtPattern(t, client, ctx, testDir)
	})

	t.Run("DeleteFileLines", func(t *testing.T) {
		testDeleteFileLines(t, client, ctx, testDir)
	})

	t.Run("ReplacePattern", func(t *testing.T) {
		testReplacePattern(t, client, ctx, testDir)
	})
}

func testUpdateFileLines(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create a test file with multiple lines
	testFilePath := filepath.Join(testDir, UpdateTestFile)
	err := os.WriteFile(testFilePath, []byte(UpdateTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	// Test updating lines 7-8
	resp, err := client.CallTool(ctx, "update_file_lines", map[string]interface{}{
		"path":       testFilePath,
		"start_line": "7",
		"end_line":   "8",
		"content":    "\tfmt.Println(\"Updated line 7\")\n\tfmt.Println(\"Updated line 8\")",
	})
	require.NoError(t, err, "Failed to call update_file_lines")
	require.Nil(t, resp.Error, "update_file_lines returned error: %v", resp.Error)

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

func testInsertFileLines(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create a test file
	testFilePath := filepath.Join(testDir, InsertTestFile)
	err := os.WriteFile(testFilePath, []byte(InsertTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	t.Run("InsertAfterLine", func(t *testing.T) {
		// Insert after line 1 (package declaration)
		resp, err := client.CallTool(ctx, "insert_file_lines", map[string]interface{}{
			"path":        testFilePath,
			"line_number": "1",
			"content":     "\nimport \"fmt\"",
			"position":    "after",
		})
		require.NoError(t, err, "Failed to call insert_file_lines")
		require.Nil(t, resp.Error, "insert_file_lines returned error: %v", resp.Error)

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "package main\n\nimport \"fmt\"", "Should contain inserted import")
	})

	t.Run("InsertBeforeLine", func(t *testing.T) {
		// Insert before the last line
		resp, err := client.CallTool(ctx, "insert_file_lines", map[string]interface{}{
			"path":        testFilePath,
			"line_number": "7", // Before the closing brace
			"content":     "\t// Added comment",
			"position":    "before",
		})
		require.NoError(t, err, "Failed to call insert_file_lines")
		require.Nil(t, resp.Error, "insert_file_lines returned error: %v", resp.Error)

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "\t// Added comment\n}", "Should contain inserted comment before closing brace")
	})
}

func testInsertAtPattern(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create a test Go file
	testFilePath := filepath.Join(testDir, PatternTestFile)
	err := os.WriteFile(testFilePath, []byte(PatternTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	t.Run("InsertAfterPattern", func(t *testing.T) {
		// Insert import after package declaration
		resp, err := client.CallTool(ctx, "insert_at_pattern", map[string]interface{}{
			"path":          testFilePath,
			"after_pattern": "package main",
			"content":       "\nimport \"fmt\"",
			"position":      "after",
		})
		require.NoError(t, err, "Failed to call insert_at_pattern")
		require.Nil(t, resp.Error, "insert_at_pattern returned error: %v", resp.Error)

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "package main\n\nimport \"fmt\"", "Should contain inserted import")
	})

	t.Run("InsertBeforePattern", func(t *testing.T) {
		// Insert comment before helper function
		resp, err := client.CallTool(ctx, "insert_at_pattern", map[string]interface{}{
			"path":           testFilePath,
			"before_pattern": "func helper()",
			"content":        "// Helper function comment",
			"position":       "before",
		})
		require.NoError(t, err, "Failed to call insert_at_pattern")
		require.Nil(t, resp.Error, "insert_at_pattern returned error: %v", resp.Error)

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "// Helper function comment\nfunc helper()", "Should contain inserted comment before function")
	})

	t.Run("RegexPattern", func(t *testing.T) {
		// Use regex to find function declaration
		resp, err := client.CallTool(ctx, "insert_at_pattern", map[string]interface{}{
			"path":           testFilePath,
			"before_pattern": "func \\w+\\(\\)",
			"content":        "// Function found by regex",
			"position":       "before",
			"regex":          true,
		})
		require.NoError(t, err, "Failed to call insert_at_pattern with regex")
		require.Nil(t, resp.Error, "insert_at_pattern with regex returned error: %v", resp.Error)

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "// Function found by regex", "Should contain regex-inserted comment")
	})
}

func testDeleteFileLines(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create a test file with numbered lines
	testFilePath := filepath.Join(testDir, DeleteTestFile)
	err := os.WriteFile(testFilePath, []byte(DeleteTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	t.Run("DeleteSingleLine", func(t *testing.T) {
		// Delete line 5
		resp, err := client.CallTool(ctx, "delete_file_lines", map[string]interface{}{
			"path":       testFilePath,
			"start_line": "5",
		})
		require.NoError(t, err, "Failed to call delete_file_lines")
		require.Nil(t, resp.Error, "delete_file_lines returned error: %v", resp.Error)

		// Verify the deletion
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.NotContains(t, string(updatedContent), "Line 5", "Line 5 should be deleted")
		assert.Contains(t, string(updatedContent), "Line 4", "Line 4 should still exist")
		assert.Contains(t, string(updatedContent), "Line 6", "Line 6 should still exist")
	})

	t.Run("DeleteLineRange", func(t *testing.T) {
		// Delete lines 7-9
		resp, err := client.CallTool(ctx, "delete_file_lines", map[string]interface{}{
			"path":       testFilePath,
			"start_line": "7",
			"end_line":   "9",
		})
		require.NoError(t, err, "Failed to call delete_file_lines")
		require.Nil(t, resp.Error, "delete_file_lines returned error: %v", resp.Error)

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

func testReplacePattern(t *testing.T, client *MCPClient, ctx context.Context, testDir string) {
	// Create a test file with patterns to replace
	testFilePath := filepath.Join(testDir, ReplaceTestFile)
	err := os.WriteFile(testFilePath, []byte(ReplaceTestContent), 0644)
	require.NoError(t, err, "Failed to create test file")

	t.Run("ReplaceAllOccurrences", func(t *testing.T) {
		// Replace all occurrences of "old" with "new"
		resp, err := client.CallTool(ctx, "replace_pattern", map[string]interface{}{
			"path":            testFilePath,
			"pattern":         "old",
			"replacement":     "new",
			"all_occurrences": true,
		})
		require.NoError(t, err, "Failed to call replace_pattern")
		require.Nil(t, resp.Error, "replace_pattern returned error: %v", resp.Error)

		// Parse the response to check replacement count
		var result map[string]interface{}
		err = json.Unmarshal(resp.Result, &result)
		require.NoError(t, err, "Failed to parse replace_pattern response")

		content := result["content"].([]interface{})
		contentItem := content[0].(map[string]interface{})
		message := contentItem["text"].(string)

		assert.Contains(t, message, "replaced", "Should indicate replacements were made")

		// Verify the replacements
		updatedContent, err := os.ReadFile(testFilePath)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "newFunctionName", "Should contain newFunctionName")
		assert.Contains(t, string(updatedContent), "newVariableName", "Should contain newVariableName")
		assert.NotContains(t, string(updatedContent), "oldFunctionName", "Should not contain oldFunctionName")
	})

	t.Run("ReplaceFirstOccurrence", func(t *testing.T) {
		// Reset file content for this test
		testFilePath2 := filepath.Join(testDir, ReplaceFirstFile)
		err := os.WriteFile(testFilePath2, []byte(ReplaceFirstTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		// Replace only first occurrence
		resp, err := client.CallTool(ctx, "replace_pattern", map[string]interface{}{
			"path":            testFilePath2,
			"pattern":         "test",
			"replacement":     "demo",
			"all_occurrences": false,
		})
		require.NoError(t, err, "Failed to call replace_pattern")
		require.Nil(t, resp.Error, "replace_pattern returned error: %v", resp.Error)

		// Verify only first occurrence was replaced
		updatedContent, err := os.ReadFile(testFilePath2)
		require.NoError(t, err, "Failed to read updated file")

		// Should contain both "demo" and "test"
		assert.Contains(t, string(updatedContent), "func demo()", "Should contain replaced function name")
		assert.Contains(t, string(updatedContent), `"test"`, "Should still contain test strings")
	})

	t.Run("RegexReplace", func(t *testing.T) {
		// Create test content with function declarations
		testFilePath3 := filepath.Join(testDir, RegexReplaceFile)
		err := os.WriteFile(testFilePath3, []byte(RegexReplaceTestContent), 0644)
		require.NoError(t, err, "Failed to create test file")

		// Use regex to add comments to all function declarations
		resp, err := client.CallTool(ctx, "replace_pattern", map[string]interface{}{
			"path":            testFilePath3,
			"pattern":         "func (\\w+)\\(",
			"replacement":     "// $1 function\nfunc $1(",
			"regex":           true,
			"all_occurrences": true,
		})
		require.NoError(t, err, "Failed to call replace_pattern with regex")
		require.Nil(t, resp.Error, "replace_pattern with regex returned error: %v", resp.Error)

		// Verify regex replacements
		updatedContent, err := os.ReadFile(testFilePath3)
		require.NoError(t, err, "Failed to read updated file")

		assert.Contains(t, string(updatedContent), "// functionOne function", "Should contain comment for functionOne")
		assert.Contains(t, string(updatedContent), "// functionTwo function", "Should contain comment for functionTwo")
		assert.Contains(t, string(updatedContent), "// functionThree function", "Should contain comment for functionThree")
	})
}

// TestGranularEditingErrorCases tests error handling for granular editing tools
func TestGranularEditingErrorCases(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	t.Run("InvalidLineNumbers", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "update_file_lines", map[string]interface{}{
			"path":       "/nonexistent/file.txt",
			"start_line": "0", // Invalid line number
			"end_line":   "5",
			"content":    "test",
		})
		require.NoError(t, err, "Failed to call update_file_lines")
		assert.NotNil(t, resp.Error, "Should return error for invalid line number")
	})

	t.Run("NonAllowedPath", func(t *testing.T) {
		resp, err := client.CallTool(ctx, "create_file", map[string]interface{}{
			"path":    "/etc/should_not_work.txt",
			"content": "This should fail",
		})
		require.NoError(t, err, "Failed to call create_file")
		assert.NotNil(t, resp.Error, "Should return error for non-allowed path")
	})

	t.Run("InvalidRegexPattern", func(t *testing.T) {
		testDir := GetTestDir()
		testFilePath := filepath.Join(testDir, RegexErrorTestFile)
		err := os.WriteFile(testFilePath, []byte("test content"), 0644)
		require.NoError(t, err, "Failed to create test file")

		resp, err := client.CallTool(ctx, "replace_pattern", map[string]interface{}{
			"path":        testFilePath,
			"pattern":     "[invalid regex", // Invalid regex
			"replacement": "test",
			"regex":       true,
		})
		require.NoError(t, err, "Failed to call replace_pattern")
		assert.NotNil(t, resp.Error, "Should return error for invalid regex")
	})
}
