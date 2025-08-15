package mcptools

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*ReplaceFilePartTool)(nil)

func init() {
	mcputil.RegisterTool(&ReplaceFilePartTool{
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "replace_file_part",
			Description: "Replace specific language constructs (functions, types, constants) by name using AST parsing",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathProperty.Required(),
				LanguageProperty.Required(),
				PartTypeProperty.Required(),
				PartNameProperty.Required(),
				NewContentProperty.Required(),
			},
		}),
	})
}

type ReplaceFilePartTool struct {
	*mcputil.ToolBase
}

func (t *ReplaceFilePartTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var language string
	var partType string
	var partName string
	var newContent string

	logger.Info("Tool called", "tool", "replace_file_part")

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

	newContent, err = NewContentProperty.String(req)
	if err != nil {
		goto end
	}

	err = t.validateInputs(language, partType, newContent)
	if err != nil {
		goto end
	}

	err = t.replaceFilePart(filePath, language, partType, partName, newContent)
	if err != nil {
		goto end
	}

	result = mcputil.NewToolResultJSON(map[string]any{
		"success":   true,
		"file_path": filePath,
		"language":  language,
		"part_type": partType,
		"part_name": partName,
		"message":   fmt.Sprintf("Successfully replaced %s '%s' in %s", partType, partName, filePath),
	})
	logger.Info("Tool completed", "tool", "replace_file_part", "path", filePath, "part_type", partType, "part_name", partName)

end:
	return result, err
}

func (t *ReplaceFilePartTool) validateInputs(language, partType, newContent string) (err error) {
	validGoTypes := []string{"const", "var", "type", "func", "import", "package"}
	valid := false

	// Validate language
	if language != "go" {
		err = fmt.Errorf("language '%s' not supported. Currently supported: go", language)
		goto end
	}

	// Validate part type for Go
	for _, validType := range validGoTypes {
		if partType == validType {
			valid = true
			break
		}
	}

	if !valid {
		err = fmt.Errorf("part_type '%s' not supported for Go. Valid types: %v", partType, validGoTypes)
		goto end
	}

	// Validate content starts appropriately for Go
	err = t.validateGoContent(partType, newContent)

end:
	return err
}

func (t *ReplaceFilePartTool) validateGoContent(partType, content string) (err error) {
	content = strings.TrimSpace(content)

	switch partType {
	case "func":
		if !strings.HasPrefix(content, "func ") {
			err = fmt.Errorf("func replacement must start with 'func ', got: %s", content[:min(20, len(content))])
		}
	case "type":
		if !strings.HasPrefix(content, "type ") {
			err = fmt.Errorf("type replacement must start with 'type ', got: %s", content[:min(20, len(content))])
		}
	case "const":
		if !strings.Contains(content, "=") && !strings.HasPrefix(content, "const") {
			err = fmt.Errorf("const replacement must contain '=' or start with 'const', got: %s", content[:min(20, len(content))])
		}
	case "var":
		if !strings.Contains(content, "=") && !strings.HasPrefix(content, "var") {
			err = fmt.Errorf("var replacement must contain '=' or start with 'var', got: %s", content[:min(20, len(content))])
		}
	case "import":
		if !strings.Contains(content, "import") && !strings.Contains(content, "\"") {
			err = fmt.Errorf("import replacement must contain 'import' or quotes, got: %s", content[:min(20, len(content))])
		}
	case "package":
		if !strings.HasPrefix(content, "package ") {
			err = fmt.Errorf("package replacement must start with 'package ', got: %s", content[:min(20, len(content))])
		}
	}

	return err
}

func (t *ReplaceFilePartTool) replaceFilePart(filePath, language, partType, partName, newContent string) (err error) {
	var originalContent string

	if !t.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	originalContent, err = ReadFile(t.Config(), filePath)
	if err != nil {
		goto end
	}

	switch language {
	case "go":
		err = t.replaceGoPart(filePath, originalContent, partType, partName, newContent)
	default:
		err = fmt.Errorf("language %s not supported", language)
	}

end:
	return err
}

func (t *ReplaceFilePartTool) replaceGoPart(filePath, originalContent, partType, partName, newContent string) (err error) {
	var fset *token.FileSet
	var file *ast.File
	var startPos, endPos token.Pos
	var found bool
	var updatedContent string

	// Parse the Go file
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, filePath, originalContent, parser.ParseComments)
	if err != nil {
		err = fmt.Errorf("failed to parse Go file: %w", err)
		goto end
	}

	// Find the part to replace
	startPos, endPos, found, err = t.findGoPart(file, partType, partName)
	if err != nil {
		goto end
	}

	if !found {
		err = fmt.Errorf("%s '%s' not found in file", partType, partName)
		goto end
	}

	// Replace the content
	updatedContent, err = t.replaceGoContent(fset, originalContent, startPos, endPos, newContent)
	if err != nil {
		goto end
	}

	// Validate the updated content parses correctly
	err = t.validateGoSyntax(updatedContent)
	if err != nil {
		goto end
	}

	// Write the updated content
	err = WriteFile(t.Config(), filePath, updatedContent)

end:
	return err
}

func (t *ReplaceFilePartTool) findGoPart(file *ast.File, partType, partName string) (startPos, endPos token.Pos, found bool, err error) {
	switch partType {
	case "package":
		if file.Name.Name == partName {
			startPos = file.Name.Pos()
			endPos = file.Name.End()
			found = true
		}
	case "import":
		startPos, endPos, found = t.findGoImport(file, partName)
	case "const":
		startPos, endPos, found = t.findGoConst(file, partName)
	case "var":
		startPos, endPos, found = t.findGoVar(file, partName)
	case "type":
		startPos, endPos, found = t.findGoType(file, partName)
	case "func":
		startPos, endPos, found = t.findGoFunc(file, partName)
	default:
		err = fmt.Errorf("unsupported part type: %s", partType)
	}

	return startPos, endPos, found, err
}

func (t *ReplaceFilePartTool) findGoImport(file *ast.File, importPath string) (startPos, endPos token.Pos, found bool) {
	for _, imp := range file.Imports {
		if imp.Path.Value == `"`+importPath+`"` || imp.Path.Value == importPath {
			startPos = imp.Pos()
			endPos = imp.End()
			found = true
			return
		}
	}
	return
}

func (t *ReplaceFilePartTool) findGoConst(file *ast.File, constName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						if name.Name == constName {
							startPos = genDecl.Pos()
							endPos = genDecl.End()
							found = true
							return
						}
					}
				}
			}
		}
	}
	return
}

func (t *ReplaceFilePartTool) findGoVar(file *ast.File, varName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						if name.Name == varName {
							startPos = genDecl.Pos()
							endPos = genDecl.End()
							found = true
							return
						}
					}
				}
			}
		}
	}
	return
}

func (t *ReplaceFilePartTool) findGoType(file *ast.File, typeName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if typeSpec.Name.Name == typeName {
						startPos = genDecl.Pos()
						endPos = genDecl.End()
						found = true
						return
					}
				}
			}
		}
	}
	return
}

func (t *ReplaceFilePartTool) findGoFunc(file *ast.File, funcName string) (startPos, endPos token.Pos, found bool) {
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			var name string

			// Handle regular functions
			if funcDecl.Recv == nil {
				name = funcDecl.Name.Name
			} else {
				// Handle methods - format as ReceiverType.MethodName
				if len(funcDecl.Recv.List) > 0 {
					var recvType string
					switch recv := funcDecl.Recv.List[0].Type.(type) {
					case *ast.StarExpr:
						if ident, ok := recv.X.(*ast.Ident); ok {
							recvType = "*" + ident.Name
						}
					case *ast.Ident:
						recvType = recv.Name
					}
					name = recvType + "." + funcDecl.Name.Name
				}
			}

			if name == funcName {
				startPos = funcDecl.Pos()
				endPos = funcDecl.End()
				found = true
				return
			}
		}
	}
	return
}

func (t *ReplaceFilePartTool) replaceGoContent(fset *token.FileSet, originalContent string, startPos, endPos token.Pos, newContent string) (result string, err error) {
	var startOffset, endOffset int

	// Convert token positions to byte offsets
	startOffset = fset.Position(startPos).Offset
	endOffset = fset.Position(endPos).Offset

	// Replace the content
	result = originalContent[:startOffset] + newContent + originalContent[endOffset:]

	return result, err
}

func (t *ReplaceFilePartTool) validateGoSyntax(content string) (err error) {
	var fset *token.FileSet

	fset = token.NewFileSet()
	_, err = parser.ParseFile(fset, "", content, parser.ParseComments)
	if err != nil {
		err = fmt.Errorf("replacement resulted in invalid Go syntax: %w", err)
	}

	return err
}
