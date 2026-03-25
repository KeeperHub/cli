# kh - KeeperHub CLI

Command-line interface for [KeeperHub](https://app.keeperhub.com/hub), the Web3 automation platform.

## Install

**Homebrew:**
```
brew install keeperhub/tap/kh
```

**Go install:**
```
go install github.com/keeperhub/cli/cmd/kh@latest
```

**Binary download:**
Download from [GitHub Releases](https://github.com/keeperhub/cli/releases).

## Auth

```
kh auth login
```

For CI/CD, set `KH_API_KEY` instead of running the browser flow.

## Common Commands

```
kh workflow list                   # List all workflows
kh workflow run <id> --wait        # Run a workflow and wait for completion
kh run status <run-id>             # Check a run's status
kh run logs <run-id>               # Stream run logs
kh execute contract-call ...       # Execute a protocol action directly
kh protocol list                   # Browse available protocols
```

## MCP Server Mode

The recommended way to connect AI assistants to KeeperHub is the remote HTTP endpoint:

```
claude mcp add --transport http keeperhub https://app.keeperhub.com/mcp
```

No local server process required. See [docs/quickstart.md](docs/quickstart.md) for full setup instructions.

The legacy `kh serve --mcp` stdio mode is still available but deprecated.

## Documentation

- [Quickstart](docs/quickstart.md) -- install, auth, and first steps
- [Concepts](docs/concepts.md) -- authentication model, output formats, configuration
- [Command reference](docs/kh.md) -- full documentation for every command

## License

MIT
