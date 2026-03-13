package serve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// buildInputSchema converts an ActionSchema's field definitions into a
// JSON Schema object (as map[string]any) suitable for mcp.Tool.InputSchema.
// All fields are typed as "string" with a description from the map value.
func buildInputSchema(action ActionSchema) map[string]any {
	properties := make(map[string]any)
	required := make([]string, 0)

	for name, desc := range action.RequiredFields {
		properties[name] = map[string]any{
			"type":        "string",
			"description": desc,
		}
		required = append(required, name)
	}

	for name, desc := range action.OptionalFields {
		properties[name] = map[string]any{
			"type":        "string",
			"description": desc,
		}
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// registerTools fetches the /api/mcp/schemas endpoint and registers one MCP
// tool per action. Tool names use underscore separators (e.g. "web3_transfer").
// If the schemas fetch fails, a warning is logged to stderr and the server
// starts with zero tools -- this is intentional per design.
func registerTools(server *mcp.Server, f *cmdutil.Factory) error {
	schemas, err := fetchMCPSchemas(f)
	if err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Warning: could not fetch MCP schemas: %v\n", err)
		fmt.Fprintln(f.IOStreams.ErrOut, "Server starting with zero tools.")
		return nil
	}

	for at, action := range schemas.Actions {
		toolName := strings.ReplaceAll(at, "/", "_")
		tool := &mcp.Tool{
			Name:        toolName,
			Description: action.Description,
			InputSchema: buildInputSchema(action),
		}
		server.AddTool(tool, makeToolHandler(f, at))
	}

	return nil
}

// makeToolHandler returns a ToolHandler that POSTs to /api/execute/{actionType}
// with the tool call arguments as the JSON body.
func makeToolHandler(f *cmdutil.Factory, actionType string) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args map[string]any
		if req.Params.Arguments != nil {
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, fmt.Errorf("unmarshaling arguments: %w", err)
			}
		}

		bodyBytes, err := json.Marshal(args)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}

		client, err := f.HTTPClient()
		if err != nil {
			return nil, fmt.Errorf("creating HTTP client: %w", err)
		}

		cfg, err := f.Config()
		if err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}

		url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/execute/" + actionType
		httpReq, err := client.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("executing request: %w", err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		if resp.StatusCode >= 400 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody))},
				},
				IsError: true,
			}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(respBody)},
			},
		}, nil
	}
}
