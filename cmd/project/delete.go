package project

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
		Use:     "delete <project-id>",
		Short:   "Delete a project",
		Aliases: []string{"d", "rm"},
		Args:    cobra.ExactArgs(1),
		Example: `  # Delete a project (will prompt for confirmation)
  kh p delete abc123

  # Delete without prompting
  kh p delete abc123 --yes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]

			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			// Fetch project name for display in the confirmation prompt.
			listURL := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/projects"
			listReq, err := client.NewRequest(http.MethodGet, listURL, nil)
			if err != nil {
				return fmt.Errorf("building list request: %w", err)
			}
			listResp, err := client.Do(listReq)
			if err != nil {
				return fmt.Errorf("fetching projects: %w", err)
			}
			defer listResp.Body.Close()

			var projects []Project
			if err := json.NewDecoder(listResp.Body).Decode(&projects); err != nil {
				return fmt.Errorf("decoding projects: %w", err)
			}

			var projectName string
			for _, proj := range projects {
				if proj.ID == projectID {
					projectName = proj.Name
					break
				}
			}
			if projectName == "" {
				projectName = projectID
			}

			// Confirmation prompt: skip if --yes or non-TTY.
			yes, _ := cmd.Flags().GetBool("yes")
			if !yes && f.IOStreams.IsTerminal() {
				fmt.Fprintf(f.IOStreams.Out, "Delete project '%s'? (y/N) ", projectName)
				scanner := bufio.NewScanner(f.IOStreams.In)
				scanner.Scan()
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					return cmdutil.CancelError{Err: fmt.Errorf("delete cancelled")}
				}
			}

			url := khhttp.BuildBaseURL(cfg.DefaultHost) + "/api/projects/" + projectID
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

			fmt.Fprintf(f.IOStreams.Out, "Deleted project %s\n", projectID)
			return nil
		},
	}

	return cmd
}
