package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	khhttp "github.com/keeperhub/cli/internal/http"
	"github.com/keeperhub/cli/internal/output"
	"github.com/keeperhub/cli/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// Workflow is the API response shape for a single workflow in the list.
type Workflow struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Enabled    bool   `json:"enabled"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

func workflowStatus(enabled bool) string {
	if enabled {
		return "active"
	}
	return "paused"
}

// maxListPageSize is the per-request page size cap GET /api/workflows
// enforces (requests above this are rejected with a 400 invalid_input). The
// endpoint supports real offset-based pagination via &offset=, with a page
// shorter than the requested limit as the authoritative end-of-list signal,
// so "kh workflow list" pages through it internally rather than being
// limited to a single request.
//
// This value is hand-copied from the API's MAX_PAGE_SIZE in
// lib/pagination.ts (keeperhub/keeperhub, imported by
// app/api/workflows/route.ts) and nothing here detects drift: if the server
// lowers the cap, every "kh wf ls" request 400s until this is updated to
// match; if it raises the cap, this just costs more round trips than
// necessary. Check that file if list requests start failing unexpectedly.
const maxListPageSize = 200

// fetchWorkflowPage performs one GET /api/workflows request at the given
// offset and decodes the result. The caller must keep limit within
// [1, maxListPageSize]; fetchWorkflows, the only caller, already guarantees
// this. limit is intentionally not clamped here: fetchWorkflows decides
// end-of-list by comparing the page it gets back against the pageSize it
// asked for, so silently sending a smaller limit than the caller requested
// would make a truncated page indistinguishable from the real end of the
// list. An out-of-range limit should surface as the 400 the API already
// returns for it, not be masked.
func fetchWorkflowPage(client *khhttp.Client, host, project, tag string, limit, offset int) ([]Workflow, error) {
	query := url.Values{}
	query.Set("limit", strconv.Itoa(limit))
	if offset > 0 {
		query.Set("offset", strconv.Itoa(offset))
	}
	if project != "" {
		query.Set("projectId", project)
	}
	if tag != "" {
		query.Set("tagId", tag)
	}
	reqURL := khhttp.BuildBaseURL(host) + "/api/workflows?" + query.Encode()

	req, err := client.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, cmdutil.NotFoundError{Err: errors.New("workflows not found")}
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, cmdutil.RateLimitError{Err: errors.New("rate limit exceeded")}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, khhttp.NewAPIError(resp)
	}

	var workflows []Workflow
	if err := json.NewDecoder(resp.Body).Decode(&workflows); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return workflows, nil
}

// fetchWorkflows pages through GET /api/workflows via &offset=, accumulating
// results until either limit results have been collected (limit <= 0 means
// unlimited, i.e. --all) or a page comes back shorter than requested, which
// the API guarantees means there is nothing left.
//
// When the loop stops because limit was reached rather than because the
// list ended, it issues one more 1-row probe request so hasMore reports
// accurately whether more workflows exist, instead of guessing from a full
// final page.
func fetchWorkflows(client *khhttp.Client, host, project, tag string, limit int) (result []Workflow, hasMore bool, err error) {
	offset := 0
	for {
		pageSize := maxListPageSize
		if limit > 0 {
			remaining := limit - len(result)
			if remaining <= 0 {
				break
			}
			if remaining < pageSize {
				pageSize = remaining
			}
		}

		page, err := fetchWorkflowPage(client, host, project, tag, pageSize, offset)
		if err != nil {
			return nil, false, err
		}
		result = append(result, page...)
		offset += len(page)

		if len(page) < pageSize {
			// A short page is the API's own end-of-list signal.
			return result, false, nil
		}
	}

	if limit > 0 {
		probe, err := fetchWorkflowPage(client, host, project, tag, 1, offset)
		if err != nil {
			return nil, false, err
		}
		hasMore = len(probe) > 0
	}
	return result, hasMore, nil
}

func NewListCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List workflows",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		Long: fmt.Sprintf(`List workflows.

GET /api/workflows caps each underlying request at %d results, but supports
real pagination via &offset=, so "kh wf ls" pages through it internally.
With --limit N, it requests up to %d results at a time, only as many times
as needed to collect N total, with the final request sized to whatever
remains rather than a full page (or fewer results overall if the org runs
out first). If more results exist beyond --limit, a note is printed to
stderr.

Pass --all to fetch every matching workflow: it drops the default --limit of
30 and pages until the API reports the end of the list. An explicit --limit
passed alongside --all still bounds the result to that count, same as
without --all. --project and --tag can be combined with either form to
scope the (possibly paginated) query to one project or tag.`, maxListPageSize, maxListPageSize),
		Example: `  # List workflows
  kh wf ls

  # List with a higher limit (paginates internally, up to 200 per request)
  kh wf ls --limit 500

  # List every workflow in the org, paginating past the API's page cap
  kh wf ls --all

  # List workflows in a project or with a tag
  kh wf ls --project proj_123
  kh wf ls --tag tag_456`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := f.HTTPClient()
			if err != nil {
				return fmt.Errorf("creating HTTP client: %w", err)
			}

			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			limit, err := cmd.Flags().GetInt("limit")
			if err != nil {
				return err
			}
			if limit < 1 {
				return cmdutil.FlagError{Err: fmt.Errorf("--limit must be at least 1, got %d", limit)}
			}

			project, err := cmd.Flags().GetString("project")
			if err != nil {
				return err
			}

			tag, err := cmd.Flags().GetString("tag")
			if err != nil {
				return err
			}

			all, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}

			// --all means "every workflow"; don't let the --limit default
			// (meant to bound a single page) cap the paginated result unless
			// the caller explicitly asked for a limit too.
			effectiveLimit := limit
			if all && !cmd.Flags().Changed("limit") {
				effectiveLimit = 0
			}

			host := cmdutil.ResolveHost(cmd, cfg)

			workflows, hasMore, err := fetchWorkflows(client, host, project, tag, effectiveLimit)
			if err != nil {
				return err
			}
			if hasMore {
				if all {
					fmt.Fprintf(f.IOStreams.ErrOut, "note: more workflows exist beyond --limit %d; remove --limit for the complete list, or increase --limit.\n", effectiveLimit)
				} else {
					fmt.Fprintf(f.IOStreams.ErrOut, "note: more workflows exist beyond --limit %d; pass --all for the complete list, or increase --limit.\n", effectiveLimit)
				}
			}

			p := output.NewPrinter(f.IOStreams, cmd)
			isTTY := f.IOStreams.IsTerminal()
			if len(workflows) == 0 && !p.IsJSON() {
				fmt.Fprintln(f.IOStreams.Out, "No workflows found.")
				return nil
			}
			return p.PrintData(workflows, func(tw table.Writer) {
				tw.AppendHeader(table.Row{"ID", "NAME", "STATUS", "VISIBILITY", "UPDATED"})
				for _, wf := range workflows {
					status := output.ColorStatus(workflowStatus(wf.Enabled), isTTY, false)
					tw.AppendRow(table.Row{wf.ID, wf.Name, status, wf.Visibility, output.TimeAgo(wf.UpdatedAt)})
				}
				tw.Render()
			})
		},
	}

	cmd.Flags().Int("limit", 30, "Maximum number of workflows to list (paginates internally past the API's per-request cap)")
	cmd.Flags().String("project", "", "Filter workflows by project ID")
	cmd.Flags().String("tag", "", "Filter workflows by tag ID")
	cmd.Flags().Bool("all", false, "List every matching workflow, dropping the default --limit (an explicit --limit still bounds the result) and paginating until the list ends")

	return cmd
}
