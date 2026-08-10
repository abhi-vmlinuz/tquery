# Fish completion for tq and tquery

for cmd in tq tquery
    complete -c $cmd -f

    # Format options
    complete -c $cmd -s f -l format -d "Output format" -r -a "auto table tree markdown csv tsv json"

    # Shape shortcuts
    complete -c $cmd -l tree -d "Force hierarchical tree view"
    complete -c $cmd -l table -d "Force tabular view"
    complete -c $cmd -l json -d "Force JSON output"
    complete -c $cmd -l raw -d "Force raw JSON output"
    complete -c $cmd -l markdown -l md -d "Force markdown table output"
    complete -c $cmd -l csv -d "Force CSV output"
    complete -c $cmd -l tsv -d "Force TSV output"

    # Grep & pattern search
    complete -c $cmd -s g -s e -l grep -d "Filter rows/tree branches by regex or string" -r
    complete -c $cmd -l strict -d "Strict multi-pattern matching (AND)"
    complete -c $cmd -l gi -d "Grep with case-insensitivity" -r
    complete -c $cmd -l gv -d "Grep with invert-match" -r
    complete -c $cmd -l gvi -d "Grep with invert-match and case-insensitivity" -r
    complete -c $cmd -s v -s V -l invert -l invert-match -d "Invert grep match"
    complete -c $cmd -s I -l ignore-case -d "Case-insensitive grep match"

    # Limit options
    complete -c $cmd -s l -s L -l limit -d "Limit number of output rows/lines" -r

    # General flags
    complete -c $cmd -s i -l interactive -l ui -d "Launch interactive TUI mode"
    complete -c $cmd -s n -l no-headers -d "Hide headers in table and CSV formats"
    complete -c $cmd -l no-unwrap -d "Disable automatic root array un-wrapping"
    complete -c $cmd -l no-color -d "Disable ANSI color formatting"
    complete -c $cmd -l version -d "Show version"
    complete -c $cmd -s h -l help -d "Show usage help"

    # File arguments (.json)
    complete -c $cmd -a "(__fish_complete_suffix .json)" -d "JSON file"
end
