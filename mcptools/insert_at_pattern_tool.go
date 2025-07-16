package mcptools

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*InsertAtPatternTool)(nil)

func init() {
	mcputil.RegisterTool(&InsertAtPatternTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "insert_at_pattern",
			Description: "Insert content before or after a code pattern match",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
			},
		}),
	})
}

type InsertAtPatternTool struct {
	*toolBase
}

func (t *InsertAtPatternTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var beforePattern string
	var afterPattern string
	var content string
	var position string
	var useRegex bool

	logger.Info("Tool called", "tool", "insert_at_pattern")

	filePath, err = req.RequireString("path")
	if err != nil {
		goto end
	}
	content, err = req.RequireString("content")
	if err != nil {
		goto end
	}

	// TODO: Fix these to have errors after we fix ToolRequest interface
	beforePattern = req.GetString("before_pattern", "")
	afterPattern = req.GetString("after_pattern", "")
	position = req.GetString("position", "before")
	useRegex = req.GetBool("regex", false)

	err = t.validatePatterns(beforePattern, afterPattern)
	if err != nil {
		goto end
	}

	err = t.validatePosition(position)
	if err != nil {
		goto end
	}

	err = t.insertAtPattern(filePath, beforePattern, afterPattern, content, position, useRegex)
	if err != nil {
		goto end
	}

	result = mcputil.NewToolResultText(fmt.Sprintf("Successfully inserted content at pattern in %s", filePath))
	logger.Info("Tool completed", "tool", "insert_at_pattern", "path", filePath)

end:
	return result, err
}

func (t *InsertAtPatternTool) validatePatterns(beforePattern, afterPattern string) (err error) {
	if beforePattern == "" && afterPattern == "" {
		err = fmt.Errorf("either before_pattern or after_pattern must be specified")
		goto end
	}

	if beforePattern != "" && afterPattern != "" {
		err = fmt.Errorf("only one of before_pattern or after_pattern should be specified")
		goto end
	}

end:
	return err
}

func (t *InsertAtPatternTool) validatePosition(position string) (err error) {
	return RelativePosition(position).Validate()
}

func (t *InsertAtPatternTool) insertAtPattern(filePath, beforePattern, afterPattern, content, position string, useRegex bool) (err error) {
	var allowed bool
	var originalContent string
	var updatedContent string
	var pattern string

	allowed, err = t.IsAllowedPath(filePath)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	originalContent, err = readFile(t.Config(), filePath)
	if err != nil {
		goto end
	}

	if beforePattern != "" {
		pattern = beforePattern
	} else {
		pattern = afterPattern
	}

	updatedContent, err = t.insertContentAtPattern(originalContent, pattern, content, position, useRegex)
	if err != nil {
		goto end
	}

	err = writeFile(t.Config(), filePath, updatedContent)

end:
	return err
}

func (t *InsertAtPatternTool) insertContentAtPattern(originalContent, pattern, content, position string, useRegex bool) (result string, err error) {
	var lines []string
	var matchLine int
	var found bool

	lines = strings.Split(originalContent, "\n")

	matchLine, found, err = t.findPatternLine(lines, pattern, useRegex)
	if err != nil {
		goto end
	}

	if !found {
		err = fmt.Errorf("pattern not found: %s", pattern)
		goto end
	}

	result = t.insertAtLineNumber(lines, matchLine, content, position)

end:
	return result, err
}

func (t *InsertAtPatternTool) findPatternLine(lines []string, pattern string, useRegex bool) (lineNumber int, found bool, err error) {
	var re *regexp.Regexp

	if useRegex {
		re, err = regexp.Compile(pattern)
		if err != nil {
			err = fmt.Errorf("invalid regex pattern: %w", err)
			goto end
		}
	}

	for i, line := range lines {
		var matches bool

		if useRegex {
			matches = re.MatchString(line)
		} else {
			matches = strings.Contains(line, pattern)
		}

		if matches {
			lineNumber = i + 1 // Convert to 1-based
			found = true
			goto end
		}
	}

end:
	return lineNumber, found, err
}

func (t *InsertAtPatternTool) insertAtLineNumber(lines []string, lineNumber int, content, position string) (result string) {
	var insertIdx int
	var newLines []string
	var combined []string

	// Convert to 0-based indexing
	baseIdx := lineNumber - 1

	if position == "before" {
		insertIdx = baseIdx
	} else {
		insertIdx = baseIdx + 1
	}

	newLines = strings.Split(content, "\n")

	combined = make([]string, 0, len(lines)+len(newLines))
	combined = append(combined, lines[:insertIdx]...)
	combined = append(combined, newLines...)
	combined = append(combined, lines[insertIdx:]...)

	result = strings.Join(combined, "\n")

	return result
}
