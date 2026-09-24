# Fish completion for tquery

complete -c tquery -f

# Format options
complete -c tquery -s f -l format -d "Output format" -r -a "auto table tree markdown md csv tsv json raw"

# Shape shortcuts
complete -c tquery -l tree -d "Force hierarchical tree view"
complete -c tquery -l table -d "Force tabular view"
complete -c tquery -l json -d "Force JSON output"
complete -c tquery -l raw -d "Force raw JSON output"
complete -c tquery -l markdown -l md -d "Force markdown table output"
complete -c tquery -l csv -d "Force CSV output"
complete -c tquery -l tsv -d "Force TSV output"

# Data shaping & projection
complete -c tquery -s c -l columns -d "Cherry-pick columns by name (comma-separated)" -r
complete -c tquery -s s -l sort -d "Sort records by column name (use -col for descending)" -r
complete -c tquery -l desc -d "Sort records in descending order"
complete -c tquery -l cb -l clipboard -d "Read JSON data directly from system clipboard"

# Grep & pattern search
complete -c tquery -s g -s e -l grep -d "Filter rows/tree branches by regex or string" -r
complete -c tquery -l strict -d "Strict multi-pattern matching (AND)"
complete -c tquery -l gi -d "Grep with case-insensitivity" -r
complete -c tquery -l gv -d "Grep with invert-match" -r
complete -c tquery -l gvi -d "Grep with invert-match and case-insensitivity" -r
complete -c tquery -s v -s V -l invert -l invert-match -d "Invert grep match"
complete -c tquery -s i -s I -l ignore-case -d "Case-insensitive search match"

# Limit options
complete -c tquery -s l -s L -l limit -d "Limit number of output rows/lines" -r

# General flags
complete -c tquery -s u -l ui -l interactive -d "Launch interactive TUI mode"
complete -c tquery -s n -l no-headers -d "Hide headers in table and CSV formats"
complete -c tquery -l no-unwrap -d "Disable automatic root array un-wrapping"
complete -c tquery -l no-color -d "Disable ANSI color formatting"
complete -c tquery -l no-pager -d "Disable automatic terminal pager (less -R)"
complete -c tquery -l pager -d "Force automatic terminal pager"
complete -c tquery -l version -d "Show version"
complete -c tquery -s h -l help -d "Show usage help"

# File arguments (.json)
complete -c tquery -a "(__fish_complete_suffix .json)" -d "JSON file"

