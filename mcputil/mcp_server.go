package mcputil

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server represents an MCP server
type Server interface {
	AddTool(Tool) error
	ServeStdio() error
	Shutdown(ctx context.Context) error
}

// mcpServer implements Server interface
type mcpServer struct {
	srv    *server.MCPServer
	Stdin  io.Reader
	Stdout io.Writer
}

// ServerOpts contains options for creating an MCP server
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

// NewServer creates a new MCP server with the given options
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
		err = tool.EnsurePreconditions(wrappedReq)
		if err != nil {
			tr = mcpNewToolResultError(err.Error())
			goto end
		}

		// Call user handler
		result, err = tool.Handle(ctx, wrappedReq)
		if err != nil {
			goto end
		}

		// Convert result
		jsonRes, ok = result.(*jsonResult)
		if ok {
			tr = mcpNewToolResultText(jsonRes.json)
			goto end
		}
		errRes, ok = result.(*errorResult)
		if ok {
			tr = mcpNewToolResultError(errRes.message)
			goto end
		}

		tr = mcpNewToolResultError("unknown result type")
	end:
		return tr, err
	})
end:
	return err
}

func (s *mcpServer) ServeStdio() error {
	sss := server.NewStdioServer(s.srv)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigChan
		cancel()
	}()

	return sss.Listen(ctx, s.Stdin, s.Stdout)
}

func (s *mcpServer) Shutdown(context.Context) error {
	// mcp-go may not have explicit shutdown - check docs
	return nil
}
