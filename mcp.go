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

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type MCPServer struct {
	config          Config
	whitelistedDirs map[string]bool
	mcpServer       *server.MCPServer
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

	// Create MCP server with stdio transport
	s.mcpServer = server.NewMCPServer(
		AppName,
		AppVersion,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
		server.WithPromptCapabilities(false),
		server.WithLogging(),
	)

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
	err = server.ServeStdio(s.mcpServer)

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
	// Register list_files tool (enhanced from search_files)
	listFilesTool := mcp.NewTool("list_files",
		mcp.WithDescription("List files and directories in whitelisted paths"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("Directory path to list"),
		),
		mcp.WithBoolean("recursive",
			mcp.DefaultBool(false),
			mcp.Description("Recursive listing"),
		),
		mcp.WithArray("extensions",
			mcp.Description("Filter by file extensions (e.g., ['.go', '.txt'])"),
		),
		mcp.WithString("pattern",
			mcp.Description("Name pattern to match"),
		),
	)
	s.mcpServer.AddTool(listFilesTool, s.handleListFiles)

	// Register read_file tool
	readFileTool := mcp.NewTool("read_file",
		mcp.WithDescription("Read contents of a file from whitelisted directories"),
		mcp.WithString("path",
			mcp.Required(),
			mcp.Description("File path to read"),
		),
	)
	s.mcpServer.AddTool(readFileTool, s.handleReadFile)

	// Register create_file tool
	createFileTool := mcp.NewTool("create_file",
		mcp.WithDescription("Create a new file in whitelisted directories"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path to create")),
		mcp.WithString("content", mcp.Required(), mcp.Description("File content")),
		mcp.WithBoolean("create_dirs", mcp.DefaultBool(false), mcp.Description("Create parent directories if needed")),
	)
	s.mcpServer.AddTool(createFileTool, s.handleCreateFile)

	// Register update_file tool
	updateFileTool := mcp.NewTool("update_file",
		mcp.WithDescription("Update existing file in whitelisted directories"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File path to update")),
		mcp.WithString("content", mcp.Required(), mcp.Description("New file content")),
	)
	s.mcpServer.AddTool(updateFileTool, s.handleUpdateFile)

	// Register delete_file tool
	deleteFileTool := mcp.NewTool("delete_file",
		mcp.WithDescription("Delete file or directory from whitelisted directories"),
		mcp.WithString("path", mcp.Required(), mcp.Description("File or directory path to delete")),
		mcp.WithBoolean("recursive", mcp.DefaultBool(false), mcp.Description("Delete directory recursively")),
	)
	s.mcpServer.AddTool(deleteFileTool, s.handleDeleteFile)
	return err
}

func (s *MCPServer) handleListFiles(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var path string
	var recursive bool
	var pattern string
	var results []FileSearchResult
	var err error

	logger.Info("Tool called", "tool", "list_files")

	path, err = req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	recursive = req.GetBool("recursive", false)
	pattern = req.GetString("pattern", "")

	logger.Info("Tool arguments parsed", "tool", "list_files", "list_files", "path", path, "recursive", recursive, "pattern", pattern)

	// Check path is allowed
	allowed, err := s.isPathAllowed(path)
	if err != nil || !allowed {
		return mcp.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", path)), nil
	}

	results, err = s.searchFiles(path, pattern, recursive)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Convert results to JSON
	jsonData, err := json.Marshal(map[string]any{
		"path":      path,
		"results":   results,
		"count":     len(results),
		"recursive": recursive,
		"pattern":   pattern,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

func (s *MCPServer) handleReadFile(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var filePath string
	var content string
	var err error

	filePath, err = req.RequireString("path")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Check path is allowed
	allowed, err := s.isPathAllowed(filePath)
	if err != nil || !allowed {
		return mcp.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", filePath)), nil
	}

	content, err = s.readFile(filePath)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(content), nil
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

// Add to registerTools() function:

// Handler functions to add:

func (s *MCPServer) handleCreateFile(_ context.Context, req mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
	var filePath string
	var content string
	var createDirs bool
	var allowed bool
	var fileDir string

	logger.Info("Tool called", "tool", "create_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcp.NewToolResultError(err.Error())
		goto end
	}

	content, err = req.RequireString("content")
	if err != nil {
		result = mcp.NewToolResultError(err.Error())
		goto end
	}

	createDirs = req.GetBool("create_dirs", false)

	logger.Info("Tool arguments parsed", "tool", "create_file", "path", filePath, "create_dirs", createDirs, "content_length", len(content))

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcp.NewToolResultError(fmt.Sprintf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcp.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file already exists
	_, err = os.Stat(filePath)
	if err == nil {
		result = mcp.NewToolResultError(fmt.Sprintf("file already exists: %s", filePath))
		goto end
	}
	if !os.IsNotExist(err) {
		result = mcp.NewToolResultError(fmt.Sprintf("error checking file: %v", err))
		goto end
	}

	// Create parent directories if requested
	if createDirs {
		fileDir = filepath.Dir(filePath)
		err = os.MkdirAll(fileDir, 0755)
		if err != nil {
			result = mcp.NewToolResultError(fmt.Sprintf("failed to create directories: %v", err))
			goto end
		}
	}

	// Create the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		result = mcp.NewToolResultError(fmt.Sprintf("failed to create file: %v", err))
		goto end
	}

	result = mcp.NewToolResultText(fmt.Sprintf("File created successfully: %s (%d bytes)", filePath, len(content)))

end:
	logger.Info("Tool completed", "tool", "create_file", "success", err == nil, "path", filePath)
	return result, err
}

func (s *MCPServer) handleUpdateFile(_ context.Context, req mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
	var filePath string
	var content string
	var allowed bool
	var fileInfo os.FileInfo
	var oldSize int64

	logger.Info("Tool called", "tool", "update_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcp.NewToolResultError(err.Error())
		goto end
	}

	content, err = req.RequireString("content")
	if err != nil {
		result = mcp.NewToolResultError(err.Error())
		goto end
	}

	logger.Info("Tool arguments parsed", "tool", "update_file", "path", filePath, "content_length", len(content))

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcp.NewToolResultError(fmt.Sprintf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcp.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcp.NewToolResultError(fmt.Sprintf("file does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcp.NewToolResultError(fmt.Sprintf("error checking file: %v", err))
		goto end
	}

	// Don't allow updating directories
	if fileInfo.IsDir() {
		result = mcp.NewToolResultError(fmt.Sprintf("cannot update directory: %s", filePath))
		goto end
	}

	oldSize = fileInfo.Size()

	// Update the file
	err = os.WriteFile(filePath, []byte(content), fileInfo.Mode())
	if err != nil {
		result = mcp.NewToolResultError(fmt.Sprintf("failed to update file: %v", err))
		goto end
	}

	result = mcp.NewToolResultText(fmt.Sprintf("File updated successfully: %s (%d -> %d bytes)", filePath, oldSize, len(content)))

end:
	logger.Info("Tool completed", "tool", "update_file", "success", err == nil, "path", filePath)
	return result, err
}

func (s *MCPServer) handleDeleteFile(_ context.Context, req mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
	var filePath string
	var recursive bool
	var allowed bool
	var fileInfo os.FileInfo
	var fileType string

	logger.Info("Tool called", "tool", "delete_file")

	filePath, err = req.RequireString("path")
	if err != nil {
		result = mcp.NewToolResultError(err.Error())
		goto end
	}

	recursive = req.GetBool("recursive", false)

	logger.Info("Tool arguments parsed", "tool", "delete_file", "path", filePath, "recursive", recursive)

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcp.NewToolResultError(fmt.Sprintf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcp.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file/directory exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcp.NewToolResultError(fmt.Sprintf("file or directory does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcp.NewToolResultError(fmt.Sprintf("error checking file: %v", err))
		goto end
	}

	// Determine what we're deleting
	if fileInfo.IsDir() {
		fileType = "directory"
		if !recursive {
			result = mcp.NewToolResultError(fmt.Sprintf("cannot delete directory without recursive flag: %s", filePath))
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
		result = mcp.NewToolResultError(fmt.Sprintf("failed to delete %s: %v", fileType, err))
		goto end
	}

	result = mcp.NewToolResultText(fmt.Sprintf("%s deleted successfully: %s",
		cases.Title(language.English).String(fileType),
		filePath,
	))
	
end:
	logger.Info("Tool completed", "tool", "delete_file", "success", err == nil, "path", filePath, "type", fileType)
	return result, err
}
