package scout

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func mustClose(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func validatePath(path string) (err error) {
	var absPath string
	var pathInfo os.FileInfo

	absPath, err = filepath.Abs(path)
	if err != nil {
		err = fmt.Errorf("invalid path '%s': %v", path, err)
		goto end
	}

	pathInfo, err = os.Stat(absPath)
	if err != nil {
		err = fmt.Errorf("path '%s' does not exist: %v", absPath, err)
		goto end
	}

	if !pathInfo.IsDir() {
		err = fmt.Errorf("path '%s' is not a directory", absPath)
		goto end
	}

end:
	return err
}

// Add this utility function (probably in util.go)
func homeRelativePath(path string) string {
	var homeDir string
	var err error

	homeDir, err = os.UserHomeDir()
	if err != nil {
		return path // Return original path if we can't get home dir
	}

	// Check if path starts with home directory
	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}

	return path
}

func convertSlice[T any](input []any) (output []T, err error) {
	var t T
	var errs []error

	output = make([]T, len(input))
	for i, item := range input {
		converted, ok := item.(T)
		if !ok {
			errs = append(errs, fmt.Errorf("error converting item %d: item a '%T', not a '%T'", i, item, t))
			continue
		}
		output[i] = converted
	}
	if len(errs) > 0 {
		err = errors.Join(errs...)
	}
	return output, err
}

func titleCase(s string) string {
	return cases.Title(language.English).String(s)
}
