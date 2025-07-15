package mcptools

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

var _ mcputil.Tool = (*ReplacePatternTool)(nil)

func init() {
	mcputil.RegisterTool(&ReplacePatternTool{
		toolBase: newToolBase(mcputil.ToolOptions{
			Name:        "replace_pattern",
			Description: "Find and replace text patterns in a file with support for regex",
		}),
	})
}

type ReplacePatternTool struct {
	*toolBase
}

func (t *ReplacePatternTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var pattern string
	var replacement string
	var useRegex bool
	var allOccurrences bool
	var replacementCount int

	logger.Info("Tool called", "tool", "replace_pattern")

	filePath, err = req.RequireString("path")
	if err != nil {
		goto end
	}

	pattern, err = req.RequireString("pattern")
	if err != nil {
		goto end
	}

	replacement, err = req.RequireString("replacement")
	if err != nil {
		goto end
	}

	// TODO Change back to checking error here
	useRegex = req.GetBool("regex", false)
	allOccurrences = req.GetBool("all_occurrences", true)

	replacementCount, err = t.replaceInFile(filePath, pattern, replacement, useRegex, allOccurrences)
	if err != nil {
		goto end
	}

	if replacementCount == 0 {
		result = mcputil.NewToolResultText(fmt.Sprintf("Pattern '%s' not found in %s", pattern, filePath))
	} else if replacementCount == 1 {
		result = mcputil.NewToolResultText(fmt.Sprintf("Successfully replaced 1 occurrence of '%s' in %s", pattern, filePath))
	} else {
		result = mcputil.NewToolResultText(fmt.Sprintf("Successfully replaced %d occurrences of '%s' in %s", replacementCount, pattern, filePath))
	}

	logger.Info("Tool completed", "tool", "replace_pattern", "path", filePath, "replacements", replacementCount)

end:
	return result, err
}

func (t *ReplacePatternTool) replaceInFile(filePath, pattern, replacement string, useRegex, allOccurrences bool) (count int, err error) {
	var allowed bool
	var originalContent string
	var updatedContent string

	allowed, err = t.IsAllowedPath(filePath)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	originalContent, err = readFile(t.Config(), filePath)
	if err != nil {
		goto end
	}

	updatedContent, count, err = t.performReplacement(originalContent, pattern, replacement, useRegex, allOccurrences)
	if err != nil {
		goto end
	}

	err = writeFile(t.Config(), filePath, updatedContent)

end:
	return count, err
}

func (t *ReplacePatternTool) performReplacement(content, pattern, replacement string, useRegex, allOccurrences bool) (result string, count int, err error) {
	if useRegex {
		result, count, err = t.regexReplace(content, pattern, replacement, allOccurrences)
	} else {
		result, count = t.stringReplace(content, pattern, replacement, allOccurrences)
	}

	return result, count, err
}

func (t *ReplacePatternTool) regexReplace(content, pattern, replacement string, allOccurrences bool) (result string, count int, err error) {
	var re *regexp.Regexp

	re, err = regexp.Compile(pattern)
	if err != nil {
		err = fmt.Errorf("invalid regex pattern: %w", err)
		goto end
	}

	if allOccurrences {
		result, count = t.regexReplaceAll(re, content, replacement)
	} else {
		result, count = t.regexReplaceFirst(re, content, replacement)
	}

end:
	return result, count, err
}

func (t *ReplacePatternTool) regexReplaceAll(re *regexp.Regexp, content, replacement string) (result string, count int) {
	matches := re.FindAllStringIndex(content, -1)
	count = len(matches)

	if count > 0 {
		result = re.ReplaceAllString(content, replacement)
	} else {
		result = content
	}

	return result, count
}

func (t *ReplacePatternTool) regexReplaceFirst(re *regexp.Regexp, content, replacement string) (result string, count int) {
	loc := re.FindStringIndex(content)

	if loc != nil {
		result = content[:loc[0]] + re.ReplaceAllString(content[loc[0]:loc[1]], replacement) + content[loc[1]:]
		count = 1
	} else {
		result = content
		count = 0
	}

	return result, count
}

func (t *ReplacePatternTool) stringReplace(content, pattern, replacement string, allOccurrences bool) (result string, count int) {
	if allOccurrences {
		result, count = t.stringReplaceAll(content, pattern, replacement)
	} else {
		result, count = t.stringReplaceFirst(content, pattern, replacement)
	}

	return result, count
}

func (t *ReplacePatternTool) stringReplaceAll(content, pattern, replacement string) (result string, count int) {
	count = strings.Count(content, pattern)

	if count > 0 {
		result = strings.ReplaceAll(content, pattern, replacement)
	} else {
		result = content
	}

	return result, count
}

func (t *ReplacePatternTool) stringReplaceFirst(content, pattern, replacement string) (result string, count int) {
	idx := strings.Index(content, pattern)

	if idx != -1 {
		result = content[:idx] + replacement + content[idx+len(pattern):]
		count = 1
	} else {
		result = content
		count = 0
	}

	return result, count
}
