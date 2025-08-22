package test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp"
	"github.com/mikeschinkel/scout-mcp/fsfix"
	"github.com/mikeschinkel/scout-mcp/jsontest"
	"github.com/tidwall/gjson"

	_ "github.com/mikeschinkel/scout-mcp/jsontest/pipefuncs"
)

func RunJSONRPCTest(t *testing.T, fixture *fsfix.RootFixture, tt test) {
	var err error
	ctx := context.Background()
	ctx = context.WithValue(ctx, "testing", true)

	// Assign ID dynamically instead of hardcoding in individual test functions
	testFunc := func(t *testing.T, fileType string, st subtest) {
		var input []byte

		argMap := make(map[string]any)
		if st.arguments != nil {
			// Convert the struct to map[string]any for JSON marshaling
			argBytes, _ := json.Marshal(st.arguments)
			err = json.Unmarshal(argBytes, &argMap)
			if err != nil {
				t.Errorf("Error unmarshalling arguments: %v", err)
				return
			}
			
			// Transform relative paths to absolute paths in temp directory
			if fixture != nil {
				transformPaths(argMap, fixture.TempDir())
			}
		}
		if st.name != "start_session" {
			argMap["session_token"] = shared.Token
		}
		tt.input.Params.Arguments = argMap
		input, err = json.Marshal(tt.input)
		if err != nil {
			t.Errorf("Error marshalling input: %v", err)
			return
		}

		// Debug: Print what we're sending
		stdin := strings.NewReader(string(input) + "\n")
		//stdin := strings.NewReader(string(input) + "\n")
		stdout := &bytes.Buffer{}
		
		// Use fixture temp directory as allowed path
		cliArgs := []string{""}
		if fixture != nil {
			cliArgs = append(cliArgs, fixture.TempDir())
		} else {
			cliArgs = append(cliArgs, tt.cliArgs...)
		}
		err = scout.RunMain(ctx, scout.RunArgs{
			Args:   cliArgs,
			Stdin:  stdin,
			Stdout: stdout,
		})
		if err != nil {
			t.Error(err.Error())
		}
		var output []byte
		for i := 0; i < 100; i++ { // Max attempts
			time.Sleep(10 * time.Millisecond)
			output = stdout.Bytes()
			if len(output) > 0 {
				break
			}
		}
		// Log the response for analysis
		err = shared.LogEntryFunc(map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"test":      st.name,
			"tool":      tt.input.Params.Name,
			"response":  json.RawMessage(output),
		})
		if err != nil {
			t.Error(err.Error())
		}

		// Test the response against expected assertions
		// Use subtest-specific expected if available, otherwise use shared expected
		expected := maps.Clone(tt.expected)
		for k, v := range st.expected {
			expected[k] = v
		}
		isError:= gjson.GetBytes(output,"result.isError")
		switch  {
		case !tt.wantErr && isError.Bool():
			errMsg:= gjson.GetBytes(output, "result.content.0.text")
			if !errMsg.Exists() {
				errMsg.Str = "Error returned but without any message: "+string(output)
			}
			t.Error(errMsg.Str)
			return
		case tt.wantErr && !isError.Bool():
			t.Error("Error expected but not received")
			return
		default:
			err = jsontest.TestJSON(output, expected)
			if err != nil {
				t.Error(err)
			}
		}
	}

	tt.input = newJsonRPC(getTestId(), tt.name)

	if tt.subtests == nil {
		// No subtests, run single test
		t.Run(tt.name, func(t *testing.T) {
			testFunc(t,"", subtest{
				name:      tt.name,
				arguments: tt.arguments,
				expected:  tt.expected,
			})
		})
		return
	}
	// Run subtests for each file type
	for fileType, subtests := range tt.subtests {
		t.Run(fileType, func(t *testing.T) {
			for i,subtest := range subtests {
				name:= strconv.Itoa(i+1)
				if subtest.name !=""{
					name = fmt.Sprintf("%s-%d", subtest.name, i+1)
				}
				t.Run(name, func(t *testing.T) {
					testFunc(t, fileType, subtest)
				})
			}
		})
	}

}

// transformPaths converts relative paths in test arguments to absolute paths in temp directory
func transformPaths(argMap map[string]any, tempDir string) {
	for field, value := range argMap {
		switch v := value.(type) {
		case string:
			if isPathField(field) && !filepath.IsAbs(v) {
				argMap[field] = filepath.Join(tempDir, v)
			}
		case []interface{}:
			for i, item := range v {
				if str, ok := item.(string); ok && isPathField(field) && !filepath.IsAbs(str) {
					v[i] = filepath.Join(tempDir, str)
				}
			}
		}
	}
}

// isPathField determines if a field name represents a file path
func isPathField(field string) bool {
	pathFields := []string{"filepath", "path", "paths", "files"}
	for _, pathField := range pathFields {
		if strings.Contains(strings.ToLower(field), pathField) {
			return true
		}
	}
	return false
}

