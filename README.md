# tquery

Interactive terminal visualizer and query engine for JSON data.

[![Release](https://img.shields.io/github/v/release/abhi-vmlinuz/tquery?style=flat-square&color=blue)](https://github.com/abhi-vmlinuz/tquery/releases)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)

---

`tquery` parses raw JSON streams, normalizes nested data structures, executes embedded `jq` filters natively, and renders ASCII/Unicode tables, hierarchical trees, and markdown exports.

```text
# Before: raw / jq pretty print
$ curl -s https://inference.dahl.global/v1/models | jq
{
  "data": [
    {
      "id": "MiniMaxAI/MiniMax-M2.7",
      "object": "model",
      "created": 1677610602,
      "owned_by": "gonka"
    },
    ...
  ]
}

# After: tquery
$ curl -s https://inference.dahl.global/v1/models | tquery
┼────────────┼────────────────────────┼────────┼──────────┼
│ created    │ id                     │ object │ owned_by │
┼────────────┼────────────────────────┼────────┼──────────┼
│ 1677610602 │ MiniMaxAI/MiniMax-M2.7 │ model  │ gonka    │
│ 1677610602 │ moonshotai/Kimi-K2.6   │ model  │ gonka    │
┼────────────┼────────────────────────┼────────┼──────────┼
```

---

## Features

- **Table formatter** — auto-detects arrays of objects and generates aligned, syntax-highlighted tables
- **Auto-unwrapping** — automatically unwraps standard REST API envelope keys (`data`, `items`, `results`, `models`, `records`) without manual querying
- **Embedded jq engine** — native Go JQ query processing powered by `gojq`; no external `jq` binary required
- **Interactive TUI mode** (`-i`) — live JQ search prompt, vim keybindings, table navigation, and row inspection drawer
- **Multi-format export** — `table`, `markdown`, `csv`, `tsv`, `tree`, `json`
- **Zero dependencies** — standalone, single static binary

---

## Installation

### From source (Go)

```bash
go install github.com/abhi-vmlinuz/tquery@latest
```

### Build and install via Make

```bash
git clone https://github.com/abhi-vmlinuz/tquery.git
cd tquery

# System-wide installation (requires sudo)
sudo make install

# User-level installation (installs to ~/.local/bin and ~/.local/share/man)
make install-user
```

### Manual binary build

```bash
git clone https://github.com/abhi-vmlinuz/tquery.git
cd tquery
go build -ldflags="-s -w" -o tquery main.go
sudo mv tquery /usr/local/bin/
```

---

## Shell completions

Completions are available for Bash, Fish, and Zsh.

### Bash
```bash
# Add to ~/.bashrc:
source <(tquery --completion bash 2>/dev/null || cat /usr/local/share/bash-completion/completions/tquery)
```
Or copy directly:
```bash
sudo cp completions/tquery.bash /etc/bash_completion.d/tquery
```

### Fish
```bash
cp completions/tquery.fish ~/.config/fish/completions/
```

### Zsh
```bash
cp completions/tquery.zsh "${fpath[1]}/_tquery"
```

---

## Quick start

**Pipe stdin directly**
```bash
curl -s https://api.github.com/repos/golang/go/releases | tquery
```

**Query JSON with JQ filters**
```bash
cat data.json | tquery '.data[] | {id, owned_by}'
```

**Convert JSON to a markdown table**
```bash
tquery -f markdown '.users' users.json
```

**Hierarchical tree mode**
```bash
tquery -f tree config.json
```

**Launch the interactive explorer**
```bash
tquery -i payload.json
```

---

## TUI keybindings

When running in interactive mode (`tquery -i`):

| Key | Action |
| --- | --- |
| Type text | Live JQ query filtering with real-time preview |
| `Tab` | Switch view mode (Table → Tree → JSON) |
| `Enter` | Open / close inspect row detail drawer |
| `Esc` | Close inspect drawer |
| `↑` / `↓` / `k` / `j` | Navigate rows and scroll viewports |
| `Ctrl+C` | Quit |

---

## Documentation

- [CLI reference](docs/cli.md) — full flag documentation, format options, and exit codes
- [Interactive TUI guide](docs/tui.md) — live visualizer, view modes, and keybindings
- [Practical examples and pipelines](docs/examples.md) — DevOps, cloud API (GitHub, Docker, Kubernetes, AWS, LLM endpoints) workflows
- [Man page](man/tquery.1) — UNIX Section 1 manual page

---

## License

Distributed under the [MIT License](LICENSE).
