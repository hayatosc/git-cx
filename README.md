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
[cx "commit"]
  useEmoji = false
  maxSubjectLength = 100
  scopes = feat
```

### TOML

``` console
$ git cx --config examples/gemini.toml
$ git cx --config examples/copilot.toml
```

### Flags

``` console
$ git cx --provider gemini --model gemini-3.0-flash
$ git cx --candidates 3 --timeout 30
$ git cx --use-emoji --max-subject-length 100
$ git cx --command "my-cli --prompt {prompt}"
```

## Providers

Select the provider with `cx.provider`.

- `gemini`: uses the `gemini` CLI
- `copilot`: uses the `copilot` CLI
- `custom`: runs the command in `cx.command` (replaces `{prompt}`)

## Development

``` console
$ mise run format
$ mise run check
$ mise run test
$ mise run build
```
