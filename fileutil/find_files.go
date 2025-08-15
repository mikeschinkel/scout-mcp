// Package fileutil provides utilities for finding and filtering files by extension patterns.
package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FindFileArgs contains the arguments for finding files with specific criteria.
type FindFileArgs struct {
	Recursive  bool           // Whether to search directories recursively
	Paths      []string       // List of paths to search
	Extensions []string       // File extensions to match (e.g., ".go", ".txt")
	extsRE     *regexp.Regexp // Compiled regex for extension matching
	Path       string         // Single path to search (merged with Paths)
}

// newFindFileArgs creates a new FindFileArgs with compiled extension regex.
func newFindFileArgs(args FindFileArgs) FindFileArgs {
	exts := make([]string, 0, len(args.Extensions))
	for _, ext := range args.Extensions {
		exts = append(exts, strings.TrimLeft(ext, `.`))
	}
	return FindFileArgs{
		Recursive: args.Recursive,
		extsRE:    regexp.MustCompile(fmt.Sprintf(`(?i)^.+\.(%s)$`, strings.Join(exts, `|`))),
	}
}

// matches returns true if the path matches the extension criteria in args.
func (args FindFileArgs) matches(path string) bool {
	return args.extsRE.MatchString(path)
}

// FindFiles searches for files matching the specified criteria and returns their paths.
func FindFiles(args FindFileArgs) (found []string, err error) {
	// Merge Path into Paths OR convert Path to a slice
	if args.Path != "" {
		args.Paths = append(args.Paths, args.Path)
	}
	found = make([]string, 0)
	var ffa = newFindFileArgs(args)
	for _, path := range args.Paths {
		var files []string
		files, err = findFiles(path, ffa)
		if err != nil {
			goto end
		}
		found = append(found, files...)
	}
end:
	return found, err
}

// findFiles is a helper function that finds files in a single path.
func findFiles(path string, args FindFileArgs) (files []string, err error) {
	var info os.FileInfo

	info, err = os.Stat(path)
	if err != nil {
		goto end
	}

	if !info.IsDir() {
		// Single file
		if args.matches(path) {
			files = []string{path}
		}
		goto end
	}

	if !args.Recursive {
		// Single directory
		files, err = listDirFiles(path, args)
		goto end
	}

	// Recursive directory scan
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, walkErr error) (err error) {
		if walkErr != nil {
			err = walkErr
			goto end
		}

		if info.IsDir() {
			goto end
		}

		if !args.matches(filePath) {
			goto end
		}

		files = append(files, filePath)

	end:
		return err
	})

end:
	return files, err
}

// listDirFiles lists files in a single directory that match the extension criteria.
func listDirFiles(path string, args FindFileArgs) (files []string, err error) {
	var entries []os.DirEntry

	entries, err = os.ReadDir(path)
	if err != nil {
		goto end
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !args.matches(name) {
			continue
		}
		files = append(files, filepath.Join(path, name))
	}

end:
	return files, err
}
