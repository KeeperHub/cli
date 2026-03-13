package serve

import (
	"context"
	"fmt"
	"os"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

func NewServeCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a server",
		Long: `Start a KeeperHub server process.

Currently only MCP stdio mode is supported. When started with --mcp, the
server speaks the Model Context Protocol over stdin/stdout and registers
tools dynamically from the /api/mcp/schemas endpoint at startup.

All diagnostic output (warnings, errors) is written to stderr. Only
valid JSON-RPC 2.0 messages appear on stdout.`,
		Example: `  # Start an MCP stdio server (for use with Claude, Cursor, etc.)
  kh serve --mcp`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			isMCP, err := cmd.Flags().GetBool("mcp")
			if err != nil {
				return err
			}

			if !isMCP {
				return fmt.Errorf("serve requires --mcp flag")
			}

			return runServeMCP(f)
		},
	}

	cmd.Flags().Bool("mcp", false, "Start MCP stdio server")

	return cmd
}

// runServeMCP starts the MCP stdio server. The very first operation redirects
// IOStreams.Out to stderr so that no accidental writes (from factory calls,
// warnings, etc.) contaminate the stdout JSON-RPC channel.
func runServeMCP(f *cmdutil.Factory) error {
	// CRITICAL: redirect all IOStreams output to stderr before any other work.
	// Any non-JSON-RPC bytes on stdout break the MCP protocol.
	f.IOStreams.Out = os.Stderr

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "keeperhub",
		Version: f.AppVersion,
	}, nil)

	if err := registerTools(server, f); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Warning: tool registration failed: %v\n", err)
	}

	return server.Run(context.Background(), &mcp.StdioTransport{})
}
