#compdef macpwr
# Zsh completion for macpwr
# Add to ~/.zshrc: fpath=(/path/to/macpwr/completions $fpath) && compinit

_macpwr() {
    local -a commands presets profile_cmds set_opts cafe_opts

    commands=(
        'status:Show quick power status'
        'show:Display detailed power settings table'
        'set:Change power settings'
        'battery:Show detailed battery information'
        'preset:Apply built-in power presets'
        'profile:Save/load custom power profiles'
        'caffeinate:Prevent system from sleeping'
        'assertions:Show what'\''s preventing sleep'
        'thermal:Show thermal and CPU information'
        'help:Show help message'
        'version:Show version'
    )

    presets=(
        'default:Restore macOS default settings'
        'presentation:Keep screen on, prevent sleep'
        'battery-saver:Aggressive power saving'
        'performance:Maximum performance, no sleep'
        'movie:Screen stays on, system can sleep'
        'list:List all presets'
    )

    profile_cmds=(
        'save:Save current settings as a profile'
        'load:Load and apply a saved profile'
        'list:List all saved profiles'
        'delete:Delete a saved profile'
    )

    set_opts=(
        '-a[Apply to AC power]::'
        '--ac[Apply to AC power]::'
        '-b[Apply to battery power]::'
        '--battery[Apply to battery power]::'
        '-d[Set display sleep]:minutes:'
        '--display[Set display sleep]:minutes:'
        '-s[Set system sleep]:minutes:'
        '--sleep[Set system sleep]:minutes:'
        '-k[Set disk sleep]:minutes:'
        '--disk[Set disk sleep]:minutes:'
    )

    cafe_opts=(
        '-t[Prevent sleep for duration]:minutes:'
        '--time[Prevent sleep for duration]:minutes:'
        '-d[Only prevent display sleep]'
        '--display[Only prevent display sleep]'
        '-i[Only prevent idle sleep]'
        '--idle[Only prevent idle sleep]'
        '-s[Only prevent system sleep]'
        '--system[Only prevent system sleep]'
    )

    _arguments -C \
        '1: :->command' \
        '*: :->args'

    case $state in
        command)
            _describe -t commands 'macpwr commands' commands
            ;;
        args)
            case $words[2] in
                preset)
                    _describe -t presets 'presets' presets
                    ;;
                profile)
                    if (( CURRENT == 3 )); then
                        _describe -t profile_cmds 'profile commands' profile_cmds
                    elif (( CURRENT == 4 )); then
                        case $words[3] in
                            load|delete)
                                local profiles_dir="$HOME/.config/macpwr/profiles"
                                if [[ -d "$profiles_dir" ]]; then
                                    local -a profiles
                                    profiles=(${(f)"$(ls "$profiles_dir"/*.profile 2>/dev/null | xargs -n1 basename 2>/dev/null | sed 's/\.profile$//')"})
                                    _describe -t profiles 'saved profiles' profiles
                                fi
                                ;;
                            save)
                                _message 'profile name'
                                ;;
                        esac
                    fi
                    ;;
                set)
                    _arguments $set_opts
                    ;;
                caffeinate|cafe)
                    _arguments $cafe_opts
                    ;;
                help)
                    local -a help_commands
                    help_commands=(show set battery preset profile caffeinate assertions thermal)
                    _describe -t help_commands 'commands' help_commands
                    ;;
            esac
            ;;
    esac
}

_macpwr "$@"
