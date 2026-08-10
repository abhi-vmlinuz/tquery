# CLI Reference

`tq` (aliased as `tquery`) provides a flexible command-line interface for piping, filtering, and converting JSON streams into human-friendly formats.

---

## Synopsis

```bash
tq [options] [-<number>] [jq_query] [file]
<command> | tq [options] [-<number>] [jq_query]
```

---

## Arguments

| Argument | Description | Required |
| --- | --- | --- |
| `[jq_query]` | Any valid `jq` query filter (e.g. `.` or `.data[]` or `.users[] \| {name, email}`). | Optional |
| `[file]` | Path to a local `.json` file. If omitted, `tq` reads from standard input (`stdin`). | Optional |

---

## Options & Flags

### `-<number>, -l, --limit <num>`
Limits the maximum number of output rows/records displayed. Works seamlessly with any number prefix like `-10`, `-25`, `-5`, etc.

```bash
# Display only the first 10 rows
curl -s https://api.example.com/models | tq -10

# Limit with query
tq -5 '.data[]' data.json
```

---

### `-g, --grep <pattern>` (Multiple patterns supported)
Filters rows or tree branches matching regex or string patterns. You can specify `-g` (or `-e`) multiple times.

- **Default (OR matching)**: Shows branches/rows matching Pattern A **OR** Pattern B.
- **Column-scoped**: `<column>:<pattern>` (e.g. `-g "status:Running"`).

```bash
# Search multiple patterns (OR)
tq -g 'nginx' -g 'running' data.json

# Regex search
tq -g 'MiniMax|meta' data.json
```

---

### `--strict`
Enforces strict **AND** matching across all supplied `-g` / `-e` patterns.

- A logical record or object is only retained if **all** specified patterns are present within that record.
- Non-matching sibling fields and unrelated branches are pruned.

```bash
# Only show containers that are both 'nginx' AND 'Running'
docker inspect my-container | tq -g 'nginx' -g 'Running' --strict
```

---

### `-V, --invert, --invert-match`
Inverts the grep filter match (selects rows that do **not** match the pattern), similar to `grep -v`.

```bash
# Filter out stopped instances
tq -g 'stopped' --invert data.json
```

---

### Direct Shape Shortcuts

- **`--tree`**: Force hierarchical tree view.
- **`--table`**: Force table view.
- **`--json, --raw`**: Force JSON output.
- **`--markdown, --md`**: Force markdown table output.
- **`--csv, --tsv`**: Force CSV/TSV output.

```bash
# Force tree view on API response
curl -s https://api.example.com/models | tq --tree

# Force JSON output of pruned grep results
docker inspect my-container | tq -g 'port' --json
```

---

### `-f, --format <format>`
Specifies the output format for rendered data (`auto`, `table`, `tree`, `markdown`, `csv`, `tsv`, `json`). Defaults to `auto` (auto-detects table for flat datasets and tree for deep objects).

```bash
# Markdown table output
tq -f markdown data.json

# CSV output
tq -f csv data.json > output.csv

# Tree structure
tq -f tree config.json
```

---

### `-i, --interactive`
Launches the full-screen interactive Terminal User Interface (TUI). Allows real-time typing of JQ filters, keyboard navigation, row inspection, and instant view switching.

```bash
tq -i data.json
curl -s https://api.example.com/items | tq -i
```

---

### `-n, --no-headers`
Hides the column header row in `table`, `markdown`, and `csv` output modes. Useful when feeding output directly into downstream text processing utilities like `awk` or `grep`.

```bash
tq -n -f csv data.json
```

---

### `--no-unwrap`
Disables automatic root envelope unwrapping.

By default, `tq` scans top-level objects for standard wrapper keys (`data`, `items`, `results`, `models`, `rows`, `records`, `list`, `value`). If found, it automatically unfolds the inner array into a table.

Passing `--no-unwrap` treats the root object as a standard key-value map.

```bash
tq --no-unwrap data.json
```

---

### `--no-color`
Disables ANSI color codes in output. Automatically engaged when output stdout is redirected or piped to a non-TTY descriptor unless overridden.

```bash
tq --no-color data.json
```

---

### `-v, --version`
Prints the current version of `tq` and exits.

```bash
tq --version
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
