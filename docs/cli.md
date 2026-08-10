# CLI Reference

`tquery` provides a flexible command-line interface for piping, filtering, and converting JSON streams into human-friendly formats.

---

## Synopsis

```bash
tquery [options] [jq_query] [file]
<command> | tquery [options] [jq_query]
```

---

## Arguments

| Argument | Description | Required |
| --- | --- | --- |
| `[jq_query]` | Any valid `jq` query filter (e.g. `.` or `.data[]` or `.users[] \| {name, email}`). | Optional |
| `[file]` | Path to a local `.json` file. If omitted, `tquery` reads from standard input (`stdin`). | Optional |

---

## Options & Flags

### `-f, --format <format>`
Specifies the output format for rendered data. Defaults to `table`.

- **`table`**: Renders an ASCII/Unicode table with column borders and ANSI syntax highlighting.
- **`markdown`**: Renders a GitHub-Flavored Markdown (GFM) pipe table.
- **`csv`**: Standard comma-separated values format.
- **`tsv`**: Tab-separated values format.
- **`tree`**: Hierarchical indented tree structure visualizing nested keys, arrays, and primitive datatypes.
- **`json`**: Pretty-printed formatted JSON.

```bash
# Markdown table output
tquery -f markdown data.json

# CSV output
tquery -f csv data.json > output.csv

# Tree structure
tquery -f tree config.json
```

---

### `-<number>, -l, --limit <num>`
Limits the maximum number of output rows/records displayed. Works seamlessly with any number prefix like `-10`, `-25`, `-5`, etc.

```bash
# Display only the first 10 rows
curl -s https://api.example.com/models | tquery -10

# Limit with query
tquery -5 '.data[]' data.json
```

---

### `-i, --interactive`
Launches the full-screen interactive Terminal User Interface (TUI). Allows real-time typing of JQ filters, keyboard navigation, row inspection, and instant view switching.

```bash
tquery -i data.json
curl -s https://api.example.com/items | tquery -i
```

---

### `-n, --no-headers`
Hides the column header row in `table`, `markdown`, and `csv` output modes. Useful when feeding output directly into downstream text processing utilities like `awk` or `grep`.

```bash
tquery -n -f csv data.json
```

---

### `--no-unwrap`
Disables automatic root envelope unwrapping.

By default, `tquery` scans top-level objects for standard wrapper keys (`data`, `items`, `results`, `models`, `rows`, `records`, `list`, `value`). If found, it automatically unfolds the inner array into a table.

Passing `--no-unwrap` treats the root object as a standard key-value map.

```bash
tquery --no-unwrap data.json
```

---

### `--no-color`
Disables ANSI color codes in output. Automatically engaged when output stdout is redirected or piped to a non-TTY descriptor unless overridden.

```bash
tquery --no-color data.json
```

---

### `-v, --version`
Prints the current version of `tquery` and exits.

```bash
tquery --version
# Output: tquery version 1.0.0
```

---

### `-h, --help`
Displays CLI usage, available flags, and examples.

---

## Exit Codes

| Code | Meaning |
| --- | --- |
| `0` | Success |
| `1` | General error (invalid JSON, file not found, bad jq query syntax, render failure) |
