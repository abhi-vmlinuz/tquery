# tquery (tq)

<p align="center">
  <strong>Interactive terminal visualizer and query engine for JSON data.</strong>
</p>

<p align="center">
  <a href="https://github.com/abhi-vmlinuz/tquery/releases"><img src="https://img.shields.io/github/v/release/abhi-vmlinuz/tquery?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License"></a>
</p>

---

`tq` parses raw JSON streams, normalizes nested data structures, executes embedded `jq` filters natively, and renders clean ASCII/Unicode tables, hierarchical trees, and markdown exports.

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

# After: tq
$ curl -s https://inference.dahl.global/v1/models | tq
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
- **Row limit flag (`-<number>`)** — quick Unix-style row limiting (e.g. `tq -10` or `tq -l 5`)
- **Embedded jq engine** — native Go JQ query processing powered by `gojq`; no external `jq` binary required
- **Interactive TUI mode** (`-i`) — live JQ search prompt, vim keybindings, table navigation, and row inspection drawer
- **Multi-format export** — `table`, `markdown`, `csv`, `tsv`, `tree`, `json`
- **Zero dependencies** — standalone, single static binary (`tq`, aliased to `tquery`)

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

# System-wide installation (installs 'tq' with 'tquery' symlink)
sudo make install

# User-level installation (installs to ~/.local/bin and ~/.local/share/man)
make install-user
```

### Manual binary build

```bash
git clone https://github.com/abhi-vmlinuz/tquery.git
cd tquery
go build -ldflags="-s -w" -o tq main.go
sudo mv tq /usr/local/bin/
sudo ln -sf /usr/local/bin/tq /usr/local/bin/tquery
```

---

## Shell completions

Completions are available for Bash, Fish, and Zsh for both `tq` and `tquery`.

### Bash
```bash
sudo cp completions/tq.bash /etc/bash_completion.d/tq
```

### Fish
```bash
cp completions/tq.fish ~/.config/fish/completions/
```

### Zsh
```bash
cp completions/tq.zsh "${fpath[1]}/_tq"
```

---

## Quick start

**Pipe stdin directly**
```bash
curl -s https://inference.dahl.global/v1/models | tq
```

**Limit output rows**
```bash
curl -s https://inference.dahl.global/v1/models | tq -10
```

**Query JSON with JQ filters**
```bash
cat data.json | tq '.data[] | {id, owned_by}'
```

**Convert JSON to a markdown table**
```bash
tq -f markdown '.users' users.json
```

**Hierarchical tree mode**
```bash
tq -f tree config.json
```

**Launch the interactive explorer**
```bash
tq -i payload.json
```

---

## TUI keybindings

When running in interactive mode (`tq -i`):

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
- [Man page](man/tq.1) — UNIX Section 1 manual page

---

## License

Distributed under the [MIT License](LICENSE).
