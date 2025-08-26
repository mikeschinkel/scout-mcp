# Instructions: Refactor check_docs Tool for File-Grouped Output

## Problem Statement

The current `check_docs` tool returns a flat array of issues, which creates inefficiencies for AI assistants:

- **Multiple file updates**: Same file may be updated 25+ times for 25 separate issues
- **Poor workflow**: Hard to prioritize files or get file-level overview of problems
- **Scanning overhead**: AI must scan flat list to find all issues for a specific file
- **Batch operation complexity**: Difficult to apply multiple fixes to one file in a single operation

## Current vs Desired Output Structure

### Current Structure (Inefficient)
```json
{
  "path": "/path/to/project",
  "issues": [
    {"file": "cmd_base.go", "line": 14, "issue": "Missing const comment", "element": "UnknownFlagType"},
    {"file": "cmd_base.go", "line": 15, "issue": "Missing const comment", "element": "StringFlag"},
    {"file": "cmd_base.go", "line": 16, "issue": "Missing const comment", "element": "BoolFlag"},
    {"file": "cmd_runner.go", "line": 13, "issue": "Missing type comment", "element": "GlobalFlagDefGetter"}
  ],
  "returned_count": 54,
  "total_count": 54
}
```

### Desired Structure (File-Grouped)
```json
{
  "path": "/path/to/project",
  "issues_by_file": {
    "cmd_base.go": {
      "file": "cmd_base.go",
      "issue_count": 25,
      "issues": [
        {"line": 14, "issue": "Missing const comment", "element": "UnknownFlagType", "multi_line": false},
        {"line": 15, "issue": "Missing const comment", "element": "StringFlag", "multi_line": false},
        {"line": 16, "issue": "Missing const comment", "element": "BoolFlag", "multi_line": false}
      ]
    },
    "cmd_runner.go": {
      "file": "cmd_runner.go",
      "issue_count": 8,
      "issues": [
        {"line": 13, "issue": "Missing type comment", "element": "GlobalFlagDefGetter", "multi_line": false}
      ]
    }
  },
  "summary": {
    "total_files_with_issues": 12,
    "total_issues": 54,
    "files_by_issue_count": [
      {"file": "cmd_base.go", "issue_count": 25},
      {"file": "prompt_for_approval.go", "issue_count": 12},
      {"file": "cmd_runner.go", "issue_count": 8}
    ]
  },
  "returned_count": 54,
  "total_count": 54,
  "remaining_count": 0,
  "size_limited": false,
  "response_size_chars": 6200
}
```

## Implementation Steps

### Step 1: Update Data Structures in `/mcptools/check_docs_tool.go`

Add these new types **before** the existing `DocsAnalysisResult` struct:

```go
type FileIssueGroup struct {
    File    string              `json:"file"`
    IssueCount  int                 `json:"issue_count"`
    Issues      []DocsAnalysisIssue `json:"issues"`
}

type FileIssueCountItem struct {
    File       string `json:"file"`
    IssueCount int    `json:"issue_count"`
}

type IssueSummary struct {
    TotalFilesWithIssues int                  `json:"total_files_with_issues"`
    TotalIssues          int                  `json:"total_issues"`
    FilesByIssueCount    []FileIssueCountItem `json:"files_by_issue_count"`
}
```

### Step 2: Update Main Result Structure

Replace the existing `DocsAnalysisResult` struct with:

```go
type DocsAnalysisResult struct {
    Path           string            `json:"path"`
    IssuesByFile   []FileIssueGroup  `json:"issues_by_file"`
    ReturnedCount  int               `json:"returned_count"`
    TotalCount     int               `json:"total_count"`
    RemainingCount int               `json:"remaining_count"`
    SizeLimited    bool              `json:"size_limited"`
    ResponseSize   int               `json:"response_size_chars"`
    Message        string            `json:"message,omitempty"`
}
```

### Step 3: Add Grouping Logic

Add this new function **after** the existing `NewDocsAnalysisIssue` function:

```go
// groupIssuesByFile groups a flat array of issues by file path, maintaining priority order within each file
func groupIssuesByFile(issues []DocsAnalysisIssue) []FileIssueGroup {
    var fileGroupMap map[string][]DocsAnalysisIssue
    var fileGroups []FileIssueGroup
    var group []DocsAnalysisIssue
    var exists bool
    var seenFiles []string
    
    fileGroupMap = make(map[string][]DocsAnalysisIssue)
    
    // Group issues by file, preserving order
    for _, issue := range issues {
        group, exists = fileGroupMap[issue.File]
        if !exists {
            seenFiles = append(seenFiles, issue.File)
        }
        group = append(group, issue)
        fileGroupMap[issue.File] = group
    }
    
    // Convert to array format, maintaining file priority order
    fileGroups = make([]FileIssueGroup, 0, len(seenFiles))
    for _, filePath := range seenFiles {
        group = fileGroupMap[filePath]
        fileGroups = append(fileGroups, FileIssueGroup{
            File:       filePath,
            IssueCount: len(group),
            Issues:     group,
        })
    }
    
    return fileGroups
}
```

### Step 4: Update NewDocsAnalysisResult Function

Replace the existing `NewDocsAnalysisResult` function with:

```go
func NewDocsAnalysisResult(args DocsAnalysisResultArgs) *DocsAnalysisResult {
    var fileGroups []FileIssueGroup
    var returnedCount int
    var sizeLimited bool
    var remainingCount int
    
    // Convert flat issues to grouped structure
    issues := NewDocsAnalysisIssues(args.Exceptions, args.Path)
    fileGroups = groupIssuesByFile(issues)
    
    returnedCount = len(args.Exceptions)
    sizeLimited = args.TotalFound > returnedCount
    
    // Calculate remaining count considering offset
    remainingCount = args.TotalFound - args.Offset - returnedCount
    if remainingCount < 0 {
        remainingCount = 0
    }
    
    result := &DocsAnalysisResult{
        Path:           args.Path,
        IssuesByFile:   fileGroups,
        ReturnedCount:  returnedCount,
        TotalCount:     args.TotalFound,
        RemainingCount: remainingCount,
        SizeLimited:    sizeLimited,
        ResponseSize:   args.ResponseSize,
    }
    
    if sizeLimited {
        result.Message = fmt.Sprintf(
            "Response limited to %d of %d total issues due to size constraints (%d chars). "+
                "Showing highest priority issues first. %d issues remaining. "+
                "Run again with offset parameter or after fixing current issues.",
            returnedCount, args.TotalFound, args.ResponseSize, result.RemainingCount)
    }
    
    return result
}
```

### Step 5: Update Size Limiting Logic

Modify `createSizedAnalysisResult` to remove entire file groups when size limiting is needed, rather than cutting individual issues. This ensures file completeness but is more complex to implement.

The current character-based cutting will work fine initially, but if you want file-group-aware cutting, add logic to:

1. Group issues by file first
2. Calculate size contribution per file group
3. Remove lowest-priority file groups when size limiting is needed
4. Maintain file group integrity

### Step 6: Update Tests

Update all tests in `/mcptools/check_docs_tool_test.go`:

1. **Update the `CheckDocsResult` struct** to match new output format
2. **Update `requireCheckDocsResult` function** to validate grouped structure
3. **Update all test expectations** to work with `issues_by_file` instead of flat `issues` array

Example test update:
```go
// Old test expectation
assert.Len(t, result.Issues, opts.ExpectedIssueCount, "Issues array should match expected count")

// New test expectation  
totalIssuesInGroups := 0
for _, fileGroup := range result.IssuesByFile {
    totalIssuesInGroups += fileGroup.IssueCount
}
assert.Equal(t, opts.ExpectedIssueCount, totalIssuesInGroups, "Total issues in groups should match expected count")
```

## Constraints and Requirements

### MAINTAIN Current Architecture
- **DO NOT CHANGE**: File processing logic (still process all files first)
- **DO NOT CHANGE**: Size limiting approach (still limit after processing)
- **DO NOT CHANGE**: Pagination/offset logic (still works the same way)
- **DO NOT CHANGE**: Tool parameters (no new parameters needed)
- **DO NOT CHANGE**: Priority sorting logic (still sort by priority first)

### Follow Clear Path Style
- Use `goto end` pattern for all functions
- Declare all variables before first `goto`
- Use named return variables
- Follow existing code style in the file

### Size Limiting Compatibility
- The existing size limiting logic should still work
- JSON marshaling and character counting remains the same
- The grouped structure may actually be more compact than flat array

## Testing Requirements

Before submitting changes:
1. **Run existing tests**: `make test` should pass
2. **Test with scout-mcp project**: Verify it still works on the full project
3. **Verify size limiting**: Test that responses stay under character limits
4. **Check pagination**: Ensure offset parameter still works correctly

## Expected Benefits

After implementation:
- **Efficient file updates**: AI can fix all issues in a file with one operation
- **Better prioritization**: Clear view of which files need most work
- **Improved workflow**: Process files systematically rather than jumping around
- **Progress tracking**: Track files completed vs individual issues completed

## Note on Backward Compatibility

This is a **breaking change** to the tool's output format. Any code that depends on the current flat `issues` array will need to be updated to use the new `issues_by_file` structure. This is acceptable since this is an internal tool primarily used with Claude Code.