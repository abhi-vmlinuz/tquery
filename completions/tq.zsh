#compdef tq tquery

_tq() {
    local -a formats
    formats=(
        'auto:Auto-detect best format (table for records, tree for deep objects)'
        'table:ANSI color-coded Unicode ASCII table'
        'tree:Hierarchical indented ASCII tree'
        'markdown:GitHub-Flavored Markdown table'
        'csv:Comma-separated values data'
        'tsv:Tab-separated values data'
        'json:Pretty-printed JSON output'
    )

    _arguments \
        '(-f --format)'{-f,--format}'[Set output format]:format:(($formats))' \
        '--tree[Force hierarchical tree view]' \
        '--table[Force tabular view]' \
        '(--json --raw)'{--json,--raw}'[Force JSON output]' \
        '(--markdown --md)'{--markdown,--md}'[Force markdown table output]' \
        '--csv[Force CSV output]' \
        '--tsv[Force TSV output]' \
        '*'{-g,-e,--grep}'[Filter by regex or string pattern]:pattern:' \
        '--strict[Strict multi-pattern matching (AND)]' \
        '(-v -V --invert --invert-match)'{-v,-V,--invert,--invert-match}'[Invert grep match]' \
        '(-I --ignore-case)'{-I,--ignore-case}'[Case-insensitive grep match]' \
        '(-l -L --limit)'{-l,-L,--limit}'[Limit number of output rows/lines]:number:' \
        '(-i --interactive --ui)'{-i,--interactive,--ui}'[Launch interactive TUI mode]' \
        '(-n --no-headers)'{-n,--no-headers}'[Hide headers in table and CSV formats]' \
        '--no-unwrap[Disable automatic root array un-wrapping]' \
        '--no-color[Disable ANSI color formatting]' \
        '--version[Print software version]' \
        '(-h --help)'{-h,--help}'[Display usage help]' \
        '1:JQ Query or JSON file:_files -g "*.json"' \
        '2:JSON file:_files -g "*.json"'
}

_tq "$@"
