#!/usr/bin/env bash

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

function summarize {
    local -a arr=("./playbooks/test_create_resources.yml" "./playbooks/test_remove_resources.yml")

    for f in "${arr[@]}"; do
        info "Test Summary for $f:" "==="
        # Replace ansible variable with value
        # Strip task name
        # Align POSITIVE and NEGATIVE test with IDEMPOTENT test
        printf "$(grep "^  - name:" $f \
            | sed \
                -e 's?- name: ??' \
                -e 's?{{ auto_prefix }}?PSUS_ANSIBLE_?g' \
                -e 's?POSITIVE test -> \(.*$\)?\\e[32mPOSITIVE test ---> \1\\e[0m?g' \
                -e 's?NEGATIVE test -> \(.*$\)?\\e[33mNEGATIVE test ---> \1\\e[0m?g' \
                -e 's?IDEMPOTENT test -> \(.*$\)?\\e[35mIDEMPOTENT test -> \1\\e[0m?g' \
                -e 's?ASSERT test -> \(.*$\)?\\e[36mASSERT test -----> \1\\e[0m?g' \
                -e 's?DEBUG test -> \(.*$\)?\\e[37mDEBUG test ------> \1\\e[0m?g' \
                )"
        printf "\n\n"
    done
}

# Main
summarize
