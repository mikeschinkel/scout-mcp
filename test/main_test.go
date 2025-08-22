package test

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/mikeschinkel/scout-mcp/langutil/golang"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/mikeschinkel/scout-mcp/testutil"
)

const (
	GoFile         = ".go"
	JavascriptFile = ".js"
	PythonFile     = ".py"
)

var shared = struct {
	LogEntryFunc func(any) error
	Token        string
}{}

func TestMain(m *testing.M) {
	// Setup code here if needed
	// For example: initialize test data, mock services, etc.
	// Create log file for JSON responses
	logFile, err := os.OpenFile("../log/test_responses.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer mustClose(logFile)
	shared.LogEntryFunc = writeLogEntry(logFile)
	shared.Token = getSessionToken()

	logger := testutil.NewTestLogger()
	golang.SetLogger(logger)

	// Run tests
	code := m.Run()

	// Cleanup code here if needed

	os.Exit(code)
}

func writeLogEntry(logFile *os.File) func(any) error {
	return func(entry any) error {
		var logJSON []byte
		logJSON, err := json.Marshal(entry)
		if err != nil {
			log.Printf("failed to marshal log entry: %v [%v]", err, entry)
			goto end
		}
		_, err = logFile.WriteString(string(logJSON) + "\n")
		if err != nil {
			log.Printf("failed to writelog entry to file %v: %v", err, string(logJSON))
			goto end
		}
	end:
		return err
	}
}

// getSessionToken gets a session token using simplified approach
func getSessionToken() string {
	session := mcputil.NewSession()
	err := session.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize session: %v", err)
	}
	return session.Token
}

var (
	testId  int
	idMutex sync.Mutex
)

func getTestId() int {
	idMutex.Lock()
	testId++
	id := testId
	idMutex.Unlock()
	return id
}
