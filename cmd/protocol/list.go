package protocol

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/internal/cache"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SchemasResponse is the shape of the /api/mcp/schemas response.
type SchemasResponse struct {
	Plugins []Protocol `json:"plugins"`
}

// Protocol represents a single blockchain protocol with its available actions.
type Protocol struct {
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	Actions     []Action `json:"actions"`
}

// Action represents a single action available on a protocol.
type Action struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields"`
}

// Field describes a single input field for an action.
type Field struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List blockchain protocols",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Example: `  # List all protocols (cached)
  kh pr ls

  # Force refresh from API
  kh pr ls --refresh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			refresh, err := cmd.Flags().GetBool("refresh")
			if err != nil {
				return err
			}

			protocols, err := loadProtocols(f, refresh, cmd)
			if err != nil {
				return err
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(protocols, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"NAME", "ACTIONS"})
				for _, proto := range protocols {
					tw.AppendRow(table.Row{proto.Name, len(proto.Actions)})
				}
				if len(protocols) == 0 {
					fmt.Fprintln(f.IOStreams.Out, "No protocols found.")
					return
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
		// Non-fatal: warn but continue with the freshly fetched data
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

// unmarshalProtocols extracts the plugins slice from a raw SchemasResponse.
func unmarshalProtocols(raw json.RawMessage) ([]Protocol, error) {
	var schemas SchemasResponse
	if err := json.Unmarshal(raw, &schemas); err != nil {
		return nil, fmt.Errorf("decoding schemas response: %w", err)
	}
	return schemas.Plugins, nil
}
