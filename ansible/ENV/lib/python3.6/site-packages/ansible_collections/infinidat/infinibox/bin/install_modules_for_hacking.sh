#!/usr/bin/env bash

#####
# A utility script to copy collection resources into a clone of ansible source code.
# This is useful if using ansible hacking to execute modules without use of playbooks.
#
# Script uses "git rev-parse --show-toplevel" to find the top of the tree in this collection.
# It will fail if this is not a git clone.
#####

set -o nounset
set -o pipefail
#set -o errexit

# To enable info and die functions logging to syslog, set to true.
declare -r enable_syslogging="false"

# Report error with $leader prepended and die.
# Override leader with an optional replacement.
# Optionally send to logger.
function die {
    local msg="$1"
    local leader="${2:-Fatal:}"
    (>&2 printf "$leader $msg\n")  # Subshell avoids interactions with other redirections
    if [ "$enable_syslogging" == "true" ]; then
        logger -p user.error -t "$(basename $0)" "$leader $msg"
    fi
    kill -SIGPIPE "$$"  # Die with exit code 141
}

# Report msg with $leader prepended.
# Override leader with an optional replacement.
# Optionally send to logger.
function info {
    local msg="$1"
    local leader="${2:-Info:}"
    (printf "$leader $msg\n")  # Subshell avoids interactions with other redirections
    if [ "$enable_syslogging" == "true" ]; then
        logger -p user.notice -t "$(basename $0)" "$leader $msg"
    fi
}

function get_git_toplevel {
    local -r git_toplevel="$(git rev-parse --show-toplevel)"
    printf "$git_toplevel"
}

function install_infindat_modules {
    local -r ansible_path="$1"
    local -r src="$(get_git_toplevel)/plugins/modules"
    local -r dest="$ansible_path/lib/ansible/modules/storage/infinidat/"
    info "Copying modules from $src to $dest" "$info_leader"
    cp "$src/infini_"*".py" "$dest" \
        || die "Cannot copy modules from $src to $dest"
}

function install_infindat_module_utils {
    local -r ansible_path="$1"
    local -r src="$(get_git_toplevel)/plugins/module_utils"
    local -r dest="$ansible_path/lib/ansible/module_utils/"
    info "Copying module_utils from $src to $dest" "$info_leader"
    cp "$src/infinibox.py" "$dest" \
        || die "Cannot copy module_utils from $src to $dest"
}

function install_infindat_doc_fragments {
    local -r ansible_path="$1"
    local -r src="$(get_git_toplevel)/plugins/doc_fragments"
    local -r dest="$ansible_path/lib/ansible/plugins/doc_fragments/"
    info "Copying doc_fragments from $src to $dest" "$info_leader"
    cp "$src/infinibox.py" "$dest" \
        || die "Cannot copy doc_fragments from $src to $dest"
}

function sanity {
    local -r ansible_path="$1"
    local -r ansible_dir="lib/ansible"
    local -r check_this="$ansible_path/$ansible_dir"

    if ! [[ -a "$check_this" ]]; then
        die "Cannot find Ansible source at \"$check_this\". Check path to Ansible source clone."
    fi
}

# Main
declare -r info_leader="  "
declare -r ansible_working_copy_path="${1:-}"
if [ -z "${ansible_working_copy_path}" ]; then
    die "Provide the path to an Ansible source clone. Example: $0 \"<path to clone>\""
fi

info "Copying Infinidat resources to $ansible_working_copy_path" "=="
sanity "$ansible_working_copy_path"
install_infindat_modules        "$ansible_working_copy_path"
install_infindat_module_utils   "$ansible_working_copy_path"
install_infindat_doc_fragments  "$ansible_working_copy_path"
info "Completed copying Infinidat resources to $ansible_working_copy_path" "=="
