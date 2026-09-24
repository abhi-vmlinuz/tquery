# Interactive TUI Guide

`tq` includes an interactive terminal user interface (TUI) powered by Charm's `bubbletea` and `lipgloss` libraries.

---

## Launching the TUI

You can enter interactive mode by passing the `-i` flag:

```bash
# Explore a local JSON file
tq -i payload.json

# Pipe live API output directly into the interactive TUI
curl -s https://integrate.api.nvidia.com/v1/models | tq -i
```

---

## Interface Layout

```text
┌─────────────────────────────────────────────────────────────┐
│ tq (tquery)   [Table] [Tree] [JSON]                         │  <-- Header & View Modes
│                                                             │
│ jq > .data[] | {id, owned_by}                               │  <-- Live Query Prompt
│                                                             │
│ ┌────────────────────────┬────────────────────────────────┐ │
│ │ id                     │ owned_by                       │ │  <-- Dynamic Data Table /
│ ├────────────────────────┼────────────────────────────────┤ │      Tree / Viewport
│ │ MiniMaxAI/MiniMax-M2.7 │ gonka                          │ │
│ │ moonshotai/Kimi-K2.6   │ gonka                          │ │
│ └────────────────────────┴────────────────────────────────┘ │
│                                                             │
│ Tab: switch view  •  Enter: inspect row  •  Ctrl+C: quit   │  <-- Status & Help Line
└─────────────────────────────────────────────────────────────┘
```

---

## Key Features

### 1. Live JQ Query Filter
Type any `jq` expression directly into the prompt bar. As you type, `tq` evaluates the query against the raw dataset in memory and updates the rendered view with zero lag.

- If a query is malformed or invalid while typing, a non-intrusive red error indicator appears below the prompt while preserving your current view.

### 2. View Mode Switcher (`Tab`)
Press `Tab` to cycle between three specialized view modes:
- **Table Mode**: Formatted columns with auto-detected keys.
- **Tree Mode**: Expandable hierarchy showing types, arrays, and sub-keys.
- **JSON Mode**: Formatted and syntax-highlighted raw JSON.

### 3. Row Detail Inspect Overlay (`Enter`)
When browsing large tables with many columns, press `Enter` in Navigation Mode on any selected row to open an inspect drawer overlay. This drawer displays all fields and full nested JSON sub-objects with complete indentation. Press `Esc` or `Enter` to close the overlay.

### 4. Direct Query Copy (`Ctrl+Y`)
Press `Ctrl+Y` at any moment in the TUI to copy the currently composed `jq` query directly to your system clipboard, allowing you to instantly paste it into shell scripts, CLI commands, or documentation.

---

## Keyboard Controls Reference

| Mode / Shortcut | Description |
| --- | --- |
| **Filter Mode** | Type text into the live `jq >` query bar |
| `Esc` / `Enter` | Switch from **Filter Mode** to **Navigation Mode** |
| `↑` / `↓` / `Ctrl+P` / `Ctrl+N` | Move table selection while typing in filter |
| **Nav Mode**: `j` / `k` / `↑` / `↓` | Move cursor down / up rows, or scroll viewport |
| **Nav Mode**: `g` / `G` | Jump to top / bottom of table or viewport |
| **Nav Mode**: `/` or `i` | Return to **Filter Mode** (focuses query prompt) |
| `Tab` | Cycles through view modes (`Table` ➔ `Tree` ➔ `JSON`) |
| `Enter` (Nav Mode) | Opens the full-detail inspection drawer for the highlighted row |
| `Esc` (in drawer) | Exits the inspection drawer and returns to the active table view |
| `Ctrl+Y` | Copies the active JQ query to your clipboard |
| `Ctrl+C` / `q` | Quit `tq` |
