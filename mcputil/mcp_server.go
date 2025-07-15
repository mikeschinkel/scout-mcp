package mcputil

import (
	"context"

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
	srv *server.MCPServer
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
		serverOpts = append(serverOpts, server.WithRecovery())
	}

	srv := server.NewMCPServer(opts.Name, opts.Version, serverOpts...)

	return &mcpServer{srv: srv}
}

func (s *mcpServer) AddTool(tool Tool) (err error) {
	var mcpTool mcp.Tool

	errs := make([]error, 0)
	opts := tool.Options()

	// Build mcp mcpTool options
	var toolOpts []mcp.ToolOption
	toolOpts = append(toolOpts, mcp.WithDescription(opts.Description))

	// Add properties using PropertyOptionsGetter interface
	for _, prop := range opts.Properties {
		toolOpts = append(toolOpts, prop.mcpToolOption(prop.mcpPropertyOptions()))
	}
	if len(errs) > 0 {
		goto end
	}

	// Create the mcpTool
	mcpTool = mcp.NewTool(opts.Name, toolOpts...)

	// Add the mcpTool with wrapper handler
	s.srv.AddTool(mcpTool, func(ctx context.Context, req mcp.CallToolRequest) (tr *mcp.CallToolResult, err error) {
		var txtRes *textResult
		var errRes *errorResult
		var ok bool

		// Wrap the request
		wrappedReq := &toolRequest{req: req}

		// Call user handler
		result, err := tool.Handle(ctx, wrappedReq)
		if err != nil {
			goto end
		}

		// Convert result
		txtRes, ok = result.(*textResult)
		if ok {
			tr = mcp.NewToolResultText(txtRes.text)
			goto end
		}
		errRes, ok = result.(*errorResult)
		if ok {
			tr = mcp.NewToolResultError(errRes.message)
			goto end
		}

		tr = mcp.NewToolResultError("unknown result type")
	end:
		return tr, err
	})
end:
	return err
}

func (s *mcpServer) ServeStdio() error {
	return server.ServeStdio(s.srv)
}

func (s *mcpServer) Shutdown(context.Context) error {
	// mcp-go may not have explicit shutdown - check docs
	return nil
}
