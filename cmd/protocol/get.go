package protocol

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/keeperhub/cli/internal/cache"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// ProtocolDetail holds a protocol name and its actions for the get command.
type ProtocolDetail struct {
	Name    string         `json:"name"`
	Actions []ActionDetail `json:"actions"`
}

// ActionDetail holds a single action's metadata for display.
type ActionDetail struct {
	ActionType  string            `json:"actionType"`
	Label       string            `json:"label"`
	Description string            `json:"description"`
	Fields      map[string]string `json:"requiredFields,omitempty"`
}

func NewGetCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <protocol-name>",
		Short:   "Get protocol details and actions",
		Aliases: []string{"g"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Get protocol reference card
  kh pr g aave

  # Get protocol details as JSON
  kh pr g morpho --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			name := strings.ToLower(args[0])
			refresh, _ := cmd.Flags().GetBool("refresh")

			detail, err := loadProtocolDetail(f, name, refresh, cmd)
			if err != nil {
				return err
			}

			if detail == nil {
				return cmdutil.NotFoundError{Err: fmt.Errorf("protocol %q not found", name)}
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(detail, func(_ table.Writer) {
				renderProtocolDetail(f, detail)
			})
		},
	}

	cmd.Flags().Bool("refresh", false, "Bypass local cache and fetch fresh data")

	return cmd
}

// loadProtocolDetail extracts actions for a specific integration from the schemas cache.
func loadProtocolDetail(f *cmdutil.Factory, name string, refresh bool, cmd *cobra.Command) (*ProtocolDetail, error) {
	var raw []byte

	if !refresh {
		entry, err := cache.ReadCache(cache.ProtocolCacheName)
		if err == nil && !cache.IsStale(entry, cache.ProtocolCacheTTL) {
			raw = entry.Data
		}
	}

	if raw == nil {
		fetched, err := fetchSchemas(f, cmd)
		if err != nil {
			entry, cacheErr := cache.ReadCache(cache.ProtocolCacheName)
			if cacheErr != nil {
				return nil, fmt.Errorf("could not fetch protocols: %w", err)
			}
			raw = entry.Data
		} else {
			raw = fetched
			_ = cache.WriteCache(cache.ProtocolCacheName, fetched)
		}
	}

	var schemas SchemasResponse
	if err := unmarshalSchemasRaw(raw, &schemas); err != nil {
		return nil, err
	}

	var actions []ActionDetail
	for _, action := range schemas.Actions {
		integration := strings.ToLower(action.Integration)
		if integration == "" {
			integration = strings.ToLower(action.Category)
		}
		if integration != name {
			continue
		}
		actions = append(actions, ActionDetail{
			ActionType:  action.ActionType,
			Label:       action.Label,
			Description: action.Description,
		})
	}

	if len(actions) == 0 {
		return nil, nil
	}

	sort.Slice(actions, func(i, j int) bool {
		return actions[i].ActionType < actions[j].ActionType
	})

	return &ProtocolDetail{Name: name, Actions: actions}, nil
}

func unmarshalSchemasRaw(raw []byte, out *SchemasResponse) error {
	return json.Unmarshal(raw, out)
}

// renderProtocolDetail writes a reference card for the protocol to stdout.
func renderProtocolDetail(f *cmdutil.Factory, detail *ProtocolDetail) {
	fmt.Fprintf(f.IOStreams.Out, "%s (%d actions)\n\n", detail.Name, len(detail.Actions))

	for _, action := range detail.Actions {
		fmt.Fprintf(f.IOStreams.Out, "  %s\n", action.Label)
		fmt.Fprintf(f.IOStreams.Out, "    %s\n", action.ActionType)
		if action.Description != "" {
			fmt.Fprintf(f.IOStreams.Out, "    %s\n", action.Description)
		}
		fmt.Fprintln(f.IOStreams.Out)
	}
}
