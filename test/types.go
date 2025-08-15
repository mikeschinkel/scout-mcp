package test

// jsonRPC represents a JSON-RPC 2.0 request for MCP tool calls.
type jsonRPC struct {
	Version string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Method  string `json:"method"`
	Params  params `json:"params"`
}

// newJsonRPC creates a new JSON-RPC request for the specified tool.
func newJsonRPC(id int, name string) jsonRPC {
	return jsonRPC{
		Version: "2.0",
		Id:      id,
		Method:  "tools/call",
		Params: params{
			Name: name,
		},
	}
}

// subtest represents a sub-test case within a larger test scenario.
type subtest struct {
	name      string
	arguments Arguments
	expected  map[string]any
}

// test represents a complete test case including input, expected output, and CLI arguments.
type test struct {
	name      string
	input     jsonRPC
	subtests  map[string][]subtest
	expected  map[string]any
	arguments Arguments
	cliArgs   []string
	wantErr   bool
}

// Arguments represents the arguments payload for MCP tool calls.
type Arguments any

// params represents the parameters section of a JSON-RPC tool call.
type params struct {
	Name      string    `json:"name"`
	Arguments Arguments `json:"arguments"`
}

// Argument types for different tool calls - SESSION TOKENS REMOVED

// filepathContent represents arguments for file content operations.
type filepathContent struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
}

// readFilesArgs represents arguments for the read_files tool.
type readFilesArgs struct {
	Paths      []string `json:"paths"`
	Extensions []string `json:"extensions,omitempty"`
	Recursive  bool     `json:"recursive,omitempty"`
	Pattern    string   `json:"pattern,omitempty"`
	MaxFiles   int      `json:"max_files,omitempty"`
}

// updateFileArgs represents arguments for the update_file tool.
type updateFileArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
}

// updateFileLinesArgs represents arguments for the update_file_lines tool.
type updateFileLinesArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
}

// insertFileLinesArgs represents arguments for the insert_file_lines tool.
type insertFileLinesArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
	Position   string `json:"position"`
	LineNumber int    `json:"line_number"`
}

// deleteFileLinesArgs represents arguments for the delete_file_lines tool.
type deleteFileLinesArgs struct {
	Filepath  string `json:"filepath"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// insertAtPatternArgs represents arguments for the insert_at_pattern tool.
type insertAtPatternArgs struct {
	Path          string `json:"path"`
	NewContent    string `json:"new_content"`
	BeforePattern string `json:"before_pattern,omitempty"`
	AfterPattern  string `json:"after_pattern,omitempty"`
	Position      string `json:"position,omitempty"`
	Regex         bool   `json:"regex,omitempty"`
}

// replacePatternArgs represents arguments for the replace_pattern tool.
type replacePatternArgs struct {
	Path           string `json:"path"`
	Pattern        string `json:"pattern"`
	Replacement    string `json:"replacement"`
	Regex          bool   `json:"regex,omitempty"`
	AllOccurrences bool   `json:"all_occurrences,omitempty"`
}

// searchFilesArgs represents arguments for the search_files tool.
type searchFilesArgs struct {
	Path        string   `json:"path"`
	Recursive   bool     `json:"recursive,omitempty"`
	Extensions  []string `json:"extensions,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	NamePattern string   `json:"name_pattern,omitempty"`
	FilesOnly   bool     `json:"files_only,omitempty"`
	DirsOnly    bool     `json:"dirs_only,omitempty"`
	MaxResults  int      `json:"max_results,omitempty"`
}

// deleteFilesArgs represents arguments for the delete_files tool.
type deleteFilesArgs struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
}

// findFilePartArgs represents arguments for the find_file_part tool.
type findFilePartArgs struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	PartType string `json:"part_type"`
	PartName string `json:"part_name"`
}

// replaceFilePartArgs represents arguments for the replace_file_part tool.
type replaceFilePartArgs struct {
	Path       string `json:"path"`
	Language   string `json:"language"`
	PartType   string `json:"part_type"`
	PartName   string `json:"part_name"`
	NewContent string `json:"new_content"`
}

// validateFilesArgs represents arguments for the validate_files tool.
type validateFilesArgs struct {
	Paths      []string `json:"paths,omitempty"`
	Files      []string `json:"files,omitempty"`
	Language   string   `json:"language,omitempty"`
	Extensions []string `json:"extensions,omitempty"`
	Recursive  bool     `json:"recursive,omitempty"`
}

// analyzeFilesArgs represents arguments for the analyze_files tool.
type analyzeFilesArgs struct {
	Files []string `json:"files"`
}

// sessionTokenArgs represents arguments for session-based tools.
type sessionTokenArgs struct {
}

// requestApprovalArgs represents arguments for the request_approval tool.
type requestApprovalArgs struct {
	Operation      string   `json:"operation"`
	Files          []string `json:"files"`
	ImpactSummary  string   `json:"impact_summary,omitempty"`
	PreviewContent string   `json:"preview_content,omitempty"`
	RiskLevel      string   `json:"risk_level,omitempty"`
}

// generateApprovalTokenArgs represents arguments for the generate_approval_token tool.
type generateApprovalTokenArgs struct {
	Operations  []string `json:"operations,omitempty"`
	FileActions []string `json:"file_actions,omitempty"`
}
