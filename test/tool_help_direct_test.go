package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToolHelpDirect tests the help functionality using direct server access
func TestToolHelpDirect(t *testing.T) {
	env := NewDirectServerTestEnv(t)
	defer env.Cleanup()

	t.Run("FullDocumentation", func(t *testing.T) {
		testFullDocumentationDirect(t, env)
	})

	t.Run("SpecificToolHelp", func(t *testing.T) {
		testSpecificToolHelpDirect(t, env)
	})

	t.Run("NonExistentTool", func(t *testing.T) {
		testNonExistentToolDirect(t, env)
	})

	t.Run("DocumentationContent", func(t *testing.T) {
		testDocumentationContentDirect(t, env)
	})
}

func testFullDocumentationDirect(t *testing.T, env *DirectServerTestEnv) {
	// Test getting full documentation without specifying a tool
	result := env.CallTool(t, "help", map[string]interface{}{})

	// Parse the response as JSON and extract content
	var helpResult map[string]interface{}
	ParseJSONResult(t, result, &helpResult)
	content, ok := helpResult["content"].(string)
	require.True(t, ok, "Help result should have content field")
	require.NotEmpty(t, content, "Content should not be empty")

	// Verify it contains the full README content
	assert.Contains(t, content, "# Scout MCP Tools Documentation", "Should contain main title")
	assert.Contains(t, content, "## Session Management", "Should contain Session Management section")
	assert.Contains(t, content, "## File Reading Tools", "Should contain File Reading Tools section")
	assert.Contains(t, content, "## Best Practices", "Should contain Best Practices section")

	// Verify it contains tool descriptions
	assert.Contains(t, content, "### `read_files`", "Should contain read_files documentation")
	assert.Contains(t, content, "### `update_file_lines`", "Should contain update_file_lines documentation")
	assert.Contains(t, content, "### `help`", "Should contain self-documentation")

	// Verify safety warnings are present
	assert.Contains(t, content, "⚠️ DANGEROUS", "Should contain danger warnings")
	assert.Contains(t, content, "Use granular editing tools", "Should recommend safer alternatives")
}

func testSpecificToolHelpDirect(t *testing.T, env *DirectServerTestEnv) {
	// Test cases for specific tool documentation
	testCases := []struct {
		toolName         string
		shouldContain    []string
		shouldNotContain []string
		description      string
	}{
		{
			toolName: "read_files",
			shouldContain: []string{
				"### `read_files`",
				"Read contents of multiple files",
				"paths` (required)",
				"\"tool\": \"read_files\"",
			},
			shouldNotContain: []string{
				"### `update_file`",
				"### `create_file`",
				"## Best Practices", // Should not include full doc sections
			},
			description: "read_files specific help",
		},
		{
			toolName: "update_file",
			shouldContain: []string{
				"### `update_file`",
				"⚠️ DANGEROUS",
				"Replaces entire file content",
				"Use granular editing tools",
				"update_file_lines",
			},
			shouldNotContain: []string{
				"### `read_file`",
				"## File Management Tools",
			},
			description: "update_file specific help with warnings",
		},
		{
			toolName: "update_file_lines",
			shouldContain: []string{
				"### `update_file_lines`",
				"Update specific lines",
				"start_line` (required)",
				"end_line` (required)",
				"Much safer than `update_file`",
			},
			shouldNotContain: []string{
				"### `insert_file_lines`",
				"## Configuration Tools",
			},
			description: "update_file_lines specific help",
		},
		{
			toolName: "replace_pattern",
			shouldContain: []string{
				"### `replace_pattern`",
				"Find and replace text patterns",
				"regex` (optional)",
				"all_occurrences` (optional)",
				"**Regex Example:**",
			},
			shouldNotContain: []string{
				"### `delete_file_lines`",
				"## Best Practices",
			},
			description: "replace_pattern specific help",
		},
		{
			toolName: "search_files",
			shouldContain: []string{
				"### `search_files`",
				"Search for files and directories",
				"recursive` (optional)",
				"extensions` (optional)",
				"pattern` (optional)",
			},
			shouldNotContain: []string{
				"### `get_config`",
				"## Granular File Editing Tools",
			},
			description: "search_files specific help",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			result := env.CallTool(t, "help", map[string]interface{}{
				"tool": tc.toolName,
			})
			require.NotNil(t, result, "help should return result for %s", tc.toolName)

			var helpResult map[string]interface{}
			ParseJSONResult(t, result, &helpResult)
			content, ok := helpResult["content"].(string)
			require.True(t, ok, "Help result should have content field for %s", tc.toolName)
			require.NotEmpty(t, content, "Content should not be empty for %s", tc.toolName)

			// Check that required content is present
			for _, shouldContain := range tc.shouldContain {
				assert.Contains(t, content, shouldContain,
					"Help for %s should contain: %s", tc.toolName, shouldContain)
			}

			// Check that other tool content is not present (tool-specific only)
			for _, shouldNotContain := range tc.shouldNotContain {
				assert.NotContains(t, content, shouldNotContain,
					"Help for %s should not contain: %s", tc.toolName, shouldNotContain)
			}

			// Verify the content is not empty and contains the tool name
			assert.NotEmpty(t, content, "Help content should not be empty for %s", tc.toolName)
			assert.True(t, len(content) > 100, "Help content should be substantial for %s", tc.toolName)
		})
	}
}

func testNonExistentToolDirect(t *testing.T, env *DirectServerTestEnv) {
	// Test requesting help for a non-existent tool
	result := env.CallTool(t, "help", map[string]interface{}{
		"tool": "nonexistent_tool",
	})
	require.NotNil(t, result, "help should return result")

	var helpResult map[string]interface{}
	ParseJSONResult(t, result, &helpResult)
	content, ok := helpResult["content"].(string)
	require.True(t, ok, "Help result should have content field")
	require.NotEmpty(t, content, "Content should not be empty")

	// Verify helpful error message
	assert.Contains(t, content, "Tool 'nonexistent_tool' not found", "Should indicate tool not found")
	assert.Contains(t, content, "Available tools:", "Should list available tools")

	// Verify it lists actual available tools
	expectedTools := []string{
		"read_files", "create_file", "update_file", "delete_files",
		"search_files", "update_file_lines", "insert_file_lines",
		"replace_pattern", "get_config", "help", "detect_current_project",
	}

	for _, tool := range expectedTools {
		assert.Contains(t, content, tool, "Should list available tool: %s", tool)
	}

	// Verify usage instructions
	assert.Contains(t, content, "Call help without parameters", "Should provide usage instructions")
	assert.Contains(t, content, `"tool": "help"`, "Should show example usage")
}

func testDocumentationContentDirect(t *testing.T, env *DirectServerTestEnv) {
	// Test that the documentation contains important safety information
	result := env.CallTool(t, "help", map[string]interface{}{})
	require.NotNil(t, result, "help should return result")

	var helpResult map[string]interface{}
	ParseJSONResult(t, result, &helpResult)
	content, ok := helpResult["content"].(string)
	require.True(t, ok, "Help result should have content field")
	require.NotEmpty(t, content, "Content should not be empty")

	t.Run("SafetyWarnings", func(t *testing.T) {
		// Verify important safety warnings are present
		assert.Contains(t, content, "⚠️ DANGEROUS", "Should contain danger warnings")
		assert.Contains(t, content, "RECOMMENDED: Use these tools for precise code editing",
			"Should recommend granular tools")
		assert.Contains(t, content, "Much safer than `update_file`",
			"Should emphasize safety of granular tools")
	})

	t.Run("ToolCategories", func(t *testing.T) {
		// Verify all tool categories are documented
		categories := []string{
			"## Session Management",
			"## File Reading Tools",
			"## File Management Tools",
			"## Granular File Editing Tools",
			"## Language-Aware Tools",
			"## Analysis Tools",
			"## Best Practices",
		}

		for _, category := range categories {
			assert.Contains(t, content, category, "Should contain category: %s", category)
		}
	})

	t.Run("ExampleFormat", func(t *testing.T) {
		// Verify JSON examples are properly formatted
		assert.Contains(t, content, `"tool":`, "Should contain JSON examples")
		assert.Contains(t, content, `"parameters":`, "Should show parameter structure")
		assert.Contains(t, content, `"path"`, "Should show common parameters")

		// Count JSON example blocks - should have multiple
		jsonBlocks := strings.Count(content, "```json")
		assert.Greater(t, jsonBlocks, 10, "Should contain multiple JSON examples")
	})

	t.Run("BestPractices", func(t *testing.T) {
		// Verify best practices section contains guidance
		assert.Contains(t, content, "## Best Practices", "Should provide tool selection guidance")
		assert.Contains(t, content, "Common Patterns", "Should provide usage patterns")
		assert.Contains(t, content, "Adding an import to a Go file", "Should show practical examples")
		assert.Contains(t, content, "Refactoring variable names", "Should show refactoring examples")
	})

	t.Run("AllToolsCovered", func(t *testing.T) {
		// Verify all major tools are documented
		majorTools := []string{
			"`read_files`", "`create_file`", "`update_file`", "`delete_files`",
			"`search_files`", "`update_file_lines`", "`insert_file_lines`",
			"`insert_at_pattern`", "`delete_file_lines`", "`replace_pattern`",
			"`get_config`",
		}

		for _, tool := range majorTools {
			assert.Contains(t, content, tool, "Should document tool: %s", tool)
		}
	})
}
