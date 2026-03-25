package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/internal/cache"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SchemasResponse is the shape of the /api/mcp/schemas response.
type SchemasResponse struct {
	Actions map[string]SchemaAction `json:"actions"`
}

// SchemaAction represents a single action from the schemas API.
type SchemaAction struct {
	ActionType  string `json:"actionType"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Integration string `json:"integration"`
}

// Protocol represents a grouped set of actions under one integration.
type Protocol struct {
	Name        string `json:"name"`
	ActionCount int    `json:"actionCount"`
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List available plugins and integrations",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Example: `  # List all plugins (cached)
  kh plugin ls

  # Force refresh from API
  kh plugin ls --refresh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			refresh, err := cmd.Flags().GetBool("refresh")
			if err != nil {
				return err
			}

			protocols, err := loadProtocols(f, refresh, cmd)
			if err != nil {
				return err
			}

			if len(protocols) == 0 {
				fmt.Fprintln(f.IOStreams.Out, "No protocols found.")
				return nil
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(protocols, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"NAME", "ACTIONS"})
				for _, proto := range protocols {
					tw.AppendRow(table.Row{proto.Name, proto.ActionCount})
				}
				tw.Render()
			})
		},
	}

	cmd.Flags().Bool("refresh", false, "Bypass local cache and fetch fresh data")

	return cmd
}

// loadProtocols loads protocol data from cache or network, following the
// cache-first strategy with stale-while-error fallback.
func loadProtocols(f *cmdutil.Factory, refresh bool, cmd *cobra.Command) ([]Protocol, error) {
	var staleEntry *cache.CacheEntry

	if !refresh {
		entry, err := cache.ReadCache(cache.ProtocolCacheName)
		if err == nil {
			if !cache.IsStale(entry, cache.ProtocolCacheTTL) {
				return unmarshalProtocols(entry.Data)
			}
			staleEntry = entry
		}
	}

	// Fetch fresh data
	raw, fetchErr := fetchSchemas(f, cmd)
	if fetchErr != nil {
		if staleEntry != nil {
			fmt.Fprintln(f.IOStreams.ErrOut, "Warning: using cached data (could not reach API)")
			return unmarshalProtocols(staleEntry.Data)
		}
		return nil, fmt.Errorf("could not fetch protocols. Run with internet connection to populate cache")
	}

	if err := cache.WriteCache(cache.ProtocolCacheName, raw); err != nil {
		fmt.Fprintf(f.IOStreams.ErrOut, "Warning: could not write cache: %v\n", err)
	}

	return unmarshalProtocols(raw)
}

// fetchSchemas performs a GET /api/mcp/schemas request and returns the raw body.
func fetchSchemas(f *cmdutil.Factory, cmd *cobra.Command) (json.RawMessage, error) {
	client, err := f.HTTPClient()
	if err != nil {
		return nil, fmt.Errorf("creating HTTP client: %w", err)
	}

	cfg, err := f.Config()
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	url := khhttp.BuildBaseURL(cmdutil.ResolveHost(cmd, cfg)) + "/api/mcp/schemas"
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

	return json.RawMessage(body), nil
}

// unmarshalProtocols groups actions by integration to produce a protocol list.
func unmarshalProtocols(raw json.RawMessage) ([]Protocol, error) {
	var schemas SchemasResponse
	if err := json.Unmarshal(raw, &schemas); err != nil {
		return nil, fmt.Errorf("decoding schemas response: %w", err)
	}

	counts := make(map[string]int)
	for _, action := range schemas.Actions {
		name := action.Integration
		if name == "" {
			name = action.Category
		}
		if name == "" {
			continue
		}
		counts[name]++
	}

	protocols := make([]Protocol, 0, len(counts))
	for name, count := range counts {
		protocols = append(protocols, Protocol{Name: name, ActionCount: count})
	}
	sort.Slice(protocols, func(i, j int) bool {
		return protocols[i].Name < protocols[j].Name
	})

	return protocols, nil
}
