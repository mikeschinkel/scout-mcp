package scout

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcputil"
)

type MCPServer struct {
	config          Config
	whitelistedDirs map[string]bool
	mcpServer       mcputil.Server
}

func NewMCPServer(additionalPaths []string, opts Opts) (s *MCPServer, err error) {
	var config Config
	var whitelistedDirs map[string]bool

	s = &MCPServer{
		whitelistedDirs: make(map[string]bool),
	}

	// Load config and validate paths
	config, whitelistedDirs, err = s.loadConfigAndPaths(additionalPaths, opts)
	if err != nil {
		goto end
	}

	s.config = config
	s.whitelistedDirs = whitelistedDirs

	// Create MCP server with stdio transport using mcputil
	s.mcpServer = mcputil.NewServer(mcputil.ServerOpts{
		Name:        AppName,
		Version:     AppVersion,
		Tools:       true,
		Subscribe:   false,
		ListChanged: false,
		Prompts:     false,
		Logging:     true,
	})

	// Register tools
	err = s.registerTools()
	if err != nil {
		goto end
	}

end:
	return s, err
}

func (s *MCPServer) StartMCP() (err error) {
	logger.Info("MCP server starting with stdio transport")
	logger.Info("Whitelisted directories:")
	for dir := range s.whitelistedDirs {
		logger.Info("Directory", "path", dir)
	}

	// Start the stdio server (blocking)
	err = s.mcpServer.ServeStdio()

	return err
}

func (s *MCPServer) WhitelistedDirs() map[string]bool {
	return s.whitelistedDirs
}

func (s *MCPServer) ReloadConfig() (err error) {
	var config Config
	var whitelistedDirs map[string]bool

	// Reload config using current additional paths (empty for reload)
	config, whitelistedDirs, err = s.loadConfigAndPaths([]string{}, Opts{OnlyMode: false})
	if err != nil {
		goto end
	}

	s.config = config
	s.whitelistedDirs = whitelistedDirs

	logger.Info("Configuration reloaded")
	logger.Info("Updated whitelisted directories:")
	for dir := range s.whitelistedDirs {
		logger.Info("Directory", "path", dir)
	}

end:
	return err
}

func (s *MCPServer) loadConfigAndPaths(additionalPaths []string, opts Opts) (config Config, dirs map[string]bool, err error) {
	var configFile *os.File
	var fileData []byte
	var configPath string
	var allPaths []string

	dirs = make(map[string]bool)

	configPath, err = GetConfigPath()
	if err != nil {
		goto end
	}

	if opts.OnlyMode {
		// Use only the additional paths, ignore config file
		config = Config{
			WhitelistedPaths: additionalPaths,
			Port:             ConfigPort,
			AllowedOrigins:   []string{"https://claude.ai", "https://*.anthropic.com"},
		}
	} else {
		// Try to load config file
		configFile, err = os.Open(configPath)
		if err != nil {
			// If no config file and no additional paths, this is an error
			if len(additionalPaths) == 0 {
				err = fmt.Errorf("no configuration file found and no paths specified")
				goto end
			}
			// Create minimal config with just the additional paths
			config = Config{
				WhitelistedPaths: additionalPaths,
				Port:             ConfigPort,
				AllowedOrigins:   []string{"https://claude.ai", "https://*.anthropic.com"},
			}
		} else {
			defer mustClose(configFile)

			fileData, err = io.ReadAll(configFile)
			if err != nil {
				goto end
			}

			err = json.Unmarshal(fileData, &config)
			if err != nil {
				goto end
			}

			// Combine config paths with additional paths
			allPaths = make([]string, 0, len(config.WhitelistedPaths)+len(additionalPaths))
			allPaths = append(allPaths, config.WhitelistedPaths...)
			allPaths = append(allPaths, additionalPaths...)
			config.WhitelistedPaths = allPaths
		}
	}

	// Check if we have any paths at all
	if len(config.WhitelistedPaths) == 0 {
		err = fmt.Errorf("no whitelisted paths specified in config file or command line")
		goto end
	}

	// Validate and normalize paths
	err = s.validateAndNormalizePaths(config.WhitelistedPaths, dirs)
	if err != nil {
		goto end
	}

end:
	return config, dirs, err
}

func (s *MCPServer) validateAndNormalizePaths(paths []string, dirs map[string]bool) (err error) {
	var absPath string
	var pathInfo os.FileInfo

	for _, path := range paths {
		absPath, err = filepath.Abs(path)
		if err != nil {
			goto end
		}

		pathInfo, err = os.Stat(absPath)
		if err != nil {
			goto end
		}

		if !pathInfo.IsDir() {
			err = fmt.Errorf("whitelisted path is not a directory: %s", absPath)
			goto end
		}

		dirs[absPath] = true
		logger.Info("Whitelisted directory", "path", absPath)
	}

end:
	return err
}

func (s *MCPServer) registerTools() (err error) {
	// Register request_approval tool
	err = s.mcpServer.AddTool(s.handleRequestApproval, mcputil.ToolOptions{
		Name:        "request_approval",
		Description: "Request user approval with rich visual formatting",
		Properties: []mcputil.Property{
			mcputil.String("operation", "Brief operation description").Required(),
			mcputil.Array("files", "List of files to be affected").Required(),
			mcputil.String("preview_content", "Code preview or diff content"),
			mcputil.String("risk_level", "Risk level: low, medium, or high"),
			mcputil.String("impact_summary", "Summary of what will change"),
		},
	})
	if err != nil {
		goto end
	}

	// Register generate_approval_token tool
	err = s.mcpServer.AddTool(s.handleGenerateApprovalToken, mcputil.ToolOptions{
		Name:        "generate_approval_token",
		Description: "Generate approval token after user confirmation",
		Properties: []mcputil.Property{
			mcputil.Array("file_actions", "File actions approved").Required(),
			mcputil.Array("operations", "Operations approved (create, update, delete)").Required(),
			mcputil.String("session_id", "Session identifier for this approval"),
		},
	})
	if err != nil {
		goto end
	}

	// Register list_files tool using mcputil
	err = s.mcpServer.AddTool(s.handleRequestApproval,
		mcputil.ToolOptions{
			Name:        "request_approval",
			Description: "Request user approval with rich visual formatting",
			Properties: []mcputil.Property{
				mcputil.String("operation", "Brief operation description").Required(),
				mcputil.Array("files", "List of files to be affected"),
				mcputil.String("preview_content", "Code preview or diff content"),
				mcputil.String("risk_level", "low, medium, or high"),
				mcputil.String("impact_summary", "Summary of what will change"),
			},
		})
	// Register list_files tool using mcputil
	err = s.mcpServer.AddTool(s.handleListFiles, mcputil.ToolOptions{
		Name:        "list_files",
		Description: "List files and directories in whitelisted paths",
		Properties: []mcputil.Property{
			mcputil.String("path", "Directory path to list").Required(),
			mcputil.Bool("recursive", "Recursive listing"),
			mcputil.Array("extensions", "Filter by file extensions (e.g., ['.go', '.txt'])"),
			mcputil.String("pattern", "Name pattern to match"),
		},
	})
	if err != nil {
		goto end
	}

	// Register read_file tool using mcputil
	err = s.mcpServer.AddTool(s.handleReadFile, mcputil.ToolOptions{
		Name:        "read_file",
		Description: "Read contents of a file from whitelisted directories",
		Properties: []mcputil.Property{
			mcputil.String("path", "File path to read").Required(),
		},
	})
	if err != nil {
		goto end
	}

	// Register create_file tool using mcputil
	err = s.mcpServer.AddTool(s.handleCreateFile, mcputil.ToolOptions{
		Name:        "create_file",
		Description: "Create a new file in whitelisted directories",
		Properties: []mcputil.Property{
			mcputil.String("path", "File path to create").Required(),
			mcputil.String("content", "File content").Required(),
			mcputil.Bool("create_dirs", "Create parent directories if needed"),
		},
	})
	if err != nil {
		goto end
	}

	// Register update_file tool using mcputil
	err = s.mcpServer.AddTool(s.handleUpdateFile, mcputil.ToolOptions{
		Name:        "update_file",
		Description: "Update existing file in whitelisted directories",
		Properties: []mcputil.Property{
			mcputil.String("path", "File path to update").Required(),
			mcputil.String("content", "New file content").Required(),
		},
	})
	if err != nil {
		goto end
	}

	// Register delete_file tool using mcputil
	err = s.mcpServer.AddTool(s.handleDeleteFile, mcputil.ToolOptions{
		Name:        "delete_file",
		Description: "Delete file or directory from whitelisted directories",
		Properties: []mcputil.Property{
			mcputil.String("path", "File or directory path to delete").Required(),
			mcputil.Bool("recursive", "Delete directory recursively"),
		},
	})
	if err != nil {
		goto end
	}

end:
	return err
}

//	func (s *MCPServer) handleRequestApproval(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error)  {
//		// MCP server builds the entire formatted approval display
//		formatted := s.buildApprovalDisplay(ApprovalRequest{
//			Operation: req.RequireString("operation"),
//			FileActions:     req.GetArray("files"),
//			Preview:   req.GetString("preview_content"),
//			Risk:      req.GetString("risk_level"),
//			Impact:    req.GetString("impact_summary"),
//		})
//
//		// Returns ready-to-display formatted text
//		result = mcputil.NewToolResultText(formatted)
//
// end:
//
//		return result, err
//	}
func (s *MCPServer) handleRequestApproval(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var operation string
	var fileActions []FileAction
	var previewContent string
	var riskLevel string
	var impactSummary string
	var formatted string

	logger.Info("Tool called", "tool", "request_approval")

	operation, err = req.RequireString("operation")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	fileActions, err = s.getFileActions(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}
	previewContent = req.GetString("preview_content", "")
	riskLevel = req.GetString("risk_level", "medium")
	impactSummary = req.GetString("impact_summary", "")

	// Build the formatted approval display
	formatted = s.buildApprovalDisplay(ApprovalRequest{
		Operation:   operation,
		FileActions: fileActions,
		Preview:     previewContent,
		Risk:        riskLevel,
		Impact:      impactSummary,
	})

	result = mcputil.NewToolResultText(formatted)

end:
	return result, err
}

type FileAction struct {
	Action  string `json:"action"`  // create, update, delete
	Path    string `json:"path"`    // file path
	Purpose string `json:"purpose"` // why this file is being modified
}

func (s *MCPServer) buildApprovalDisplay(req ApprovalRequest) string {
	var riskIcon string

	var b strings.Builder

	b.WriteString("┌─ APPROVAL REQUIRED ─────────────────────────────────────┐\n")
	b.WriteString("│\n")
	b.WriteString(fmt.Sprintf("│ 📋 Operation: %s\n", req.Operation))
	b.WriteString("│\n")

	// Files section
	b.WriteString("│ ┌─ FILES TO MODIFY ───────────────────────────────────┐\n")
	for _, fa := range req.FileActions {
		var icon string
		icon = s.getFileActionIcon(fa.Action)
		b.WriteString(fmt.Sprintf("│ │ %s %s %s\n", icon, fa.Action, fa.Path))
		if fa.Purpose != "" {
			b.WriteString(fmt.Sprintf("│ │    └── 🎯 %s\n", fa.Purpose))
		}
	}
	b.WriteString("│ └─────────────────────────────────────────────────────┘\n")

	// Code preview section
	if req.Preview != "" {
		b.WriteString("│\n")
		b.WriteString("│ ┌─ CODE PREVIEW ──────────────────────────────────────┐\n")
		lines := strings.Split(req.Preview, "\n")
		for _, line := range lines {
			b.WriteString(fmt.Sprintf("│ │ %s\n", line))
		}
		b.WriteString("│ └─────────────────────────────────────────────────────┘\n")
	}

	// Risk assessment
	riskIcon = s.getRiskIcon(req.Risk)
	b.WriteString("│\n")
	b.WriteString(fmt.Sprintf("│ %s Risk Level: %s\n", riskIcon, titleCase(req.Risk)))
	b.WriteString(fmt.Sprintf("│ 📊 Impact: %s\n", req.Impact))
	b.WriteString("│\n")
	b.WriteString("│ ❓ Reply 'approve' to continue, 'deny' to cancel\n")
	b.WriteString("└─────────────────────────────────────────────────────────┘")

	return b.String()
}

func (s *MCPServer) handleListFiles(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var path string
	var recursive bool
	var pattern string
	var results []FileSearchResult
	var allowed bool

	logger.Info("Tool called", "tool", "list_files")

	path, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	recursive = req.GetBool("recursive", false)
	pattern = req.GetString("pattern", "")

	logger.Info("Tool arguments parsed", "tool", "list_files", "path", path, "recursive", recursive, "pattern", pattern)

	// Check path is allowed
	allowed, err = s.isPathAllowed(path)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("path validation failed: %v", err))
		goto end
	}

	if !allowed {
		result = mcputil.NewToolResultError(fmt.Errorf("access denied: path not whitelisted: %s", path))
		goto end
	}

	results, err = s.searchFiles(path, pattern, recursive)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	// Convert results to JSON using mcputil
	result = mcputil.NewToolResultJSON(map[string]any{
		"path":      path,
		"results":   results,
		"count":     len(results),
		"recursive": recursive,
		"pattern":   pattern,
	})
end:
	return result, err
}

func (s *MCPServer) handleReadFile(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var allowed bool

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not whitelisted: %s", filePath)
		goto end
	}

	content, err = s.readFile(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	result = mcputil.NewToolResultText(content)

end:
	return result, err
}
func (s *MCPServer) isPathAllowed(targetPath string) (allowed bool, err error) {
	var absPath string

	absPath, err = filepath.Abs(targetPath)
	if err != nil {
		goto end
	}

	for whitelistedDir := range s.whitelistedDirs {
		if strings.HasPrefix(absPath, whitelistedDir) {
			allowed = true
			goto end
		}
	}

	allowed = false

end:
	return allowed, err
}

func (s *MCPServer) searchFiles(searchPath, pattern string, recursive bool) (results []FileSearchResult, err error) {
	var allowed bool
	var searchDir string

	allowed, err = s.isPathAllowed(searchPath)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not whitelisted: %s", searchPath)
		goto end
	}

	searchDir, err = filepath.Abs(searchPath)
	if err != nil {
		goto end
	}

	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, walkErr error) error {
		var shouldInclude bool
		var result FileSearchResult

		if walkErr != nil {
			return nil
		}

		// Skip subdirectories if not recursive
		if !recursive && info.IsDir() && path != searchDir {
			return filepath.SkipDir
		}

		shouldInclude = pattern == "" || strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern))
		if !shouldInclude {
			return nil
		}

		result = FileSearchResult{
			Path:     path,
			Name:     info.Name(),
			Size:     info.Size(),
			Modified: info.ModTime().Format(time.RFC3339),
			IsDir:    info.IsDir(),
		}

		results = append(results, result)
		return nil
	})

end:
	return results, err
}

func (s *MCPServer) readFile(filePath string) (content string, err error) {
	var allowed bool
	var fileData []byte

	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		goto end
	}

	if !allowed {
		err = fmt.Errorf("access denied: path not whitelisted: %s", filePath)
		goto end
	}

	fileData, err = os.ReadFile(filePath)
	if err != nil {
		goto end
	}

	content = string(fileData)

end:
	return content, err
}

func StartMCP(additionalPaths []string, opts Opts) (err error) {
	var mcpServer *MCPServer

	mcpServer, err = NewMCPServer(additionalPaths, opts)
	if err != nil {
		goto end
	}

	err = mcpServer.StartMCP()

end:
	return err
}

func (s *MCPServer) handleCreateFile(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var createDirs bool
	var allowed bool
	var fileDir string

	logger.Info("Tool called", "tool", "create_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	content, err = req.RequireString("content")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	createDirs = req.GetBool("create_dirs", false)

	logger.Info("Tool arguments parsed", "tool", "create_file", "path", filePath, "create_dirs", createDirs, "content_length", len(content))

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcputil.NewToolResultError(fmt.Errorf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file already exists
	_, err = os.Stat(filePath)
	if err == nil {
		result = mcputil.NewToolResultError(fmt.Errorf("file already exists: %s", filePath))
		goto end
	}
	if !os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Errorf("error checking file: %v", err))
		goto end
	}

	// Create parent directories if requested
	if createDirs {
		fileDir = filepath.Dir(filePath)
		err = os.MkdirAll(fileDir, 0755)
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to create directories: %v", err))
		goto end
	}

	// Create the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to create file: %v", err))
		goto end
	}

	logger.Info("Tool completed", "tool", "create_file", "success", true, "path", filePath)
	result = mcputil.NewToolResultText(fmt.Sprintf("File created successfully: %s (%d bytes)", filePath, len(content)))
end:
	return result, err
}

func (s *MCPServer) handleUpdateFile(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var content string
	var allowed bool
	var fileInfo os.FileInfo
	var oldSize int64

	logger.Info("Tool called", "tool", "update_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	content, err = req.RequireString("content")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	logger.Info("Tool arguments parsed", "tool", "update_file", "path", filePath, "content_length", len(content))

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcputil.NewToolResultError(fmt.Errorf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Errorf("file does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("error checking file: %v", err))
		goto end
	}

	// Don't allow updating directories
	if fileInfo.IsDir() {
		result = mcputil.NewToolResultError(fmt.Errorf("cannot update directory: %s", filePath))
		goto end
	}

	oldSize = fileInfo.Size()

	// Update the file
	err = os.WriteFile(filePath, []byte(content), fileInfo.Mode())
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to update file: %v", err))
		goto end
	}

	logger.Info("Tool completed", "tool", "update_file", "success", true, "path", filePath)
	result = mcputil.NewToolResultText(fmt.Sprintf("File updated successfully: %s (%d -> %d bytes)", filePath, oldSize, len(content)))
end:
	return result, err
}

func (s *MCPServer) handleDeleteFile(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var filePath string
	var recursive bool
	var allowed bool
	var fileInfo os.FileInfo
	var fileType string

	logger.Info("Tool called", "tool", "delete_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	recursive = req.GetBool("recursive", false)

	logger.Info("Tool arguments parsed", "tool", "delete_file", "path", filePath, "recursive", recursive)

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcputil.NewToolResultError(fmt.Errorf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file/directory exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Errorf("file or directory does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("error checking file: %v", err))
		goto end
	}

	// Determine what we're deleting
	if fileInfo.IsDir() {
		fileType = "directory"
		if !recursive {
			result = mcputil.NewToolResultError(fmt.Errorf("cannot delete directory without recursive flag: %s", filePath))
			goto end
		}
		// Use RemoveAll for recursive directory deletion
		err = os.RemoveAll(filePath)
	} else {
		fileType = "file"
		// Use Remove for single file deletion
		err = os.Remove(filePath)
	}

	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to delete %s: %v", fileType, err))
		goto end
	}

	logger.Info("Tool completed", "tool", "delete_file", "success", true, "path", filePath, "type", fileType)
	result = mcputil.NewToolResultText(fmt.Sprintf("%s deleted successfully: %s",
		titleCase(fileType),
		filePath,
	))
end:
	return result, err
}

func (s *MCPServer) getFileActions(req mcputil.ToolRequest) ([]FileAction, error) {
	return convertSlice[FileAction](req.GetArray("file_actions", nil))
}
func (s *MCPServer) getOperations(req mcputil.ToolRequest) ([]string, error) {
	return convertSlice[string](req.GetArray("operations", nil))
}

// Tool: generate_approval_token
func (s *MCPServer) handleGenerateToken(req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var token string
	var fileActions []FileAction
	var operations []string

	fileActions, err = s.getFileActions(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}
	operations, err = s.getOperations(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	// MCP server handles all JWT logic
	token, err = s.generateApprovalToken(TokenRequest{
		FileActions: fileActions,
		Operations:  operations,
		ExpiresIn:   time.Hour,
	})

	result = mcputil.NewToolResultText(fmt.Sprintf("Approval token generated: %s", token))

end:
	return result, err
}

// Tool: analyze_files_for_approval
func (s *MCPServer) handleAnalyzeFiles(req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var fileActions []FileAction
	var analysis FileAnalysis

	fileActions, err = s.getFileActions(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}
	analysis = FileAnalysis{
		TotalLines:   s.countTotalLines(fileActions),
		Complexity:   s.assessComplexity(fileActions),
		Dependencies: s.findNewDependencies(fileActions),
		RiskFactors:  s.identifyRiskFactors(fileActions),
	}
	result = mcputil.NewToolResultJSON(analysis)
end:
	return result, err
}

type FileAnalysis struct {
	TotalLines   int      `json:"total_lines"`
	Complexity   string   `json:"complexity"`   // "low", "medium", "high"
	Dependencies []string `json:"dependencies"` // New imports/packages
	RiskFactors  []string `json:"risk_factors"` // Security, breaking changes, etc.
}

func (s *MCPServer) countTotalLines(files []FileAction) int {
	// Implementation to count lines across files
	return 0
}

func (s *MCPServer) assessComplexity(files []FileAction) string {
	// Implementation to assess complexity level
	return ""
}

func (s *MCPServer) findNewDependencies(files []FileAction) []string {
	// Implementation to find new imports
	return []string{}
}

func (s *MCPServer) identifyRiskFactors(files []FileAction) []string {
	// Implementation to identify potential risks
	return []string{}
}
func (s *MCPServer) handleGenerateApprovalToken(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var fileActions []FileAction
	var operations []string
	var sessionID string
	var token string

	logger.Info("Tool called", "tool", "generate_approval_token")

	fileActions, err = s.getFileActions(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}

	operations, err = s.getOperations(req)
	if err != nil {
		result = mcputil.NewToolResultError(err)
		goto end
	}
	sessionID = req.GetString("session_id", "")

	// Generate the JWT token
	token, err = s.generateApprovalToken(TokenRequest{
		FileActions: fileActions,
		Operations:  operations,
		SessionID:   sessionID,
		ExpiresIn:   time.Hour,
	})
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Errorf("failed to generate token: %v", err))
		goto end
	}

	result = mcputil.NewToolResultText(fmt.Sprintf("✅ Approval token generated (expires in 1 hour)\n🔑 Token: %s", token))

end:
	return result, err
}

type ApprovalRequest struct {
	Operation   string
	FileActions []FileAction
	Preview     string
	Risk        string
	Impact      string
}

type TokenRequest struct {
	FileActions []FileAction
	Operations  []string
	SessionID   string
	ExpiresIn   time.Duration
}

func (s *MCPServer) generateApprovalToken(req TokenRequest) (token string, err error) {
	// JWT token generation logic
	return token, err
}

func (s *MCPServer) getFileActionIcon(action string) string {
	switch action {
	case "create":
		return "✨"
	case "update", "modify":
		return "📝"
	case "delete":
		return "🗑️"
	case "move", "rename":
		return "📦"
	default:
		return "📄"
	}
}

func (s *MCPServer) getRiskIcon(riskLevel string) string {
	switch strings.ToLower(riskLevel) {
	case "low":
		return "🟢"
	case "medium":
		return "🟡"
	case "high":
		return "🔴"
	default:
		return "⚪"
	}
}
