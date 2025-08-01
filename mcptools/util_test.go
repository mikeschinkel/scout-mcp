package mcptools_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mikeschinkel/scout-mcp/mcptools"
	"github.com/mikeschinkel/scout-mcp/mcputil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testToken = "test-session-Token" // Unit tests don't validate tokens

// Create set of expected tools
var toolNamesMap = mcptools.ToolNamesMap

func withinTimeframe(t *testing.T, t1, t2 time.Time, d time.Duration) bool {
	t.Helper()
	delta := t1.Sub(t2)
	if delta < 0 {
		delta = -delta
	}
	return delta <= d
}

// containsMatch returns true if the item is contained in the list of items per the func f()
func containsMatch[T any, U any](item U, items []T, f func(U, T) bool) (contains bool) {
	for _, i := range items {
		if f(item, i) {
			contains = true
			goto end
		}
	}
end:
	return contains
}

type CallResult struct {
	mcputil.ToolResult
	Error error
}

func callResult(tr mcputil.ToolResult, err error) CallResult {
	return CallResult{
		ToolResult: tr,
		Error:      err,
	}
}

func getToolResult[R any](t *testing.T, cr CallResult, errMsg string) (r *R, err error) {
	var b []byte
	t.Helper()

	// If there's an error, return it without trying to parse JSON
	if cr.Error != nil {
		err = cr.Error
		goto end
	}

	// Only assert ToolResult is not nil when there's no error
	assert.NotNil(t, cr.ToolResult, errMsg)
	r = new(R)
	b = []byte(cr.ToolResult.Value())
	err = json.Unmarshal(b, &r)
	require.NoError(t, err, "Should be able to parse JSON result")
end:
	return r, err
}
