package test

// getJSONRPCDetectCurrentProjectTest returns the test definition for detect_current_project tool
import (
	"testing"
	"time"
	
	"github.com/mikeschinkel/scout-mcp/testutil"
)

func TestDetectCurrentProjectToolWithJSONRPC(t *testing.T) {
	fixture := testutil.NewTestFixture("detect-current-project-test")
	// Create multiple project directories with .git to test project detection
	fixture.AddProjectFixture("project1", testutil.ProjectFixtureArgs{
		HasGit: true,
		Permissions: 0755,
		ModifiedTime: time.Now().Add(-24 * time.Hour), // Older project
	})
	fixture.AddProjectFixture("project2", testutil.ProjectFixtureArgs{
		HasGit: true,
		Permissions: 0755,
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
