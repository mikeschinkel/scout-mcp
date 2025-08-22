package mcputil

// JSON-RPC 2.0 specification error codes as defined in RFC 7159.
// These constants provide standardized error codes for MCP protocol communication
// and are used throughout the server for consistent error reporting.
const (
	// JSONRPCParseError indicates invalid JSON was received by the server.
	// An error occurred on the server while parsing the JSON text.
	JSONRPCParseError = -32700

	// JSONRPCInvalidRequest indicates the JSON sent is not a valid Request object.
	// The request object is not a valid JSON-RPC 2.0 request.
	JSONRPCInvalidRequest = -32600

	// JSONRPCMethodNotFound indicates the method does not exist or is not available.
	// The requested remote-procedure does not exist or is not available.
	JSONRPCMethodNotFound = -32601

	// JSONRPCInvalidParams indicates invalid method parameter(s) were provided.
	// Invalid method parameter(s) or wrong parameter types.
	JSONRPCInvalidParams = -32602

	// JSONRPCInternalError indicates an internal JSON-RPC error occurred.
	// Server-side error that is not related to the JSON-RPC protocol.
	JSONRPCInternalError = -32603

	// Custom error codes can be defined in range -32000 to -32099 as per JSON-RPC 2.0 spec.
)
