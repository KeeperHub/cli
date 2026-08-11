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

## Reason, scope, plan

Every issue carries three things. An issue missing any of them cannot be
answered, only discussed, and discussion is what this policy exists to replace.

**Reason** - why it matters, in evidence. For a bug: the exact command, its full
output, what you expected, *what told you to expect it*, and what it costs
someone who hits it. Name the source of the expectation - help text, a docs
page, a flag name - because that tells us straight away whether the code is
wrong or the source is.

**Scope** - what this covers and what it does not. Which commands you checked
and found fine. If the same fault plausibly affects sibling commands, say which.
Then confirm it is one issue: if any part could be fixed and shipped while
another stays broken, those are separate issues.

**Plan** - what should happen next. Your proposal, not a commitment we have
made; triage may replace it. If you do not know the fix, *"I do not know; here
is what I would need to determine"* is a complete plan. Blank is not - a problem
with no proposed next step puts the whole cost of thinking on whoever reads it.

### Already filed an issue

Nothing here applies retroactively. Issues filed before this page existed are
triaged on what they contain, and you will never be asked to resubmit one to
match a template that did not exist when you wrote it.

More generally, and for new issues too: **you will not be asked to restate
something you have already said.** If triage needs one more fact, it asks for
that fact, on your issue.

If an issue turns out to hold several problems, we split it and credit you on
each part.

## What happens to your issue

| Label | Meaning |
|---|---|
| `needs-triage` | Received, not yet read. Applied automatically. |
| `confirmed` | Someone reproduced it. Says nothing yet about whether we will fix it. |
| `accepted` | Reason, scope and plan all stand. Write the pull request. |
| `needs-discussion` | Real, but the scope or the plan is not settled. Do not start yet. |
| `wontfix` / `duplicate` / `invalid` | Closed, with the reason in a comment. |

**`accepted` is the signal to start.** It is what the pull request gate checks
for. Nothing else means "go" - `confirmed` in particular does not.

**`accepted` accepts a specific plan.** If triage keeps your reason and scope but
replaces your plan, it says so in a comment before applying the label, and that
comment is the plan. Build against it, not the one you filed.

We aim to triage within two working days. If an issue has sat longer, comment on
it - that is the fastest way to get it moving.

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
