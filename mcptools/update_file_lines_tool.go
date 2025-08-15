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
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "update_file_lines",
			Description: "Update specific lines in a file by line number range",
			QuickHelp:   "Edit specific line ranges safely",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				FilepathProperty.Required(),
				NewContentProperty.Required(),
				StartLineProperty.Required(),
				EndLineProperty.Required(),
			},
		}),
	})
}

// UpdateFileLinesTool updates specific lines in a file by line number range.
type UpdateFileLinesTool struct {
	*mcputil.ToolBase
}

// Handle processes the update_file_lines tool request and replaces the specified line range.
func (t *UpdateFileLinesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var startLine, endLine int
	var newContent string

	logger.Info("Tool called", "tool", "update_file_lines")

	filePath, err = FilepathProperty.String(req)
	if err != nil {
		goto end
	}

	newContent, err = NewContentProperty.String(req)
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

	err = t.updateFileLines(filePath, startLine, endLine, newContent)
	if err != nil {
		goto end
	}

	result = mcputil.NewToolResultJSON(map[string]any{
		"success":    true,
		"file_path":  filePath,
		"start_line": startLine,
		"end_line":   endLine,
		"message":    fmt.Sprintf("Successfully updated lines %d-%d in %s", startLine, endLine, filePath),
	})
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

	updatedContent = t.replaceLines(lines, startLine, endLine, newContent)

	err = WriteFile(t.Config(), filePath, updatedContent)

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
