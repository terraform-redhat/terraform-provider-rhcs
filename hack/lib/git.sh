#!/usr/bin/env bash
# Copyright Red Hat
# SPDX-License-Identifier: Apache-2.0

# newest_merge_base HEAD ref...
# Prints the newest merge-base among merge-base(HEAD, ref) for each reachable ref.
newest_merge_base() {
  local head=$1
  shift

  local best_mb=""
  local ref mb

  for ref in "$@"; do
    [ -z "$ref" ] && continue
    if ! git rev-parse --verify "${ref}^{commit}" >/dev/null 2>&1; then
      continue
    fi
    if ! mb=$(git merge-base "$head" "${ref}" 2>/dev/null); then
      continue
    fi
    if [ -z "$best_mb" ]; then
      best_mb=$mb
      continue
    fi
    if git merge-base --is-ancestor "$best_mb" "$mb" 2>/dev/null; then
      best_mb=$mb
    fi
  done

  printf '%s' "$best_mb"
}

# merge_base_range [HEAD] ref...
# Prints merge_base...HEAD when a merge base exists.
merge_base_range() {
  local head=${1:-HEAD}
  shift

  local mb
  mb=$(newest_merge_base "$head" "$@")
  if [ -n "$mb" ]; then
    printf '%s...%s' "$mb" "$head"
  fi
}
