package mcptools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*DeleteFileLinesTool)(nil)

func init() {
	mcputil.RegisterTool(&DeleteFileLinesTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "delete_file_lines",
			Description: "Delete specific lines from a file by line number range",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				FilepathProperty.Required(),
				StartLineProperty.Required(),
				EndLineProperty.Required(),
			},
		}),
	})
}

type DeleteFileLinesTool struct {
	*mcputil.ToolBase
}

func (t *DeleteFileLinesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var startLine, endLine int
	var message string

	logger.Info("Tool called", "tool", "delete_file_lines")

	filePath, err = FilepathProperty.String(req)
	if err != nil {
		goto end
	}

	startLine, err = StartLineProperty.Int(req)
	if err != nil {
		err = fmt.Errorf("start_line must be a valid number: %w", err)
		goto end
	}

	endLine, err = EndLineProperty.Int(req)
	if err != nil {
		err = fmt.Errorf("end_line must be a valid number: %w", err)
		goto end
	}

	err = t.validateLineRange(startLine, endLine)
	if err != nil {
		goto end
	}

	err = t.deleteFileLines(filePath, startLine, endLine)
	if err != nil {
		goto end
	}

	if startLine == endLine {
		message = fmt.Sprintf("Successfully deleted line %d from %s", startLine, filePath)
	} else {
		message = fmt.Sprintf("Successfully deleted lines %d-%d from %s", startLine, endLine, filePath)
	}

	result = mcputil.NewToolResultJSON(map[string]interface{}{
		"success":       true,
		"file_path":     filePath,
		"start_line":    startLine,
		"end_line":      endLine,
		"lines_deleted": endLine - startLine + 1,
		"message":       message,
	})

	logger.Info("Tool completed", "tool", "delete_file_lines", "path", filePath, "start_line", startLine, "end_line", endLine)

end:
	return result, err
}

func (t *DeleteFileLinesTool) validateLineRange(startLine, endLine int) (err error) {
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

func (t *DeleteFileLinesTool) deleteFileLines(filePath string, startLine, endLine int) (err error) {
	var originalContent string
	var lines []string
	var updatedContent string

	if !t.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	originalContent, err = ReadFile(t.Config(), filePath)
	if err != nil {
		goto end
	}

	lines = strings.Split(originalContent, "\n")

	err = t.validateLineNumbers(lines, startLine, endLine)
	if err != nil {
		goto end
	}

	updatedContent = t.removeLines(lines, startLine, endLine)

	err = WriteFile(t.Config(), filePath, updatedContent)

end:
	return err
}

func (t *DeleteFileLinesTool) validateLineNumbers(lines []string, startLine, endLine int) (err error) {
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

func (t *DeleteFileLinesTool) removeLines(lines []string, startLine, endLine int) (result string) {
	var before, after []string
	var combined []string

	// Convert to 0-based indexing
	startIdx := startLine - 1
	endIdx := endLine - 1

	before = lines[:startIdx]
	after = lines[endIdx+1:]

	combined = make([]string, 0, len(before)+len(after))
	combined = append(combined, before...)
	combined = append(combined, after...)

	result = strings.Join(combined, "\n")

	return result
}
