package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FindFileArgs struct {
	Recursive  bool
	Paths      []string
	Extensions []string
	extsRE     *regexp.Regexp
	Path       string
}

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

// matches returns true of the path matches the criteria in args
func (args FindFileArgs) matches(path string) bool {
	return args.extsRE.MatchString(path)
}

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
