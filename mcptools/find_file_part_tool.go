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
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "find_file_part",
			Description: "Find specific language constructs by name and return their location and content",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathProperty.Required(),
				LanguageProperty.Required(),
				PartTypeProperty.Required(),
				PartNameProperty.Required(),
			},
		}),
	})
}

// FindFilePartTool finds specific language constructs by name and returns their location and content.
type FindFilePartTool struct {
	*mcputil.ToolBase
}

// Handle processes the find_file_part tool request and searches for language constructs using AST parsing.
func (t *FindFilePartTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var language string
	var partType string
	var partName string
	var originalContent string
	var partInfo *langutil.PartInfo

	logger.Info("Tool called", "tool", "find_file_part")

	filePath, err = PathProperty.String(req)
	if err != nil {
		goto end
	}

	language, err = LanguageProperty.String(req)
	if err != nil {
		goto end
	}

	partType, err = PartTypeProperty.String(req)
	if err != nil {
		goto end
	}

	partName, err = PartNameProperty.String(req)
	if err != nil {
		goto end
	}

	err = t.validateInputs(langutil.Language(language), partType)
	if err != nil {
		goto end
	}

	originalContent, err = t.findFilePart(langutil.PartArgs{
		Language: langutil.Language(language),
		Filepath: filePath,
		PartType: langutil.PartType(partType),
		PartName: partName,
	})
	if err != nil {
		goto end
	}

	partInfo, err = langutil.FindPart(langutil.PartArgs{
		Language: langutil.Language(language),
		Content:  originalContent,
		PartType: langutil.PartType(partType),
		PartName: partName,
	})
	if err != nil {
		goto end
	}

	if !partInfo.Found {
		err = fmt.Errorf("%s '%s' not found in file", partType, partName)
		goto end
	}

	result = mcputil.NewToolResultJSON(map[string]any{
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

func (t *FindFilePartTool) validateInputs(language langutil.Language, partType string) (err error) {
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

func (t *FindFilePartTool) findFilePart(args langutil.PartArgs) (content string, err error) {

	if !t.IsAllowedPath(args.Filepath) {
		err = fmt.Errorf("access denied: path not allowed: %s", args.Filepath)
		goto end
	}

	// TODO: This is unfinished

	content, err = ReadFile(t.Config(), args.Filepath)

end:
	return content, err
}
