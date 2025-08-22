package golang

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type GoFile struct {
	Directory    *GoDirectory // New directory-based approach
	astFile      *ast.File
	dirEntry     os.DirEntry
	declarations []GoDeclaration
	packageName  string // Package name when using Directory
}

func (gf *GoFile) Parse(ctx context.Context) (err error) {
	gf.astFile, err = parser.ParseFile(
		gf.FileSet(),
		gf.Fullpath(),
		nil,
		parser.ParseComments|parser.SkipObjectResolution|parser.DeclarationErrors|parser.AllErrors,
	)
	return err
}

func NewGoFile(de os.DirEntry, dir *GoDirectory) *GoFile {
	return &GoFile{
		dirEntry:  de,
		Directory: dir,
	}
}

func (gf *GoFile) FileSet() (fs *token.FileSet) {
	if gf.Directory != nil {
		fs = gf.Directory.fileSet
	}
	return fs
}

func (gf *GoFile) Name() string {
	return gf.dirEntry.Name()
}

func (gf *GoFile) Fullpath() string {
	return gf.Dir(gf.Name())
}

func (gf *GoFile) Dir(filename string) string {
	return filepath.Join(gf.Directory.Path, filename)
}

func (gf *GoFile) Declarations() []GoDeclaration {
	if gf.declarations == nil {
		gf.declarations = make([]GoDeclaration, len(gf.astFile.Decls))
		for i, decl := range gf.astFile.Decls {
			gf.declarations[i] = NewGoDeclaration(decl, gf)
		}
	}
	return gf.declarations
}

func (gf *GoFile) Exceptions(ctx context.Context) (exceptions []DocException) {
	exception := gf.PackageException()
	if exception != nil {
		exceptions = append(exceptions, *exception)
	}
	for _, d := range gf.Declarations() {
		exceptions = append(exceptions, d.Exceptions(ctx)...)
	}
	return exceptions
}

func (gf *GoFile) PackagePos() token.Pos {
	return gf.astFile.Package
}

func (gf *GoFile) PackageException() (exception *DocException) {
	var fe DocException
	// Require a file-level doc associated with the package clause: starts with "Package <name>"
	doc := gf.astFile.Doc
	if doc != nil && gf.HasProperPackagePrefix(doc.Text()) {
		goto end
	}
	fe = NewDocException(gf.Fullpath(), FileException, &DocExceptionArgs{
		Line: gf.Line(gf.PackagePos()),
	})
	exception = &fe
end:
	return exception
}

func (gf *GoFile) HasProperPackagePrefix(text string) (hasPrefix bool) {
	var pkgName string
	s := strings.TrimSpace(text)
	if s == "" {
		goto end
	}
	pkgName = gf.getPackageName()
	hasPrefix = strings.HasPrefix(firstLine(s), fmt.Sprintf("Package %s", pkgName))
end:
	return hasPrefix
}

// getPackageName returns the package name, supporting both old Package and new Directory approaches
func (gf *GoFile) getPackageName() (name string) {
	name = gf.astFile.Name.Name
	if gf.Directory != nil && gf.Directory.PackageName == "" {
		gf.Directory.PackageName = name
	}
	return name
}
func (gf *GoFile) hasDoc(gen *ast.GenDecl, name string, i int) (hasDoc bool) {
	if gen.Doc == nil {
		goto end
	}
	if !hasProperIdentifierPrefix(gen.Doc.Text(), name) {
		goto end
	}
	hasDoc = true
end:
	return hasDoc
}
func (gf *GoFile) TypeExceptions(gen *ast.GenDecl) (exceptions []DocException) {

	for i, spec := range gen.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if ts == nil {
			continue
		}
		name := ts.Name.Name
		if gf.hasDoc(gen, name, i) {
			continue
		}
		exceptions = append(exceptions, NewDocException(gf.Fullpath(), TypeException, &DocExceptionArgs{
			Element: name,
			Line:    gf.Line(ts.Name.Pos()),
		}))
	}
	return exceptions
}

func (gf *GoFile) SpecName(spec ast.Spec) (name string) {
	switch s := spec.(type) {
	case *ast.TypeSpec:
		name = s.Name.Name
	case *ast.ValueSpec:
		name = s.Names[0].Name
	default:
		name = fmt.Sprintf("SpecName() not implemented yet for type '%T'", s)
		logger.Warn(name)
	}
	return name
}

func (gf *GoFile) ConstVarExceptions(gen *ast.GenDecl) (exceptions []DocException) {

	fileSet := gf.FileSet()

	kind := VarException
	if gen.Tok == token.CONST {
		kind = ConstException
	}

	if gen.Lparen.IsValid() {
		// Grouped declaration: require group-level doc AND per-name EOL comments.
		// 1) Group-level doc: require non-empty text (strict but general).
		groupDocOK := gen.Doc != nil && hasNonEmptyFirstLine(gen.Doc.Text())
		if !groupDocOK {
			exceptions = append(exceptions, NewDocException(gf.Fullpath(), kind|GroupException, &DocExceptionArgs{
				Line:    gf.Line(gen.Pos()),
				EndLine: gf.LinePtr(gen.End()),
			}))
		}

		// Per-name end-of-line comments required
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			if vs == nil {
				continue
			}
			multi := len(vs.Names) > 1
			for _, name := range vs.Names {
				if name == nil {
					continue
				}
				if gf.HasRightSideComment(name) {
					continue
				}
				exceptions = append(exceptions, NewDocException(
					gf.Fullpath(),
					kind,
					&DocExceptionArgs{
						Line:      gf.Line(name.Pos()),
						MultiLine: multi,
						Element:   gf.SpecName(spec),
					},
				))
			}
		}
		return exceptions
	}

	// Ungrouped declaration: require leading doc on the ValueSpec.
	for _, spec := range gen.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		if vs == nil {
			continue
		}
		if len(vs.Names) == 0 {
			continue
		}
		multi := len(vs.Names) > 1
		if gen.Doc == nil || !hasNonEmptyFirstLine(gen.Doc.Text()) {
			// Report once per spec, using first name's line
			line := fileSet.Position(vs.Names[0].Pos()).Line
			exceptions = append(exceptions, NewDocException(gf.Fullpath(), kind, &DocExceptionArgs{
				Line:      line,
				EndLine:   nil,
				MultiLine: multi,
			}))
		}
	}

	return exceptions
}

func (gf *GoFile) FuncException(fd *ast.FuncDecl) (exception *DocException) {
	var except DocException
	var hasException bool
	name := fd.Name.Name
	if fd.Doc == nil {
		hasException = true
		goto end
	}
	if hasProperIdentifierPrefix(fd.Doc.Text(), name) {
		goto end
	}
	hasException = true
end:
	if hasException {
		except = NewDocException(gf.Fullpath(), FuncException, &DocExceptionArgs{
			Line:    gf.Line(fd.Name.Pos()),
			Element: name,
		})
		exception = &except
	}
	return exception
}

func (gf *GoFile) Comments() []*ast.CommentGroup {
	return gf.astFile.Comments
}

func (gf *GoFile) Line(pos token.Pos) int {
	return gf.Position(pos).Line
}
func (gf *GoFile) LinePtr(pos token.Pos) *int {
	line := gf.Line(pos)
	return &line
}

func (gf *GoFile) Column(pos token.Pos) int {
	return gf.Position(pos).Column
}

func (gf *GoFile) Position(pos token.Pos) token.Position {
	return gf.FileSet().Position(pos)
}

func (gf *GoFile) HasRightSideComment(name *ast.Ident) (hasComment bool) {
	line := gf.Line(name.Pos())
	colStartAfter := gf.Column(name.End())
	for _, cg := range gf.Comments() {
		for _, c := range cg.List {
			pos := gf.Position(c.Slash)
			if pos.Filename == "" {
				continue
			}
			if pos.Line != line {
				continue
			}
			if pos.Column <= colStartAfter {
				continue
			}
			// comment that begins to the right of the identifier on same line
			hasComment = true
			goto end
		}
	}
end:
	return hasComment
}
