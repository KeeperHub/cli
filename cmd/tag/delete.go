package tag

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

func NewDeleteCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <tag-id>",
		Short:   "Delete a tag",
		Aliases: []string{"d", "rm"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Delete a tag (will prompt for confirmation)
  kh t delete abc123

  # Delete without prompting
  kh t delete abc123 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tagID := args[0]

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			// Fetch tag name for display in the confirmation prompt.
			listURL := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/tags"
			listReq, err := client.NewRequest(http.MethodGet, listURL, nil)
			if err != nil {
				return fmt.Errorf("building list request: %w", err)
			}
			listResp, err := client.Do(listReq)
			if err != nil {
				return fmt.Errorf("fetching tags: %w", err)
			}
			defer listResp.Body.Close()

			var tags []Tag
			if err := json.NewDecoder(listResp.Body).Decode(&tags); err != nil {
				return fmt.Errorf("decoding tags: %w", err)
			}

			var tagName string
			for _, t := range tags {
				if t.ID == tagID {
					tagName = t.Name
					break
				}
			}
			if tagName == "" {
				tagName = tagID
			}

			// Confirmation prompt: skip if --yes or non-TTY.
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes && f.IOStreams.IsTerminal() {
				fmt.Fprintf(f.IOStreams.Out, "Delete tag '%s'? (y/N) ", tagName)
				scanner := bufio.NewScanner(f.IOStreams.In)
				scanner.Scan()
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					return cmdutil.CancelError{Err: fmt.Errorf("delete cancelled")}
				}
			}

			url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/tags/" + tagID
			req, err := client.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				return fmt.Errorf("building request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("executing request: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return khhttp.NewAPIError(resp)
			}

			fmt.Fprintf(f.IOStreams.Out, "Deleted tag %s\n", tagID)
			return nil
		},
	}

	return cmd
}
