# Fish completion for tq and tquery

for cmd in tq tquery
    complete -c $cmd -f

    # Format options
    complete -c $cmd -s f -l format -d "Output format" -r -a "table markdown csv tsv tree json"

    # Limit options
    complete -c $cmd -s l -s L -l limit -d "Limit number of output rows" -r

    # Flags
    complete -c $cmd -s i -l interactive -d "Launch interactive TUI mode"
    complete -c $cmd -s n -l no-headers -d "Hide headers in table and CSV formats"
    complete -c $cmd -l no-unwrap -d "Disable automatic root array un-wrapping"
    complete -c $cmd -l no-color -d "Disable ANSI color formatting"
    complete -c $cmd -s v -l version -d "Show version"
    complete -c $cmd -s h -l help -d "Show usage help"

    # File arguments (.json)
    complete -c $cmd -a "(__fish_complete_suffix .json)" -d "JSON file"
end
