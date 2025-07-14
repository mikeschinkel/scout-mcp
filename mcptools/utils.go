package mcptools

import (
	"errors"
	"fmt"

	"github.com/mikeschinkel/scout-mcp/mcputil"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func getFileActions(req mcputil.ToolRequest) ([]FileAction, error) {
	return convertSlice[FileAction](req.GetArray("file_actions", nil))
}
func getOperations(req mcputil.ToolRequest) ([]string, error) {
	return convertSlice[string](req.GetArray("operations", nil))
}
func getFiles(req mcputil.ToolRequest) ([]string, error) {
	return convertSlice[string](req.GetArray("files", nil))
}
func getExtensions(req mcputil.ToolRequest) ([]string, error) {
	return convertSlice[string](req.GetArray("extensions", nil))
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
