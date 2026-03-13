package org

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewSwitchCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "switch <org-slug>",
		Short:   "Switch to an organization",
		Aliases: []string{"sw"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Switch to an organization by slug
  kh o sw my-org

  # Find org slugs first
  kh o ls`,
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := args[0]

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			baseURL := khhttp.BuildBaseURL(cfg.DefaultHost)

			// Step 1: Resolve slug to organization ID.
			listReq, err := client.NewRequest(http.MethodGet, baseURL+"/api/organizations", nil)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			listResp, err := client.Do(listReq)
			if err != nil {
				return fmt.Errorf("fetching organizations: %w", err)
			}
			defer listResp.Body.Close()

			if listResp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(listResp)
			}

			var orgs []Organization
			if err := json.NewDecoder(listResp.Body).Decode(&orgs); err != nil {
				return fmt.Errorf("decoding organizations: %w", err)
			}

			var target *Organization
			for i := range orgs {
				if orgs[i].Slug == slug {
					target = &orgs[i]
					break
				}
			}
			if target == nil {
				return cmdutil.NotFoundError{Err: fmt.Errorf("organization '%s' not found", slug)}
			}

			// Step 2: POST to set-active with the resolved org ID.
			bodyBytes, err := json.Marshal(map[string]string{"organizationId": target.ID})
			if err != nil {
				return fmt.Errorf("encoding request body: %w", err)
			}

			setActiveReq, err := client.NewRequest(http.MethodPost, baseURL+"/api/auth/organization/set-active", bytes.NewReader(bodyBytes))
			if err != nil {
				return fmt.Errorf("building set-active request: %w", err)
			}
			setActiveReq.Header.Set("Content-Type", "application/json")

			setActiveResp, err := client.Do(setActiveReq)
			if err != nil {
				return fmt.Errorf("switching organization: %w", err)
			}
			defer setActiveResp.Body.Close()

			if setActiveResp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(setActiveResp)
			}

			var sessionResult map[string]interface{}
			if err := json.NewDecoder(setActiveResp.Body).Decode(&sessionResult); err != nil {
				return fmt.Errorf("decoding set-active response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			return p.PrintData(target, func(tw table.Writer) {
				confirmation := fmt.Sprintf("Switched to organization '%s'", target.Name)
				if target.Metadata != nil {
					memberCount, hasMemberCount := target.Metadata["memberCount"]
					plan, hasPlan := target.Metadata["plan"]
					if hasMemberCount && hasPlan {
						count := 0
						switch v := memberCount.(type) {
						case float64:
							count = int(v)
						case int:
							count = v
						}
						confirmation = fmt.Sprintf("Switched to organization '%s' (%d members, %v tier)", target.Name, count, plan)
					}
				}
				fmt.Fprintln(f.IOStreams.Out, confirmation)
			})
		},
	}

	return cmd
}
