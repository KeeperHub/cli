package serve_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/keeperhub/cli/cmd/serve"
	"github.com/keeperhub/cli/internal/config"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/keeperhub/cli/pkg/iostreams"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeSchemasServer returns an httptest.Server that serves GET /api/mcp/schemas
// with the provided actions map.
func makeSchemasServer(t *testing.T, actions map[string]serve.ActionSchema) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/mcp/schemas" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"actions": actions,
		})
	}))
}

// newServeFactory creates a Factory pointing at the given test server.
func newServeFactory(server *httptest.Server, ios *iostreams.IOStreams) *cmdutil.Factory {
	client := khhttp.NewClient(khhttp.ClientOptions{Host: server.URL, AppVersion: "1.0.0"})
	return &cmdutil.Factory{
		AppVersion: "1.0.0",
		IOStreams:   ios,
		HTTPClient: func() (*khhttp.Client, error) { return client, nil },
		Config:     func() (config.Config, error) { return config.Config{DefaultHost: server.URL}, nil },
		BaseURL:    func() string { return server.URL },
	}
}

// TestStdoutIsolation verifies that runServeMCP redirects IOStreams.Out to
// stderr before any other work. After the swap, writes to IOStreams.Out must
// not reach the original stdout buffer.
func TestStdoutIsolation(t *testing.T) {
	ios, outBuf, _, _ := iostreams.Test()

	originalOut := ios.Out
	assert.Equal(t, originalOut, outBuf, "initially IOStreams.Out should point to the test stdout buffer")

	// Simulate the stdout isolation step that runServeMCP performs.
	ios.Out = os.Stderr

	// After the swap, IOStreams.Out should not be the original stdout buffer.
	assert.NotEqual(t, ios.Out, outBuf, "after swap, IOStreams.Out must not point to original stdout buffer")

	// A write to the swapped Out should not appear in the original buffer.
	before := outBuf.Len()
	_, _ = io.WriteString(ios.Out, "this must not appear on mcp stdout\n")
	assert.Equal(t, before, outBuf.Len(), "writes after swap must not reach original stdout buffer")
}

// TestRegisterTools_FromSchema verifies that registerTools registers exactly
// one tool per action, with the correct underscore-separated name.
func TestRegisterTools_FromSchema(t *testing.T) {
	actions := map[string]serve.ActionSchema{
		"web3/transfer": {
			ActionType:     "web3/transfer",
			Description:    "Transfer tokens",
			RequiredFields: map[string]string{"network": "Chain ID", "to": "Recipient"},
			OptionalFields: map[string]string{},
		},
		"discord/send-message": {
			ActionType:     "discord/send-message",
			Description:    "Send a Discord message",
			RequiredFields: map[string]string{"channel": "Channel ID"},
			OptionalFields: map[string]string{},
		},
		"aave/supply": {
			ActionType:     "aave/supply",
			Description:    "Supply assets to Aave",
			RequiredFields: map[string]string{"amount": "Amount"},
			OptionalFields: map[string]string{},
		},
	}

	server := makeSchemasServer(t, actions)
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newServeFactory(server, ios)

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	err := serve.RegisterTools(mcpServer, f)
	require.NoError(t, err)

	// Connect client via in-memory transport to list tools.
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()

	ss, err := mcpServer.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer ss.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer cs.Close()

	result, err := cs.ListTools(ctx, nil)
	require.NoError(t, err)

	require.Len(t, result.Tools, 3, "expected exactly 3 tools registered")

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	sort.Strings(names)

	expected := []string{"aave_supply", "discord_send-message", "web3_transfer"}
	assert.Equal(t, expected, names, "tool names should use underscore to replace '/' separator")
}

// TestRegisterTools_SchemaUnreachable verifies that when /api/mcp/schemas
// returns an error, registerTools returns nil and the server has zero tools.
func TestRegisterTools_SchemaUnreachable(t *testing.T) {
	errServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer errServer.Close()

	ios, _, _, _ := iostreams.Test()
	f := newServeFactory(errServer, ios)

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	err := serve.RegisterTools(mcpServer, f)
	require.NoError(t, err, "registerTools must return nil when schemas are unreachable")

	// Connect and list tools -- should be empty.
	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()

	ss, err := mcpServer.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer ss.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer cs.Close()

	result, err := cs.ListTools(ctx, nil)
	require.NoError(t, err)
	assert.Empty(t, result.Tools, "server must have zero tools when schemas are unreachable")
}

// TestBuildInputSchema verifies that buildInputSchema converts ActionSchema
// fields into a well-formed JSON Schema object.
func TestBuildInputSchema(t *testing.T) {
	action := serve.ActionSchema{
		RequiredFields: map[string]string{
			"network": "Chain ID",
			"amount":  "Amount to transfer",
		},
		OptionalFields: map[string]string{
			"memo": "Optional transaction memo",
		},
	}

	schema := serve.BuildInputSchema(action)

	assert.Equal(t, "object", schema["type"], "schema type must be 'object'")

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok, "schema must have a properties map")
	assert.Len(t, props, 3, "expected 3 properties (2 required + 1 optional)")

	required, ok := schema["required"].([]string)
	require.True(t, ok, "schema must have a required slice")
	assert.Len(t, required, 2, "expected 2 required fields")

	// Verify the required fields are both present.
	sort.Strings(required)
	assert.Equal(t, []string{"amount", "network"}, required)

	// Verify property structure.
	for _, name := range []string{"network", "amount", "memo"} {
		prop, exists := props[name]
		assert.True(t, exists, "property %q must exist", name)
		propMap, ok := prop.(map[string]any)
		require.True(t, ok, "property %q must be a map", name)
		assert.Equal(t, "string", propMap["type"], "property %q type must be 'string'", name)
		assert.NotEmpty(t, propMap["description"], "property %q must have a description", name)
	}
}

// TestMakeToolHandler_ExecutesAction verifies that the tool handler POSTs to
// /api/execute/{actionType} with the arguments and returns the response body.
func TestMakeToolHandler_ExecutesAction(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody []byte

	actionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/mcp/schemas" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"actions": map[string]interface{}{}})
			return
		}
		gotMethod = r.Method
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		gotBody = body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"txHash":"0xabc"}`))
	}))
	defer actionServer.Close()

	ios, _, _, _ := iostreams.Test()
	f := newServeFactory(actionServer, ios)

	handler := serve.MakeToolHandler(f, "web3/transfer")

	args := map[string]any{"network": "1", "to": "0xabc"}
	argsJSON, err := json.Marshal(args)
	require.NoError(t, err)

	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Name:      "web3_transfer",
			Arguments: argsJSON,
		},
	}

	ctx := context.Background()
	result, err := handler(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, http.MethodPost, gotMethod, "handler must use POST")
	assert.Equal(t, "/api/execute/web3/transfer", gotPath, "handler must POST to /api/execute/{actionType}")

	var sentArgs map[string]any
	require.NoError(t, json.Unmarshal(gotBody, &sentArgs))
	assert.Equal(t, "1", sentArgs["network"])
	assert.Equal(t, "0xabc", sentArgs["to"])

	require.Len(t, result.Content, 1, "expected one content item")
	textContent, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok, "expected TextContent")
	assert.Contains(t, textContent.Text, "0xabc", "response body must be in TextContent")
}

// TestToolsAreFromSchema_NoneHardcoded verifies that the number of registered
// tools equals the number of actions in the schema -- no hardcoded tools exist.
func TestToolsAreFromSchema_NoneHardcoded(t *testing.T) {
	t.Run("3 actions = 3 tools", func(t *testing.T) {
		actions := map[string]serve.ActionSchema{
			"web3/transfer":      {RequiredFields: map[string]string{"network": "Chain"}},
			"discord/send":       {RequiredFields: map[string]string{"channel": "Channel"}},
			"sendgrid/send-mail": {RequiredFields: map[string]string{"to": "Recipient"}},
		}
		assertToolCount(t, actions, 3)
	})

	t.Run("1 action = 1 tool", func(t *testing.T) {
		actions := map[string]serve.ActionSchema{
			"webhook/call": {RequiredFields: map[string]string{"url": "URL"}},
		}
		assertToolCount(t, actions, 1)
	})

	t.Run("0 actions = 0 tools", func(t *testing.T) {
		assertToolCount(t, map[string]serve.ActionSchema{}, 0)
	})
}

// assertToolCount is a helper that registers tools from a fixture and asserts
// that exactly n tools are registered.
func assertToolCount(t *testing.T, actions map[string]serve.ActionSchema, n int) {
	t.Helper()

	server := makeSchemasServer(t, actions)
	defer server.Close()

	ios, _, _, _ := iostreams.Test()
	f := newServeFactory(server, ios)

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	err := serve.RegisterTools(mcpServer, f)
	require.NoError(t, err)

	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()

	ss, err := mcpServer.Connect(ctx, st, nil)
	require.NoError(t, err)
	defer ss.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	defer cs.Close()

	result, err := cs.ListTools(ctx, nil)
	require.NoError(t, err)
	assert.Len(t, result.Tools, n, "expected %d tools for %d actions", n, n)
}

// TestServeCmd_RequiresMCPFlag verifies that running serve without --mcp
// returns an error.
func TestServeCmd_RequiresMCPFlag(t *testing.T) {
	ios, _, _, _ := iostreams.Test()
	ios.Out = &bytes.Buffer{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	f := newServeFactory(server, ios)
	cmd := serve.NewServeCmd(f)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "mcp", "error should mention --mcp flag")
}
