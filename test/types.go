package test

// jsonRPC represents a JSON-RPC 2.0 request for MCP tool calls.
type jsonRPC struct {
	Version string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Method  string `json:"method"`
	Params  params `json:"params"`
}

// newJsonRPC creates a new JSON-RPC request for the specified tool.
func newJsonRPC(id int, name string) jsonRPC {
	return jsonRPC{
		Version: "2.0",
		Id:      id,
		Method:  "tools/call",
		Params: params{
			Name: name,
		},
	}
}

// subtest represents a sub-test case within a larger test scenario.
type subtest struct {
	name      string
	arguments Arguments
	expected  map[string]any
}

// test represents a complete test case including input, expected output, and CLI arguments.
type test struct {
	name      string
	input     jsonRPC
	subtests  map[string][]subtest
	expected  map[string]any
	arguments Arguments
	cliArgs   []string
	wantErr   bool
}

// Arguments represents the arguments payload for MCP tool calls.
type Arguments any

// params represents the parameters section of a JSON-RPC tool call.
type params struct {
	Name      string    `json:"name"`
	Arguments Arguments `json:"arguments"`
}

// Note: Individual tool argument types have been moved to their respective test files.
// This file now contains only shared types used across multiple tests.
