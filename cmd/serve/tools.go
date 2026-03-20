package serve

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BuildInputSchema converts an ActionSchema's field definitions into a
// JSON Schema object (as map[string]any) suitable for mcp.Tool.InputSchema.
// All fields are typed as "string" with a description from the map value.
func BuildInputSchema(action ActionSchema) map[string]any {
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

// RegisterTools fetches the /api/mcp/schemas endpoint and registers one MCP
// tool per action. Tool names use underscore separators (e.g. "web3_transfer").
// If the schemas fetch fails, a warning is logged to stderr and the server
// starts with zero tools -- this is intentional per design.
func RegisterTools(server *mcp.Server, f *cmdutil.Factory) error {
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
			InputSchema: BuildInputSchema(action),
		}
		server.AddTool(tool, MakeToolHandler(f, at))
	}

	return nil
}

// MakeToolHandler returns a ToolHandler that POSTs to /api/execute/{actionType}
// with the tool call arguments as the JSON body.
func MakeToolHandler(f *cmdutil.Factory, actionType string) mcp.ToolHandler {
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

		url := f.BaseURL() + "/api/execute/" + actionType
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

// getStringArg extracts a string value from the args map.
func getStringArg(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// makeStaticHandler creates a handler for a static tool that calls the
// KeeperHub API directly with a configurable HTTP method and URL pattern.
func makeStaticHandler(
	f *cmdutil.Factory,
	method string,
	buildURL func(args map[string]any, baseURL string) string,
	buildBody func(args map[string]any) ([]byte, error),
) mcp.ToolHandler {
	return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var args map[string]any
		if req.Params.Arguments != nil {
			if unmarshalErr := json.Unmarshal(req.Params.Arguments, &args); unmarshalErr != nil {
				return nil, fmt.Errorf("unmarshaling arguments: %w", unmarshalErr)
			}
		}
		if args == nil {
			args = make(map[string]any)
		}

		client, err := f.HTTPClient()
		if err != nil {
			return nil, fmt.Errorf("creating HTTP client: %w", err)
		}

		baseURL := f.BaseURL()
		targetURL := buildURL(args, baseURL)

		var body io.Reader
		if buildBody != nil {
			bodyBytes, bodyErr := buildBody(args)
			if bodyErr != nil {
				return nil, fmt.Errorf("building request body: %w", bodyErr)
			}
			body = bytes.NewReader(bodyBytes)
		}

		httpReq, err := client.NewRequest(method, targetURL, body)
		if err != nil {
			return nil, fmt.Errorf("building request: %w", err)
		}
		if body != nil {
			httpReq.Header.Set("Content-Type", "application/json")
		}

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

// registerStaticTools registers workflow management and execution tools that
// call KeeperHub API endpoints directly (not via /api/execute/).
func registerStaticTools(server *mcp.Server, f *cmdutil.Factory) {
	// workflow_list -- GET /api/workflows
	server.AddTool(&mcp.Tool{
		Name:        "workflow_list",
		Description: "List all workflows in the current organization",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"limit": map[string]any{
					"type":        "string",
					"description": "Maximum number of workflows to return",
				},
			},
		},
	}, makeStaticHandler(f, http.MethodGet, func(args map[string]any, baseURL string) string {
		u := baseURL + "/api/workflows"
		if limit := getStringArg(args, "limit"); limit != "" {
			u += "?limit=" + limit
		}
		return u
	}, nil))

	// workflow_get -- GET /api/workflows/{id}
	server.AddTool(&mcp.Tool{
		Name:        "workflow_get",
		Description: "Get a workflow by ID, including its nodes and edges",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workflow_id": map[string]any{
					"type":        "string",
					"description": "The workflow ID",
				},
			},
			"required": []string{"workflow_id"},
		},
	}, makeStaticHandler(f, http.MethodGet, func(args map[string]any, baseURL string) string {
		return baseURL + "/api/workflows/" + getStringArg(args, "workflow_id")
	}, nil))

	// workflow_create -- POST /api/workflows/create
	server.AddTool(&mcp.Tool{
		Name:        "workflow_create",
		Description: "Create a new workflow",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Workflow name",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Workflow description",
				},
				"nodes": map[string]any{
					"type":        "string",
					"description": "JSON string of the nodes array",
				},
				"edges": map[string]any{
					"type":        "string",
					"description": "JSON string of the edges array",
				},
			},
			"required": []string{"name"},
		},
	}, makeStaticHandler(f, http.MethodPost, func(args map[string]any, baseURL string) string {
		return baseURL + "/api/workflows/create"
	}, func(args map[string]any) ([]byte, error) {
		body := map[string]any{
			"name": getStringArg(args, "name"),
		}
		if desc := getStringArg(args, "description"); desc != "" {
			body["description"] = desc
		}
		if nodesStr := getStringArg(args, "nodes"); nodesStr != "" {
			var nodes []interface{}
			if err := json.Unmarshal([]byte(nodesStr), &nodes); err != nil {
				return nil, fmt.Errorf("parsing nodes JSON: %w", err)
			}
			body["nodes"] = nodes
		}
		if edgesStr := getStringArg(args, "edges"); edgesStr != "" {
			var edges []interface{}
			if err := json.Unmarshal([]byte(edgesStr), &edges); err != nil {
				return nil, fmt.Errorf("parsing edges JSON: %w", err)
			}
			body["edges"] = edges
		}
		return json.Marshal(body)
	}))

	// workflow_update -- PATCH /api/workflows/{id}
	server.AddTool(&mcp.Tool{
		Name:        "workflow_update",
		Description: "Update an existing workflow",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workflow_id": map[string]any{
					"type":        "string",
					"description": "The workflow ID to update",
				},
				"name": map[string]any{
					"type":        "string",
					"description": "New workflow name",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "New workflow description",
				},
				"nodes": map[string]any{
					"type":        "string",
					"description": "JSON string of the nodes array",
				},
				"edges": map[string]any{
					"type":        "string",
					"description": "JSON string of the edges array",
				},
			},
			"required": []string{"workflow_id"},
		},
	}, makeStaticHandler(f, http.MethodPatch, func(args map[string]any, baseURL string) string {
		return baseURL + "/api/workflows/" + getStringArg(args, "workflow_id")
	}, func(args map[string]any) ([]byte, error) {
		body := make(map[string]any)
		if name := getStringArg(args, "name"); name != "" {
			body["name"] = name
		}
		if desc := getStringArg(args, "description"); desc != "" {
			body["description"] = desc
		}
		if nodesStr := getStringArg(args, "nodes"); nodesStr != "" {
			var nodes []interface{}
			if err := json.Unmarshal([]byte(nodesStr), &nodes); err != nil {
				return nil, fmt.Errorf("parsing nodes JSON: %w", err)
			}
			body["nodes"] = nodes
		}
		if edgesStr := getStringArg(args, "edges"); edgesStr != "" {
			var edges []interface{}
			if err := json.Unmarshal([]byte(edgesStr), &edges); err != nil {
				return nil, fmt.Errorf("parsing edges JSON: %w", err)
			}
			body["edges"] = edges
		}
		return json.Marshal(body)
	}))

	// workflow_delete -- DELETE /api/workflows/{id}
	server.AddTool(&mcp.Tool{
		Name:        "workflow_delete",
		Description: "Delete a workflow by ID. Use force=true to delete workflows that have execution history.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workflow_id": map[string]any{
					"type":        "string",
					"description": "The workflow ID to delete",
				},
				"force": map[string]any{
					"type":        "boolean",
					"description": "Force delete even if the workflow has execution history. This will permanently delete all runs and logs.",
				},
			},
			"required": []string{"workflow_id"},
		},
	}, makeStaticHandler(f, http.MethodDelete, func(args map[string]any, baseURL string) string {
		u := baseURL + "/api/workflows/" + getStringArg(args, "workflow_id")
		if force, ok := args["force"]; ok && force == true {
			u += "?force=true"
		}
		return u
	}, nil))

	// workflow_execute -- POST /api/workflow/{id}/execute
	server.AddTool(&mcp.Tool{
		Name:        "workflow_execute",
		Description: "Execute a workflow by ID",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"workflow_id": map[string]any{
					"type":        "string",
					"description": "The workflow ID to execute",
				},
				"input": map[string]any{
					"type":        "string",
					"description": "JSON string of input data for the execution",
				},
			},
			"required": []string{"workflow_id"},
		},
	}, makeStaticHandler(f, http.MethodPost, func(args map[string]any, baseURL string) string {
		return baseURL + "/api/workflow/" + getStringArg(args, "workflow_id") + "/execute"
	}, func(args map[string]any) ([]byte, error) {
		if inputStr := getStringArg(args, "input"); inputStr != "" {
			var input map[string]interface{}
			if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
				return nil, fmt.Errorf("parsing input JSON: %w", err)
			}
			return json.Marshal(input)
		}
		return []byte("{}"), nil
	}))

	// execution_status -- GET /api/workflows/executions/{id}/status
	server.AddTool(&mcp.Tool{
		Name:        "execution_status",
		Description: "Get the status of a workflow execution",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"execution_id": map[string]any{
					"type":        "string",
					"description": "The execution ID to check",
				},
			},
			"required": []string{"execution_id"},
		},
	}, makeStaticHandler(f, http.MethodGet, func(args map[string]any, baseURL string) string {
		return baseURL + "/api/workflows/executions/" + getStringArg(args, "execution_id") + "/status"
	}, nil))

	// execution_logs -- GET /api/workflows/executions/{id}/logs
	server.AddTool(&mcp.Tool{
		Name:        "execution_logs",
		Description: "Get the logs for a workflow execution",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"execution_id": map[string]any{
					"type":        "string",
					"description": "The execution ID to get logs for",
				},
			},
			"required": []string{"execution_id"},
		},
	}, makeStaticHandler(f, http.MethodGet, func(args map[string]any, baseURL string) string {
		return baseURL + "/api/workflows/executions/" + getStringArg(args, "execution_id") + "/logs"
	}, nil))
}
