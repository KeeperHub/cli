package billing

import (
	"encoding/json"
	"fmt"
	"net/http"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// SubscriptionResponse is the API response shape for GET /api/billing/subscription.
type SubscriptionResponse struct {
	Subscription struct {
		Plan   string `json:"plan"`
		Status string `json:"status"`
	} `json:"subscription"`
	Usage struct {
		Executions int `json:"executions"`
		Limit      int `json:"limit"`
	} `json:"usage"`
	OverageCharges float64                `json:"overageCharges"`
	Limits         map[string]interface{} `json:"limits"`
}

func NewStatusCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status",
		Short:   "Show billing status",
		Aliases: []string{"st"},
		Args:    cobra.NoArgs,
		Example: `  # Show billing plan and usage
  kh b st

  # Show as JSON
  kh b st --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			host := cmdutil.ResolveHost(cmd, cfg)
			url := khhttp.BuildBaseURL(host) + "/api/billing/subscription"

			req, err := client.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				fmt.Fprintln(f.IOStreams.Out, "Billing is not enabled for this instance.")
				return nil
			}
			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			var sub SubscriptionResponse
			if err := json.NewDecoder(resp.Body).Decode(&sub); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			if p.IsJSON() {
				return p.PrintJSON(sub)
			}

			fmt.Fprintf(f.IOStreams.Out, "Plan:        %s\n", sub.Subscription.Plan)
			fmt.Fprintf(f.IOStreams.Out, "Status:      %s\n", sub.Subscription.Status)
			fmt.Fprintf(f.IOStreams.Out, "Executions:  %d / %d\n", sub.Usage.Executions, sub.Usage.Limit)
			fmt.Fprintf(f.IOStreams.Out, "Overage:     $%.2f\n", sub.OverageCharges)
			return nil
		},
	}

	return cmd
}
