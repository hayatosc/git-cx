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
$ git config --global cx.provider gemini
$ git config --global cx.model gemini-3.0-flash
$ git config --global cx.candidates 3
$ git config --global cx.timeout 30
$ git config --global cx.api.baseUrl https://api.openai.com/v1
$ git config --global cx.api.key YOUR_API_KEY
$ git config --global cx.commit.useEmoji false
$ git config --global cx.commit.maxSubjectLength 100
$ git config --global cx.commit.scopes feat
```

### .gitconfig example

``` ini
[cx]
  provider = gemini
  model = gemini-3.0-flash
  candidates = 3
  timeout = 30
[cx "api"]
  baseUrl = https://api.openai.com/v1
  key = YOUR_API_KEY
[cx "commit"]
  useEmoji = false
  maxSubjectLength = 100
  scopes = feat
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

## Providers

Select the provider with `cx.provider`.

- `gemini`: uses the `gemini` CLI
- `copilot`: uses the `copilot` CLI
- `claude`: uses the `claude` CLI
- `codex`: uses the `codex exec` CLI
- `api`: uses an OpenAI-compatible API endpoint (set `cx.api.baseUrl` and `cx.api.key`)
- `custom`: runs the command in `cx.command` (replaces `{prompt}`)

For the `api` provider, set `cx.api.baseUrl` to your OpenAI-compatible endpoint (e.g. `https://api.openai.com/v1`, `https://openrouter.ai/api/v1`, `http://localhost:8000/v1`) and set `cx.api.key` or `OPENAI_API_KEY`.

## Development

``` console
$ mise run format
$ mise run check
$ mise run test
$ mise run build
```
