#!/usr/bin/env bash

set -o nounset
set -o pipefail
set -o noclobber
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

# Search for a file glob.
function check_glob {
	local -r fglob="$1"
    #info "$fglob" "\nTesting"
    result="$(find . -name "$fglob")"
    if [ -n "$result" ]; then
        echo -n "$result\n"
    fi
}

# MAIN
declare found=""
declare foundone=""
declare -ar fglobs=("*.gz" ".*.un~" ".*.swp" ".*.swo" "*.pyc" "venv")

for fg in "${fglobs[@]}"; do
	foundone=$(check_glob "$fg")
	if [ -n "$foundone" ]; then
		found="${found}${foundone}"
	fi
done

if [ -n "${found}" ]; then
	die "Found files that should not be included in collections:\n$found"
fi
