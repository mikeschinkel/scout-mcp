package mcptools

import (
	"os"
	"path/filepath"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func titleCase(s string) string {
	return cases.Title(language.English).String(s)
}

// Add this utility function (probably in util.go)
func makeRelativePath(path, root string) (rel string, err error) {

	rel, err = filepath.Rel(root, path)
	if err != nil {
		logger.Warn("Path is not relative to root", "path", path, "root", root)
		goto end
	}
	path = filepath.Join("~", rel)
end:
	return path, err
}

func errorsStringSlice(errs []error) (es []string) {
	es = make([]string, len(errs))
	for i, err := range errs {
		es[i] = err.Error()
	}
	return es
}

// checkFileExists returns a file error no matter what
//
//	os.ErrExists — when the file exists
//	os.ErrNotExists — when the file does not exist
//	os.Err??? — some other error
func checkFileExists(fp string) (err error) {
	// Check if file already exists
	_, err = os.Stat(fp)
	if err == nil {
		err = os.ErrExist
	}
	return err
}
