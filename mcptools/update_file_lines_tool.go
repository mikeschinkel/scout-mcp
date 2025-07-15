package mcptools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*UpdateFileLinesTool)(nil)

func init() {
	mcputil.RegisterTool(&UpdateFileLinesTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "update_file_lines",
			Description: "Update specific lines in a file by line number range",
		}),
	})
}

type UpdateFileLinesTool struct {
	*toolBase
}

func (t *UpdateFileLinesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var startLine, endLine int
	var newContent string

	logger.Info("Tool called", "tool", "update_file_lines")

	filePath, err = req.RequireString("path")
	if err != nil {
		goto end
	}

	newContent, err = req.RequireString("content")
	if err != nil {
		goto end
	}

	startLine, err = getNumberAsInt(req, "start_line")
	if err != nil {
		err = fmt.Errorf("start_line must be a valid number: %w", err)
		goto end
	}

	endLine, err = getNumberAsInt(req, "end_line")
	if err != nil {
		err = fmt.Errorf("end_line must be a valid number: %w", err)
		goto end
	}

	err = t.validateLineRange(startLine, endLine)
	if err != nil {
		goto end
	}

	err = t.updateFileLines(filePath, startLine, endLine, newContent)
	if err != nil {
		goto end
	}

	result = mcputil.NewToolResultText(fmt.Sprintf("Successfully updated lines %d-%d in %s", startLine, endLine, filePath))
	logger.Info("Tool completed", "tool", "update_file_lines", "path", filePath, "start_line", startLine, "end_line", endLine)

end:
	return result, err
}

func (t *UpdateFileLinesTool) validateLineRange(startLine, endLine int) (err error) {
	if startLine < 1 {
		err = fmt.Errorf("start_line must be >= 1, got %d", startLine)
		goto end
	}

	if endLine < startLine {
		err = fmt.Errorf("end_line (%d) must be >= start_line (%d)", endLine, startLine)
		goto end
	}

end:
	return err
}

func (t *UpdateFileLinesTool) updateFileLines(filePath string, startLine, endLine int, newContent string) (err error) {
	var allowed bool
	var originalContent string
	var lines []string
	var updatedContent string

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

	lines = strings.Split(originalContent, "\n")

	err = t.validateLineNumbers(lines, startLine, endLine)
	if err != nil {
		goto end
	}

	updatedContent = t.replaceLines(lines, startLine, endLine, newContent)

	err = writeFile(t.Config(), filePath, updatedContent)

end:
	return err
}

func (t *UpdateFileLinesTool) validateLineNumbers(lines []string, startLine, endLine int) (err error) {
	totalLines := len(lines)

	if startLine > totalLines {
		err = fmt.Errorf("start_line %d exceeds file length %d", startLine, totalLines)
		goto end
	}

	if endLine > totalLines {
		err = fmt.Errorf("end_line %d exceeds file length %d", endLine, totalLines)
		goto end
	}

end:
	return err
}

func (t *UpdateFileLinesTool) replaceLines(lines []string, startLine, endLine int, newContent string) (result string) {
	var before, after []string
	var newLines []string
	var combined []string

	// Convert to 0-based indexing
	startIdx := startLine - 1
	endIdx := endLine - 1

	before = lines[:startIdx]
	after = lines[endIdx+1:]

	newLines = strings.Split(newContent, "\n")

	combined = make([]string, 0, len(before)+len(newLines)+len(after))
	combined = append(combined, before...)
	combined = append(combined, newLines...)
	combined = append(combined, after...)

	result = strings.Join(combined, "\n")

	return result
}
