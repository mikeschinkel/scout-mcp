package golang

import (
	"context"
	"errors"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// GoDirectory represents a directory containing Go files. This simplified works with individual files
// and doesn't require valid Go module structure.
type GoDirectory struct {
	Path        string // Path to the directory
	PackageName string
	HasReadme   bool
	files       []*GoFile // Go files in this directory
	subDirs     []*GoDirectory
	fileSet     *token.FileSet
	parent      *GoDirectory
}

// NewGoDirectory creates a new GoDirectory for the specified directory path
func NewGoDirectory(path string, parent *GoDirectory) *GoDirectory {
	return &GoDirectory{
		Path:    path,
		files:   make([]*GoFile, 0),
		subDirs: make([]*GoDirectory, 0),
		fileSet: token.NewFileSet(),
		parent:  parent,
	}
}

// AddFile adds a GoFile to this directory
func (dir *GoDirectory) AddFile(file *GoFile) {
	dir.files = append(dir.files, file)
}

func (dir *GoDirectory) AddSubDir(gd *GoDirectory) {
	dir.subDirs = append(dir.subDirs, gd)
}

// Exceptions returns documentation exceptions for all files in this directory
func (dir *GoDirectory) Exceptions(ctx context.Context, args *DocsExceptionsArgs) (exceptions []DocException) {
	// Skip invalid directories
	if dir == nil {
		goto end
	}

	// README.md presence per directory with Go files
	// But not in the root; we ignore those
	if !dir.HasReadme && len(dir.Path) > len(args.Path) {
		exceptions = append(exceptions, NewDocException(dir.Path+"/README.md", ReadmeException, nil))
	}

	// Check exceptions for all files in the directory
	for _, f := range dir.files {
		exceptions = append(exceptions, f.Exceptions(ctx)...)
	}
	if args.Recursive == DoNotRecurse {
		goto end
	}

	// Check exceptions for subdirectories
	for _, sd := range dir.subDirs {
		exceptions = append(exceptions, sd.Exceptions(ctx, args)...)
	}

end:
	return exceptions
}

type RecurseDirective int

const (
	NonSpecifiedRecurse RecurseDirective = 0
	DoNotRecurse        RecurseDirective = 1
	DoRecurse           RecurseDirective = 2
)

func GetRecurseDirective(recursive bool) (rd RecurseDirective) {
	if recursive {
		rd = DoRecurse
		goto end
	}
	rd = DoNotRecurse
end:
	return rd
}

// ExcludeMode defines how to handle the exclude list
type ExcludeMode int

const (
	// UseDefaults uses the default exclude list
	UseDefaults ExcludeMode = iota
	// AddToDefaults adds to the default exclude list
	AddToDefaults
	// ReplaceDefaults replaces the default exclude list completely
	ReplaceDefaults
)

// TraverseArgs contains arguments for directory traversal operations.
// This structure provides flexible control over how directories are traversed,
// including recursion behavior and file/directory exclusion patterns.
type TraverseArgs struct {
	// RecurseDirectory controls whether to descend into subdirectories
	RecurseDirectory RecurseDirective

	// Exclude specifies file and directory names to exclude during traversal
	Exclude []string

	// ExcludeMode controls how Exclude is interpreted relative to defaults
	ExcludeMode ExcludeMode
}

// DefaultExcludes returns the default list of files and directories to exclude during traversal.
// These typically don't contain Go files or documentation that needs validation.
func DefaultExcludes() []string {
	return []string{
		".git",          // Git version control directory
		"node_modules",  // Node.js dependencies
		".svn",          // Subversion version control
		".hg",           // Mercurial version control
		".bzr",          // Bazaar version control
		"vendor",        // Go vendor directory (older Go versions)
		".vscode",       // Visual Studio Code settings
		".idea",         // IntelliJ IDEA settings
		"build",         // Common build output directory
		"dist",          // Common distribution directory
		"target",        // Maven/Gradle build directory
		"bin",           // Binary output directory
		"obj",           // Object file directory
		".DS_Store",     // macOS metadata (if someone names a dir this)
		"__pycache__",   // Python cache directory
		".pytest_cache", // Python test cache
		"coverage",      // Coverage report directory
		"tmp",           // Temporary files directory
		"temp",          // Temporary files directory
	}
}

// GetEffectiveExcludes returns the effective list of files and directories to exclude
// based on the ExcludeMode and Exclude settings.
func (args *TraverseArgs) GetEffectiveExcludes() []string {
	switch args.ExcludeMode {
	case UseDefaults:
		return DefaultExcludes()
	case AddToDefaults:
		defaults := DefaultExcludes()
		result := make([]string, 0, len(defaults)+len(args.Exclude))
		result = append(result, defaults...)
		result = append(result, args.Exclude...)
		return result
	case ReplaceDefaults:
		return args.Exclude
	default:
		return DefaultExcludes()
	}
}

// shouldExclude checks if a file or directory name should be excluded from traversal.
func (args *TraverseArgs) shouldExclude(name string) bool {
	excludes := args.GetEffectiveExcludes()
	lowerName := strings.ToLower(name)

	for _, exclude := range excludes {
		if strings.ToLower(exclude) == lowerName {
			return true
		}
	}
	return false
}

// Traverse uses parser.ParseDir to traverse all Go files in a directory tree
// with support for directory exclusions to avoid descending into irrelevant directories.
func (dir *GoDirectory) Traverse(ctx context.Context, args *TraverseArgs) (err error) {
	var entries []os.DirEntry
	var errs []error
	var sd *GoDirectory
	var gf *GoFile

	entries, err = os.ReadDir(dir.Path)
	if err != nil {
		goto end
	}

	// Check for .go files in this directory
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip excluded directories
			if args.shouldExclude(entry.Name()) {
				continue
			}
			sd = NewGoDirectory(filepath.Join(dir.Path, entry.Name()), dir)
			if args.RecurseDirectory == DoRecurse {
				err = sd.Traverse(ctx, args)
				if err != nil {
					errs = append(errs, err)
					continue
				}
			}
			dir.AddSubDir(sd)
			continue
		}
		if strings.ToUpper(entry.Name()) == "README.MD" {
			dir.HasReadme = true
			continue
		}
		if strings.ToLower(filepath.Ext(entry.Name())) != ".go" {
			continue
		}
		// Skip excluded files
		if args.shouldExclude(entry.Name()) {
			continue
		}
		gf = NewGoFile(entry, dir)
		err = gf.Parse(ctx)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		dir.AddFile(gf)
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
		goto end
	}
end:
	return err
}
