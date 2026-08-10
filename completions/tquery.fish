# Fish completion for tquery

complete -c tquery -f

# Format options
complete -c tquery -s f -l format -d "Output format" -r -a "table markdown csv tsv tree json"

# Flags
complete -c tquery -s i -l interactive -d "Launch interactive TUI mode"
complete -c tquery -s n -l no-headers -d "Hide headers in table and CSV formats"
complete -c tquery -l no-unwrap -d "Disable automatic root array un-wrapping"
complete -c tquery -l no-color -d "Disable ANSI color formatting"
complete -c tquery -s v -l version -d "Show version"
complete -c tquery -s h -l help -d "Show usage help"

# File arguments (.json)
complete -c tquery -a "(__fish_complete_suffix .json)" -d "JSON file"
