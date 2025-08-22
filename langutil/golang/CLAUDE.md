# Refactoring Instructions for golang.DocExceptions()

## Current Problem

golang.DocExceptions() currently uses packages.Load() which requires valid Go modules/workspaces and loads entire packages. This creates
complexity with module structure, workspace files, and makes testing difficult.

## Solution             

Refactor to parse individual .go files directly using Go's AST parser.

## Changes Needed

1. Replace packages.Load() with parser.ParseFile()
   - Use go/parser.ParseFile() for single files
   - Use go/parser.ParseDir() for directory scanning when recursive
   - Remove dependency on valid Go modules/workspaces
2. Update DocsExceptionsArgs
   - Keep Path parameter (can be file or directory)
   - Keep Recursive parameter
   - Keep MaxFiles parameter
   - Remove any package-specific parameters
3. File Discovery Logic
   - Non-recursive: If Path is file, analyze that file; if directory, find .go files in that directory only
   - Recursive: If Path is file, analyze that file; if directory, recursively find all .go files
4. AST Analysis Changes
   - Instead of iterating over pkg.CompiledGoFiles, iterate over discovered .go files
   - Parse each file individually with parser.ParseFile(fset, filename, nil, parser.ParseComments)
   - Apply same documentation analysis logic to each parsed AST
5. Error Handling
   - Parse errors should not stop analysis of other files
   - Return aggregate results from all successfully parsed files
   - Consider whether to return parse errors or just skip unparseable files
6. Testing Benefits
   - Tests can create simple .go files without module structure
   - No need for go.mod, go.work, or valid project structure
   - Each test file can focus on specific documentation violations
   - Much simpler test fixtures

## Implementation Notes
- Please keep GoFile and GoDeclaration types in-tact.
- Change GoPackage to be GoDirectory 
- Keep the same DocException return type and analysis logic
- File path handling should work with both absolute and relative paths
- Consider using filepath.Walk() for recursive directory traversal
- Maintain same README.md detection logic for directories

This refactoring will eliminate all the module/workspace complexity and make the function much more reliable and testable.