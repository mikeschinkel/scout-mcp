package mcptools

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getStringSlice(req mcputil.ToolRequest, prop string) ([]string, error) {
	return convertSliceOfAny[string](req.GetArray(prop, nil))
}

func convertSliceOfAny[T any](input []any) (output []T, err error) {
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

func errorsStringSlice(errs []error) (es []string) {
	es = make([]string, len(errs))
	for i, err := range errs {
		es[i] = err.Error()
	}
	return es
}
