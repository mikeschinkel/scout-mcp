package mcptools_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateFileLinesTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("update_file_lines")
	require.NotNil(t, tool, "update_file_lines tool should be registered")

	tool.SetConfig(config)

	t.Run("UpdateSingleLine", func(t *testing.T) {
		// Create test file with multiple lines
		testFile := filepath.Join(tempDir, "update_lines_test.txt")
		content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"filepath":      testFile,
			"start_line":    "2",
			"end_line":      "2",
			"new_content":   "Updated Line 2",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error updating single line")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the update
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "Line 1\nUpdated Line 2\nLine 3\nLine 4\nLine 5\n"
		assert.Equal(t, expected, string(updatedContent), "Line should be updated")
	})

	t.Run("UpdateMultipleLines", func(t *testing.T) {
		// Create test file with multiple lines
		testFile := filepath.Join(tempDir, "update_multi_lines_test.txt")
		content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"filepath":      testFile,
			"start_line":    "2",
			"end_line":      "4",
			"new_content":   "Updated Lines 2-4",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error updating multiple lines")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the update
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "Line 1\nUpdated Lines 2-4\nLine 5\n"
		assert.Equal(t, expected, string(updatedContent), "Lines should be updated")
	})

	t.Run("InvalidLineRange", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"filepath":      testFile,
			"start_line":    "0", // Invalid line number
			"end_line":      "2",
			"content":       "Updated content",
		})

		result, err := testutil.CallTool(tool, req)
		assert.Error(t, err, "Should error with invalid line range")
		assert.Nil(t, result, "Result should be nil on error")
	})
}

func TestInsertFileLinesTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("insert_file_lines")
	require.NotNil(t, tool, "insert_file_lines tool should be registered")

	tool.SetConfig(config)

	t.Run("InsertAfterLine", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "insert_test.txt")
		content := "Line 1\nLine 2\nLine 3\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"filepath":      testFile,
			"position":      "after",
			"line_number":   "1", // String should be parsed as int
			"new_content":   "Inserted after line 1",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should successfully insert after line")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "Line 1\nInserted after line 1\nLine 2\nLine 3\n"
		assert.Equal(t, expected, string(updatedContent), "Content should be inserted after line 1")
	})

	t.Run("InsertBeforeLine", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "insert_before_test.txt")
		content := "Line 1\nLine 2\nLine 3\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"filepath":      testFile,
			"position":      "before",
			"line_number":   "2", // String should be parsed as int
			"new_content":   "Inserted before line 2",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should successfully insert before line")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "Line 1\nInserted before line 2\nLine 2\nLine 3\n"
		assert.Equal(t, expected, string(updatedContent), "Content should be inserted before line 2")
	})
}

func TestDeleteFileLinesTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("delete_file_lines")
	require.NotNil(t, tool, "delete_file_lines tool should be registered")

	tool.SetConfig(config)

	t.Run("DeleteSingleLine", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "delete_test.txt")
		content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"filepath":      testFile,
			"start_line":    "3",
			"end_line":      "3",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error deleting single line")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the deletion
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "Line 1\nLine 2\nLine 4\nLine 5\n"
		assert.Equal(t, expected, string(updatedContent), "Line should be deleted")
	})

	t.Run("DeleteMultipleLines", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "delete_multi_test.txt")
		content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"filepath":      testFile,
			"start_line":    "2",
			"end_line":      "4",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error deleting multiple lines")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the deletion
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "Line 1\nLine 5\n"
		assert.Equal(t, expected, string(updatedContent), "Lines should be deleted")
	})
}

func TestInsertAtPatternTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("insert_at_pattern")
	require.NotNil(t, tool, "insert_at_pattern tool should be registered")

	tool.SetConfig(config)

	t.Run("InsertAfterPattern", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "pattern_test.go")
		content := "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"after_pattern": "func main() {",
			"content":       "\n\t// Added comment",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error inserting after pattern")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the insertion
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		assert.Contains(t, string(updatedContent), "\t// Added comment", "Comment should be inserted")
		assert.Contains(t, string(updatedContent), "func main() {", "Function should still be there")
	})

	t.Run("InsertBeforePattern", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "pattern_before_test.go")
		content := "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":  token,
			"path":           testFile,
			"before_pattern": "func main()",
			"content":        "// Main function\n",
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error inserting before pattern")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the insertion - should be inserted before the function with a blank line
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		assert.Contains(t, string(updatedContent), "// Main function", "Comment should be inserted")
		assert.Contains(t, string(updatedContent), "func main()", "Function should still be there")
	})

	t.Run("RegexPattern", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "regex_pattern_test.go")
		content := "package main\n\nfunc test() {\n\treturn\n}\n\nfunc another() {\n\treturn\n}\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":  token,
			"path":           testFile,
			"before_pattern": "func \\w+\\(\\)",
			"content":        "// Function comment\n",
			"regex":          true,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error with regex pattern")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the insertion (should insert before the first match)
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		assert.Contains(t, string(updatedContent), "// Function comment", "Comment should be inserted")
		assert.Contains(t, string(updatedContent), "func test()", "First function should still be there")
	})
}

func TestReplacePatternTool(t *testing.T) {
	tempDir, cleanup := setupTestEnv(t)
	token := "test-session-token" // Unit tests don't validate tokens
	defer cleanup()

	// Create config that allows our temp directory
	config := testutil.NewMockConfig([]string{tempDir})

	// Get the tool and set config
	tool := mcputil.GetRegisteredTool("replace_pattern")
	require.NotNil(t, tool, "replace_pattern tool should be registered")

	tool.SetConfig(config)

	t.Run("SimpleTextReplace", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "replace_test.txt")
		content := "Hello old world\nThis is old content\nold values everywhere"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":   token,
			"path":            testFile,
			"pattern":         "old",
			"replacement":     "new",
			"all_occurrences": true,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing text")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify the replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "Hello new world\nThis is new content\nnew values everywhere"
		assert.Equal(t, expected, string(updatedContent), "All occurrences should be replaced")
	})

	t.Run("ReplaceFirstOnly", func(t *testing.T) {
		// Create test file
		testFile := filepath.Join(tempDir, "replace_first_test.txt")
		content := "test value\ntest again\ntest final"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":   token,
			"path":            testFile,
			"pattern":         "test",
			"replacement":     "demo",
			"all_occurrences": false,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error replacing first occurrence")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify only first occurrence was replaced
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		expected := "demo value\ntest again\ntest final"
		assert.Equal(t, expected, string(updatedContent), "Only first occurrence should be replaced")
	})

	t.Run("RegexReplace", func(t *testing.T) {
		// Create test file - use a simple text file to avoid Go syntax validation
		testFile := filepath.Join(tempDir, "regex_replace_test.txt")
		content := "func functionOne() {\n\treturn\n}\n\nfunc functionTwo(param string) {\n\treturn\n}\n"
		err := os.WriteFile(testFile, []byte(content), 0644)
		require.NoError(t, err, "Failed to create test file")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token":   token,
			"path":            testFile,
			"pattern":         "func (\\w+)\\(",
			"replacement":     "function $1(",
			"regex":           true,
			"all_occurrences": true,
		})

		result, err := testutil.CallTool(tool, req)
		require.NoError(t, err, "Should not error with regex replacement")
		assert.NotNil(t, result, "Result should not be nil")

		// Verify regex replacement
		updatedContent, err := os.ReadFile(testFile)
		require.NoError(t, err, "Should be able to read updated file")
		assert.Contains(t, string(updatedContent), "function functionOne(", "Should replace functionOne")
		assert.Contains(t, string(updatedContent), "function functionTwo(", "Should replace functionTwo")
	})

	t.Run("InvalidRegex", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")

		req := testutil.NewMockRequest(map[string]interface{}{
			"session_token": token,
			"path":          testFile,
			"pattern":       "[invalid regex",
			"replacement":   "replacement",
			"regex":         true,
		})

		result, err := testutil.CallTool(tool, req)
		assert.Error(t, err, "Should error with invalid regex")
		assert.Nil(t, result, "Result should be nil on error")
	})
}
