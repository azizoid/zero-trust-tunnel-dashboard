#!/bin/bash
# Interactive connection setup

# Function to prompt for connection
interactive_connection() {
    echo "How would you like to connect?"
    echo "  1) SSH config host"
    echo "  2) Direct connection"
    read -p "Choose [1-2]: " connection_method
    
    local args=()
    
    case "$connection_method" in
        1)
            local ssh_host
            local hosts
            hosts=$(list_ssh_hosts)
            if [ $? -eq 0 ] && [ -n "$hosts" ]; then
                echo ""
                local host_array=()
                local i=1
                while IFS= read -r host; do
                    if [ -n "$host" ]; then
                        echo "  $i) $host"
                        host_array+=("$host")
                        i=$((i + 1))
                    fi
                done <<< "$hosts"
                echo ""
                read -p "Choose host [1-$((i-1))] or enter custom: " host_choice
                
                if [ -n "$host_choice" ] && echo "$host_choice" | grep -qE '^[0-9]+$'; then
                    if [ "$host_choice" -ge 1 ] && [ "$host_choice" -lt $i ]; then
                        ssh_host="${host_array[$((host_choice-1))]}"
                    else
                        ssh_host="$host_choice"
                    fi
                else
                    ssh_host="$host_choice"
                fi
            else
                read -p "Enter SSH host alias: " ssh_host
            fi
            
            if [ -z "$ssh_host" ]; then
                echo "Error: SSH host is required"
                exit 1
            fi
            args+=("--host" "$ssh_host")
            ;;
        2)
            read -p "Server address: " server
            read -p "Username: " user
            
            if [ -z "$server" ] || [ -z "$user" ]; then
                echo "Error: Server and user are required"
                exit 1
            fi
            
            args+=("--server" "$server" "--user" "$user")
            ;;
        *)
            echo "Invalid choice"
            exit 1
            ;;
    esac
    
    build_tunnel_dash
    exec ./tunnel-dash "${args[@]}"
}

