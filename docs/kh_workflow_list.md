## kh workflow list

List workflows

### Synopsis

List workflows.

GET /api/workflows caps each underlying request at 200 results, but supports
real pagination via &offset=, so "kh wf ls" pages through it internally: with
--limit N, it fetches as many 200-result pages as needed to collect N results
(or fewer if the org runs out first). If more results exist beyond --limit,
a note is printed to stderr.

Pass --all to fetch every matching workflow: it drops the default --limit of
30 and pages until the API reports the end of the list. An explicit --limit
passed alongside --all still bounds the result to that count, same as
without --all. --project and --tag can be combined with either form to
scope the (possibly paginated) query to one project or tag.

```
kh workflow list [flags]
```

### Examples

```
  # List workflows
  kh wf ls

  # List with a higher limit (paginates internally past 200 if needed)
  kh wf ls --limit 500

  # List every workflow in the org, paginating past the API's page cap
  kh wf ls --all

  # List workflows in a project or with a tag
  kh wf ls --project proj_123
  kh wf ls --tag tag_456
```

### Options

```
      --all              List every matching workflow, dropping the default --limit (an explicit --limit still bounds the result) and paginating until the list ends
  -h, --help             help for list
      --limit int        Maximum number of workflows to list (paginates internally past the API's per-request cap) (default 30)
      --project string   Filter workflows by project ID
      --tag string       Filter workflows by tag ID
```

### Options inherited from parent commands

```
  -H, --host string   KeeperHub host (default: app.keeperhub.com)
      --jq string     Filter JSON output with a jq expression
      --json          Output as JSON
      --no-color      Disable color output
      --org string    Organization ID to use (overrides default from auth)
  -y, --yes           Skip confirmation prompts
```

### SEE ALSO

* [kh workflow](kh_workflow.md)	 - Manage workflows

