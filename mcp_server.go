package scout

//
//type MCPRequest struct {
//	Id      int            `json:"id"`
//	JsonRPC string         `json:"jsonrpc"`
//	Method  string         `json:"method"`
//	Params  map[string]any `json:"params"`
//}
//
//type MCPResponse struct {
//	Id     int       `json:"id"`
//	Result any       `json:"result,omitempty"`
//	Error  *MCPError `json:"error,omitempty"`
//}
//
//type MCPError struct {
//	Code    int    `json:"code"`
//	Message string `json:"message"`
//}
//
//func NewMCPServer(additionalPaths []string, opts MCPServerOpts) (server *MCPServer, err error) {
//	var config Config
//	var configFile *os.File
//	var fileData []byte
//	var configPath string
//	var allPaths []string
//
//	configPath, err = GetConfigPath()
//	if err != nil {
//		goto end
//	}
//
//	if opts.OnlyMode {
//		// Use only the additional paths, ignore config file
//		config = Config{
//			WhitelistedPaths: additionalPaths,
//			Port:             "8080",
//			AllowedOrigins:   []string{"https://claude.ai", "https://*.anthropic.com"},
//		}
//	} else {
//		// Try to load config file
//		configFile, err = os.Open(configPath)
//		if err != nil {
//			// If no config file and no additional paths, this is an error
//			if len(additionalPaths) == 0 {
//				err = fmt.Errorf("no configuration file found and no paths specified")
//				goto end
//			}
//			// Create minimal config with just the additional paths
//			config = Config{
//				WhitelistedPaths: additionalPaths,
//				Port:             "8080",
//				AllowedOrigins:   []string{"https://claude.ai", "https://*.anthropic.com"},
//			}
//		} else {
//			defer mustClose(configFile)
//
//			fileData, err = io.ReadAll(configFile)
//			if err != nil {
//				goto end
//			}
//
//			err = json.Unmarshal(fileData, &config)
//			if err != nil {
//				goto end
//			}
//
//			// Combine config paths with additional paths
//			allPaths = make([]string, 0, len(config.WhitelistedPaths)+len(additionalPaths))
//			allPaths = append(allPaths, config.WhitelistedPaths...)
//			allPaths = append(allPaths, additionalPaths...)
//			config.WhitelistedPaths = allPaths
//		}
//	}
//
//	// Check if we have any paths at all
//	if len(config.WhitelistedPaths) == 0 {
//		err = fmt.Errorf("no whitelisted paths specified in config file or command line")
//		goto end
//	}
//
//	server = &MCPServer{
//		config:          config,
//		whitelistedDirs: make(map[string]bool),
//	}
//
//	err = server.validateAndNormalizePaths()
//	if err != nil {
//		goto end
//	}
//
//end:
//	return server, err
//}
//
//func (s *MCPServer) validateAndNormalizePaths() (err error) {
//	var absPath string
//	var pathInfo os.FileInfo
//
//	for _, path := range s.config.WhitelistedPaths {
//		absPath, err = filepath.Abs(path)
//		if err != nil {
//			goto end
//		}
//
//		pathInfo, err = os.Stat(absPath)
//		if err != nil {
//			goto end
//		}
//
//		if !pathInfo.IsDir() {
//			err = fmt.Errorf("whitelisted path is not a directory: %s", absPath)
//			goto end
//		}
//
//		s.whitelistedDirs[absPath] = true
//		logger.Info("Whitelisted directory", "path", absPath)
//	}
//
//end:
//	return err
//}
//
//func (s *MCPServer) isPathAllowed(targetPath string) (allowed bool, err error) {
//	var absPath string
//
//	absPath, err = filepath.Abs(targetPath)
//	if err != nil {
//		goto end
//	}
//
//	for whitelistedDir := range s.whitelistedDirs {
//		if strings.HasPrefix(absPath, whitelistedDir) {
//			allowed = true
//			goto end
//		}
//	}
//
//	allowed = false
//
//end:
//	return allowed, err
//}
//
//func (s *MCPServer) searchFiles(searchPath, pattern string) (results []FileSearchResult, err error) {
//	var allowed bool
//	var searchDir string
//
//	allowed, err = s.isPathAllowed(searchPath)
//	if err != nil {
//		goto end
//	}
//
//	if !allowed {
//		err = fmt.Errorf("access denied: path not whitelisted: %s", searchPath)
//		goto end
//	}
//
//	searchDir, err = filepath.Abs(searchPath)
//	if err != nil {
//		goto end
//	}
//
//	err = filepath.Walk(searchDir, func(path string, info os.FileInfo, walkErr error) error {
//		var shouldInclude bool
//		var result FileSearchResult
//
//		if walkErr != nil {
//			return nil
//		}
//
//		shouldInclude = pattern == "" || strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pattern))
//		if !shouldInclude {
//			return nil
//		}
//
//		result = FileSearchResult{
//			Path:     path,
//			Name:     info.Name(),
//			Size:     info.Size(),
//			Modified: info.ModTime().Format(time.RFC3339),
//			IsDir:    info.IsDir(),
//		}
//
//		results = append(results, result)
//		return nil
//	})
//
//end:
//	return results, err
//}
//
//func (s *MCPServer) readFile(filePath string) (content string, err error) {
//	var allowed bool
//	var fileData []byte
//
//	allowed, err = s.isPathAllowed(filePath)
//	if err != nil {
//		goto end
//	}
//
//	if !allowed {
//		err = fmt.Errorf("access denied: path not whitelisted: %s", filePath)
//		goto end
//	}
//
//	fileData, err = os.ReadFile(filePath)
//	if err != nil {
//		goto end
//	}
//
//	content = string(fileData)
//
//end:
//	return content, err
//}
//
//func (s *MCPServer) handleToolsListRequest() (result []ToolDefinition) {
//	return []ToolDefinition{
//		{
//			Name:        "search_files",
//			Description: "Search for files in whitelisted directories",
//			InputSchema: map[string]any{
//				"type": "object",
//				"properties": map[string]any{
//					"path": map[string]any{
//						"type":        "string",
//						"description": "Directory path to search in",
//					},
//					"pattern": map[string]any{
//						"type":        "string",
//						"description": "Search pattern (optional)",
//					},
//				},
//				"required": []string{"path"},
//			},
//		},
//		{
//			Name:        "read_file",
//			Description: "Read contents of a file from whitelisted directories",
//			InputSchema: map[string]any{
//				"type": "object",
//				"properties": map[string]any{
//					"path": map[string]any{
//						"type":        "string",
//						"description": "File path to read",
//					},
//				},
//				"required": []string{"path"},
//			},
//		},
//	}
//
//}
//
//func (s *MCPServer) handleToolCall(toolName string, params map[string]any) (result any, err error) {
//	switch toolName {
//	case "search_files":
//		var searchPath string
//		var pattern string
//		var searchResults []FileSearchResult
//		var ok bool
//
//		searchPath, ok = params["path"].(string)
//		if !ok {
//			err = fmt.Errorf("missing or invalid path parameter")
//			goto end
//		}
//
//		if patternVal, exists := params["pattern"]; exists {
//			pattern, _ = patternVal.(string)
//		}
//
//		searchResults, err = s.searchFiles(searchPath, pattern)
//		if err != nil {
//			goto end
//		}
//
//		result = map[string]any{
//			"content": searchResults,
//		}
//
//	case "read_file":
//		var filePath string
//		var content string
//		var ok bool
//
//		filePath, ok = params["path"].(string)
//		if !ok {
//			err = fmt.Errorf("missing or invalid path parameter")
//			goto end
//		}
//
//		content, err = s.readFile(filePath)
//		if err != nil {
//			goto end
//		}
//
//		result = map[string]any{
//			"content": content,
//		}
//
//	default:
//		err = fmt.Errorf("unknown tool: %s", toolName)
//	}
//
//end:
//	return result, err
//}
//
//func (s *MCPServer) handleMCPRequest(req MCPRequest) (response MCPResponse, err error) {
//	var result any
//
//	switch req.Method {
//	case "initialize":
//		result, err = s.handleInitializeRequest(req.Params)
//		if err != nil {
//			goto end
//		}
//
//	case "tools/list":
//		result = ServerCapabilities{
//			Tools: s.handleToolsListRequest(),
//		}
//
//	case "tools/call":
//		var toolName string
//		var params map[string]any
//		var ok bool
//
//		toolName, ok = req.Params["name"].(string)
//		if !ok {
//			err = fmt.Errorf("missing tool name")
//			goto end
//		}
//
//		if argsVal, exists := req.Params["arguments"]; exists {
//			params, ok = argsVal.(map[string]any)
//			if !ok {
//				err = fmt.Errorf("invalid arguments format")
//				goto end
//			}
//		}
//
//		result, err = s.handleToolCall(toolName, params)
//		if err != nil {
//			goto end
//		}
//
//	default:
//		err = fmt.Errorf("unsupported method: %s", req.Method)
//	}
//
//end:
//	response = MCPResponse{Id: req.Id}
//
//	if err != nil {
//		response.Error = &MCPError{
//			Code:    -1,
//			Message: err.Error(),
//		}
//	} else {
//		response.Result = result
//	}
//
//	return response, err
//}
//
//func (s *MCPServer) corsMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		var allowedOrigin string
//
//		origin := r.Header.Get("Origin")
//
//		for _, allowed := range s.config.AllowedOrigins {
//			if origin == allowed || allowed == "*" {
//				allowedOrigin = origin
//				break
//			}
//		}
//
//		if allowedOrigin != "" {
//			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
//		}
//
//		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
//		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
//		w.Header().Set("Access-Control-Allow-Credentials", "true")
//
//		if r.Method == "OPTIONS" {
//			w.WriteHeader(http.StatusOK)
//			return
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}
//
//func (s *MCPServer) handleMCP(w http.ResponseWriter, r *http.Request) {
//	var decoder *json.Decoder
//	var req MCPRequest
//	var response MCPResponse
//	var responseData []byte
//	var err error
//
//	w.Header().Set("Content-Type", "application/json")
//	w.Header().Set("Cache-Control", "no-cache")
//	w.Header().Set("Connection", "keep-alive")
//
//	if r.Method != "POST" {
//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
//		return
//	}
//	logger.Info(fmt.Sprintf("%s %s", r.Method, r.Pattern))
//
//	decoder = json.NewDecoder(r.Body)
//	defer mustClose(r.Body)
//
//	err = decoder.Decode(&req)
//	if err != nil {
//		if err == io.EOF {
//			goto end
//		}
//		logger.Error("Error decoding request", "error", err)
//		goto end
//	}
//
//	logger.Info("Received request", "method", req.Method, "id", req.Id)
//	response, err = s.handleMCPRequest(req)
//	if err != nil {
//		logger.Error("Error handling MCP request", "error", err)
//		goto end
//	}
//
//	responseData, err = json.Marshal(response)
//	if err != nil {
//		logger.Error("Error marshaling response", "error", err)
//		goto end
//	}
//
//	logger.Info("Sending", "response", string(responseData))
//
//	w.WriteHeader(http.StatusOK)
//	_, err = fmt.Fprintf(w, "%s\n\n", responseData)
//	if err != nil {
//		logger.Error("Error writing response data", "error", err)
//		goto end
//	}
//
//	if flusher, ok := w.(http.Flusher); ok {
//		flusher.Flush()
//	}
//end:
//	return
//}
//
//func (s *MCPServer) Start() (err error) {
//	var handler http.Handler
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/mcp", s.handleMCP)
//
//	handler = s.corsMiddleware(mux)
//
//	logger.Info("Starting MCP server", "port", s.config.Port)
//	logger.Info("Whitelisted directories", "paths", s.config.WhitelistedPaths)
//
//	err = http.ListenAndServe(":"+s.config.Port, handler)
//
//	return err
//}
//
//// handleInitializeRequest handles MCP initialization requests
////
////	 JSON
////			{
////			   "method" : "initialize",
////			   "params" : {
////			      "protocolVersion" : "2024-11-05",
////			      "capabilities" : { },
////			      "clientInfo" : {
////			         "name" : "claude-ai",
////			         "version" : "0.1.0"
////			      }
////			   },
////			   "jsonrpc" : "2.0",
////			   "id" : 0
////			}
//func (s *MCPServer) handleInitializeRequest(params map[string]any) (result InitializeResult, err error) {
//	var initParams InitializeParams
//	var jsonBytes []byte
//
//	jsonBytes, err = json.Marshal(params)
//	if err != nil {
//		goto end
//	}
//
//	err = json.Unmarshal(jsonBytes, &initParams)
//	if err != nil {
//		goto end
//	}
//
//	if initParams.ProtocolVersion != "2024-11-05" {
//		err = fmt.Errorf("unsupported protocol version: %s", initParams.ProtocolVersion)
//		goto end
//	}
//
//	result = InitializeResult{
//		ProtocolVersion: "2024-11-05",
//		Capabilities: ServerCapabilities{
//			Tools: s.handleToolsListRequest(), // Empty object, not nil
//		},
//		ServerInfo: Implementation{
//			Name:    AppName,
//			Version: AppVersion,
//		},
//	}
//
//end:
//	return result, err
//}
