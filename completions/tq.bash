# Bash completion for tq and tquery

_tq_completions() {
    local cur prev opts formats
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="-f --format --tree --table --json --raw --markdown --md --csv --tsv -g -e --grep --strict -gi -gv -gvi -giv -iv -vi -v -V --invert --invert-match -i -I --ignore-case -u --ui --interactive -n --no-headers --no-unwrap --no-color --no-pager --pager -l -L --limit --version -h --help"
    formats="auto table tree markdown csv tsv json"

    case "$prev" in
        -f|--format)
            COMPREPLY=( $(compgen -W "${formats}" -- "$cur") )
            return 0
            ;;
        -g|-e|--grep|-gi|-gv|-gvi)
            return 0
            ;;
        -l|-L|--limit)
            return 0
            ;;
    esac

    if [[ "$cur" == -* ]]; then
        COMPREPLY=( $(compgen -W "${opts}" -- "$cur") )
        return 0
    fi

    # Complete files if not typing an option
    COMPREPLY=( $(compgen -f -X '!*.json' -- "$cur") $(compgen -f -- "$cur") )
}

complete -F _tq_completions tq
complete -F _tq_completions tquery
