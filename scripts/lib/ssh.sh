#!/bin/bash
# SSH configuration functions

# Function to list SSH config hosts
list_ssh_hosts() {
    local ssh_config="${HOME}/.ssh/config"
    if [ ! -f "$ssh_config" ]; then
        return 1
    fi
    
    local hosts=()
    while IFS= read -r line; do
        line=$(echo "$line" | sed 's/#.*//' | xargs)
        if [ -n "$line" ] && echo "$line" | grep -qiE "^host\s+"; then
            local host=$(echo "$line" | awk '{print $2}')
            if [ "$host" != "*" ] && [ -n "$host" ]; then
                hosts+=("$host")
            fi
        fi
    done < "$ssh_config"
    
    if [ ${#hosts[@]} -eq 0 ]; then
        return 1
    fi
    
    printf '%s\n' "${hosts[@]}"
    return 0
}

