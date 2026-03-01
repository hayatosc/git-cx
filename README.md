# git-cx

> Stage your changes, generate Conventional Commits with AI, pick one in a TUI.

## Quick Start

**1. Install**

```console
go install github.com/hayatosc/git-cx@latest
```

**2. Set your provider (example: Gemini)**

```console
git config --global cx.provider gemini
```

**3. Stage changes and run**

```console
git add -A
git cx
```

## TUI Keyboard Shortcuts

| Screen | Key | Action |
|---|---|---|
| Select Type / Message | `↑` `↓` | Move |
| Select Type / Message | `Enter` | Confirm |
| Input Scope / Footer | `Enter` | Next |
| Input Body | `Ctrl+D` | Done |
| Input Body | `Enter` (empty) | Skip |
| Confirm | `y` | Commit |
| Confirm | `n` / `q` | Abort |

## Providers

| Provider | Requirements | Key config |
|---|---|---|
| `gemini` | [Gemini CLI](https://github.com/google-gemini/gemini-cli) | `cx.provider = gemini` |
| `copilot` | `gh` CLI + Copilot subscription | `cx.provider = copilot` |
| `claude` | [Claude Code](https://github.com/anthropics/claude-code) | `cx.provider = claude` |
| `codex` | [Codex CLI](https://github.com/openai/codex) | `cx.provider = codex` |
| `api` | OpenAI-compatible endpoint + API key | `cx.apiBaseUrl` + `OPENAI_API_KEY` |
| `custom` | Any CLI with stdout output | `cx.command = "mycli --prompt {prompt}"` |

## Configuration

All options can be set via `git config`, a config file (`--config`), or flags per invocation.

### git config keys

| Key | Type | Default | Description |
|---|---|---|---|
| `cx.provider` | string | `gemini` | AI provider |
| `cx.model` | string | — | Model name |
| `cx.candidates` | int | `3` | Number of candidates |
| `cx.timeout` | int | `30` | Request timeout (seconds) |
| `cx.command` | string | — | Command template for `custom` provider (`{prompt}` is replaced) |
| `cx.apiBaseUrl` | string | — | Base URL for `api` provider |
| `cx.commit.useEmoji` | bool | `false` | Prefix commit type with emoji |
| `cx.commit.maxSubjectLength` | int | `100` | Max subject line length |
| `cx.commit.scopes` | string (multi) | — | Scope candidates |

**Environment:** `OPENAI_API_KEY` — required for `api` provider.

### Flags

| Flag | Description |
|---|---|
| `--config <path>` | gitconfig-format config file |
| `--provider <name>` | AI provider |
| `--model <name>` | Model name |
| `--candidates <n>` | Number of candidates |
| `--timeout <n>` | Timeout in seconds |
| `--command <template>` | Command template for `custom` provider |
| `--api-base-url <url>` | Base URL for `api` provider |
| `--use-emoji` | Prefix commit type with emoji |
| `--max-subject-length <n>` | Max subject line length |

## Config file (`--config`)

```console
git cx --config examples/gemini.gitconfig
git cx --config examples/copilot.gitconfig
git cx --config examples/claude.gitconfig
git cx --config examples/codex.gitconfig
git cx --config examples/api.gitconfig
```

## Development

```console
mise run format
mise run check
mise run test
mise run build
```

## Release

1. Run the "release-pr" workflow with a version like `1.2.3`.
2. Merge the generated PR.

The release workflow runs on the VERSION update and publishes GitHub Releases via GoReleaser.
