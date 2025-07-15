package test

// File paths used throughout the test suite
const (
	// Binary paths

	ServerBinaryPath         = "../bin/scout-mcp"
	ServerBinaryPathFallback = "../../bin/scout-mcp" // For running from subdirectories
	ServerSourcePath         = "../cmd/main.go"

	// Test file names and paths

	TestConfigFile    = "test_config.json"
	TestReadmeFile    = "README.md"
	TestMainGoFile    = "src/main.go"
	TestUtilsGoFile   = "src/util.go"
	TestGuideFile     = "docs/guide.md"
	TestAPIFile       = "docs/api.md"
	TestGitignoreFile = ".gitignore"
	TestSampleFile    = "data/sample.txt"

	// Test directories

	TestSrcDir   = "src"
	TestDocsDir  = "docs"
	TestDataDir  = "data"
	TestEmptyDir = "empty"
	TestBuildDir = "build"
	TestTmpDir   = "tmp"

	// Test file content

	TestProjectName = "scout-mcp-test"
	TestFilePrefix  = "scout-mcp-test-"

	// Build commands and messages

	BuildCommand       = "go build -o bin/scout-mcp cmd/main.go"
	BinaryNotFoundMsg  = "Scout-mcp binary not found. Please run: " + BuildCommand
	SourceNotFoundMsg  = "Cannot find scout-mcp binary or source"
	BuildFromSourceMsg = "Binary not found, will need to build from source"

	// Test file extensions for filtering

	GoExtension   = ".go"
	JSONExtension = ".json"
	MDExtension   = ".md"
	TXTExtension  = ".txt"
)

// Test file content templates
const (
	TestProjectReadme = `# Test Project

This is a test project for Scout MCP.`

	TestConfigJSON = `{"name": "test-config", "version": "1.0.0"}`

	TestMainGoContent = `package main

func main() {
	println("Hello, World!")
}`

	TestUtilsGoContent = `package main

func helper() string {
	return "helper"
}`

	TestUserGuide = `# User Guide

How to use this application.`

	TestAPIDocumentation = `# API Documentation

API endpoints and usage.`

	TestGitignoreContent = `*.log
*.tmp
bin/`

	TestSampleData = `Sample data file
with multiple lines
for testing`
)
