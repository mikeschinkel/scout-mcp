package mcptools

import (
	"context"
	"fmt"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*InsertFileLinesTool)(nil)

func init() {
	mcputil.RegisterTool(&InsertFileLinesTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "insert_file_lines",
			Description: "Insert content at a specific line number in a file",
			QuickHelp:   "Insert content at specific lines",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				FilepathProperty.Required(),
				NewContentProperty.Required(),
				PositionProperty.Description("Position at which to insert").Required(),
				LineNumberProperty.Description("Line number where to insert content").Required(),
			},
		}),
	})
}

type InsertFileLinesTool struct {
	*mcputil.ToolBase
}

func (t *InsertFileLinesTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var lineNumber int
	var content string
	var position string

	logger.Info("Tool called", "tool", "insert_file_lines")

	filePath, err = FilepathProperty.String(req)
	if err != nil {
		goto end
	}

	content, err = NewContentProperty.String(req)
	if err != nil {
		goto end
	}

	lineNumber, err = LineNumberProperty.Int(req)
	if err != nil {
		goto end
	}

	position, err = PositionProperty.SetDefault(string(AfterPosition)).String(req)
	if err != nil {
		goto end
	}

	err = t.validatePosition(position)
	if err != nil {
		goto end
	}

	err = t.insertAtLine(filePath, lineNumber, content, position)
	if err != nil {
		goto end
	}

	result = mcputil.NewToolResultJSON(map[string]interface{}{
		"success":     true,
		"file_path":   filePath,
		"line_number": lineNumber,
		"position":    position,
		"message":     fmt.Sprintf("Successfully inserted content %s line %d in %s", position, lineNumber, filePath),
	})
	logger.Info("Tool completed", "tool", "insert_file_lines", "path", filePath, "line_number", lineNumber, "position", position)

end:
	return result, err
}

func (t *InsertFileLinesTool) validatePosition(position string) (err error) {
	return RelativePosition(position).Validate()
}

func (t *InsertFileLinesTool) insertAtLine(filePath string, lineNumber int, content, position string) (err error) {
	var originalContent string
	var lines []string
	var updatedContent string

	if !t.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	originalContent, err = readFile(t.Config(), filePath)
	if err != nil {
		goto end
	}

	lines = strings.Split(originalContent, "\n")

	err = t.validateLineNumber(lines, lineNumber)
	if err != nil {
		goto end
	}

	updatedContent = t.insertContent(lines, lineNumber, content, position)

	err = writeFile(t.Config(), filePath, updatedContent)

end:
	return err
}

func (t *InsertFileLinesTool) validateLineNumber(lines []string, lineNumber int) (err error) {
	totalLines := len(lines)

	if lineNumber < 1 {
		err = fmt.Errorf("line_number must be >= 1, got %d", lineNumber)
		goto end
	}

	if lineNumber > totalLines {
		err = fmt.Errorf("line_number %d exceeds file length %d", lineNumber, totalLines)
		goto end
	}

end:
	return err
}

func (t *InsertFileLinesTool) insertContent(lines []string, lineNumber int, content, position string) (result string) {
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
