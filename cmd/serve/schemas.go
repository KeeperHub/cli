package serve

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
)

// SchemasResponse is the correct shape of the /api/mcp/schemas response.
// The Actions field is a map keyed by actionType (e.g. "web3/check-balance").
// NOTE: This is intentionally separate from cmd/protocol/list.go which uses a
// different (incorrect) response shape for the protocol discovery commands.
type SchemasResponse struct {
	Actions map[string]ActionSchema `json:"actions"`
}

// ActionSchema describes a single action available from the MCP schemas endpoint.
type ActionSchema struct {
	ActionType          string            `json:"actionType"`
	Label               string            `json:"label"`
	Description         string            `json:"description"`
	Category            string            `json:"category"`
	Integration         string            `json:"integration"`
	RequiresCredentials bool              `json:"requiresCredentials"`
	RequiredFields      map[string]string `json:"requiredFields"`
	OptionalFields      map[string]string `json:"optionalFields"`
	OutputFields        map[string]string `json:"outputFields"`
}

// fetchMCPSchemas fetches and decodes the /api/mcp/schemas response.
func fetchMCPSchemas(f *cmdutil.Factory) (*SchemasResponse, error) {
	client, err := f.HTTPClient()
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client: %w", err)
	}

	url := f.BaseURL() + "/api/mcp/schemas"
	req, err := client.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, khhttp.NewAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	var schemas SchemasResponse
	if err := json.Unmarshal(body, &schemas); err != nil {
		return nil, fmt.Errorf("decoding schemas response: %w", err)
	}

	return &schemas, nil
}
