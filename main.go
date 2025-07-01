package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	WhitelistedPaths []string `json:"whitelisted_paths"`
	Port             string   `json:"port"`
	AllowedOrigins   []string `json:"allowed_origins"`
}

type MCPServer struct {
	config          Config
	whitelistedDirs map[string]bool
}

type MCPRequest struct {
	ID     string                 `json:"id"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
}

type MCPResponse struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type FileSearchResult struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
	IsDir    bool   `json:"is_directory"`
}

func NewMCPServer(configPath string) (server *MCPServer, err error) {
	var config Config
	var configFile *os.File
	var fileData []byte

	configFile, err = os.Open(configPath)
	if err != nil {
		goto end
	}
	defer configFile.Close()

	fileData, err = io.ReadAll(configFile)
	if err != nil {
		goto end
	}

	err = json.Unmarshal(fileData, &config)
	if err != nil {
		goto end
	}

	server = &MCPServer{
		config:          config,
		whitelistedDirs: make(map[string]bool),
	}

	err = server.validateAndNormalizePaths()
	if err != nil {
		goto end
	}

end:
	return server, err
}

func (s *MCPServer) validateAndNormalizePaths() (err error) {
	var absPath string
	var pathInfo os.FileInfo

	for _, path := range s.config.WhitelistedPaths {
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

		s.whitelistedDirs[absPath] = true
		log.Printf("Whitelisted directory: %s", absPath)
	}

end:
	return err
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

func (s *MCPServer) searchFiles(searchPath, pattern string) (results []FileSearchResult, err error) {
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

func (s *MCPServer) handleToolsListRequest() (result interface{}) {
	tools := []ToolDefinition{
		{
			Name:        "search_files",
			Description: "Search for files in whitelisted directories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Directory path to search in",
					},
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Search pattern (optional)",
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read contents of a file from whitelisted directories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File path to read",
					},
				},
				"required": []string{"path"},
			},
		},
	}

	result = map[string]interface{}{
		"tools": tools,
	}

	return result
}

func (s *MCPServer) handleToolCall(toolName string, params map[string]interface{}) (result interface{}, err error) {
	switch toolName {
	case "search_files":
		var searchPath string
		var pattern string
		var searchResults []FileSearchResult
		var ok bool

		searchPath, ok = params["path"].(string)
		if !ok {
			err = fmt.Errorf("missing or invalid path parameter")
			goto end
		}

		if patternVal, exists := params["pattern"]; exists {
			pattern, _ = patternVal.(string)
		}

		searchResults, err = s.searchFiles(searchPath, pattern)
		if err != nil {
			goto end
		}

		result = map[string]interface{}{
			"content": searchResults,
		}

	case "read_file":
		var filePath string
		var content string
		var ok bool

		filePath, ok = params["path"].(string)
		if !ok {
			err = fmt.Errorf("missing or invalid path parameter")
			goto end
		}

		content, err = s.readFile(filePath)
		if err != nil {
			goto end
		}

		result = map[string]interface{}{
			"content": content,
		}

	default:
		err = fmt.Errorf("unknown tool: %s", toolName)
	}

end:
	return result, err
}

func (s *MCPServer) handleMCPRequest(req MCPRequest) (response MCPResponse) {
	var result interface{}
	var err error

	switch req.Method {
	case "tools/list":
		result = s.handleToolsListRequest()

	case "tools/call":
		var toolName string
		var params map[string]interface{}
		var ok bool

		toolName, ok = req.Params["name"].(string)
		if !ok {
			err = fmt.Errorf("missing tool name")
			goto end
		}

		if argsVal, exists := req.Params["arguments"]; exists {
			params, ok = argsVal.(map[string]interface{})
			if !ok {
				err = fmt.Errorf("invalid arguments format")
				goto end
			}
		}

		result, err = s.handleToolCall(toolName, params)
		if err != nil {
			goto end
		}

	default:
		err = fmt.Errorf("unsupported method: %s", req.Method)
	}

end:
	response = MCPResponse{ID: req.ID}

	if err != nil {
		response.Error = &MCPError{
			Code:    -1,
			Message: err.Error(),
		}
	} else {
		response.Result = result
	}

	return response
}

func (s *MCPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var allowedOrigin string

		origin := r.Header.Get("Origin")

		for _, allowed := range s.config.AllowedOrigins {
			if origin == allowed || allowed == "*" {
				allowedOrigin = origin
				break
			}
		}

		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *MCPServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	var decoder *json.Decoder
	var req MCPRequest
	var response MCPResponse
	var responseData []byte
	var err error

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	decoder = json.NewDecoder(r.Body)

	for {
		err = decoder.Decode(&req)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error decoding request: %v", err)
			continue
		}

		response = s.handleMCPRequest(req)

		responseData, err = json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}

		fmt.Fprintf(w, "data: %s\n\n", responseData)

		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

func (s *MCPServer) Start() (err error) {
	var handler http.Handler

	mux := http.NewServeMux()
	mux.HandleFunc("/sse", s.handleSSE)

	handler = s.corsMiddleware(mux)

	log.Printf("Starting MCP server on port %s", s.config.Port)
	log.Printf("Whitelisted directories: %v", s.config.WhitelistedPaths)

	err = http.ListenAndServe(":"+s.config.Port, handler)

	return err
}

func createDefaultConfig() (err error) {
	var homeDir string
	var projectsDir string
	var config Config
	var configData []byte
	var configFile *os.File

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	projectsDir = filepath.Join(homeDir, "Projects")

	config = Config{
		WhitelistedPaths: []string{projectsDir},
		Port:             "8080",
		AllowedOrigins:   []string{"https://claude.ai", "https://*.anthropic.com"},
	}

	configData, err = json.MarshalIndent(config, "", "  ")
	if err != nil {
		goto end
	}

	configFile, err = os.Create("config.json")
	if err != nil {
		goto end
	}
	defer mustClose(configFile)

	_, err = configFile.Write(configData)
	if err != nil {
		goto end
	}

	fmt.Printf("Created default config.json\n")
	fmt.Printf("Default whitelisted directory: %s\n", projectsDir)
	fmt.Printf("Please review and modify config.json as needed.\n")

end:
	return err
}

func main() {
	var server *MCPServer
	var err error
	var input string

	if len(os.Args) > 1 && os.Args[1] == "init" {
		err = createDefaultConfig()
		if err != nil {
			log.Fatalf("Failed to create config: %v", err)
		}
		return
	}

	server, err = NewMCPServer("config.json")
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("MCP File Search Server\n")
	fmt.Printf("Whitelisted directories:\n")
	for dir := range server.whitelistedDirs {
		fmt.Printf("  - %s\n", dir)
	}
	fmt.Printf("\nServer will start on port %s\n", server.config.Port)
	fmt.Printf("Press Enter to start server, or 'q' to quit: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "q" || input == "quit" {
		return
	}

	err = server.Start()
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
