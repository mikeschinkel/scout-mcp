package langutil

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

// Go language part types
const (
	FuncGoPart    PartType = "func"
	TypeGoPart    PartType = "type"
	ConstGoPart   PartType = "const"
	VarGoPart     PartType = "var"
	ImportGoPart  PartType = "import"
	PackageGoPart PartType = "package"
)

// GoLanguage implements the Language interface for Go
type GoLanguage struct{}

func init() {
	RegisterLanguage(&GoLanguage{})
}

func (g *GoLanguage) Name() string {
	return "go"
}

func (g *GoLanguage) SupportedPartTypes() []PartType {
	return []PartType{
		FuncGoPart,
		TypeGoPart,
		ConstGoPart,
		VarGoPart,
		ImportGoPart,
		PackageGoPart,
	}
}

func (g *GoLanguage) FindPart(source string, partType PartType, partName string) (pi *PartInfo, err error) {
	var fs *token.FileSet
	var file *ast.File
	var start, end token.Pos
	var found bool
	var startPos, endPos token.Position

	pi = &PartInfo{Found: false}

	// Parse the Go file
	fs = token.NewFileSet()
	file, err = parser.ParseFile(fs, "", source, parser.ParseComments)
	if err != nil {
		err = fmt.Errorf("failed to parse Go file: %w", err)
		goto end
	}

	// Find the part
	start, end, found, err = g.findGoPart(file, partType, partName)
	if err != nil {
		goto end
	}

	if !found {
		goto end
	}

	// Convert positions to line numbers and offsets
	startPos = fs.Position(start)
	endPos = fs.Position(end)

	pi.Found = true
	pi.StartLine = startPos.Line
	pi.EndLine = endPos.Line
	pi.StartOffset = startPos.Offset
	pi.EndOffset = endPos.Offset
	pi.Content = source[startPos.Offset:endPos.Offset]

end:
	return pi, err
}

func (g *GoLanguage) ReplacePart(source string, partType PartType, partName string, newContent string) (result string, err error) {
	var partInfo *PartInfo

	// Find the part first
	partInfo, err = g.FindPart(source, partType, partName)
	if err != nil {
		goto end
	}

	if !partInfo.Found {
		err = fmt.Errorf("%s '%s' not found in file", partType, partName)
		goto end
	}

	// Replace the content
	result = source[:partInfo.StartOffset] + newContent + source[partInfo.EndOffset:]

	// Validate the result
	err = g.ValidateSyntax(result)
	if err != nil {
		err = fmt.Errorf("replacement resulted in invalid Go syntax: %w", err)
		goto end
	}

end:
	return result, err
}

func (g *GoLanguage) ValidateContent(partType PartType, content string) (err error) {
	content = strings.TrimSpace(content)

	switch partType {
	case FuncGoPart:
		if !strings.HasPrefix(content, "func ") {
			err = fmt.Errorf("func replacement must start with 'func ', got: %s", content[:min(20, len(content))])
		}
	case TypeGoPart:
		if !strings.HasPrefix(content, "type ") {
			err = fmt.Errorf("type replacement must start with 'type ', got: %s", content[:min(20, len(content))])
		}
	case ConstGoPart:
		if !strings.Contains(content, "=") && !strings.HasPrefix(content, "const") {
			err = fmt.Errorf("const replacement must contain '=' or start with 'const', got: %s", content[:min(20, len(content))])
		}
	case VarGoPart:
		if !strings.Contains(content, "=") && !strings.HasPrefix(content, "var") {
			err = fmt.Errorf("var replacement must contain '=' or start with 'var', got: %s", content[:min(20, len(content))])
		}
	case ImportGoPart:
		if !strings.Contains(content, "import") && !strings.Contains(content, "\"") {
			err = fmt.Errorf("import replacement must contain 'import' or quotes, got: %s", content[:min(20, len(content))])
		}
	case PackageGoPart:
		if !strings.HasPrefix(content, "package ") {
			err = fmt.Errorf("package replacement must start with 'package ', got: %s", content[:min(20, len(content))])
		}
	default:
		err = fmt.Errorf("unsupported part type for Go: %s", partType)
	}

	return err
}

func (g *GoLanguage) ValidateSyntax(source string) (err error) {
	var fs *token.FileSet

	fs = token.NewFileSet()
	_, err = parser.ParseFile(fs, "", source, parser.ParseComments)

	return err
}

func (g *GoLanguage) findGoPart(file *ast.File, partType PartType, partName string) (startPos, endPos token.Pos, found bool, err error) {
	switch partType {
	case PackageGoPart:
		if file.Name.Name == partName {
			startPos = file.Name.Pos()
			endPos = file.Name.End()
			found = true
		}
	case ImportGoPart:
		startPos, endPos, found = g.findGoImport(file, partName)
	case ConstGoPart:
		startPos, endPos, found = g.findGoConst(file, partName)
	case VarGoPart:
		startPos, endPos, found = g.findGoVar(file, partName)
	case TypeGoPart:
		startPos, endPos, found = g.findGoType(file, partName)
	case FuncGoPart:
		startPos, endPos, found = g.findGoFunc(file, partName)
	default:
		err = fmt.Errorf("unsupported part type: %s", partType)
	}

	return startPos, endPos, found, err
}

func (g *GoLanguage) findGoImport(file *ast.File, importPath string) (startPos, endPos token.Pos, found bool) {
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

func (g *GoLanguage) findGoConst(file *ast.File, constName string) (startPos, endPos token.Pos, found bool) {
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

func (g *GoLanguage) findGoVar(file *ast.File, varName string) (startPos, endPos token.Pos, found bool) {
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

func (g *GoLanguage) findGoType(file *ast.File, typeName string) (startPos, endPos token.Pos, found bool) {
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

func (g *GoLanguage) findGoFunc(file *ast.File, funcName string) (startPos, endPos token.Pos, found bool) {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
