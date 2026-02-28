# git-cx

An AI-powered `git` subcommand that generates Conventional Commits messages and lets you pick one in a TUI.

## Usage

``` console
$ git add -A
$ git cx
```

Pick a generated candidate, then optionally enter a body and footer before committing.

### Commands

``` console
$ git cx version
$ git cx config
```

## Install

**go install:**

``` console
$ go install github.com/hayatosc/git-cx@latest
```

**manually:**

``` console
$ go build -o output/git-cx .
$ ./output/git-cx
```

## Configuration

Configure via `git config` or TOML. All options can be overridden by flags per invocation.

### git config

``` console
# CLI providers (gemini/copilot/claude/codex/custom)
$ git config --global cx.provider gemini
$ git config --global cx.model gemini-3.0-flash
$ git config --global cx.candidates 3
$ git config --global cx.timeout 30
$ git config --global cx.commit.useEmoji false
$ git config --global cx.commit.maxSubjectLength 100
$ git config --global cx.commit.scopes feat

# API provider
$ git config --global cx.provider api
$ git config --global cx.model gpt-5
$ git config --global cx.apiBaseUrl https://api.openai.com/v1
$ git config --global cx.apiKey YOUR_API_KEY
```

### .gitconfig example

``` ini
# CLI providers (gemini/copilot/claude/codex/custom)
[cx]
  provider = gemini
  model = gemini-3.0-flash
  candidates = 3
  timeout = 30
[cx "commit"]
  useEmoji = false
  maxSubjectLength = 100
  scopes = feat
```

``` ini
# API provider
[cx]
  provider = api
  model = gpt-5
  apiBaseUrl = https://api.openai.com/v1
  apiKey = YOUR_API_KEY
```

### TOML

``` console
$ git cx --config examples/gemini.toml
$ git cx --config examples/copilot.toml
$ git cx --config examples/claude.toml
$ git cx --config examples/codex.toml
$ git cx --config examples/api.toml
```

### Flags

``` console
$ git cx --provider gemini --model gemini-3.0-flash
$ git cx --candidates 3 --timeout 30
$ git cx --use-emoji --max-subject-length 100
$ git cx --command "my-cli --prompt {prompt}"
$ git cx --provider api --model gpt-5 --api-base-url https://api.openai.com/v1 --api-key YOUR_API_KEY
```

### Configuration mapping

| Setting | git config | .gitconfig | TOML | Flag | Notes |
| --- | --- | --- | --- | --- | --- |
| provider | `cx.provider` | `[cx] provider` | `provider` | `--provider` | `gemini`, `copilot`, `claude`, `codex`, `api`, `custom` |
| model | `cx.model` | `[cx] model` | `model` | `--model` | required for `api` |
| candidates | `cx.candidates` | `[cx] candidates` | `candidates` | `--candidates` | number of suggestions |
| timeout | `cx.timeout` | `[cx] timeout` | `timeout` | `--timeout` | seconds |
| command | `cx.command` | `[cx] command` | `command` | `--command` | custom provider only |
| apiBaseUrl | `cx.apiBaseUrl` | `[cx] apiBaseUrl` | `api_base_url` | `--api-base-url` | OpenAI-compatible base URL |
| apiKey | `cx.apiKey` | `[cx] apiKey` | `api_key` | `--api-key` | fallback: `OPENAI_API_KEY` |
| commit.useEmoji | `cx.commit.useEmoji` | `[cx "commit"] useEmoji` | `[commit] use_emoji` | `--use-emoji` | adds emoji prefix |
| commit.maxSubjectLength | `cx.commit.maxSubjectLength` | `[cx "commit"] maxSubjectLength` | `[commit] max_subject_length` | `--max-subject-length` | 0 disables limit |
| commit.scopes | `cx.commit.scopes` | `[cx "commit"] scopes` | `[commit] scopes` | (none) | repeatable in git config |

Legacy keys still accepted: `cx.api.baseUrl`, `cx.api.key`, and `[api] base_url/key`.

## Providers

Select the provider with `cx.provider`.

- `gemini`: uses the `gemini` CLI
- `copilot`: uses the `copilot` CLI
- `claude`: uses the `claude` CLI
- `codex`: uses the `codex exec` CLI
- `api`: uses an OpenAI-compatible API endpoint (set `cx.apiBaseUrl` and `cx.apiKey`)
- `custom`: runs the command in `cx.command` (replaces `{prompt}`)

For the `api` provider, set `cx.apiBaseUrl` to your OpenAI-compatible endpoint (e.g. `https://api.openai.com/v1`, `https://openrouter.ai/api/v1`, `http://localhost:8000/v1`) and set `cx.apiKey` or `OPENAI_API_KEY`.

## Development

``` console
$ mise run format
$ mise run check
$ mise run test
$ mise run build
```
