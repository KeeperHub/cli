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

func NewUsageCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "usage",
		Short:   "Show billing usage",
		Aliases: []string{"u"},
		Args:    cobra.NoArgs,
		Example: `  # Show current period usage
  kh b u

  # Show usage for a specific period
  kh b u --period 2026-03`,
		RunE: func(cmd *cobra.Command, args []string) error {
			period, err := cmd.Flags().GetString("period")
			if err != nil {
				return err
			}

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
			if period != "" && period != "current" {
				url += "?period=" + period
			}

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

			pct := 0
			if sub.Usage.Limit > 0 {
				pct = (sub.Usage.Executions * 100) / sub.Usage.Limit
			}
			fmt.Fprintf(f.IOStreams.Out, "Executions:  %d / %d (%d%% used)\n", sub.Usage.Executions, sub.Usage.Limit, pct)
			fmt.Fprintf(f.IOStreams.Out, "Overage:     $%.2f\n", sub.OverageCharges)
			return nil
		},
	}

	cmd.Flags().String("period", "current", "Billing period (e.g. 2026-03)")

	return cmd
}
