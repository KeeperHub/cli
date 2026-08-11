# Issues before pull requests

Open an issue before you write code. We answer it, and once the problem and the
shape of the fix are agreed, the pull request is a short step rather than a
negotiation.

## When an issue is required

**Required** for anything that changes behaviour:

- Command behaviour, output format, or exit codes
- New commands, flags, or defaults
- Anything that changes what an existing flag does
- Go module dependencies added, removed, or upgraded
- CI, release, or build configuration
- Authentication and credential handling

**Not required** - open a pull request directly:

- Typos, broken links, and formatting
- Help text and error-message wording that corrects an existing statement
- Documentation that matches what the code already does

If you are unsure, open the issue.

## What happens to your issue

| Label | Meaning |
|---|---|
| `needs-triage` | Received, not yet read. Applied automatically. |
| `accepted` | The problem is real and the proposed approach is sound and correctly scoped. Write the pull request. |
| `needs-discussion` | Real, but the approach or the scope is not settled. Do not start yet. |
| `wontfix` / `duplicate` / `invalid` | Closed, with the reason in a comment. |

**`accepted` is the signal to start.** It is what the pull request gate checks
for. Nothing else on the issue means "go".

We aim to triage within two working days. If an issue has sat longer, comment on
it - that is the fastest way to get it moving.

## What makes an issue answerable

**For a bug**: the exact command, its full output, what you expected, your `kh
version`, and your OS and architecture. A pasted terminal session settles a
report faster than any description of one.

**For a change in behaviour**: what you are trying to do that you currently
cannot, before what to build. If a flag exists that nearly does it, say what it
falls short on.

**Scope it.** Name what the change touches and what it does not.

## Should this be one change

Apply this to each seam in what you are proposing:

> Can piece A ship, deploy, and be correct with piece B absent or reverted?

If yes for every pair, they are separate issues and separate pull requests. If
the pieces are only correct together, they are one unit regardless of size.

A worked example from this repo: a pull request added receipt rendering with
`--require-verified` **and** a `--timeout` deadline for `--watch`. Neither needed
the other to be correct, and the timeout changed behaviour for every existing
`--watch` caller. Two issues, two pull requests.

## Opening the pull request

Once your issue carries `accepted`:

1. **Reference the issue in the pull request title**, after the conventional
   commit type:

   ```
   feat: #97 add --require-verified to execute status
   fix(execute): #98 bound a hung request under --watch --timeout
   ```

   The type prefix drives release-please, so keep it accurate.

2. Fill in the pull request template.

3. Target `main`. This repo has no `staging` branch.

`docs`, `chore`, and `style` pull requests are exempt from the issue check
automatically. A maintainer can apply `no-issue-required` to exempt anything else.

## Security

Do not open an issue for a vulnerability. Report it privately through
[GitHub Private Vulnerability Reporting](https://github.com/KeeperHub/cli/security).

## Related

- [CONTRIBUTING.md](CONTRIBUTING.md) - setup, build, test, release
- [docs.keeperhub.com/cli](https://docs.keeperhub.com/cli) - command reference
