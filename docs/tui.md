# Interactive TUI Guide

`tquery` includes a terminal user interface (TUI) powered by Charm's `bubbletea` and `lipgloss` libraries.

---

## Launching the TUI

You can enter interactive mode by passing the `-i` flag:

```bash
# Explore a local JSON file
tquery -i payload.json

# Pipe live API output directly into the interactive TUI
curl -s https://inference.dahl.global/v1/models | tquery -i
```

---

## Interface Layout

```text
┌─────────────────────────────────────────────────────────────┐
│ tquery   [Table] [Tree] [JSON]                              │  <-- Header & View Modes
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
Type any `jq` expression directly into the prompt bar. As you type, `tquery` evaluates the query against the raw dataset in memory and updates the rendered view with zero lag.

- If a query is malformed or invalid while typing, a non-intrusive red error indicator appears below the prompt while preserving your current view.

### 2. View Mode Switcher (`Tab`)
Press `Tab` to cycle between three specialized view modes:
- **Table Mode**: Formatted columns with auto-detected keys.
- **Tree Mode**: Expandable hierarchy showing types, arrays, and sub-keys.
- **JSON Mode**: Formatted and syntax-highlighted raw JSON.

### 3. Row Detail Inspect Overlay (`Enter`)
When browsing large tables with many columns, press `Enter` on any selected row to open an inspect drawer overlay. This drawer displays all fields and nested sub-objects for that specific row. Press `Esc` or `Enter` to close the overlay.

---

## Keyboard Controls Reference

| Shortcut | Description |
| --- | --- |
| `Typing` | Inputs text into the live `jq >` query bar |
| `Tab` | Cycles through view modes (`Table` ➔ `Tree` ➔ `JSON`) |
| `Enter` | Opens the full-detail inspection drawer for the highlighted row |
| `Esc` | Exits the inspection drawer and returns to the active table view |
| `↑` / `k` | Move cursor / scroll up |
| `↓` / `j` | Move cursor / scroll down |
| `Ctrl+C` | Quit `tquery` |
