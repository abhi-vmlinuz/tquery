#compdef tquery

_tquery() {
    local -a formats
    formats=(
        'table:ANSI color-coded Unicode ASCII table'
        'markdown:GitHub-Flavored Markdown table'
        'csv:Comma-separated values data'
        'tsv:Tab-separated values data'
        'tree:Hierarchical indented ASCII tree'
        'json:Pretty-printed JSON output'
    )

    _arguments \
        '(-f --format)'{-f,--format}'[Set output format]:format:(($formats))' \
        '(-i --interactive)'{-i,--interactive}'[Launch interactive TUI mode]' \
        '(-n --no-headers)'{-n,--no-headers}'[Hide headers in table and CSV formats]' \
        '--no-unwrap[Disable automatic root array un-wrapping]' \
        '--no-color[Disable ANSI color formatting]' \
        '(-v --version)'{-v,--version}'[Print software version]' \
        '(-h --help)'{-h,--help}'[Display usage help]' \
        '1:JQ Query or JSON file:_files -g "*.json"' \
        '2:JSON file:_files -g "*.json"'
}

_tquery "$@"
