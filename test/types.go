package test

type jsonRPC struct {
	Version string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Method  string `json:"method"`
	Params  params `json:"params"`
}

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

type subtest struct {
	name      string
	arguments Arguments
	expected  map[string]any
}

type test struct {
	name      string
	input     jsonRPC
	subtests  map[string][]subtest
	expected  map[string]any
	arguments Arguments
	cliArgs   []string
	wantErr   bool
}

type Arguments any

type params struct {
	Name      string    `json:"name"`
	Arguments Arguments `json:"arguments"`
}

// Argument types for different tool calls - SESSION TOKENS REMOVED

type filepathContent struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
}

type readFilesArgs struct {
	Paths      []string `json:"paths"`
	Extensions []string `json:"extensions,omitempty"`
	Recursive  bool     `json:"recursive,omitempty"`
	Pattern    string   `json:"pattern,omitempty"`
	MaxFiles   int      `json:"max_files,omitempty"`
}

type updateFileArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
}

type updateFileLinesArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
	StartLine  int    `json:"start_line"`
	EndLine    int    `json:"end_line"`
}

type insertFileLinesArgs struct {
	Filepath   string `json:"filepath"`
	NewContent string `json:"new_content"`
	Position   string `json:"position"`
	LineNumber int    `json:"line_number"`
}

type deleteFileLinesArgs struct {
	Filepath  string `json:"filepath"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type insertAtPatternArgs struct {
	Path          string `json:"path"`
	NewContent    string `json:"new_content"`
	BeforePattern string `json:"before_pattern,omitempty"`
	AfterPattern  string `json:"after_pattern,omitempty"`
	Position      string `json:"position,omitempty"`
	Regex         bool   `json:"regex,omitempty"`
}

type replacePatternArgs struct {
	Path           string `json:"path"`
	Pattern        string `json:"pattern"`
	Replacement    string `json:"replacement"`
	Regex          bool   `json:"regex,omitempty"`
	AllOccurrences bool   `json:"all_occurrences,omitempty"`
}

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

type deleteFilesArgs struct {
	Path      string `json:"path"`
	Recursive bool   `json:"recursive,omitempty"`
}

type findFilePartArgs struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	PartType string `json:"part_type"`
	PartName string `json:"part_name"`
}

type replaceFilePartArgs struct {
	Path       string `json:"path"`
	Language   string `json:"language"`
	PartType   string `json:"part_type"`
	PartName   string `json:"part_name"`
	NewContent string `json:"new_content"`
}

type validateFilesArgs struct {
	Paths      []string `json:"paths,omitempty"`
	Files      []string `json:"files,omitempty"`
	Language   string   `json:"language,omitempty"`
	Extensions []string `json:"extensions,omitempty"`
	Recursive  bool     `json:"recursive,omitempty"`
}

type analyzeFilesArgs struct {
	Files []string `json:"files"`
}

// Session token only arguments
type sessionTokenArgs struct {
}

// Request approval arguments
type requestApprovalArgs struct {
	Operation      string   `json:"operation"`
	Files          []string `json:"files"`
	ImpactSummary  string   `json:"impact_summary,omitempty"`
	PreviewContent string   `json:"preview_content,omitempty"`
	RiskLevel      string   `json:"risk_level,omitempty"`
}

// Generate approval token arguments
type generateApprovalTokenArgs struct {
	Operations  []string `json:"operations,omitempty"`
	FileActions []string `json:"file_actions,omitempty"`
}

