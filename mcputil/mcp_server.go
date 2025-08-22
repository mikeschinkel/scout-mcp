package mcputil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents an MCP server instance that can register tools
// and serve requests via stdio transport.
type Server interface {
	AddTool(Tool) error
	ServeStdio(ctx context.Context) error
	Shutdown(context.Context) error
}

// mcpServer implements Server interface using mark3labs/mcp-go library.
// It wraps the underlying MCP server with additional functionality for
// session validation and tool registration.
type mcpServer struct {
	srv    *server.MCPServer
	Stdin  io.Reader
	Stdout io.Writer
}

// ServerOpts contains options for creating an MCP server including
// capability flags and IO configuration for stdio transport.
type ServerOpts struct {
	Name        string
	Version     string
	Tools       bool
	Subscribe   bool // Resource subscribe capability
	ListChanged bool // Resource list changed capability
	Prompts     bool
	Logging     bool
	Stdin       io.Reader
	Stdout      io.Writer
}

// NewServer creates a new MCP server with the given options.
// It configures capabilities and sets up stdio transport for
// communication with MCP clients like Claude Desktop.
func NewServer(opts ServerOpts) Server {
	var serverOpts []server.ServerOption

	if opts.Tools {
		serverOpts = append(serverOpts, server.WithToolCapabilities(true))
	}
	if opts.Subscribe || opts.ListChanged {
		serverOpts = append(serverOpts, server.WithResourceCapabilities(opts.Subscribe, opts.ListChanged))
	}
	if opts.Prompts {
		serverOpts = append(serverOpts, server.WithPromptCapabilities(true))
	}
	if opts.Logging {
		serverOpts = append(serverOpts, withRecovery())
	}

	srv := server.NewMCPServer(opts.Name, opts.Version, serverOpts...)

	return &mcpServer{
		srv:    srv,
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
	}
}

// withRecovery creates a middleware that recovers from panics in tool handlers
// and converts them to proper error responses with stack traces for debugging.
// This ensures that tool panics don't crash the entire MCP server.
func withRecovery() server.ServerOption {
	return server.WithToolHandlerMiddleware(func(next server.ToolHandlerFunc) server.ToolHandlerFunc {
		return func(ctx context.Context, request mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					err = fmt.Errorf(
						"panic recovered in %s tool handler: %v\n\nStack trace:\n%s",
						request.Params.Name,
						r,
						stack,
					)
				}
			}()
			return next(ctx, request)
		}
	})
}

// AddTool registers a tool with the MCP server, validating preconditions
// and wrapping the tool handler with session enforcement and error handling.
func (s *mcpServer) AddTool(tool Tool) (err error) {
	var mcpTool mcpTool

	errs := make([]error, 0)
	opts := tool.Options()

	// Build mcp mcpTool options
	var toolOpts []mcpToolOption
	toolOpts = append(toolOpts, mcpWithDescription(opts.Description))

	// Add properties using PropertyOptionsGetter interface
	for _, prop := range opts.Properties {
		toolOpts = append(toolOpts, prop.mcpToolOption(prop.mcpPropertyOptions()))
	}
	if len(errs) > 0 {
		goto end
	}

	// Create the mcpTool
	mcpTool = mcpNewTool(opts.Name, toolOpts...)

	// Add the mcpTool with wrapper handler
	s.srv.AddTool(mcpTool, func(ctx context.Context, req CallToolRequest) (tr *CallToolResult, err error) {
		var jsonRes *jsonResult
		var errRes *errorResult
		var ok bool
		var result ToolResult

		// Wrap the request
		wrappedReq := &toolRequest{req: req}

		// Check preconditions first
		err = tool.EnsurePreconditions(ctx, wrappedReq)
		if err != nil {
			tr = mcpNewToolResultError(err.Error())
			goto end
		}

		// Call user handler
		result, err = tool.Handle(ctx, wrappedReq)
		if err != nil {
			var internalError *InternalError
			if errors.As(err, &internalError) {
				// This is a system error - log it and return as error to become mcp.INTERNAL_ERROR
				if logger != nil {
					logger.Error("Internal tool error", "tool", req.Params.Name, "error", err)
				}
				goto end
			} else {
				// This is an application error - convert to tool result
				tr = mcpNewToolResultError(err.Error())
				err = nil
				goto end
			}
		}

		// Convert result
		jsonRes, ok = result.(*jsonResult)
		if ok {
			tr = mcpNewToolResultText(jsonRes.json)
			goto end
		}
		errRes, ok = result.(*errorResult)
		if !ok {
			tr = mcpNewToolResultError(errRes.message)
			goto end
		}
		tr = mcpNewToolResultError(fmt.Sprintf("unknown result type: %v", result))
	end:
		return tr, err
	})
end:
	return err
}

// ServeStdio starts the MCP server listening on stdio transport.
// This is a blocking call that handles MCP protocol communication
// until the context is canceled or an error occurs.
func (s *mcpServer) ServeStdio(ctx context.Context) error {
	sss := server.NewStdioServer(s.srv)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()
	cr := NewCapturingReader(s.Stdin)
	err := sss.Listen(ctx, cr, s.Stdout)
	if err != nil {
		err = fmt.Errorf("ERROR: %w [REQUEST: %s]", err, cr.Capture)
	}
	return err
}

// Shutdown gracefully shuts down the MCP server.
// Currently a no-op as mcp-go may not have explicit shutdown.
func (s *mcpServer) Shutdown(context.Context) error {
	// mcp-go may not have explicit shutdown - check docs
	return nil
}

var _ io.Reader = (*CapturingReader)(nil)

// CapturingReader wraps an io.Reader to capture all read data
// for debugging and error reporting purposes.
type CapturingReader struct {
	io.Reader
	Capture []byte
}

// NewCapturingReader creates a new CapturingReader that captures
// all data read from the underlying reader.
func NewCapturingReader(reader io.Reader) *CapturingReader {
	return &CapturingReader{Reader: reader}
}

// Read implements io.Reader by reading from the underlying reader
// while capturing the read data for later inspection.
func (c *CapturingReader) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	if err != nil {
		goto end
	}
	c.Capture = append(c.Capture, p...)
end:
	return n, err
}
