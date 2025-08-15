package mcputil

// JSON-RPC 2.0 specification error codes
const (
	JSONRPCParseError     = -32700 // Parse error: Invalid JSON was received by the server
	JSONRPCInvalidRequest = -32600 // Invalid Request: The JSON sent is not a valid Request object
	JSONRPCMethodNotFound = -32601 // Method not found: The method does not exist / is not available
	JSONRPCInvalidParams  = -32602 // Invalid params: Invalid method parameter(s)
	JSONRPCInternalError  = -32603 // Internal error: Internal JSON-RPC error
	// Custom error codes can be defined in range -32000 to -32099
)
