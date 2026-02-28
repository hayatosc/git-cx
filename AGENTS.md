# AGENTS

git-cx is a CLI that generates Conventional Commits messages from staged diffs using AI and lets you choose in a TUI.

## Key Components

- `main.go`: CLI entrypoint (flags/commands)
- `internal/app`: commit flow orchestration
- `internal/ai`: providers (gemini/copilot/custom)
- `internal/tui`: TUI UI/state transitions
- `internal/config`: git config configuration loading
- `internal/git`: git command runner

## Development Commands

```bash
mise run dev
mise run format
mise run check
mise run test
mise run build
```
