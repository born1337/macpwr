# Bash completion for macpwr
# Add to ~/.bashrc: source /path/to/macpwr/completions/macpwr.bash

_macpwr_completions() {
    local cur prev commands presets profile_cmds

    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    commands="status show set battery preset profile caffeinate assertions thermal help version"
    presets="default presentation battery-saver performance movie list"
    profile_cmds="save load list delete"

    case "$prev" in
        macpwr)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
        preset)
            COMPREPLY=($(compgen -W "$presets" -- "$cur"))
            return 0
            ;;
        profile)
            COMPREPLY=($(compgen -W "$profile_cmds" -- "$cur"))
            return 0
            ;;
        set)
            COMPREPLY=($(compgen -W "-a --ac -b --battery -d --display -s --sleep -k --disk" -- "$cur"))
            return 0
            ;;
        caffeinate|cafe)
            COMPREPLY=($(compgen -W "-t --time -d --display -i --idle -s --system --" -- "$cur"))
            return 0
            ;;
        help)
            COMPREPLY=($(compgen -W "$commands" -- "$cur"))
            return 0
            ;;
        load|delete)
            # Complete with saved profile names
            local profiles_dir="$HOME/.config/macpwr/profiles"
            if [[ -d "$profiles_dir" ]]; then
                local profiles=$(ls "$profiles_dir"/*.profile 2>/dev/null | xargs -n1 basename 2>/dev/null | sed 's/\.profile$//')
                COMPREPLY=($(compgen -W "$profiles" -- "$cur"))
            fi
            return 0
            ;;
        -a|--ac|-b|--battery)
            COMPREPLY=($(compgen -W "-d --display -s --sleep -k --disk" -- "$cur"))
            return 0
            ;;
        -d|--display|-s|--sleep|-k|--disk|-t|--time)
            # Expect a number
            COMPREPLY=()
            return 0
            ;;
    esac

    # If we're past the first argument, provide context-specific completions
    if [[ ${COMP_CWORD} -gt 1 ]]; then
        case "${COMP_WORDS[1]}" in
            set)
                COMPREPLY=($(compgen -W "-a --ac -b --battery -d --display -s --sleep -k --disk" -- "$cur"))
                ;;
            caffeinate|cafe)
                COMPREPLY=($(compgen -W "-t --time -d --display -i --idle -s --system" -- "$cur"))
                ;;
        esac
    fi

    return 0
}

complete -F _macpwr_completions macpwr
