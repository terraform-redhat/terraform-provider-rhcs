#!/usr/bin/env bash

set -euo pipefail

command_name=${1:-}

if [ -n "${2:-}" ] && [ -n "${3:-}" ] && [ -n "${4:-}" ]; then
    before_unstaged=$2
    before_staged=$3
    before_untracked=$4

    current_unstaged=$(mktemp)
    current_staged=$(mktemp)
    current_untracked=$(mktemp)
    trap 'rm -f "$current_unstaged" "$current_staged" "$current_untracked"' EXIT

    git diff --binary --no-ext-diff > "$current_unstaged"
    git diff --cached --binary --no-ext-diff > "$current_staged"
    git ls-files --others --exclude-standard > "$current_untracked"

    if cmp -s "$before_unstaged" "$current_unstaged" && cmp -s "$before_staged" "$current_staged" && cmp -s "$before_untracked" "$current_untracked"; then
        exit 0
    fi
fi

if [[ -n "$(git status --porcelain)" ]]; then
    echo "It seems like you need to run 'make $command_name'. Please run it and commit the changes"
    git status --porcelain
    exit 1
fi