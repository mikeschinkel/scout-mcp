package test

import (
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp/fsfix"
)

// sessionTokenArgs represents arguments for session-based tools.
type sessionTokenArgs struct {
}

// TestDetectCurrentProjectToolWithJSONRPC tests the detect_current_project tool via JSON-RPC.
func TestDetectCurrentProjectToolWithJSONRPC(t *testing.T) {
	fixture := fsfix.NewRootFixture("detect-current-project-test")
	// Create multiple project directories with .git to test project detection
	fixture.AddRepoFixture("project1", &fsfix.RepoFixtureArgs{
		ModifiedTime: time.Now().Add(-24 * time.Hour), // Older project
	})
	fixture.AddRepoFixture("project2", &fsfix.RepoFixtureArgs{
		ModifiedTime: time.Now().Add(-1 * time.Hour), // More recent project
	})
	fixture.Setup(t)
	defer fixture.Cleanup()
	
	RunJSONRPCTest(t, fixture, test{
		name: "detect_current_project",
		expected: map[string]any{
			"jsonrpc":               "2.0",
			"result.content.#":      1,
			"result.content.0.type": "text",
		},
		subtests: map[string][]subtest{
			GoFile: {
				{
					arguments: sessionTokenArgs{},
					expected:  nil,
				},
			},
		},
	})
}
