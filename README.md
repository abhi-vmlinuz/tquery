# tquery

<p align="center">
  <strong>Interactive terminal visualizer and query engine for JSON data.</strong>
</p>

<p align="center">
  <a href="https://github.com/abhi-vmlinuz/tquery/releases"><img src="https://img.shields.io/github/v/release/abhi-vmlinuz/tquery?style=flat-square&color=blue" alt="Release"></a>
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go" alt="Go Version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-green?style=flat-square" alt="License"></a>
</p>

---

`tquery` solves the **human interface problem** when working with JSON data in the terminal.

Instead of dumping dense or ugly formatted text, `tquery` parses raw JSON streams, normalizes nested data structures, executes embedded `jq` filters natively, and renders human-readable ASCII/Unicode tables, hierarchical trees, and markdown exports.

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

## ✨ Features

- 📊 **Smart Table Formatter**: Auto-detects arrays of objects and generates aligned, ANSI syntax-highlighted tables.
- ⚡ **Auto-Unwrapping**: Automatically unwraps standard REST API envelope keys (`data`, `items`, `results`, `models`, `records`) without manual querying.
- 🔍 **Embedded `jq` Engine**: Native Go JQ query processing powered by `gojq` — no external `jq` binary required.
- 🖥️ **Interactive TUI Mode (`-i`)**: Live real-time JQ search prompt, vim keybindings, table navigation, and row inspection drawer.
- 📁 **Multi-Format Export**: One-flag export to `table`, `markdown`, `csv`, `tsv`, `tree`, and `json`.
- 🚀 **Zero Dependencies**: Standalone, single static binary with ultra-fast startup.

---

## 📦 Installation

### From Source (Go)

```bash
go install github.com/abhi-vmlinuz/tquery@latest
```

### Build & Install via Make

```bash
git clone https://github.com/abhi-vmlinuz/tquery.git
cd tquery

# System-wide installation (requires sudo)
sudo make install

# User-level installation (installs to ~/.local/bin and ~/.local/share/man)
make install-user
```

### Manual Binary Build

```bash
git clone https://github.com/abhi-vmlinuz/tquery.git
cd tquery
go build -ldflags="-s -w" -o tquery main.go
sudo mv tquery /usr/local/bin/
```

---

## 🐚 Shell Completions

Shell completions are available for **Bash**, **Fish**, and **Zsh**.

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

## 🚀 Quick Start

### 1. Pipe Stdin Directly
```bash
curl -s https://api.github.com/repos/golang/go/releases | tquery
```

### 2. Query JSON with JQ Filters
```bash
cat data.json | tquery '.data[] | {id, owned_by}'
```

### 3. Convert JSON to Markdown Table
```bash
tquery -f markdown '.users' users.json
```

### 4. Hierarchical Tree Mode
```bash
tquery -f tree config.json
```

### 5. Launch Interactive Explorer
```bash
tquery -i payload.json
```

---

## ⌨️ TUI Keybindings

When running in interactive mode (`tquery -i`):

| Key | Action |
| --- | --- |
| `Type text` | Live JQ query filtering with real-time preview |
| `Tab` | Switch view mode (`Table` ➔ `Tree` ➔ `JSON`) |
| `Enter` | Open / Close Inspect Row Detail drawer |
| `Esc` | Close Inspect Drawer |
| `↑` / `↓` / `k` / `j` | Navigate rows and scroll viewports |
| `Ctrl+C` | Quit |

---

## 📚 Documentation

Detailed documentation and practical references:

- 📖 **[CLI Reference](docs/cli.md)**: Full flag documentation, format options, and exit codes.
- 🖥️ **[Interactive TUI Guide](docs/tui.md)**: Deep dive into the live visualizer, view modes, and keybindings.
- 💡 **[Practical Examples & Pipelines](docs/examples.md)**: Real-world DevOps, cloud APIs (GitHub, Docker, Kubernetes, AWS, LLM endpoints) workflows.
- 📄 **[Man Page](man/tquery.1)**: UNIX Section 1 manual page.

---

## 📄 License

Distributed under the [MIT License](LICENSE).
