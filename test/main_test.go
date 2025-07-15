package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var (
	// Global test client shared across all tests
	testClient *MCPClient
	testDir    string
	ctx        context.Context
	cancel     context.CancelFunc
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	var exitCode int

	// Setup
	if err := setupTestEnvironment(); err != nil {
		log.Printf("Failed to setup test environment: %v", err)
		exitCode = 1
	} else {
		// Run tests
		exitCode = m.Run()
	}

	// Cleanup
	teardownTestEnvironment()

	os.Exit(exitCode)
}

// setupTestEnvironment initializes the test environment
func setupTestEnvironment() (err error) {
	var serverPath string

	log.Println("🚀 Setting up Scout MCP test environment...")

	// Create context with timeout for all tests
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Minute)

	// Get server binary path
	serverPath, err = findServerBinary()
	if err != nil {
		goto end
	}
	log.Printf("📦 Using server binary: %s", serverPath)

	// Create test directory
	testDir, err = createTestDirectory()
	if err != nil {
		goto end
	}
	log.Printf("📁 Created test directory: %s", testDir)

	// Start MCP client connected to server
	testClient, err = NewMCPClient(serverPath, testDir)
	if err != nil {
		goto end
	}
	log.Println("🔌 Started MCP client")

	// Initialize MCP session
	err = testClient.Initialize(ctx)
	if err != nil {
		goto end
	}
	log.Println("✅ MCP session initialized")

	log.Println("🎯 Test environment ready!")

end:
	return err
}

// teardownTestEnvironment cleans up the test environment
func teardownTestEnvironment() {
	log.Println("🧹 Cleaning up test environment...")

	if testClient != nil {
		mustClose(testClient)
		log.Println("🔌 MCP client closed")
	}

	if testDir != "" {
		logOnError(os.RemoveAll(testDir))
		log.Printf("📁 Removed test directory: %s", testDir)
	}

	if cancel != nil {
		cancel()
		log.Println("⏰ Context cancelled")
	}

	log.Println("✅ Cleanup complete")
}

// findServerBinary locates the scout-mcp binary
func findServerBinary() (serverPath string, err error) {
	// Try the compiled binary first
	if _, err = os.Stat(ServerBinaryPath); err == nil {
		serverPath = ServerBinaryPath
		goto end
	}

	// Try fallback path
	if _, err = os.Stat(ServerBinaryPathFallback); err == nil {
		serverPath = ServerBinaryPathFallback
		goto end
	}

	err = fmt.Errorf(BinaryNotFoundMsg)

end:
	return serverPath, err
}

// createTestDirectory creates a temporary directory with test files
func createTestDirectory() (testDirPath string, err error) {
	var fullPath string
	var filePath, content string
	var dir string
	var testFiles map[string]string
	var emptyDirs []string

	testDirPath, err = os.MkdirTemp("", TestFilePrefix)
	if err != nil {
		goto end
	}

	// Create test file structure using constants
	testFiles = map[string]string{
		TestReadmeFile:    TestProjectReadme,
		TestConfigFile:    TestConfigJSON,
		TestMainGoFile:    TestMainGoContent,
		TestUtilsGoFile:   TestUtilsGoContent,
		TestGuideFile:     TestUserGuide,
		TestAPIFile:       TestAPIDocumentation,
		TestGitignoreFile: TestGitignoreContent,
		TestSampleFile:    TestSampleData,
	}

	for filePath, content = range testFiles {
		fullPath = filepath.Join(testDirPath, filePath)

		// Create parent directories if needed
		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			goto end
		}

		// Create the file
		err = os.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			goto end
		}
	}

	// Create empty directories using constants
	emptyDirs = []string{
		TestEmptyDir,
		TestBuildDir,
		TestTmpDir,
	}

	for _, dir = range emptyDirs {
		err = os.MkdirAll(filepath.Join(testDirPath, dir), 0755)
		if err != nil {
			goto end
		}
	}

end:
	return testDirPath, err
}

// GetTestClient returns the shared test client
func GetTestClient() *MCPClient {
	return testClient
}

// GetTestDir returns the test directory path
func GetTestDir() string {
	return testDir
}

// GetTestContext returns the test context
func GetTestContext() context.Context {
	return ctx
}
