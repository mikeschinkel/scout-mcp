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
		ToolBase: mcputil.NewToolBase(mcputil.ToolOptions{
			Name:        "replace_pattern",
			Description: "Find and replace text patterns in a file with support for regex",
			QuickHelp:   "Find and replace text",
			Properties: []mcputil.Property{
				RequiredSessionTokenProperty,
				PathProperty.Required(),
				PatternProperty.Required(),
				ReplacementProperty.Required(),
				RegexProperty,
				AllOccurrencesProperty,
			},
		}),
	})
}

type ReplacePatternTool struct {
	*mcputil.ToolBase
}

func (t *ReplacePatternTool) Handle(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var pattern string
	var replacement string
	var useRegex bool
	var allOccurrences bool
	var replacementCount int
	var message string

	logger.Info("Tool called", "tool", "replace_pattern")

	filePath, err = PathProperty.String(req)
	if err != nil {
		goto end
	}

	pattern, err = PatternProperty.String(req)
	if err != nil {
		goto end
	}

	replacement, err = ReplacementProperty.String(req)
	if err != nil {
		goto end
	}

	useRegex, err = RegexProperty.Bool(req)
	if err != nil {
		goto end
	}

	allOccurrences, err = AllOccurrencesProperty.Bool(req)
	if err != nil {
		goto end
	}

	replacementCount, err = t.replaceInFile(filePath, pattern, replacement, useRegex, allOccurrences)
	if err != nil {
		goto end
	}

	if replacementCount == 0 {
		message = fmt.Sprintf("Pattern '%s' not found in %s", pattern, filePath)
	} else if replacementCount == 1 {
		message = fmt.Sprintf("Successfully replaced 1 occurrence of '%s' in %s", pattern, filePath)
	} else {
		message = fmt.Sprintf("Successfully replaced %d occurrences of '%s' in %s", replacementCount, pattern, filePath)
	}

	result = mcputil.NewToolResultJSON(map[string]interface{}{
		"success":           replacementCount > 0,
		"file_path":         filePath,
		"pattern":           pattern,
		"replacement":       replacement,
		"replacement_count": replacementCount,
		"use_regex":         useRegex,
		"all_occurrences":   allOccurrences,
		"message":           message,
	})

	logger.Info("Tool completed", "tool", "replace_pattern", "path", filePath, "replacements", replacementCount)

end:
	return result, err
}

func (t *ReplacePatternTool) replaceInFile(filePath, pattern, replacement string, useRegex, allOccurrences bool) (count int, err error) {
	var originalContent string
	var updatedContent string

	if !t.IsAllowedPath(filePath) {
		err = fmt.Errorf("access denied: path not allowed: %s", filePath)
		goto end
	}

	originalContent, err = ReadFile(t.Config(), filePath)
	if err != nil {
		goto end
	}

	updatedContent, count, err = t.performReplacement(originalContent, pattern, replacement, useRegex, allOccurrences)
	if err != nil {
		goto end
	}

	err = WriteFile(t.Config(), filePath, updatedContent)

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
