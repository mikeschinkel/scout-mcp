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

func (s *MCPServer) handleListFiles(_ context.Context, req mcputil.ToolRequest) (result mcputil.ToolResult, err error) {
	var path string
	var recursive bool
	var pattern string
	var results []FileSearchResult
	var allowed bool

	logger.Info("Tool called", "tool", "list_files")

	path, err = req.RequireString("path")
	if err != nil {
		result = mcputil.NewToolResultError(err.Error())
		goto end
	}

	recursive = req.GetBool("recursive", false)
	pattern = req.GetString("pattern", "")

	logger.Info("Tool arguments parsed", "tool", "list_files", "path", path, "recursive", recursive, "pattern", pattern)

	// Check path is allowed
	allowed, err = s.isPathAllowed(path)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("path validation failed: %v", err))
		goto end
	}

	if !allowed {
		result = mcputil.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", path))
		goto end
	}

	results, err = s.searchFiles(path, pattern, recursive)
	if err != nil {
		result = mcputil.NewToolResultError(err.Error())
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
		result = mcputil.NewToolResultError(err.Error())
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
		result = mcputil.NewToolResultError(err.Error())
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
		result = mcputil.NewToolResultError(err.Error())
		goto end
	}

	content, err = req.RequireString("content")
	if err != nil {
		result = mcputil.NewToolResultError(err.Error())
		goto end
	}

	createDirs = req.GetBool("create_dirs", false)

	logger.Info("Tool arguments parsed", "tool", "create_file", "path", filePath, "create_dirs", createDirs, "content_length", len(content))

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcputil.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file already exists
	_, err = os.Stat(filePath)
	if err == nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("file already exists: %s", filePath))
		goto end
	}
	if !os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Sprintf("error checking file: %v", err))
		goto end
	}

	// Create parent directories if requested
	if createDirs {
		fileDir = filepath.Dir(filePath)
		err = os.MkdirAll(fileDir, 0755)
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("failed to create directories: %v", err))
		goto end
	}

	// Create the file
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("failed to create file: %v", err))
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
		result = mcputil.NewToolResultError(err.Error())
		goto end
	}

	content, err = req.RequireString("content")
	if err != nil {
		result = mcputil.NewToolResultError(err.Error())
		goto end
	}

	logger.Info("Tool arguments parsed", "tool", "update_file", "path", filePath, "content_length", len(content))

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcputil.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Sprintf("file does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("error checking file: %v", err))
		goto end
	}

	// Don't allow updating directories
	if fileInfo.IsDir() {
		result = mcputil.NewToolResultError(fmt.Sprintf("cannot update directory: %s", filePath))
		goto end
	}

	oldSize = fileInfo.Size()

	// Update the file
	err = os.WriteFile(filePath, []byte(content), fileInfo.Mode())
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("failed to update file: %v", err))
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
		result = mcputil.NewToolResultError(err.Error())
		goto end
	}

	recursive = req.GetBool("recursive", false)

	logger.Info("Tool arguments parsed", "tool", "delete_file", "path", filePath, "recursive", recursive)

	// Check path is allowed
	allowed, err = s.isPathAllowed(filePath)
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("path validation failed: %v", err))
		goto end
	}
	if !allowed {
		result = mcputil.NewToolResultError(fmt.Sprintf("access denied: path not whitelisted: %s", filePath))
		goto end
	}

	// Check if file/directory exists
	fileInfo, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		result = mcputil.NewToolResultError(fmt.Sprintf("file or directory does not exist: %s", filePath))
		goto end
	}
	if err != nil {
		result = mcputil.NewToolResultError(fmt.Sprintf("error checking file: %v", err))
		goto end
	}

	// Determine what we're deleting
	if fileInfo.IsDir() {
		fileType = "directory"
		if !recursive {
			result = mcputil.NewToolResultError(fmt.Sprintf("cannot delete directory without recursive flag: %s", filePath))
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
		result = mcputil.NewToolResultError(fmt.Sprintf("failed to delete %s: %v", fileType, err))
		goto end
	}

	logger.Info("Tool completed", "tool", "delete_file", "success", true, "path", filePath, "type", fileType)
	result = mcputil.NewToolResultText(fmt.Sprintf("%s deleted successfully: %s",
		cases.Title(language.English).String(fileType),
		filePath,
	))
end:
	return result, err
}
