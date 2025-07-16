package test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToolHelp tests the tool_help functionality
func TestToolHelp(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	t.Run("FullDocumentation", func(t *testing.T) {
		testFullDocumentation(t, client, ctx)
	})

	t.Run("SpecificToolHelp", func(t *testing.T) {
		testSpecificToolHelp(t, client, ctx)
	})

	t.Run("NonExistentTool", func(t *testing.T) {
		testNonExistentTool(t, client, ctx)
	})

	t.Run("DocumentationContent", func(t *testing.T) {
		testDocumentationContent(t, client, ctx)
	})
}

func testFullDocumentation(t *testing.T, client *MCPClient, ctx context.Context) {
	// Test getting full documentation without specifying a tool
	resp, err := client.CallTool(ctx, "tool_help", map[string]interface{}{})
	require.NoError(t, err, "Failed to call tool_help")
	require.Nil(t, resp.Error, "tool_help returned error: %v", resp.Error)

	// Parse the response
	var content string
	parseToolResponse(t, resp, &content)

	// Verify it contains the full README content
	assert.Contains(t, content, "# Scout MCP Tools Documentation", "Should contain main title")
	assert.Contains(t, content, "## File Management Tools", "Should contain File Management section")
	assert.Contains(t, content, "## Granular File Editing Tools", "Should contain Granular Editing section")
	assert.Contains(t, content, "## Best Practices", "Should contain Best Practices section")

	// Verify it contains tool descriptions
	assert.Contains(t, content, "### `read_file`", "Should contain read_file documentation")
	assert.Contains(t, content, "### `update_file_lines`", "Should contain update_file_lines documentation")
	assert.Contains(t, content, "### `tool_help`", "Should contain self-documentation")

	// Verify safety warnings are present
	assert.Contains(t, content, "⚠️ DANGEROUS", "Should contain danger warnings")
	assert.Contains(t, content, "Use granular editing tools", "Should recommend safer alternatives")
}

func testSpecificToolHelp(t *testing.T, client *MCPClient, ctx context.Context) {
	// Test cases for specific tool documentation
	testCases := []struct {
		toolName         string
		shouldContain    []string
		shouldNotContain []string
		description      string
	}{
		{
			toolName: "read_file",
			shouldContain: []string{
				"### `read_file`",
				"Reads the contents of a file",
				"path` (required)",
				"\"tool\": \"read_file\"",
			},
			shouldNotContain: []string{
				"### `update_file`",
				"### `create_file`",
				"## Best Practices", // Should not include full doc sections
			},
			description: "read_file specific help",
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
			resp, err := client.CallTool(ctx, "tool_help", map[string]interface{}{
				"tool": tc.toolName,
			})
			require.NoError(t, err, "Failed to call tool_help for %s", tc.toolName)
			require.Nil(t, resp.Error, "tool_help returned error for %s: %v", tc.toolName, resp.Error)

			var content string
			parseToolResponse(t, resp, &content)

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

func testNonExistentTool(t *testing.T, client *MCPClient, ctx context.Context) {
	// Test requesting help for a non-existent tool
	resp, err := client.CallTool(ctx, "tool_help", map[string]interface{}{
		"tool": "nonexistent_tool",
	})
	require.NoError(t, err, "Failed to call tool_help")
	require.Nil(t, resp.Error, "tool_help returned error: %v", resp.Error)

	var content string
	parseToolResponse(t, resp, &content)

	// Verify helpful error message
	assert.Contains(t, content, "Tool 'nonexistent_tool' not found", "Should indicate tool not found")
	assert.Contains(t, content, "Available tools:", "Should list available tools")

	// Verify it lists actual available tools
	expectedTools := []string{
		"read_file", "create_file", "update_file", "delete_files",
		"search_files", "update_file_lines", "insert_file_lines",
		"replace_pattern", "get_config", "tool_help",
	}

	for _, tool := range expectedTools {
		assert.Contains(t, content, tool, "Should list available tool: %s", tool)
	}

	// Verify usage instructions
	assert.Contains(t, content, "Call tool_help without parameters", "Should provide usage instructions")
	assert.Contains(t, content, `"tool": "tool_help"`, "Should show example usage")
}

func testDocumentationContent(t *testing.T, client *MCPClient, ctx context.Context) {
	// Test that the documentation contains important safety information
	resp, err := client.CallTool(ctx, "tool_help", map[string]interface{}{})
	require.NoError(t, err, "Failed to call tool_help")
	require.Nil(t, resp.Error, "tool_help returned error: %v", resp.Error)

	var content string
	parseToolResponse(t, resp, &content)

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
			"## File Management Tools",
			"## File Search Tools",
			"## Granular File Editing Tools",
			"## Configuration Tools",
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
		assert.Contains(t, content, "When to Use Each Tool", "Should provide tool selection guidance")
		assert.Contains(t, content, "Common Patterns", "Should provide usage patterns")
		assert.Contains(t, content, "Adding an import to a Go file", "Should show practical examples")
		assert.Contains(t, content, "Refactoring variable names", "Should show refactoring examples")
	})

	t.Run("AllToolsCovered", func(t *testing.T) {
		// Verify all major tools are documented
		majorTools := []string{
			"`read_file`", "`create_file`", "`update_file`", "`delete_files`",
			"`search_files`", "`update_file_lines`", "`insert_file_lines`",
			"`insert_at_pattern`", "`delete_file_lines`", "`replace_pattern`",
			"`get_config`",
		}

		for _, tool := range majorTools {
			assert.Contains(t, content, tool, "Should document tool: %s", tool)
		}
	})
}

// TestToolHelpIntegration tests tool_help integration with the test suite
func TestToolHelpIntegration(t *testing.T) {
	client := GetTestClient()
	ctx := GetTestContext()

	// Verify tool_help appears in tools list
	resp, err := client.ListTools(ctx)
	require.NoError(t, err, "Failed to list tools")
	require.Nil(t, resp.Error, "ListTools returned error: %v", resp.Error)

	var result map[string]interface{}
	err = json.Unmarshal(resp.Result, &result)
	require.NoError(t, err, "Failed to parse tools list response")

	tools, ok := result["tools"].([]interface{})
	require.True(t, ok, "Tools should be an array")

	// Check that tool_help is in the list
	foundToolHelp := false
	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		require.True(t, ok, "Tool should be an object")

		name, ok := toolMap["name"].(string)
		require.True(t, ok, "Tool should have a name")

		if name == "tool_help" {
			foundToolHelp = true

			// Verify tool description
			description, ok := toolMap["description"].(string)
			assert.True(t, ok, "tool_help should have a description")
			assert.Contains(t, description, "documentation", "Description should mention documentation")
			break
		}
	}

	assert.True(t, foundToolHelp, "tool_help should be available in tools list")
}
