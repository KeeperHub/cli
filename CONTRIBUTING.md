# Contributing to the KeeperHub CLI

`kh` is the KeeperHub command-line interface. Go, distributed through Homebrew
and `go install`.

## Start with an issue

Anything that changes behaviour needs an issue first, accepted by a maintainer,
before the pull request. **[ISSUES.md](ISSUES.md) is the policy** - what needs an
issue, what goes straight to a pull request, and what happens after you file one.

The short version: open an issue, wait for the `accepted` label, then reference
it in your pull request title (`feat: #97 description`). Typos, help-text
wording, and docs matching existing behaviour skip all of that.

## Development setup

Go 1.25+ (`go.mod` pins the exact version CI uses) and `golangci-lint` for the
lint target.

```bash
make build     # builds bin/kh with version metadata
make test      # go test -race ./...
make lint      # golangci-lint run ./...
make install   # go install into your GOPATH
make clean
```

`make build` and `make install` run `sync-version` first, which copies
`.release-please-manifest.json` into `internal/version/`. Build with plain
`go build` and the binary reports the wrong version.

## Generated command docs

`docs/` is generated from the cobra command tree, and CI fails on drift:

```bash
go generate ./docs/
```

Run it after any change to a command, a flag, or its help text, and commit the
result. The `docs-check` job runs `git diff --exit-code docs/` and turns red
otherwise.

## Tests

```bash
go test -race ./...                              # unit tests, what CI runs
go test -tags integration ./tests/integration/   # integration, needs credentials
```

Integration tests are behind the `integration` build tag and need
`KH_TEST_HOST`, `KH_TEST_EMAIL`, `KH_TEST_PASSWORD` and `KH_API_KEY`. CI only
runs them on push, not on pull requests, so they will not run on your fork.
Unit-test any behaviour you change - a command test alongside the changed
command is the expectation.

Timeouts and deadlines deserve a specific note: a test that only exercises a
server responding promptly does not cover a request that hangs. If you add a
deadline, test a handler that blocks past it.

## Pull requests

1. **Title**: `<type>: #<issue> <description>`, for example
   `feat: #97 add --require-verified to execute status`. The type prefix drives
   release-please, so keep it accurate. Types: `feat`, `fix`, `chore`, `docs`,
   `refactor`, `test`, `ci`, `build`, `perf`, `style`.
2. **Base branch**: `main`. This repo has no `staging` branch.
3. **Scope**: one change per pull request. If a part of it could ship and be
   correct with the rest reverted, split it.
4. **Before submitting**: `make lint`, `make test`, and `go generate ./docs/`
   all clean, and the backing issue carries `accepted`.

Breaking changes use `!` (`feat!:`) or a `BREAKING CHANGE:` trailer, which
release-please turns into a major version.

## Related

- [ISSUES.md](ISSUES.md) - the issue-first policy
- [docs.keeperhub.com/cli](https://docs.keeperhub.com/cli) - command reference
