package mcptools

import (
	"context"
	"fmt"

	"github.com/mikeschinkel/scout-mcp/langutil"
	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*FindFilePartTool)(nil)

func init() {
	mcputil.RegisterTool(&FindFilePartTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "find_file_part",
			Description: "Find specific language constructs by name and return their location and content",
		}),
	})
}

type FindFilePartTool struct {
	*toolBase
}

func (t *FindFilePartTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var language string
	var partType string
	var partName string
	var originalContent string
	var partInfo *langutil.PartInfo

	logger.Info("Tool called", "tool", "find_file_part")

	filePath, err = req.RequireString("path")
	if err != nil {
		goto end
	}

	language, err = req.RequireString("language")
	if err != nil {
		goto end
	}

	partType, err = req.RequireString("part_type")
	if err != nil {
		goto end
	}

	partName, err = req.RequireString("part_name")
	if err != nil {
		goto end
	}

	err = t.validateInputs(language, partType)
	if err != nil {
		goto end
	}

	originalContent, err = t.findFilePart(filePath, language, partType, partName)
	if err != nil {
		goto end
	}

	partInfo, err = langutil.FindPart(language, originalContent, langutil.PartType(partType), partName)
	if err != nil {
		goto end
	}

	if !partInfo.Found {
		err = fmt.Errorf("%s '%s' not found in file", partType, partName)
		goto end
	}

	result = mcputil.NewToolResultJSON(map[string]interface{}{
		"found":        true,
		"part_type":    partType,
		"part_name":    partName,
		"start_line":   partInfo.StartLine,
		"end_line":     partInfo.EndLine,
		"start_offset": partInfo.StartOffset,
		"end_offset":   partInfo.EndOffset,
		"content":      partInfo.Content,
		"file_path":    filePath,
	})

	logger.Info("Tool completed", "tool", "find_file_part", "path", filePath, "part_type", partType, "part_name", partName, "found", true)

end:
	return result, err
}

func (t *FindFilePartTool) validateInputs(language, partType string) (err error) {
	var supportedTypes []langutil.PartType
	var validType bool

	// Validate language is supported
	supportedTypes, err = langutil.GetSupportedPartTypes(language)
	if err != nil {
		goto end
	}

	// Validate part type is supported for this language
	for _, supportedType := range supportedTypes {
		if string(supportedType) == partType {
			validType = true
			break
		}
	}

	if !validType {
		err = fmt.Errorf("part_type '%s' not supported for language '%s'. Valid types: %v", partType, language, supportedTypes)
	}

end:
	return err
}

func (t *FindFilePartTool) findFilePart(filePath, language, partType, partName string) (content string, err error) {
	var allowed bool

	allowed, err = t.IsAllowedPath(filePath)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	content, err = readFile(t.Config(), filePath)

end:
	return content, err
}
