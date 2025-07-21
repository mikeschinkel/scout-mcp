package mcptools

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getFileActions(req mcputil.ToolRequest) ([]FileAction, error) {
	return convertSlice[FileAction](req.GetArray("file_actions", nil))
}
func getStringSlice(req mcputil.ToolRequest, prop string) ([]string, error) {
	return convertSlice[string](req.GetArray(prop, nil))
}

func getNumberAsInt(req mcputil.ToolRequest, prop string, nonZero bool) (n int, err error) {
	n = req.GetInt(prop, 0)
	if n != 0 {
		goto end
	}
	if nonZero {
		err = fmt.Errorf("'%s' must be a valid number: %w", prop, err)
		goto end
	}
end:
	return n, err
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
