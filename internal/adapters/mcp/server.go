package mcp

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	mcp_golang "github.com/metoro-io/mcp-golang"
	mcptransport "github.com/metoro-io/mcp-golang/transport"
	mcphttp "github.com/metoro-io/mcp-golang/transport/http"
	"github.com/metoro-io/mcp-golang/transport/stdio"
	"github.com/spf13/cobra"

	"github.com/pmaojo/kthulu-go/internal/adapters/cli/parser"
)

// TransportOptions configure how the MCP server listens for requests.
type TransportOptions struct {
	Kind       string
	ListenAddr string
	HTTPPath   string
}

// ServerOptions describe how to assemble a new MCP server instance.
type ServerOptions struct {
	WorkingDir   string
	AllowList    []string
	DenyList     []string
	Transport    TransportOptions
	Name         string
	Version      string
	Instructions string
}

// ServerInstance captures the assembled MCP server and metadata useful for callers.
type ServerInstance struct {
	Server   *mcp_golang.Server
	Endpoint string
	Tools    []RegisteredTool
	Done     <-chan struct{}
}

// ServerBuilder assembles MCP servers that expose kthulu CLI commands.
type ServerBuilder struct {
	factory *ToolFactory
}

// ServerBuilderDependencies configure the builder with shared dependencies.
type ServerBuilderDependencies struct {
	RootCmd   *cobra.Command
	Executor  CommandExecutor
	TagParser *parser.TagParser
}

// NewServerBuilder constructs a builder capable of producing MCP servers.
func NewServerBuilder(deps ServerBuilderDependencies) *ServerBuilder {
	factory := NewToolFactory(deps.RootCmd, deps.Executor, deps.TagParser)
	return &ServerBuilder{factory: factory}
}

// BuildServer assembles a ready-to-serve MCP server based on the provided options.
func (b *ServerBuilder) BuildServer(options ServerOptions) (*ServerInstance, error) {
	if b == nil || b.factory == nil {
		return nil, fmt.Errorf("mcp server builder is not initialized")
	}

	filter := NewAllowDenyFilter(options.AllowList, options.DenyList)
	tools := b.factory.BuildTools(options.WorkingDir, filter)

	transport, endpoint, done, err := BuildTransport(options.Transport)
	if err != nil {
		return nil, err
	}

	var serverOptions []mcp_golang.ServerOptions
	if trimmed := strings.TrimSpace(options.Name); trimmed != "" {
		serverOptions = append(serverOptions, mcp_golang.WithName(trimmed))
	}
	if trimmed := strings.TrimSpace(options.Version); trimmed != "" {
		serverOptions = append(serverOptions, mcp_golang.WithVersion(trimmed))
	}
	if trimmed := strings.TrimSpace(options.Instructions); trimmed != "" {
		serverOptions = append(serverOptions, mcp_golang.WithInstructions(trimmed))
	}

	server := mcp_golang.NewServer(transport, serverOptions...)
	for _, tool := range tools {
		if err := server.RegisterTool(tool.Name, tool.Description, tool.Handler); err != nil {
			return nil, fmt.Errorf("register tool %s: %w", tool.Name, err)
		}
	}

	return &ServerInstance{Server: server, Endpoint: endpoint, Tools: tools, Done: done}, nil
}

// StdinKeeper wraps an io.Reader and closes Done channel on EOF.
type StdinKeeper struct {
	io.Reader
	Done chan struct{}
	once sync.Once
}

func (k *StdinKeeper) Read(p []byte) (n int, err error) {
	n, err = k.Reader.Read(p)
	if err == io.EOF {
		k.once.Do(func() {
			close(k.Done)
		})
	}
	return n, err
}

// BuildTransport creates the transport requested by the caller and a human readable endpoint string.
// It returns a done channel that is closed when the transport is finished (only applicable for stdio).
func BuildTransport(options TransportOptions) (mcptransport.Transport, string, <-chan struct{}, error) {
	kind := strings.ToLower(strings.TrimSpace(options.Kind))
	switch kind {
	case "", "stdio":
		done := make(chan struct{})
		keeper := &StdinKeeper{Reader: os.Stdin, Done: done}
		return stdio.NewStdioServerTransportWithIO(keeper, os.Stdout), "stdio", done, nil
	case "http":
		path := normalizeHTTPPath(options.HTTPPath)
		transport := mcphttp.NewHTTPTransport(path)
		addr := strings.TrimSpace(options.ListenAddr)
		if addr != "" {
			transport = transport.WithAddr(addr)
		}
		displayAddr := renderHTTPAddress(addr)
		return transport, fmt.Sprintf("%s%s", displayAddr, path), nil, nil
	default:
		return nil, "", nil, fmt.Errorf("unsupported MCP transport %s", options.Kind)
	}
}

func normalizeHTTPPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return "/mcp"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func renderHTTPAddress(addr string) string {
	display := strings.TrimSpace(addr)
	if display == "" {
		display = ":8080"
	}
	if strings.HasPrefix(display, "http://") || strings.HasPrefix(display, "https://") {
		return display
	}
	if strings.HasPrefix(display, ":") {
		display = "localhost" + display
	}
	return "http://" + display
}
